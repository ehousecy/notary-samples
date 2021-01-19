package sdkutil

import (
	reqContext "context"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	selectopts "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"log"
	"path/filepath"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/pkg/errors"
	"math/rand"
)

func SignPayload(ctx contextApi.Client, payload *common.Payload) (*fab.SignedEnvelope, error) {
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "marshaling of payload failed")
	}

	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payloadBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of payload failed")
	}
	return &fab.SignedEnvelope{Payload: payloadBytes, Signature: signature}, nil
}

func CreatePayload(tx *fab.Transaction) (*common.Payload, error) {
	hdr := &common.Header{}
	if err := proto.Unmarshal(tx.Proposal.Proposal.Header, hdr); err != nil {
		return nil, err
	}
	// serialize the tx
	txBytes, err := proto.Marshal(tx.Transaction)
	if err != nil {
		return nil, err
	}
	// create the payload
	return &common.Payload{Header: hdr, Data: txBytes}, nil
}

func BroadcastEnvelope(reqCtx reqContext.Context, envelope *fab.SignedEnvelope, orderers []fab.Orderer) (*fab.TransactionResponse, error) {
	// Check if orderers are defined
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	// Copy aside the ordering service endpoints
	randOrderers := []fab.Orderer{}
	randOrderers = append(randOrderers, orderers...)

	// get a context client instance to create child contexts with timeout read from the config in sendBroadcast()
	ctxClient, ok := context.RequestClientContext(reqCtx)

	if !ok {
		return nil, errors.New("failed get client context from reqContext for SendTransaction")
	}

	// Iterate them in a random order and try broadcasting 1 by 1
	var errResp error
	for _, i := range rand.Perm(len(randOrderers)) {
		resp, err := sendBroadcast(reqCtx, envelope, randOrderers[i], ctxClient)
		if err != nil {
			errResp = err
		} else {
			return resp, nil
		}
	}
	return nil, errResp
}
func sendBroadcast(reqCtx reqContext.Context, envelope *fab.SignedEnvelope, orderer fab.Orderer, client contextApi.Client) (*fab.TransactionResponse, error) {
	log.Printf("Broadcasting envelope to orderer: %s\n", orderer.URL())
	// create a childContext for this sendBroadcast orderer using the config's timeout value
	// the parent context (reqCtx) should not have a timeout value
	childCtx, cancel := context.NewRequest(client, context.WithTimeoutType(fab.OrdererResponse), context.WithParent(reqCtx))
	defer cancel()

	// Send request
	if _, err := orderer.SendBroadcast(childCtx, envelope); err != nil {
		log.Printf("Receive Error Response from orderer: %s\n", err)
		return nil, errors.Wrapf(err, "calling orderer '%s' failed", orderer.URL())
	}

	log.Printf("Receive Success Response from orderer\n")
	return &fab.TransactionResponse{Orderer: orderer.URL()}, nil
}

func GetOrders(ctx contextApi.Client, channelCfg fab.ChannelCfg) ([]fab.Orderer, error) {
	//chNetworkConfig := ctx.EndpointConfig().ChannelConfig(channelCfg.ID())

	orderers := []fab.Orderer{}
	for _, chOrderer := range channelCfg.Orderers() {

		ordererConfig, found, ignoreOrderer := ctx.EndpointConfig().OrdererConfig(chOrderer)
		if !found || ignoreOrderer {
			//continue if given channel orderer not found in endpoint config
			continue
		}

		orderer, err := ctx.InfraProvider().CreateOrdererFromConfig(ordererConfig)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create orderer from config")
		}

		orderers = append(orderers, orderer)
	}
	return orderers, nil
}

func GetEndorsers(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext, opts ...options.Opt) ([]*fab.ChaincodeCall, []fab.Peer, error) {
	var selectionOpts []options.Opt
	selectionOpts = append(selectionOpts, opts...)
	if requestContext.SelectionFilter != nil {
		selectionOpts = append(selectionOpts, selectopts.WithPeerFilter(requestContext.SelectionFilter))
	}
	if requestContext.PeerSorter != nil {
		selectionOpts = append(selectionOpts, selectopts.WithPeerSorter(requestContext.PeerSorter))
	}

	ccCalls := newInvocationChain(requestContext)
	peers, err := clientContext.Selection.GetEndorsersForChaincode(newInvocationChain(requestContext), selectionOpts...)
	return ccCalls, peers, err
}
func newInvocationChain(requestContext *invoke.RequestContext) []*fab.ChaincodeCall {
	invocChain := []*fab.ChaincodeCall{{ID: requestContext.Request.ChaincodeID}}
	for _, ccCall := range requestContext.Request.InvocationChain {
		if ccCall.ID == invocChain[0].ID {
			invocChain[0].Collections = ccCall.Collections
		} else {
			invocChain = append(invocChain, ccCall)
		}
	}
	return invocChain
}
func SendSignedProposal(reqCtx reqContext.Context, signedProposal *peer.SignedProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	request := fab.ProcessProposalRequest{SignedProposal: signedProposal}
	var responseMtx sync.Mutex
	var transactionProposalResponses []*fab.TransactionProposalResponse
	var wg sync.WaitGroup
	errs := multi.Errors{}

	for _, p := range targets {
		wg.Add(1)
		go func(processor fab.ProposalProcessor) {
			defer wg.Done()

			// TODO: The RPC should be timed-out.
			//resp, err := processor.ProcessTransactionProposal(context.NewRequestOLD(ctx), request)
			resp, err := processor.ProcessTransactionProposal(reqCtx, request)
			if err != nil {
				log.Printf("Received error response from txn proposal processing: %s", err)
				responseMtx.Lock()
				errs = append(errs, err)
				responseMtx.Unlock()
				return
			}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, resp)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponses, errs.ToError()
}

func HandleProposalResponse(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) (*invoke.Response, error) {
	transactionProposalResponses := requestContext.Response.Responses
	if len(transactionProposalResponses) > 0 {
		requestContext.Response.Payload = transactionProposalResponses[0].ProposalResponse.GetResponse().Payload
		requestContext.Response.ChaincodeStatus = transactionProposalResponses[0].ChaincodeStatus
	}
	handler := newExecuteHandler()
	complete := make(chan bool, 1)
	go func() {
		handler.Handle(requestContext, clientContext)
		complete <- true
	}()
	select {
	case <-complete:
		return &requestContext.Response, requestContext.Error
	case <-requestContext.Ctx.Done():
		return nil, status.New(status.ClientStatus, status.Timeout.ToInt32(),
			"request timed out or been cancelled", nil)
	}
}
func newExecuteHandler(next ...invoke.Handler) invoke.Handler {
	return invoke.NewEndorsementValidationHandler(
		invoke.NewSignatureValidationHandler(next...),
	)
}

func CreateTransaction(proposal *fab.TransactionProposal, resps []*fab.TransactionProposalResponse) (*fab.Transaction, error) {
	txnRequest := fab.TransactionRequest{
		Proposal:          proposal,
		ProposalResponses: resps,
	}

	//tx, err := sender.CreateTransaction(txnRequest)
	tx, err := txn.New(txnRequest)
	if err != nil {
		return nil, err
	}
	return tx, err
}

//ValidateSignedEnvelope 校验SignedEnvelope是否被修改
func ValidateSignedEnvelope(envelope *fab.SignedEnvelope, txID fab.TransactionID) (bool, error) {
	payload := common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, &payload); err != nil {
		return false, nil
	}
	channelHeader := common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, &channelHeader); err != nil {
		return false, nil
	}
	return txID == fab.TransactionID(channelHeader.TxId), nil
}

//=========提取公共类
func InitSDK(channelID string) (*fabsdk.FabricSDK, error) {
	ccpPath := filepath.Join("config.yaml")
	return fabsdk.New(config.FromFile(filepath.Clean(ccpPath)))
}
func GetContextOptionsByChannelID(channelID string) ([]fabsdk.ContextOption, error) {
	var contextOptions []fabsdk.ContextOption
	contextOptions = append(contextOptions, fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	return contextOptions, nil
}
func GetClient(sdk *fabsdk.FabricSDK, options ...fabsdk.ContextOption) (contextApi.Client, error) {
	userContext := sdk.Context(options...)
	client, err := userContext()
	if err != nil {
		log.Printf("获取client失败, err=%v", err)
	}
	return client, err
}

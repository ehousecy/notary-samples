package client

import (
	"github.com/ehousecy/notary-samples/notary-server/fabric/business"
	"github.com/ehousecy/notary-samples/notary-server/fabric/sdkutil"
	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var clientMap = make(map[string]*Client, 8)
var lock sync.Mutex

type Client struct {
	// Fabric network information
	ConfigPath string
	OrgName    string
	OrgUser    string

	// sdk clients
	SDK            *fabsdk.FabricSDK
	rc             *resmgmt.Client
	CC             *channel.Client
	lc             *ledger.Client
	ContextClient  contextApi.Client
	eventClientMap *event.Client

	//channel info
	ChannelID       string
	ChannelProvider contextApi.ChannelProvider
	ChannelService  fab.ChannelService
	ChannelCfg      fab.ChannelCfg
	Discovery       fab.DiscoveryService
	Selection       fab.SelectionService
	Membership      fab.ChannelMembership
	Event           fab.EventService

	// Same for each peer
}

func GetClientByChannelID(channelID string) (*Client, error) {
	client, ok := clientMap[channelID]
	if ok {
		return client, nil
	}
	client, err := New(channelID)
	if err != nil {
		return nil, err
	}
	lock.Lock()
	clientMap[channelID] = client
	lock.Unlock()
	return client, nil
}

func New(channelID string) (*Client, error) {
	client := &Client{ChannelID: channelID}
	sdk, err := business.New().InitSDK(channelID)
	if err != nil {
		return nil, err
	}
	client.SDK = sdk
	contextOptions, err := business.New().GetContextOptions(channelID)
	if err != nil {
		return nil, err
	}
	if client.ContextClient, err = sdkutil.GetClient(sdk, contextOptions...); err != nil {
		return nil, err
	}

	//init channel info
	client.ChannelProvider = sdk.ChannelContext(channelID, contextOptions...)
	c, err := client.ChannelProvider()
	if err != nil {
		return nil, err
	}
	client.ChannelService = c.ChannelService()
	if client.ChannelCfg, err = client.ChannelService.ChannelConfig(); err != nil {
		return nil, err
	}
	if client.Discovery, err = client.ChannelService.Discovery(); err != nil {
		return nil, err
	}
	if client.Selection, err = client.ChannelService.Selection(); err != nil {
		return nil, err
	}
	if client.Membership, err = client.ChannelService.Membership(); err != nil {
		return nil, err
	}
	if client.Event, err = client.ChannelService.EventService(); err != nil {
		return nil, err
	}
	client.CC, err = channel.New(client.ChannelProvider)
	if err != nil {
		return nil, err
	}
	return client, err
}

func (c *Client) CreateClientContext(transactor *channelImpl.Transactor) *invoke.ClientContext {
	clientContext := &invoke.ClientContext{
		Discovery:    c.Discovery,
		Selection:    c.Selection,
		Membership:   c.Membership,
		Transactor:   transactor,
		EventService: c.Event,
	}
	return clientContext
}

func (c *Client) CreateTransactionProposal(chrequest *channel.Request, creator []byte) (*fab.TransactionProposal, error) {
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID:  chrequest.ChaincodeID,
		Fcn:          chrequest.Fcn,
		Args:         chrequest.Args,
		TransientMap: chrequest.TransientMap,
		IsInit:       chrequest.IsInit,
	}
	reqCtx, cancel := context.NewRequest(c.ContextClient, context.WithTimeout(10*time.Minute))
	defer cancel()
	transactor, err := channelImpl.NewTransactor(reqCtx, c.ChannelCfg)
	if err != nil {
		return nil, err
	}
	txh, err := transactor.CreateTransactionHeader(fab.WithCreator(creator))
	if err != nil {
		return nil, err
	}
	proposal, err := txn.CreateChaincodeInvokeProposal(txh, request)
	return proposal, err
}

func (c *Client) CreateTransactionPayload(request channel.Request, signedProposal *pb.SignedProposal) ([]byte, error) {
	reqCtx, cancel := context.NewRequest(c.ContextClient, context.WithTimeout(10*time.Minute))
	defer cancel()
	transactor, err := channelImpl.NewTransactor(reqCtx, c.ChannelCfg)
	if err != nil {
		return nil, err
	}
	clientContext := c.CreateClientContext(transactor)

	requestContext := &invoke.RequestContext{Response: invoke.Response{},
		Opts: invoke.Opts{}, Request: invoke.Request(request), Ctx: reqCtx}

	if _, requestContext.Opts.Targets, err = sdkutil.GetEndorsers(requestContext, clientContext); err != nil {
		return nil, err
	}
	defer cancel()
	transactionProposalResponses, err := sdkutil.SendSignedProposal(reqCtx, signedProposal, peer.PeersToTxnProcessors(requestContext.Opts.Targets))
	if err != nil {
		return nil, errors.Wrap(err, "发送签名Proposal失败")
	}

	proposal := &pb.Proposal{}

	if err = proto.Unmarshal(signedProposal.ProposalBytes, proposal); err != nil {
		return nil, err
	}
	txProposal := &fab.TransactionProposal{
		Proposal: proposal,
	}
	//5.处理提案响应
	requestContext.Response.Responses = transactionProposalResponses

	_, err = sdkutil.HandleProposalResponse(requestContext, clientContext)
	if err != nil {
		return nil, errors.Wrap(err, "处理ProposalResponse失败")
	}

	//6.创建交易
	tx, err := sdkutil.CreateTransaction(txProposal, transactionProposalResponses)
	if err != nil {
		return nil, errors.Wrap(err, "创建交易失败")
	}

	//7.创建payload
	payload, err := sdkutil.CreatePayload(tx)
	if err != nil {
		return nil, errors.Wrap(err, "创建payload失败")
	}
	payloadBytes, err := proto.Marshal(payload)

	return payloadBytes, err
}

func (c *Client) SendSignedEnvelopTx(envelope *fab.SignedEnvelope) (*fab.TransactionResponse, error) {
	reqCtx, cancel := context.NewRequest(c.ContextClient, context.WithTimeout(10*time.Minute))
	defer cancel()
	orders, _ := sdkutil.GetOrders(c.ContextClient, c.ChannelCfg)
	f, err := sdkutil.BroadcastEnvelope(reqCtx, envelope, orders)
	return f, err
}

func (c *Client) CreateTransaction(request channel.Request) (string, *fab.SignedEnvelope, error) {
	var options []channel.RequestOption
	options = append(options, channel.WithTimeout(fab.Execute, time.Minute*10))
	options = append(options, channel.WithRetry(retry.DefaultChannelOpts))
	response, err := c.CC.InvokeHandler(newExecuteHandler(), request, options...)
	if err != nil {
		return "", nil, err
	}

	tx, err := sdkutil.CreateTransaction(response.Proposal, response.Responses)
	payload, err := sdkutil.CreatePayload(tx)
	envelope, _ := sdkutil.SignPayload(c.ContextClient, payload)
	return string(response.Proposal.TxnID), envelope, nil
}

func newExecuteHandler(next ...invoke.Handler) invoke.Handler {
	return invoke.NewSelectAndEndorseHandler(
		invoke.NewEndorsementValidationHandler(
			invoke.NewSignatureValidationHandler(next...),
		),
	)
}

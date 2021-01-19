package tx

import (
	"github.com/ehousecy/notary-samples/fabric/client"
	"github.com/ehousecy/notary-samples/fabric/sdkutil"
	"github.com/ehousecy/notary-samples/notary-server/constant"
	"github.com/ehousecy/notary-samples/notary-server/services"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

type Handler interface {
	HandleOfflineTx(srv pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error
	HandleLocalTx(ticketId string) error
	HandleTxStatusBlock(channelID string, fb *peer.FilteredBlock)
}

var confirmingTxIDMap = make(map[string]map[string]txInfo, 8)
var ticketIDMap = make(map[string]bool, 8)

type txInfo struct {
	ticketId    string
	isOfflineTx bool
}

type txHandler struct {
	db           services.CrossTxDataService
	ticketIDChan chan string
}

func New() *txHandler {
	return &txHandler{db: services.NewCrossTxDataServiceProvider(), ticketIDChan: make(chan string, 1)}
}

func (th *txHandler) HandleOfflineTx(srv pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error {
	//通过ticketId查询跨链交易信息
	crossTxInfo, err := th.db.QueryCrossTxInfoByCID(recv.CTxId)
	if err != nil {
		return err
	}
	//todo:构造request
	request := channel.Request{}
	//获取fabric skd client对象
	c, err := client.GetClientByChannelID(crossTxInfo.FabricChannel)
	if err != nil {
		return err
	}
	//构造proposal
	proposal, err := c.CreateTransactionProposal(&request, recv.Data)
	if err != nil {
		return err
	}
	proposalBytes, err := proto.Marshal(proposal.Proposal)
	//发送proposal到客户端进行签名，srv.Send()
	err = srv.Send(&pb.TransferPropertyResponse{TxData: proposalBytes})
	if err != nil {
		return err
	}

	//接收signedProposal对象,srv.Recv()
	recv, err = srv.Recv()
	if err != nil {
		return err
	}
	//构造交易Payload
	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     recv.Data,
	}
	payloadBytes, err := c.CreateTransactionPayload(request, signedProposal)
	if err != nil {
		return err
	}
	//发送交易Payload到客户端进行签名,srv.Send()
	err = srv.Send(&pb.TransferPropertyResponse{TxData: payloadBytes})
	if err != nil {
		return err
	}
	//接收SignedEnvelope,srv.Recv()
	recv, err = srv.Recv()
	if err != nil {
		return err
	}
	signedEnvelope := &fab.SignedEnvelope{
		Payload:   payloadBytes,
		Signature: recv.Data,
	}
	//验证
	ok, err := sdkutil.ValidateSignedEnvelope(signedEnvelope, proposal.TxnID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("签名验证失败")
	}
	//保存交易id到db
	err = th.db.CreateTransferFromTx(crossTxInfo.ID, string(proposal.TxnID), constant.TypeFabric)
	if err != nil {
		return err
	}
	//发送SignedEnvelope到orderer
	_, err = c.SendSignedEnvelopTx(signedEnvelope)
	if err != nil {
		return err
	}
	//将验证交易id放到map中
	putTxID(crossTxInfo.FabricChannel, string(proposal.TxnID), txInfo{
		ticketId:    crossTxInfo.ID,
		isOfflineTx: true,
	})
	return nil
}

func putTxID(channelID, txID string, ti txInfo) {
	txMap, ok := confirmingTxIDMap[channelID]
	if !ok {
		confirmingTxIDMap[channelID] = map[string]txInfo{
			txID: ti,
		}
	}
	txMap[txID] = ti
}

func (th *txHandler) HandleLocalTx(ticketId string) error {

	//通过ticketId查询跨链交易信息
	crossTxInfo, err := th.db.QueryCrossTxInfoByCID(ticketId)
	if err != nil {
		return err
	}
	//todo: 构造request
	request := channel.Request{}
	//获取fabric skd client对象
	c, err := client.GetClientByChannelID(crossTxInfo.FabricChannel)
	if err != nil {
		return err
	}

	//验证是否能交易
	if err = th.db.ValidateEnableBoundTransferToTx(ticketId, crossTxInfo.FabricTx.FromTxID); err != nil {
		return err
	}
	txID, signedEnvelope, err := c.CreateTransaction(request)
	if err != nil {
		return err
	}
	if err = th.db.BoundTransferToTx(ticketId, crossTxInfo.FabricTx.FromTxID, txID); err != nil {
		return err
	}

	//发送交易
	if _, err = c.SendSignedEnvelopTx(signedEnvelope); err != nil {
		return err
	}
	putTxID(crossTxInfo.FabricChannel, txID, txInfo{
		ticketId:    crossTxInfo.ID,
		isOfflineTx: false,
	})

	return nil
}

func (th *txHandler) HandleTxStatusBlock(channelID string, fb *peer.FilteredBlock) {
	for _, ft := range fb.FilteredTransactions {
		th.handleTx(channelID, ft)
	}

}

func (th *txHandler) handleTx(channelID string, ft *peer.FilteredTransaction) {
	//判断交易id是否有效
	if ft.Txid == "" || ft.TxValidationCode != peer.TxValidationCode_VALID {
		return
	}

	//判断交易id是否需要处理
	txMap, ok := confirmingTxIDMap[channelID]
	if !ok || txMap == nil {
		return
	}
	info, ok := txMap[ft.Txid]
	if !ok {
		return
	}

	//todo: 调用db完成交易
	if info.isOfflineTx {
		err := th.db.CompleteTransferFromTx(info.ticketId, ft.Txid)
		if err != nil {
			return
		}
		crossTxInfo, err := th.db.QueryCrossTxInfoByCID(info.ticketId)
		if err != nil {
			return
		}
		if crossTxInfo.Status == constant.StatusHosted {
			th.HandleLocalTx(info.ticketId)
		}
	} else {
		err := th.db.CompleteTransferToTx(info.ticketId, ft.Txid)
		if err != nil {
			return
		}
	}

	//处理完删除交易id
	delete(txMap, ft.Txid)
}

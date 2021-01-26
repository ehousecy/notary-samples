package tx

import (
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/db/constant"
	"github.com/ehousecy/notary-samples/notary-server/db/services"
	"github.com/ehousecy/notary-samples/notary-server/fabric/business"
	"github.com/ehousecy/notary-samples/notary-server/fabric/client"
	"github.com/ehousecy/notary-samples/notary-server/fabric/sdkutil"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

type Handler interface {
	ConstructAndSignTx(srv pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error
	Approve(ticketId string) error
	HandleTxStatusBlock(channelID string, fb *peer.FilteredBlock)
	ValidateEnableSupport(channelID, chaincodeName, assetType, asset string) error
	QueryLastFabricBlockNumber(channelID string) (uint64, error)
}

var confirmingTxIDMap = make(map[string]map[string]txInfo, 8)
var ticketIDMap = make(map[string]bool, 8)

type txInfo struct {
	ticketId    string
	isOfflineTx bool
}

type txHandler struct {
	db           services.CrossTxDataService
	bl           services.FabricBlockLogService
	ticketIDChan chan string
	bs           business.Support
}

func New() *txHandler {
	//todo: 查询待确认交易列表
	handler := txHandler{db: services.NewCrossTxDataServiceProvider(),
		bl:           services.NewFabricBlockLogServiceProvider(),
		bs:           business.New(),
		ticketIDChan: make(chan string, 1)}
	return &handler
}

func (th *txHandler) ConstructAndSignTx(srv pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error {
	//通过ticketId查询跨链交易信息
	crossTxInfo, err := th.db.QueryCrossTxInfoByCID(recv.CTxId)
	if err != nil {
		return err
	}
	request, err := th.bs.CreateFromRequest(crossTxInfo.FabricChannel, business.RequestParams{
		ChaincodeName: crossTxInfo.FabricChaincode,
		Asset:         crossTxInfo.FabricAmount,
		From:          crossTxInfo.FabricFrom,
	})
	if err != nil {
		return err
	}
	//获取fabric skd client对象
	c, err := client.GetClientByChannelID(crossTxInfo.FabricChannel)
	if err != nil {
		return err
	}
	//构造proposal
	proposal, err := c.CreateTransactionProposal(request, recv.Data)
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
	payloadBytes, err := c.CreateTransactionPayload(*request, signedProposal)
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
		txMap = confirmingTxIDMap[channelID]
	}
	txMap[txID] = ti
}

func (th *txHandler) Approve(ticketId string) error {

	//通过ticketId查询跨链交易信息
	crossTxInfo, err := th.db.QueryCrossTxInfoByCID(ticketId)
	if err != nil {
		return err
	}
	request, err := th.bs.CreateToRequest(crossTxInfo.FabricChannel, business.RequestParams{
		ChaincodeName: crossTxInfo.FabricChaincode,
		Asset:         crossTxInfo.FabricAmount,
		To:            crossTxInfo.FabricTo,
	})
	if err != nil {
		return err
	}
	//获取fabric skd client对象
	c, err := client.GetClientByChannelID(crossTxInfo.FabricChannel)
	if err != nil {
		return err
	}

	//验证是否能交易
	if err = th.db.ValidateEnableBoundTransferToTx(crossTxInfo.FabricTx.FromTxID, nil); err != nil {
		return err
	}
	txID, signedEnvelope, err := c.CreateTransaction(*request)
	if err != nil {
		return err
	}
	if err = th.db.BoundTransferToTx(crossTxInfo.FabricTx.FromTxID, txID); err != nil {
		return err
	}

	//发送交易
	if _, err = c.SendSignedEnvelopTx(signedEnvelope); err != nil {
		//todo:取消交易绑定
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
	//记录处理的区块
	th.bl.AddFabricBlockLog(fb.Number, channelID)
}

func (th *txHandler) handleTx(channelID string, ft *peer.FilteredTransaction) {
	//判断交易id是否有效
	if ft.Txid == "" {
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

	if ft.TxValidationCode != peer.TxValidationCode_VALID {
		//todo:交易无效修改跨链交易记录
	} else {
		err := th.db.CompleteTransferTx(ft.Txid)
		if err != nil {
			return
		}
		if info.isOfflineTx {
			_ = th.db.ValidateEnableBoundTransferToTx(ft.Txid, th.ticketIDChan)
		}
	}

	//处理完删除交易id
	delete(txMap, ft.Txid)
}

func (th *txHandler) ValidateEnableSupport(channelID, chaincodeName, assetType, asset string) error {
	ok, err := th.bs.ValidateEnableSupport(channelID, chaincodeName, assetType, asset)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("the specified fabric transaction is not supported, "+
			"channelID=%s,chaincodeID=%s,asset=%s", channelID, chaincodeName, asset)
	}
	return nil
}

func (th *txHandler) QueryLastFabricBlockNumber(channelID string) (uint64, error) {
	return th.bl.QueryLastFabricBlockNumber(channelID)
}

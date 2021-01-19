package tx

import (
	"github.com/ehousecy/notary-samples/notary-server/services"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Handler interface {
	HandleOfflineTx(srv pb.NotaryService_ConstructTxServer, ticketId string) error
	HandleLocalTx(ticketId string) error
	HandleTxStatusBlock(channelID string, fb *peer.FilteredBlock)
}

var confirmingTxIDMap = make(map[string]map[string]byte, 8)

type txHandler struct {
	db services.CrossTxDataService
}

func New() *txHandler {
	return &txHandler{db: services.NewCrossTxDataServiceProvider()}
}

func (th *txHandler) HandleOfflineTx(srv pb.NotaryService_ConstructTxServer, ticketId string) error {
	//通过ticketId查询跨链交易信息
	//构造request
	//获取fabric skd client对象
	//构造proposal
	//发送proposal到客户端进行签名，srv.Send()
	//接收signedProposal对象,srv.Recv()
	//构造交易Payload
	//发送交易Payload到客户端进行签名,srv.Send()
	//接收SignedEnvelope,srv.Recv()
	//验证
	//发送SignedEnvelope到orderer
	return nil
}

func (th *txHandler) HandleLocalTx(ticketId string) error {
	//通过ticketId查询跨链交易信息
	//构造request
	//获取fabric skd client对象
	//发送交易
	return nil
}

func (th *txHandler) HandleTxStatusBlock(channelID string, fb *peer.FilteredBlock) {
	for _, ft := range fb.FilteredTransactions {
		th.handleTx(channelID, ft)
	}

}

func (th *txHandler) handleTx(channelID string, ft *peer.FilteredTransaction) bool {
	//判断交易id是否有效
	if ft.Txid == "" || ft.TxValidationCode != peer.TxValidationCode_VALID {
		return true
	}

	//判断交易id是否需要处理
	txMap, ok := confirmingTxIDMap[channelID]
	if !ok || txMap == nil {
		return true
	}
	if _, ok := txMap[ft.Txid]; !ok {
		return true
	}

	//todo: 调用db完成交易
	//处理完删除交易id
	delete(txMap, ft.Txid)
	return false
}

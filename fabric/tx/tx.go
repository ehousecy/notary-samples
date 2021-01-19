package tx

import (
	"github.com/ehousecy/notary-samples/notary-server/services"
	pb "github.com/ehousecy/notary-samples/proto"
)

type Handler interface {
	Handle(pb.NotaryService_ConstructTxServer, string) error
}

type txHandler struct {
	db services.CrossTxDataService
}

func New(db services.CrossTxDataService) *txHandler {
	return &txHandler{db: db}
}

func (th *txHandler) offlineTxHandle(srv pb.NotaryService_ConstructTxServer, ticketId string) error {
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

func (th *txHandler) localTxHandle(ticketId string) error {
	//通过ticketId查询跨链交易信息
	//构造request
	//获取fabric skd client对象
	//发送交易
	return nil
}

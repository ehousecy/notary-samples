package main

import (
	"context"
	"github.com/ehousecy/notary-samples/fabric/tx"
	"github.com/ehousecy/notary-samples/notary-server/services"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/go-playground/validator/v10"
	"log"
)

var validate *validator.Validate

type NotaryService struct {
	provider services.CrossTxDataService
	fh       tx.Handler
}

//service NotaryService {
//rpc CreateCTX(CreateCrossTxReq) returns (CreateCrossTxResp) {}
//rpc SubmitTx(TransferPropertyRequest) returns(TransferPropertyResponse) {}
//rpc ListTickets(Empty) returns (ListTxResponse) {}
//rpc GetTicket(QueryTxRequest) returns (QueryTxResponse){}
//rpc OpTicket(AdminOpTicketReq) returns (AdminOpTicketResp) {}
//}

func NewNotaryService() *NotaryService {
	validate = validator.New()
	return &NotaryService{
		provider: services.NewCrossTxDataServiceProvider(),
		fh: tx.New(),
	}
}

func (n *NotaryService) CreateCTX(ctx context.Context, in *pb.CreateCrossTxReq) (*pb.CreateCrossTxResp, error) {
	crossTxBase, err := covertCreateCrossTxReq(in)
	if err != nil {
		return nil, err
	}
	cid, err := n.provider.CreateCrossTx(crossTxBase)
	return &pb.CreateCrossTxResp{
		CTxId: cid,
	}, nil
}

func (n *NotaryService) SubmitTx(srv pb.NotaryService_SubmitTxServer) error {
	recv, err := srv.Recv()
	if err != nil {
		return err
	}
	switch recv.NetworkType {
	case pb.TransferPropertyRequest_fabric:
		err := n.fh.HandleOfflineTx(srv, recv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *NotaryService) ListTickets(ctx context.Context, in *pb.Empty) (*pb.ListTxResponse, error) {
	crossTxInfos, err := n.provider.QueryAllCrossTxInfo()
	if err != nil {
		return nil, err
	}
	var cts = make([]*pb.CrossTx, 0, len(crossTxInfos))
	for _, cti := range crossTxInfos {
		cts = append(cts, convertToCrossTx(cti))
	}
	return &pb.ListTxResponse{
		Detail: cts,
	}, nil
}

func (n *NotaryService) GetTicket(ctx context.Context, in *pb.QueryTxRequest) (*pb.QueryTxResponse, error) {
	crossTxInfo, err := n.provider.QueryCrossTxInfoByCID(in.TicketId)
	if err != nil {
		return nil, err
	}
	return &pb.QueryTxResponse{
		Error:         nil,
		Detail:        convertToCrossTx(*crossTxInfo),
		BlockchainTxs: convertToTxIdsInBlock(*crossTxInfo),
	}, nil
}

func (n *NotaryService) OpTicket(ctx context.Context, in *pb.AdminOpTicketReq) (*pb.AdminOpTicketResp, error) {
	return &pb.AdminOpTicketResp{

	}, nil
}

func covertCreateCrossTxReq(req *pb.CreateCrossTxReq) (services.CrossTxBase, error) {
	detail := req.Detail
	crossTxBase := services.CrossTxBase{
		EthFrom:         detail.EFrom,
		EthTo:           detail.ETo,
		EthAmount:       detail.EAmount,
		FabricFrom:      detail.FFrom,
		FabricTo:        detail.FTo,
		FabricAmount:    detail.FAmount,
		FabricChannel:   detail.FChannel,
		FabricChaincode: detail.FChaincodeName,
	}
	err := validate.Struct(crossTxBase)
	return crossTxBase, err
}

func convertToCrossTx(cti services.CrossTxInfo) *pb.CrossTx {
	return &pb.CrossTx{
		CTxId: cti.ID,
		Detail: &pb.CrossTxDetail{
			EFrom:          cti.EthFrom,
			ETo:            cti.EthTo,
			EAmount:        cti.EthAmount,
			FFrom:          cti.FabricFrom,
			FTo:            cti.FabricTo,
			FAmount:        cti.FabricAmount,
			FChannel:       cti.FabricChannel,
			FChaincodeName: cti.FabricChaincode,
		},
		Status:         &pb.CrossTxStatus{Status: &pb.CrossTxStatus_TStatus{TStatus: pb.TicketStatus_finished}},
		CreateTime:     nil,
		LastUpdateTime: nil,
	}
}

func convertToTxIdsInBlock(cti services.CrossTxInfo) *pb.TxIdsInBlock {
	txIds := &pb.TxIdsInBlock{}
	if cti.FabricTx != nil {
		txIds.FETid = cti.FabricTx.FromTxID
		txIds.FVTid = cti.FabricTx.ToTxID
	}
	if cti.EthTx != nil {
		txIds.UETid = cti.EthTx.FromTxID
		txIds.VETid = cti.EthTx.ToTxID
	}
	return txIds
}

func (n *NotaryService) TestDial(ctx context.Context, in *pb.Ping) (*pb.Pong, error) {
	log.Printf("Receive ping: %s\n", in.Ping)
	log.Printf("sending pong\n")
	return &pb.Pong{
		Pong: "ping received",
	}, nil
}

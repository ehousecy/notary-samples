package main

import (
	"context"
	pb "github.com/ehousecy/notary-samples/proto"
)

type NotaryService struct {
	pb.UnimplementedNotaryServiceServer
}

//service NotaryService {
//rpc CreateCTX(CreateCrossTxReq) returns (CreateCrossTxResp) {}
//rpc SubmitTx(TransferPropertyRequest) returns(TransferPropertyResponse) {}
//rpc ListTickets(Empty) returns (ListTxResponse) {}
//rpc GetTicket(QueryTxRequest) returns (QueryTxResponse){}
//rpc OpTicket(AdminOpTicketReq) returns (AdminOpTicketResp) {}
//}

func NewNotaryService() *NotaryService {
	return &NotaryService{

	}
}

func (n *NotaryService) CreateCTX(ctx context.Context, in *pb.CreateCrossTxReq) (*pb.CreateCrossTxResp, error) {
	return &pb.CreateCrossTxResp{
		CTxId: "12345",
	}, nil
}

func (n *NotaryService) SubmitTx(ctx context.Context, in *pb.TransferPropertyRequest) (*pb.TransferPropertyResponse, error) {
	return &pb.TransferPropertyResponse{
		Error: nil,
		ETxid: "45678",
	}, nil
}

func (n *NotaryService) ListTickets(ctx context.Context, in *pb.Empty) (*pb.ListTxResponse, error) {
	return &pb.ListTxResponse{

	}, nil
}

func (n *NotaryService) GetTicket(ctx context.Context, in *pb.QueryTxRequest) (*pb.QueryTxResponse, error) {

	return &pb.QueryTxResponse{
		Error: nil,
		Detail: &pb.CrossTx{
			CTxId: "111",
			Detail: &pb.CrossTxDetail{
				EFrom:          "123",
				ETo:            "456",
				EAmount:        "789",
				FFrom:          "ffrom",
				FTo:            "fto",
				FAmount:        "fa",
				FChannel:       "channel",
				FChaincodeName: "chaincode",
			},
			Status:         &pb.CrossTxStatus{Status: &pb.CrossTxStatus_TStatus{TStatus: pb.TicketStatus_finished}},
			CreateTime:     nil,
			LastUpdateTime: nil,
		},
		BlockchainTxs: &pb.TxIdsInBlock{
			UETid: "123456",
			VETid: "654321",
			FETid: "789456",
			FVTid: "987654",
		},
	}, nil
}

func (n *NotaryService) OpTicket(ctx context.Context, in *pb.AdminOpTicketReq) (*pb.AdminOpTicketResp, error) {
	return &pb.AdminOpTicketResp{

	}, nil
}

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

func NewNotaryService() *NotaryService  {
	return &NotaryService{

	}
}

func (n *NotaryService)CreateCTX(ctx context.Context, in *pb.CreateCrossTxReq) (*pb.CreateCrossTxResp, error)  {
	return &pb.CreateCrossTxResp{

	}, nil
}

func (n *NotaryService)SubmitTx(ctx context.Context, in *pb.TransferPropertyRequest)(*pb.TransferPropertyResponse, error)  {
	return &pb.TransferPropertyResponse{

	},nil
}

func (n *NotaryService)ListTickets(ctx context.Context, in *pb.Empty)(*pb.ListTxResponse, error)  {
	return &pb.ListTxResponse{

	}, nil
}

func (n *NotaryService)GetTicket(ctx context.Context, in *pb.QueryTxRequest)(*pb.QueryTxResponse, error)  {
	return &pb.QueryTxResponse{

	}, nil
}

func (n *NotaryService)OpTicket(ctx context.Context, in *pb.AdminOpTicketReq)(*pb.AdminOpTicketResp, error)  {
	return &pb.AdminOpTicketResp{

	}, nil
}
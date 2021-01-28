package main

import (
	"context"
	"github.com/ehousecy/notary-samples/notary-server/db/services"
	"github.com/ehousecy/notary-samples/notary-server/eth"
	"github.com/ehousecy/notary-samples/notary-server/fabric"
	"github.com/ehousecy/notary-samples/notary-server/fabric/tx"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/go-playground/validator/v10"
	"log"
)

var validate *validator.Validate

type NotaryService struct {
	provider services.CrossTxDataService
	fh       fabric.Handler
	handlers map[pb.TransferPropertyRequest_NetworkType]TxHandler
}

type TxExecResult struct {
	Err       error
	TxReceipt string
}

type TxHandler interface {
	Approve(ticketID string) error // notary admin op interface
	ConstructAndSignTx(src pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error
}

func NewNotaryService() *NotaryService {
	validate = validator.New()
	n := &NotaryService{
		provider: services.NewCrossTxDataServiceProvider(),
		fh:       tx.NewFabricHandler(),
		handlers: make(map[pb.TransferPropertyRequest_NetworkType]TxHandler, 8),
	}
	n.AddHandler(pb.TransferPropertyRequest_eth, eth.NewEthHandler("http://localhost:8545"))
	n.AddHandler(pb.TransferPropertyRequest_fabric, n.fh)
	return n
}

func (n *NotaryService) AddHandler(t pb.TransferPropertyRequest_NetworkType, handler TxHandler) *NotaryService {
	_, exist := n.handlers[t]
	if exist {
		log.Printf("[Warning] handler already exist, updating handler %v\n", t.String())
	}
	n.handlers[t] = handler
	return n
}

func (n *NotaryService) GetHandler(code pb.TransferPropertyRequest_NetworkType) TxHandler {
	handler, ok := n.handlers[code]
	if !ok {
		log.Printf("[Waring] Unkonw transaction type %v\n", code)
		return nil
	}
	return handler
}

func (n *NotaryService) CreateCTX(ctx context.Context, in *pb.CreateCrossTxReq) (*pb.CreateCrossTxResp, error) {
	crossTxBase, err := covertCreateCrossTxReq(in)
	if err != nil {
		return nil, err
	}

	if err = n.fh.ValidateEnableSupport(crossTxBase.FabricChannel, crossTxBase.FabricChaincode, "", crossTxBase.FabricAmount); err != nil {
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
	handler := n.GetHandler(recv.NetworkType)
	return handler.ConstructAndSignTx(srv, recv)
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
	ticketId := in.CTxTicketId
	switch in.Op {
	case pb.TicketOps_approve:
		err := n.approveCtx(ticketId)
		var pbErr *pb.Error = nil
		if err != nil {
			pbErr = &pb.Error{
				Code:   -1,
				ErrMsg: err.Error(),
			}
		}
		return &pb.AdminOpTicketResp{
			Err: pbErr,
		}, nil
	case pb.TicketOps_reject:
		log.Printf("Reject cross transaction, ticket-id: %s\n", ticketId)
		return &pb.AdminOpTicketResp{
			Err: nil,
		}, nil
	case pb.TicketOps_quite:
		log.Printf("Quite cross transaction, ticket-id: %s\n", ticketId)
		return &pb.AdminOpTicketResp{
			Err: nil,
		}, nil

	}
	return &pb.AdminOpTicketResp{
		Err: nil,
	}, nil
}

func (n *NotaryService) approveCtx(ticketId string) error {
	for _, handler := range n.handlers {
		err := handler.Approve(ticketId)
		if err != nil {
			return err
		}
	}
	return nil
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

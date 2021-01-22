package main

import (
	"context"
	"github.com/ehousecy/notary-samples/notary-server/db/services"
	"github.com/ehousecy/notary-samples/notary-server/eth"
	"github.com/ehousecy/notary-samples/notary-server/fabric/tx"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/go-playground/validator/v10"
	"log"
)

var validate *validator.Validate

// handler names in the service system
const (
	ETHHandler = "ETHHandler"
	FabricHandler = "FabricHandler"
)

type NotaryService struct {
	provider services.CrossTxDataService
	fh       tx.Handler
	handlers map[string]TxHandler
}

type TxExecResult struct {
	Err       error
	TxReceipt string
}

type TxHandler interface {
	ValidateTx([]byte, string) bool
	SendTx(txData []byte)
	BuildTx(args ...string) []byte
	SignTx(priv string, ticketId string) []byte
	ConstructAndSignTx(src pb.NotaryService_SubmitTxServer) error
}



func NewNotaryService() *NotaryService {
	validate = validator.New()
	n :=  &NotaryService{
		provider: services.NewCrossTxDataServiceProvider(),
		fh:       tx.New(),
	}
	n.AddHandler(ETHHandler, eth.NewEthHandler("todo"))
	return n
}

func (n *NotaryService)AddHandler(handlerName string, handler TxHandler) *NotaryService {
	_, exist := n.handlers[handlerName]
	if exist{
		log.Printf("[Warning] handler already exist, updating handler %s\n", handlerName)
	}

	n.handlers[handlerName] = handler
	return n
}

func (n *NotaryService)GetHandler(code pb.TransferPropertyRequest_NetworkType) TxHandler  {
	switch code {
	case pb.TransferPropertyRequest_fabric:
		return n.handlers[FabricHandler]
	case pb.TransferPropertyRequest_eth:
		return n.handlers[ETHHandler]
	default:
		log.Printf("[Waring] Unkonw transaction type %d\n", code)
		return nil
	}

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
	return handler.ConstructAndSignTx(srv)
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

package fabric

import (
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Handler interface {
	ConstructAndSignTx(srv pb.NotaryService_SubmitTxServer, recv *pb.TransferPropertyRequest) error
	Approve(ticketId string) error
	HandleTxStatusBlock(channelID string, fb *peer.FilteredBlock)
	ValidateEnableSupport(channelID, chaincodeName, assetType, asset string) error
	QueryLastFabricBlockNumber(channelID string) (uint64, error)
	QueryConfirmingTx() error
}

package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/ehousecy/notary-samples/cli/grpc"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"math/big"
)

// this file implement the folowing cmd
//notarycli submit-tx --ticket-id --privatekey --network-type

var submitTxCmd = &cobra.Command{
	Use: "submit",
	Run: execSubmitCmd,
}

const (
	subTicketKey  = "subTicketKey"
	subPrivKey    = "subPrivKey"
	subNetworkKey = "subNetworkKey"
	ETHType = "ethereum"
	FabricType = "fabric"
)

func init() {
	err := addStringOption(submitTxCmd, subTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subPrivKey, privateKeyOption, "", "", privateKeyOption, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subNetworkKey, networkTypeOption, "", "", networkDescription, required)
}

// get notary service details according the target ticket id, construct raw transaction and sign
func execSubmitCmd(cmd *cobra.Command, args []string) {
	ticketId := viper.GetString(subTicketKey)
	priv := viper.GetString(subPrivKey)
	client := grpc.NewClient()
	stream, err := client.SubmitTx(context.Background(), nil)
	if err != nil {
		log.Printf("Submit transaction failed, %v\n", err)
		return
	}
	txType := viper.GetString(subNetworkKey)
	builder := NewTxBuilder(txType)
	if builder == nil {
		log.Printf("Unrecoganised transaction type, %v\n", err)
		return
	}
	err = builder.BuildTx(ticketId, priv, stream)
	if err != nil {
		log.Printf("Build transaction failed, %v\n", err)
		return
	}

}

// transaction builder
type TxBuilder interface {
	BuildTx(ticketId, priv string, stream pb.NotaryService_SubmitTxClient)  error
}

func NewTxBuilder(txType string) TxBuilder  {
	switch txType {
	case ETHType:
		e := &EthBuilder{}
		return e
	default:
		return nil
	}
}

type EthBuilder struct {

}

func (e *EthBuilder) BuildTx(ticketId, priv string, stream pb.NotaryService_SubmitTxClient)  error {
	err := stream.Send(&pb.TransferPropertyRequest{
		CTxId: ticketId,
		NetworkType: pb.TransferPropertyRequest_eth,
	})
	if err != nil {
		return err
	}
	src, err := stream.Recv()
	if err != nil {
		return err
	}

	// validate received stream
	if src.Error != nil {
		return errors.New(src.Error.ErrMsg)
	}
	txRawData := src.TxData
	tx := decodeTx(txRawData)
	if tx == nil {
		errMsg := fmt.Sprintf("Invalid transaction data：[%x]", txRawData)
		return errors.New(errMsg)
	}
	privKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		return errors.New("Invalid private key ")
	}
	chainId := new(big.Int).SetUint64(1)
	signed, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privKey)
	if err != nil {
		return err
	}

	signedBytes, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return err
	}
	err = stream.Send(&pb.TransferPropertyRequest{
		Data: signedBytes,
	})
	return err
}

//helper functions here

// decode ethereum transaction
func decodeTx(tx []byte) *types.Transaction {
	var decoded *types.Transaction
	if err := rlp.DecodeBytes(tx, &decoded); err != nil {
		return nil
	}
	return decoded
}
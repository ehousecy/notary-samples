package cmd

import (
	"context"
	"io"

	"github.com/ehousecy/notary-samples/cli/fabutil"
	"github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"

	"errors"
	"fmt"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
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

	subCertKey = "subCertKey"
	subMSPKey  = "subMSPKey"

	ETHType    = "ethereum"
	FabricType = "fabric"
)

func init() {
	err := addStringOption(submitTxCmd, subTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subPrivKey, privateKeyOption, "", "", privateKeyDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subNetworkKey, networkTypeOption, "", "", networkDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subCertKey, signCertOption, "", "", signCertDescription, optional)
	exitErr(err)
	err = addStringOption(submitTxCmd, subMSPKey, mspIDOption, "", "", mspIDDescription, optional)
	exitErr(err)
}

// get notary service details according the target ticket id, construct raw transaction and sign
func execSubmitCmd(cmd *cobra.Command, args []string) {

	ticketID := viper.GetString(subTicketKey)
	privateKey := viper.GetString(subPrivKey)
	network := viper.GetString(subNetworkKey)
	if network == FabricType {
		execFabricSubmit(ticketID, privateKey)
	} else {
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
		err = builder.BuildTx(ticketID, privateKey, stream)
		if err != nil {
			log.Printf("Build transaction failed, %v\n", err)
			return
		}
	}

}

type TxBuilder interface {
	BuildTx(ticketId, priv string, stream pb.NotaryService_SubmitTxClient) error
}

func execFabricSubmit(ticketID, privateKeyPath string) {
	signCert := viper.GetString(subCertKey)
	mspID := viper.GetString(subMSPKey)
	if signCert == "" {
		log.Fatalf("submit fabric tx %s is necessary", signCertOption)
	}
	if mspID == "" {
		log.Fatalf("submit fabric tx %s is necessary", mspIDOption)
	}
	client := grpc.NewClient()
	srv, err := client.SubmitTx(context.Background())
	exitErr(err)
	//获取creator
	creator, err := fabutil.GetCreator(mspID, signCert)
	exitErr(err)
	//获取签名
	privateKey, err := fabutil.GetPrivateKey(privateKeyPath)
	exitErr(err)
	err = srv.Send(&proto.TransferPropertyRequest{
		Data:        creator,
		CTxId:       ticketID,
		NetworkType: proto.TransferPropertyRequest_fabric,
	})
	exitErr(err)
	var i = 1
	for {
		recv, err := srv.Recv()
		if err == io.EOF {
			break
		}
		exitErr(err)
		if i <= 2 {
			fmt.Printf("第[%v]次收到消息:%v", i, string(recv.TxData))
			//签名proposal
			sign, err := fabutil.Sign(recv.TxData, privateKey)
			exitErr(err)
			err = srv.Send(&proto.TransferPropertyRequest{Data: sign})
			exitErr(err)
			i++
		}
	}

	// err = srv.CloseSend()
	// exitErr(err)
	log.Println("end transfer FabricSubmitTx method=====================")

}

func NewTxBuilder(txType string) TxBuilder {
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

func (e *EthBuilder) BuildTx(ticketId, priv string, stream pb.NotaryService_SubmitTxClient) error {
	err := stream.Send(&pb.TransferPropertyRequest{
		CTxId:       ticketId,
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

package cmd

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ehousecy/notary-samples/cli/fabutil"
	"github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"

	"errors"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
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

	subCertKey   = "subCertKey"
	subMSPKey    = "subMSPKey"
	subMSPDirKey = "subMSPDirKey"

	ETHType    = "ethereum"
	FabricType = "fabric"
)

func init() {
	err := addStringOption(submitTxCmd, subTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subPrivKey, privateKeyOption, "p", "", privateKeyDescription, optional)
	exitErr(err)
	err = addStringOption(submitTxCmd, subNetworkKey, networkTypeOption, "t", "", networkDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subCertKey, signCertOption, "c", "", signCertDescription, optional)
	exitErr(err)
	err = addStringOption(submitTxCmd, subMSPKey, mspIDOption, "", "", mspIDDescription, optional)
	exitErr(err)
	err = addStringOption(submitTxCmd, subMSPDirKey, mspPathOption, "", "", mspPathDescription, optional)
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
		stream, err := client.SubmitTx(context.Background())
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
	log.Println("start transfer FabricSubmitTx method=====================")
	mspDir := viper.GetString(subMSPDirKey)
	signCert := viper.GetString(subCertKey)
	mspID := viper.GetString(subMSPKey)
	if mspID == "" {
		log.Fatalf("submit fabric tx mspid is necessary, please use --%v specify", mspIDOption)
	}
	if mspDir != "" {
		pkDir, err := ioutil.ReadDir(filepath.Join(mspDir, "keystore"))
		exitErr(err)
		for _, file := range pkDir {
			if strings.HasSuffix(file.Name(), "_sk") {
				privateKeyPath = filepath.Join(mspDir, "keystore", file.Name())
				break
			}
		}
		certs, err := ioutil.ReadDir(filepath.Join(mspDir, "signcerts"))
		exitErr(err)
		for _, cert := range certs {
			if strings.HasSuffix(cert.Name(), ".pem") {
				signCert = filepath.Join(mspDir, "signcerts", cert.Name())
				break
			}
		}
	}
	if signCert == "" {
		log.Fatalf("submit fabric tx sign cert is necessary, please use --%s or -c specify, or use --%s specify msp dir path", signCertOption, mspPathOption)
	}
	if privateKeyPath == "" {
		log.Fatalf("submit fabric tx private key is necessary, please use --%s or -p specify, or use --%s specify msp dir path", privateKeyOption, mspPathOption)
	}
	//获取creator
	creator, err := fabutil.GetCreator(mspID, signCert)
	exitErr(err)
	//获取签名
	privateKey, err := fabutil.GetPrivateKey(privateKeyPath)
	exitErr(err)

	client := grpc.NewClient()
	srv, err := client.SubmitTx(context.Background())
	exitErr(err)
	err = srv.Send(&pb.TransferPropertyRequest{
		Data:        creator,
		TicketId:       ticketID,
		NetworkType: pb.NetworkType_fabric,
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
			//签名proposal
			sign, err := fabutil.Sign(recv.TxData, privateKey)
			exitErr(err)
			err = srv.Send(&pb.TransferPropertyRequest{Data: sign})
			exitErr(err)
			i++
		} else {
			log.Println(string(recv.TxData))
		}
	}
	_ = srv.CloseSend()
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
	privKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		return errors.New("Invalid private key ")
	}
	err = stream.Send(&pb.TransferPropertyRequest{
		TicketId:       ticketId,
		NetworkType: pb.NetworkType_eth,
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
	txContent := src.TxData
	var h common.Hash
	h.SetBytes(txContent)
	sig, err := crypto.Sign(h[:], privKey)
	if err != nil {
		return err
	}
	err = stream.Send(&pb.TransferPropertyRequest{
		Data: sig,
	})
	if err != nil {
		return err
	}
	_, err = stream.Recv()
	if err == io.EOF {
		return nil
	}
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

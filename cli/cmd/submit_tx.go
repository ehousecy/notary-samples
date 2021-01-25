package cmd

import (
	"context"
	"github.com/ehousecy/notary-samples/cli/fabutil"
	"github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
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
	subCertKey    = "subCertKey"
	subMSPKey     = "subMSPKey"
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
	if network == "fabric" {
		signCert := viper.GetString(subCertKey)
		mspID := viper.GetString(subMSPKey)
		if signCert == "" {
			log.Fatalf("submit fabric tx %s is necessary", signCertOption)
		}
		if mspID == "" {
			log.Fatalf("submit fabric tx %s is necessary", mspIDOption)
		}
		execFabricSubmit(ticketID, privateKey, signCert, mspID)
	}

}

//todo
// construct raw transaction for the user, display the info and let user sign tx data

func execFabricSubmit(ticketID, privateKeyPath, signCert, mspID string) {
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
	recv, err := srv.Recv()
	exitErr(err)
	//签名proposal
	sign, err := fabutil.Sign(recv.TxData, privateKey)
	exitErr(err)
	err = srv.Send(&proto.TransferPropertyRequest{Data: sign})
	exitErr(err)
	recv, err = srv.Recv()
	exitErr(err)
	//签名交易
	sign, err = fabutil.Sign(recv.TxData, privateKey)
	exitErr(err)
	err = srv.Send(&proto.TransferPropertyRequest{Data: sign})
	exitErr(err)
	log.Println("end transfer FabricSubmitTx method=====================")

}

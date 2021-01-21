package cmd

import (
	"context"
	"github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"math/big"
)

// createTx implements the flowing command
// ```notarycli create-ticket --efrom --eto --emount --ffrom --fto --famount --fchannel --fcc```

const (
	//bind key names
	crXeFromKey      = "crXeFromKey"
	crXeToKey        = "crXeToKey"
	crXeAmountKey    = "crXeAmountKey"
	crXfFromKey      = "crXfFromKey"
	crXfToKey        = "crXfToKey"
	crXfAmountKey    = "crXfAmountKey"
	crXfChannelKey   = "crXfChannelKey"
	crXfChaincodeKey = "crXfChaincodeKey"
)

// add command options when initialize the command
func init() {
	err := addStringOption(createTxCmd, crXeFromKey, eFromOption, "", "", fromDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXeToKey, eToOption, "", "", toDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXeAmountKey, eAmountOption, "", "", amountDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXfFromKey, fFromOption, "", "", fromDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXfToKey, fToOption, "", "", toDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXfAmountKey, fAmountOption, "", "", amountDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXfChannelKey, fchannelOption, "", "", channelDescription, required)
	exitErr(err)
	err = addStringOption(createTxCmd, crXfChaincodeKey, fChaincodeOption, "", "", chaincodeDescription, required)
	exitErr(err)
}

// createTxCmd submits the initial cross transaction ticket to notary service with the required fields
var createTxCmd = &cobra.Command{
	Use: "create-ticket",
	Run: execCreateCmd,
}

// execute create cross transaction
func execCreateCmd(cmd *cobra.Command, args []string) {
	eAmount := viper.GetString(crXeAmountKey)
	if !isValidAmount(eAmount) {
		log.Fatalf("Invalid Ethereum Amount received: %s", eAmount)
	}
	fAmount := viper.GetString(crXfAmountKey)
	if !isValidAmount(fAmount) {
		log.Fatalf("Invalid fabric Amount received: %s", fAmount)
	}
	efrom := viper.GetString(crXeFromKey)
	eto := viper.GetString(crXeToKey)
	ffrom := viper.GetString(crXfFromKey)
	fto := viper.GetString(crXfToKey)
	chaincodeName := viper.GetString(crXfChaincodeKey)
	channelName := viper.GetString(crXfChannelKey)

	client := grpc.NewClient()
	var ticketDetail = proto.CrossTxDetail{
		EFrom:          efrom,
		ETo:            eto,
		EAmount:        eAmount,
		FFrom:          ffrom,
		FTo:            fto,
		FAmount:        fAmount,
		FChannel:       channelName,
		FChaincodeName: chaincodeName,
	}
	resp, err := client.CreateCTX(context.Background(), &proto.CreateCrossTxReq{
		Detail: &ticketDetail,
	})
	exitErr(err)
	log.Printf("Successfully created ticket!\n")
	log.Printf(resp.String())

}

func isValidAmount(amount string) bool {
	bigFloat := new(big.Float)
	_, ok := bigFloat.SetString(amount)
	return ok
}

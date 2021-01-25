package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

//1. notarycli create-ticket --efrom --eto --emount --ffrom --fto --famount --fchannel --fcc
//2. notarycli submit-tx --ticket-id --privatekey --network-type
//3. notarycli list tickets
//4. notarycli list ticket --ticket-id
//5. notarycli approve --ticket-id
//6. notarycli reject --ticket-id

const (
	// cli flags bind to sub-commands
	// command options are listed here
	eFromOption       = "efrom"
	eToOption         = "eto"
	eAmountOption     = "eamount"
	fFromOption       = "ffrom"
	fToOption         = "fto"
	fAmountOption     = "famount"
	fchannelOption    = "fchannel"
	fChaincodeOption  = "fcc"
	ticketIdOption    = "ticket-id"
	privateKeyOption  = "private-key"
	networkTypeOption = "network-type"
	signCertOption    = "sign-cert"
	mspIDOption       = "msp-id"
)

// option description or command description
const (
	fromDescription       = "The account address or account name, which is used to send property to the escrow account"
	toDescription         = "The account address or account name, which is used to receive the property on blockchain"
	amountDescription     = "The amount of the property"
	channelDescription    = "Channel name of the chaincode installed"
	chaincodeDescription  = "Chaincode name where the fabric property is recorded"
	ticketDescription     = "Notary service created cross transaction ticket id"
	privateKeyDescription = "Private key used to sign the transaction, should match with the cross transaction ticket related public key"
	networkDescription    = "On which blockchain network this transaction is sending to, ie. ethereum, fabric, btc"
	signCertDescription   = "Sign cert provide truly public key verification signature, fabric is necessary"
	mspIDDescription      = "fabric MSP ID"
)
const (
	required = true
	optional = false
)

func addStringOption(cmd *cobra.Command, bindKeyName, keyName, shortName, defaultValue, description string, required bool) error {
	cmd.Flags().StringP(keyName, shortName, defaultValue, description)
	err := viper.BindPFlag(bindKeyName, cmd.Flag(keyName))
	if err != nil {
		return err
	}
	if required {
		return cmd.MarkFlagRequired(keyName)
	}
	return nil
}

func exitErr(err error) {
	if err != nil {
		log.Fatalf("Error accured: %v", err)
	}
}

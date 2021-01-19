package cmd

import "github.com/spf13/cobra"

// this file implement the folowing cmd
//notarycli submit-tx --ticket-id --privatekey --network-type

var submitTxCmd = &cobra.Command{
	Use: "submit",
	Run: execSubmitCmd,
}

const (
	subTicketKey = "subTicketKey"
	subPrivKey = "subPrivKey"
	subNetworkKey = "subNetworkKey"
)

func init()  {
	err := addStringOption(submitTxCmd, subTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subPrivKey, privateKeyOption, "", "", privateKeyOption, required)
	exitErr(err)
	err = addStringOption(submitTxCmd, subNetworkKey, networkTypeOption, "", "", networkDescription, required)
}

// get notary service details according the target ticket id, construct raw transaction and sign
func execSubmitCmd(cmd *cobra.Command, args []string)  {

}

//todo
// construct raw transaction for the user, display the info and let user sign tx data
package cmd

import (
	"context"
	"github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const (
	queryTicketKey = "queryTicketKey"
)

func init() {
	err := addStringOption(queryCmd, queryTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
}

var queryCmd = &cobra.Command{
	Use: "query",
	Run: execQueryCmd,
}

// execut query commands
func execQueryCmd(cmd *cobra.Command, args []string) {
	ticketId := viper.GetString(queryTicketKey)
	client := grpc.NewClient()
	resp, err := client.GetTicket(context.Background(), &proto.QueryTxRequest{
		TicketId: ticketId,
	})
	exitErr(err)
	// todo
	// gracefully display response
	log.Printf("%s", resp.String())
}

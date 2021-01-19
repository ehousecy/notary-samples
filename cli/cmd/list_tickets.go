package cmd

import (
	"context"
	"github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"log"
)
// this file implement the folowing command
// notarycli list tickets
var listTicketCmd = &cobra.Command{
	Use: "list",
	Run: execListCmd,
}

// execute list all the tickets on-going
func execListCmd(cmd *cobra.Command, args []string) {
	client := grpc.NewClient()
	resp, err := client.ListTickets(context.Background(), &proto.Empty{})
	exitErr(err)
	log.Printf("%s", resp.String())
}
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
	appTicketKey = "appTicketKey"
)

func init()  {
	err := addStringOption(approveCmd, appTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
}

var approveCmd = &cobra.Command{
	Use: "approve",
	Run: execApproveCmd,
}

func execApproveCmd(cmd *cobra.Command, args []string) {
	ticketId := viper.GetString(appTicketKey)
	client := grpc.NewClient()
	op := proto.TicketOps_approve
	resp, err := client.OpTicket(context.Background(), &proto.AdminOpTicketReq{
		CTxTicketId: ticketId,
		Op: op,
	})
	exitErr(err)
	log.Printf("Received resp: %v\n", resp.String())
}
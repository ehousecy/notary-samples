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
	rejectTicketKey = "rejectTicketKey"
)

func init()  {
	err := addStringOption(approveCmd, rejectTicketKey, ticketIdOption, "", "", ticketDescription, required)
	exitErr(err)
}

var rejectCmd = &cobra.Command{
	Use: "reject",
	Run: execRejectCmd,
}

func execRejectCmd(cmd *cobra.Command, args []string) {
	ticketId := viper.GetString(appTicketKey)
	client := grpc.NewClient()
	op := proto.TicketOps_reject
	resp, err := client.OpTicket(context.Background(), &proto.AdminOpTicketReq{
		CTxTicketId: ticketId,
		Op: op,
	})
	exitErr(err)
	log.Printf("Received resp: %v\n", resp.String())
}
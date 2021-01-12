package cmd

import (
	"context"
	. "github.com/ehousecy/notary-samples/cli/grpc"
	"github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"log"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewClient()
		resp, err := client.TestDial(context.Background(), &proto.Ping{
			Ping: "ping",
		})
		if err != nil {
			panic(err)
		}
		log.Printf("received pong: %s\n", resp.GetPong())
	},
}

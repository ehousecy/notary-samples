package cmd

import (
	"context"
	"fmt"
	"github.com/ehousecy/notary-samples/cli/grpc"
	pb "github.com/ehousecy/notary-samples/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var queryAccountCmd = &cobra.Command{
	Use: "account",
	Run: execQueryAccountCmd,
}

const (
	queryAccountKey   = "queryAccountKey"
	queryChannelKey   = "queryChannelKey"
	queryChaincodeKey = "queryChaincodeKey"
	queryNetworkKey   = "queryNetworkKey"
)

func init() {
	err := addStringOption(queryAccountCmd, queryAccountKey, accountOption, "", "", accountDescription, required)
	exitErr(err)

	err = addStringOption(queryAccountCmd, queryChannelKey, fchannelOption, "", "", channelDescription, optional)
	exitErr(err)

	err = addStringOption(queryAccountCmd, queryChaincodeKey, fChaincodeOption, "", "", chaincodeDescription, optional)
	exitErr(err)

	err = addStringOption(queryAccountCmd, queryNetworkKey, networkTypeOption, "", "", networkDescription, required)
	exitErr(err)
}

func execQueryAccountCmd(cmd *cobra.Command, args []string) {
	client := grpc.NewClient()
	account := viper.GetString(queryAccountKey)
	netType := viper.GetString(queryNetworkKey)
	var req = &pb.QueryBlockReq{}
	switch netType {
	case ETHType:
		req.Network = pb.NetworkType_eth
		req.Account = &pb.QueryBlockReq_EthAcc{EthAcc: account}
	case FabricType:
		channelName := viper.GetString(queryChannelKey)
		chaincodeName := viper.GetString(queryChaincodeKey)
		fabricAccount := &pb.FabricAccout{
			AccountInfo:   account,
			ChaincodeName: chaincodeName,
			ChannelName:   channelName,
		}
		req.Network = pb.NetworkType_fabric
		req.Account = &pb.QueryBlockReq_FabricAcc{FabricAcc: fabricAccount}
	default:
		fmt.Println("Unknown network type: %s", netType)
		return
	}

	resp, err := client.QueryBlock(context.Background(), req)
	if err != nil {
		fmt.Println("Query Account error:", err)
		return
	}
	fmt.Println(resp)

}

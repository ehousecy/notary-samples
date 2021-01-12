package cli

import (
	"github.com/ehousecy/notary-samples/cli/cmd"
	"log"
)

// this is a cli cli
// this cli is used to interact with notary server
// supported grpc request is list below

func main()  {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalf("execute command failed %v", err)
	}
}
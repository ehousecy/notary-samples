package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is the root command for client commands.
var RootCmd = &cobra.Command{
	Use:   "ebaas-cli",
	Short: "ebaas-cli is used to interact with ehouse different services",
	ValidArgs:nil,
}

func init()  {
	RootCmd.AddCommand(testCmd)
}

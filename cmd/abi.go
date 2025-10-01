package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/cmd/abi"
)

func init() {
	abiCmd.AddCommand(abi.DownloadCmd)
	abiCmd.AddCommand(abi.DeleteCmd)
}

// abiCmd centralizes the ABI related commands.
var abiCmd = &cobra.Command{
	Use:     "abi",
	Aliases: []string{"a"},
	Short:   "Manage smart contract ABIs",
}

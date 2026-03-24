package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/cmd/abi"
)

func init() {
	// Parent args applied to all subcommands
	// Chain ID
	abiCmd.PersistentFlags().IntP("chainid", "c", 1, "Chain ID (e.g., 1 for Ethereum Mainnet)")
	// ABI storage
	abiCmd.PersistentFlags().StringP("abi-store", "s", "$HOME/.config/abitool/abis/", "Directory to store ABI files")

	// Viper config to bind args
	if err := viper.BindPFlag("chainid", abiCmd.PersistentFlags().Lookup("chainid")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-store", abiCmd.PersistentFlags().Lookup("abi-store")); err != nil {
		panic(err)
	}

	abiCmd.AddCommand(abi.DownloadCmd)
	abiCmd.AddCommand(abi.DeleteCmd)
	abiCmd.AddCommand(abi.ViewCmd)
	abiCmd.AddCommand(abi.ListCmd)
	abiCmd.AddCommand(abi.ImportCmd)
}

// abiCmd centralizes the ABI related commands.
var abiCmd = &cobra.Command{
	Use:     "abi",
	Aliases: []string{"a"},
	Short:   "Manage smart contract ABIs",
}

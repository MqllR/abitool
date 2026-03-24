package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/cmd/rpc"
)

func init() {
	// Persistent flags shared by all rpc subcommands.
	rpcCmd.PersistentFlags().IntP("chainid", "c", 1, "Chain ID (e.g., 1 for Ethereum Mainnet)")
	rpcCmd.PersistentFlags().StringP("abi-store", "s", "$HOME/.config/abitool/abis/", "Directory where ABI files are stored")
	rpcCmd.PersistentFlags().String("rpc-url", "", "JSON-RPC endpoint URL (overrides rpc.url in config)")

	if err := viper.BindPFlag("chainid", rpcCmd.PersistentFlags().Lookup("chainid")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-store", rpcCmd.PersistentFlags().Lookup("abi-store")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("rpc-url", rpcCmd.PersistentFlags().Lookup("rpc-url")); err != nil {
		panic(err)
	}

	rpcCmd.AddCommand(rpc.CallCmd)
}

// rpcCmd groups commands that interact with an Ethereum JSON-RPC node.
var rpcCmd = &cobra.Command{
	Use:     "rpc",
	Aliases: []string{"r"},
	Short:   "Interact with an Ethereum JSON-RPC node",
}

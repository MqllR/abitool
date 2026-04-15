package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/cmd/abitool/abi"
	"github.com/MqllR/abitool/internal/abitool"
)

func init() {
	// Parent args applied to all subcommands
	// Chain ID
	abiCmd.PersistentFlags().Int("chainid", 1, "Chain ID (e.g., 1 for Ethereum Mainnet)")
	// ABI storage
	abiCmd.PersistentFlags().StringP("abi-store", "s", "$HOME/.config/abitool/abis/", "Directory to store ABI files")

	// Viper config to bind args at init time as a fallback.
	// The PersistentPreRunE below rebinds at execution time to ensure the parsed
	// flag values (not defaults) are reflected in viper.
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
	abiCmd.AddCommand(abi.RenameCmd)
}

// abiCmd centralizes the ABI related commands.
var abiCmd = &cobra.Command{
	Use:     "abi",
	Aliases: []string{"a"},
	Short:   "Manage smart contract ABIs",
	// PersistentPreRunE re-binds flags at execution time so viper picks up the
	// actual parsed values. This is the canonical fix for the Cobra/Viper issue
	// where BindPFlag at init() time does not see flag values changed by a child
	// command's flag traversal. This also replaces the root PersistentPreRun for
	// all abi subcommands.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("binding flags: %w", err)
		}
		if err := abitool.Load(); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
}

// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/internal/contract"
)

func init() {
	encodeCmd.PersistentFlags().Int("chainid", 1, "Chain ID (e.g., 1 for Ethereum Mainnet)")
	encodeCmd.PersistentFlags().StringP("abi-store", "s", "$HOME/.config/abitool/abis/", "Directory where ABI files are stored")
	encodeCmd.Flags().StringP("output", "o", "hex", "Output format: hex or json")

	if err := viper.BindPFlag("chainid", encodeCmd.PersistentFlags().Lookup("chainid")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-store", encodeCmd.PersistentFlags().Lookup("abi-store")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("encode-output", encodeCmd.Flags().Lookup("output")); err != nil {
		panic(err)
	}
}

// encodeCmd ABI-encodes function call arguments into EVM calldata.
var encodeCmd = &cobra.Command{
	Use:     "encode <address> <function> [arg...]",
	Aliases: []string{"e"},
	Short:   "Encode a function call into ABI calldata",
	Long: `ABI-encode a function call for a stored contract and print the calldata hex to stdout.

The contract ABI must already be stored locally (run 'abitool abi download' first).
Arguments are provided as positional parameters after the function name.

Array and tuple arguments must be given as JSON arrays, e.g. '[1,2,3]' or '["0xAddr","0xAddr2"]'.

Examples:
  # ERC-20 transfer — encode calldata only
  abitool encode 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 transfer \
    0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 1000000

  # JSON output (includes signature and selector)
  abitool encode --output json 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 transfer \
    0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 1000000

  # Pipe into cast send
  abitool encode 0xAddr approve 0xSpender 115792089237316195423570985008687907853269984665640564039457584007913129639935 \
    | xargs -I{} cast send 0xAddr {} --private-key $KEY`,
	Args: cobra.MinimumNArgs(2),
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("binding flags: %w", err)
		}
		if err := abitool.Load(); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
	RunE: runEncode,
}

func runEncode(cmd *cobra.Command, args []string) error {
	logger := log.Default()

	address := args[0]
	functionName := args[1]
	fnArgs := args[2:]

	manager, err := contract.NewEncodeManager(logger)
	if err != nil {
		logger.Printf("Error creating encode manager: %v", err)
		os.Exit(1)
	}

	opts := contract.EncodeOptions{
		OutputJSON: viper.GetString("encode-output") == "json",
	}

	if err := manager.EncodeCalldata(address, functionName, fnArgs, opts, os.Stdout); err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	return nil
}

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
	decodeCmd.PersistentFlags().Int("chainid", 1, "Chain ID (e.g., 1 for Ethereum Mainnet)")
	decodeCmd.PersistentFlags().StringP("abi-store", "s", "$HOME/.config/abitool/abis/", "Directory where ABI files are stored")

	decodeCmd.Flags().String("eth-call", "", `Decode an eth_call JSON request body (e.g. '{"to":"0x...","data":"0x..."}')`)
	decodeCmd.Flags().String("from-tx", "", "Decode a RLP-encoded signed transaction hex (as sent via eth_sendRawTransaction)")
	decodeCmd.Flags().Bool("return-data", false, "Decode return data: provide <address> <function-name> <return-hex> as positional args")
	decodeCmd.Flags().StringP("output", "o", "text", "Output format: text or json")

	if err := viper.BindPFlag("chainid", decodeCmd.PersistentFlags().Lookup("chainid")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-store", decodeCmd.PersistentFlags().Lookup("abi-store")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("decode-eth-call", decodeCmd.Flags().Lookup("eth-call")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("decode-from-tx", decodeCmd.Flags().Lookup("from-tx")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("decode-return-data", decodeCmd.Flags().Lookup("return-data")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("decode-output", decodeCmd.Flags().Lookup("output")); err != nil {
		panic(err)
	}
}

// decodeCmd decodes ABI-encoded calldata or return data using locally stored ABIs.
var decodeCmd = &cobra.Command{
	Use:     "decode",
	Aliases: []string{"d"},
	Short:   "Decode ABI-encoded calldata or return data",
	Long: `Decode EVM calldata or return data into human-readable form using a locally
stored contract ABI. The contract ABI must be downloaded first with 'abitool abi download'.

Modes:

  Raw calldata (default):
    abitool decode <address> <calldata-hex>

  eth_call JSON request body:
    abitool decode --eth-call '{"to":"0xAddr","data":"0xcalldata"}'

  RLP-encoded signed transaction (eth_sendRawTransaction):
    abitool decode --from-tx <raw-tx-hex>

  Return data from a function call:
    abitool decode --return-data <address> <function-name> <return-hex>

Examples:
  abitool decode 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 \
    0xa9059cbb000000000000000000000000d8da6bf26964af9d7eed9e03e53415d37aa960450000000000000000000000000000000000000000000000000000000005f5e100

  abitool decode --eth-call '{"to":"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48","data":"0xa9059cbb..."}'

  abitool decode --return-data 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 totalSupply \
    0x0000000000000000000000000000000000000000000000000000000005f5e100`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("binding flags: %w", err)
		}
		if err := abitool.Load(); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
	RunE: runDecode,
}

func runDecode(cmd *cobra.Command, args []string) error {
	logger := log.Default()

	manager, err := contract.NewDecodeManager(logger)
	if err != nil {
		logger.Printf("Error creating decode manager: %v", err)
		os.Exit(1)
	}

	opts := contract.DecodeOptions{
		OutputJSON: viper.GetString("decode-output") == "json",
	}

	ethCallJSON := viper.GetString("decode-eth-call")
	fromTx := viper.GetString("decode-from-tx")
	returnData := viper.GetBool("decode-return-data")

	switch {
	case ethCallJSON != "":
		if err := manager.DecodeFromEthCall(ethCallJSON, opts, os.Stdout); err != nil {
			logger.Println(err)
			os.Exit(1)
		}

	case fromTx != "":
		if err := manager.DecodeFromRawTx(fromTx, opts, os.Stdout); err != nil {
			logger.Println(err)
			os.Exit(1)
		}

	case returnData:
		if len(args) != 3 {
			return cmd.Help()
		}
		if err := manager.DecodeReturnData(args[0], args[1], args[2], opts, os.Stdout); err != nil {
			logger.Println(err)
			os.Exit(1)
		}

	default:
		// Raw calldata mode: abitool decode <address> <calldata-hex>
		if len(args) != 2 {
			return cmd.Help()
		}
		if err := manager.DecodeCalldata(args[0], args[1], opts, os.Stdout); err != nil {
			logger.Println(err)
			os.Exit(1)
		}
	}

	return nil
}


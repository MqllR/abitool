package rpc

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/contract"
)

func init() {
	CallCmd.Flags().BoolP("interactive", "i", false, "Prompt for each argument interactively (auto-enabled when inputs are expected but no args given)")
	CallCmd.Flags().String("block", "latest", "Block tag or hex number for eth_call (e.g. latest, pending, 0x10d4f)")
	CallCmd.Flags().StringP("output", "o", "text", "Output format: text or json")

	if err := viper.BindPFlag("rpc-call-interactive", CallCmd.Flags().Lookup("interactive")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("rpc-call-block", CallCmd.Flags().Lookup("block")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("rpc-call-output", CallCmd.Flags().Lookup("output")); err != nil {
		panic(err)
	}
}

// CallCmd sends an eth_call for a stored contract and prints the decoded result.
var CallCmd = &cobra.Command{
	Use:   "call <address> <function> [arg...]",
	Short: "Call a read-only contract function via eth_call",
	Long: `Call a view or pure contract function and display the decoded return value.

The contract ABI must already be stored locally (run 'abitool abi download' first).
Arguments can be provided on the command line or entered interactively.

Examples:
  # Provide arguments directly
  abitool rpc call 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 balanceOf 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045

  # Interactive mode (prompted for each argument)
  abitool rpc call 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 balanceOf --interactive`,
	Args: cobra.MinimumNArgs(2),
	Run:  callContract,
}

func callContract(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := log.Default()

	address := args[0]
	functionName := args[1]
	fnArgs := args[2:]

	callManager, err := contract.NewCallManager(logger)
	if err != nil {
		logger.Fatalf("Error creating call manager: %v", err)
	}

	opts := contract.CallOptions{
		Interactive: viper.GetBool("rpc-call-interactive"),
		Block:       viper.GetString("rpc-call-block"),
		OutputJSON:  viper.GetString("rpc-call-output") == "json",
	}

	if err := callManager.CallContract(ctx, address, functionName, fnArgs, opts, os.Stdout); err != nil {
		logger.Println(err)
		os.Exit(1)
	}
}

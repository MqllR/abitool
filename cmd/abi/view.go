package abi

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/contract"
	scontract "github.com/MqllR/abitool/pkg/storage/contract"
)

func init() {
	ViewCmd.Flags().StringP("output", "o", "json", "Output format: json or table")
	ViewCmd.Flags().StringP("type", "t", "all", "Filter by function type: all, function, event, constructor, fallback, receive")

	viper.BindPFlag("abi-view-output", ViewCmd.Flags().Lookup("output"))
	viper.BindPFlag("abi-view-type", ViewCmd.Flags().Lookup("type"))
}

// viewCmd displays a contract's ABI by its address.
// The command can display the ABI in pretty-printed JSON or with a table format. The table shows function names, types, and inputs.
// It can also allow the user to filter the ABI by function type.
var ViewCmd = &cobra.Command{
	Use:     "view <address>",
	Aliases: []string{"v"},
	Args:    cobra.ExactArgs(1),
	Short:   "View a contract's ABI by its address",
	Run:     view,
}

func view(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	log := log.Default()

	abiManager, err := contract.NewABIManager(log)
	if err != nil {
		log.Fatalf("Error creating ABI manager: %v", err)
	}

	abi, err := abiManager.GetABI(ctx, args[0])
	if err != nil {
		if err == scontract.ErrNotFound {
			log.Println("Contract not found. Download the ABI first.")
			os.Exit(1)
		}

		log.Println("Unexpected error:", err)
	}

	prettyABI, err := contract.Print(abi)
	if err != nil {
		log.Fatalf("Error pretty printing ABI: %v", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, prettyABI)
}

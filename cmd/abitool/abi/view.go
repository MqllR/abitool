package abi

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/contract"
)

func init() {
	ViewCmd.Flags().StringP("output", "o", "table", "Output format: json or table")
	ViewCmd.Flags().StringP("type", "t", "all", "Filter by function type: all, function, event, constructor, fallback, receive")
	ViewCmd.Flags().Bool("with-input-name", false, "Display input parameter names in table output")
	ViewCmd.Flags().Bool("with-output-name", false, "Display output parameter names in table output")

	if err := viper.BindPFlag("abi-view-output", ViewCmd.Flags().Lookup("output")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-view-type", ViewCmd.Flags().Lookup("type")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-view-with-intput-name", ViewCmd.Flags().Lookup("with-input-name")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag("abi-view-with-output-name", ViewCmd.Flags().Lookup("with-output-name")); err != nil {
		panic(err)
	}
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

	if err := abiManager.ViewABI(ctx, args[0], os.Stdout); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

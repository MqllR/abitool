package abi

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/internal/contract"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all stored contract ABIs",
	Run:     list,
}

func list(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	log := log.Default()
	abiManager, err := contract.NewABIManager(log)
	if err != nil {
		log.Fatalf("Error creating ABI manager: %v", err)
	}

	if err := abiManager.ListABIs(ctx, os.Stdout); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

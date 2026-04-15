package abi

import (
	"log"

	"github.com/MqllR/abitool/internal/contract"
	"github.com/spf13/cobra"
)

// deleteCmd deletes a contract by its address.
var DeleteCmd = &cobra.Command{
	Use:     "delete <address>",
	Aliases: []string{"del", "d"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete a contract by its address",
	Run:     delete,
}

func delete(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	log := log.Default()

	abiManager, err := contract.NewABIManager(log)
	if err != nil {
		log.Fatalf("Error creating ABI manager: %v", err)
	}

	err = abiManager.DeleteWithABI(ctx, args[0])
	if err != nil {
		log.Fatalf("Error downloading and storing ABI: %v", err)
	}
}

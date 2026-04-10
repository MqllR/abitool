package abi

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/internal/contract"
)

// RenameCmd sets or updates the user-defined label for a stored contract.
var RenameCmd = &cobra.Command{
	Use:   "rename <address> <label>",
	Args:  cobra.ExactArgs(2),
	Short: "Set or update the display label for a stored contract",
	Long: `Set a human-friendly label for a stored contract.
The label is shown instead of the Etherscan contract name in list and TUI views.
The original Etherscan name is preserved and shown in brackets when a label is set.`,
	Run: rename,
}

func rename(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	log := log.Default()

	abiManager, err := contract.NewABIManager(log)
	if err != nil {
		log.Fatalf("Error creating ABI manager: %v", err)
	}

	if err := abiManager.RenameContract(ctx, args[0], args[1]); err != nil {
		log.Fatalf("Error renaming contract: %v", err)
	}
}

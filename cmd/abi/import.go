package abi

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/internal/contract"
)

var ImportCmd = &cobra.Command{
	Use:     "import <address> <path>",
	Aliases: []string{"i", "imp"},
	Args:    cobra.ExactArgs(2),
	Short:   "Import a contract's ABI from a local JSON file",
	Run:     importABI,
}

func init() {
	ImportCmd.Flags().StringP("name", "n", "", "Contract name (defaults to address if omitted)")
	ImportCmd.Flags().BoolP("force", "f", false, "Overwrite if an ABI for this address already exists")
}

func importABI(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := log.Default()

	address := args[0]
	filePath := args[1]

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		logger.Fatalf("Error reading --name flag: %v", err)
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		logger.Fatalf("Error reading --force flag: %v", err)
	}

	abiManager, err := contract.NewABIManager(logger)
	if err != nil {
		logger.Fatalf("Error creating ABI manager: %v", err)
	}

	if err := abiManager.ImportABI(ctx, address, filePath, name, force); err != nil {
		logger.Fatalf("Error importing ABI: %v", err)
	}
}

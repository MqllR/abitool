package abi

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/internal/contract"
)

// downloadCmd downloads a contract's ABI by its address.
var DownloadCmd = &cobra.Command{
	Use:     "download <address>",
	Aliases: []string{"dl"},
	Args:    cobra.ExactArgs(1),
	Short:   "Download a contract's ABI by its address",
	Run:     download,
}

// download is the handler for the download command.
func download(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	log := log.Default()

	abiManager, err := contract.NewABIManager(log)
	if err != nil {
		log.Fatalf("Error creating ABI manager: %v", err)
	}

	err = abiManager.DownloadAndStoreABI(ctx, args[0])
	if err != nil {
		log.Fatalf("Error downloading and storing ABI: %v", err)
	}
}

package contract

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/MqllR/abitool/pkg/etherscan"
	"github.com/MqllR/abitool/pkg/storage"
)

/*
The ABI implementaiton manages smart contract ABIs,
including downloading, parsing, and storing them.
*/

var ErrEtherscanAPIKeyNotSet = errors.New("ETHERSCAN_API_KEY environment variable is not set")

type ABIManager struct {
	log *log.Logger

	etherscanClient *etherscan.Client
	store           *storage.Local
}

// NewABIManager creates a new instance of ABIManager. It instanciates the storage and etherscan client.
func NewABIManager(logger *log.Logger) (*ABIManager, error) {
	etherscanKey := os.Getenv("ETHERSCAN_API_KEY")
	if etherscanKey == "" {
		return nil, ErrEtherscanAPIKeyNotSet
	}

	etherscanClient := etherscan.NewClient(etherscanKey, etherscan.Mainnet)

	store, err := storage.NewLocal()
	if err != nil {
		return nil, err
	}

	return &ABIManager{
		log:             logger,
		etherscanClient: etherscanClient,
		store:           store,
	}, nil
}

// DownloadAndStoreABI downloads the ABI for a given contract address from Etherscan and stores it locally
func (a *ABIManager) DownloadAndStoreABI(ctx context.Context, address string) error {
	a.log.Printf("Downloading ABI for contract address: %s", address)
	a.log.Println("Fetching ABI from Etherscan...")

	abi, err := a.etherscanClient.GetABI(ctx, address)
	if err != nil {
		return err
	}

	a.log.Println("ABI fetched successfully. Saving locally...")

	err = a.store.SaveContractWithABI(address, abi, nil)
	if err != nil {
		return err
	}

	a.log.Println("ABI saved successfully.")
	return nil
}

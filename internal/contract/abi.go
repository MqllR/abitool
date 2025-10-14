package contract

import (
	"context"
	"errors"
	"log"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/pkg/etherscan"
	"github.com/MqllR/abitool/pkg/storage/abi"
	"github.com/MqllR/abitool/pkg/storage/contract"
)

/*
The ABI implementaiton manages smart contract ABIs,
including downloading, parsing, and storing them.
*/

var ErrEtherscanAPIKeyNotSet = errors.New("ETHERSCAN_API_KEY environment variable is not set")

type ABIManager struct {
	log *log.Logger

	etherscanClient *etherscan.Client
	contractStore   *contract.Local
	abiStore        *abi.Local
}

// NewABIManager creates a new instance of ABIManager. It instanciates the storage and etherscan client.
func NewABIManager(logger *log.Logger) (*ABIManager, error) {
	cfg := abitool.ConfigInstance()

	if cfg.EtherScan.APIKey == "" {
		return nil, ErrEtherscanAPIKeyNotSet
	}

	storeCfg := viper.GetString("abi-store")
	chainIdCfg := viper.GetInt("chainid")

	if _, ok := SupportedChainIDs[chainIdCfg]; !ok {
		return nil, errors.New("unsupported chain ID")
	}

	etherscanClient := etherscan.NewClient(cfg.EtherScan.APIKey, etherscan.FromInt(chainIdCfg))

	contractStore, err := contract.NewLocal(filepath.Join(storeCfg, strconv.Itoa(chainIdCfg)))
	if err != nil {
		return nil, err
	}

	abiStore, err := abi.NewLocal(filepath.Join(storeCfg, strconv.Itoa(chainIdCfg)))
	if err != nil {
		return nil, err
	}

	return &ABIManager{
		log:             logger,
		etherscanClient: etherscanClient,
		contractStore:   contractStore,
		abiStore:        abiStore,
	}, nil
}

// DownloadAndStoreABI downloads the ABI for a given contract address from Etherscan and stores it locally
func (a *ABIManager) DownloadAndStoreABI(ctx context.Context, address string) error {
	contractInfo, err := a.GetContract(address) // Check if contract already exists
	if err == nil && contractInfo != nil {
		a.log.Printf("Contract with address %s already exists. Skipping download.", address)

		return nil
	}

	a.log.Printf("Downloading ABI for contract address: %s", address)
	a.log.Println("Fetching ABI from Etherscan...")

	contract, err := a.etherscanClient.GetSourceCode(ctx, address)
	if err != nil {
		return err
	}

	a.log.Println("ABI fetched successfully. Saving locally...")

	meta := Metadata{
		ContractName: contract.ContractName,
	}

	err = a.SaveContractWithABI(
		address,
		&meta,
		contract.ABI,
	)
	if err != nil {
		return err
	}

	a.log.Println("ABI saved successfully.")
	return nil
}

func (a *ABIManager) DeleteWithABI(ctx context.Context, address string) error {
	_, err := a.GetContract(address) // Check if contract already exists
	if err != nil {
		if err != contract.ErrNotFound {
			return err
		}

		a.log.Printf("Contract with address %s does not exist. Nothing to delete.", address)
	}

	a.log.Printf("Deleting ABI and metadata for contract address: %s", address)

	if err := a.contractStore.Delete(address); err != nil {
		return err
	}

	if err := a.abiStore.Delete(address); err != nil {
		return err
	}

	a.log.Println("ABI and metadata deleted successfully.")

	return nil
}

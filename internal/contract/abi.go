// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/pkg/abiparser"
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

// NewABIManager creates a new instance of ABIManager. It initialises the storage backends and,
// if an Etherscan API key is configured, the Etherscan client. Commands that only need local
// storage (e.g. import) work without an API key; DownloadAndStoreABI will return an error if
// the client is not available.
func NewABIManager(logger *log.Logger) (*ABIManager, error) {
	cfg := abitool.ConfigInstance()

	storeCfg := viper.GetString("abi-store")
	chainIdCfg := viper.GetInt("chainid")

	contractStore, err := contract.NewLocal(filepath.Join(storeCfg, strconv.Itoa(chainIdCfg)))
	if err != nil {
		return nil, err
	}

	abiStore, err := abi.NewLocal(filepath.Join(storeCfg, strconv.Itoa(chainIdCfg)))
	if err != nil {
		return nil, err
	}

	m := &ABIManager{
		log:           logger,
		contractStore: contractStore,
		abiStore:      abiStore,
	}

	if cfg.EtherScan.APIKey != "" {
		m.etherscanClient = etherscan.NewClient(cfg.EtherScan.APIKey, etherscan.FromInt(chainIdCfg))
	}

	return m, nil
}

// DownloadAndStoreABI downloads the ABI for a given contract address from Etherscan and stores it
func (a *ABIManager) DownloadAndStoreABI(ctx context.Context, address string) error {
	if a.etherscanClient == nil {
		return ErrEtherscanAPIKeyNotSet
	}

	contractInfo, err := a.getContract(address) // Check if contract already exists
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

	err = a.saveContractWithABI(
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

// DeleteWithABI deletes the ABI and metadata for a given contract address from the storage
func (a *ABIManager) DeleteWithABI(ctx context.Context, address string) error {
	_, err := a.getContract(address)
	if err != nil {
		if errors.Is(err, contract.ErrNotFound) {
			a.log.Printf("Contract with address %s does not exist. Nothing to delete.", address)
			return nil
		}

		return err
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

// ViewABI retrieves the ABI for a given contract address from the storage
func (a *ABIManager) ViewABI(ctx context.Context, address string, out io.Writer) error {
	_, err := a.getContract(address) // Check if contract already exists
	if err != nil {
		if err == contract.ErrNotFound {
			return err
		}

		return fmt.Errorf("failed to get contract: %w", err)
	}

	bABI, err := a.abiStore.Read(address)
	if err != nil {
		return fmt.Errorf("failed to read ABI from store: %w", err)
	}

	parsedABI, err := abiparser.ParseABI(bABI)
	if err != nil {
		return fmt.Errorf("failed to parse ABI: %w", err)
	}

	PrintedABI, err := Print(parsedABI)
	if err != nil {
		return fmt.Errorf("failed to print ABI: %w", err)
	}

	_, _ = fmt.Fprintln(out, PrintedABI)

	return nil
}

func (a *ABIManager) ListABIs(ctx context.Context, out io.Writer) error {
	contracts, err := a.listContracts()
	if err != nil {
		return fmt.Errorf("listing contracts: %w", err)
	}

	_, _ = fmt.Fprintln(out, PrintContractList(contracts, viper.GetInt("chainid")))

	return nil
}

// ImportABI reads an ABI from a local file and stores it under the given address.
// If name is empty, the address is used as the contract name.
// If an entry for the address already exists and force is false, an error is returned.
func (a *ABIManager) ImportABI(ctx context.Context, address, filePath, name string, force bool) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read ABI file: %w", err)
	}

	if _, err := abiparser.ParseABI(string(data)); err != nil {
		return fmt.Errorf("invalid ABI file %s: %w", filePath, err)
	}

	if name == "" {
		name = address
	}

	existing, err := a.getContract(address)
	if err == nil && existing != nil {
		if !force {
			return fmt.Errorf("contract with address %s already exists; use --force to overwrite", address)
		}

		if err := a.DeleteWithABI(ctx, address); err != nil {
			return fmt.Errorf("failed to remove existing contract before overwrite: %w", err)
		}
	}

	meta := Metadata{
		ContractName: name,
	}

	if err := a.saveContractWithABI(address, &meta, string(data)); err != nil {
		return err
	}

	a.log.Println("ABI imported successfully.")

	return nil
}

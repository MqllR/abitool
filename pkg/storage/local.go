package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Local struct {
	path         string
	contractFile string
}

type contracts map[string]any

func NewLocal() (*Local, error) {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "abitool")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Local{
		path:         dir,
		contractFile: "contracts.json",
	}, nil
}

// SaveContractWithABI updates the contracts.json file with the contract address and metadata and store the ABI
func (l *Local) SaveContractWithABI(address, abi string, meta any) error {
	if err := l.updateContract(address, meta); err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}

	if err := l.saveABI(address, abi); err != nil {
		return fmt.Errorf("saving ABI: %w", err)
	}

	return nil
}

// updateContract saves or updates the contract address
func (l *Local) updateContract(address string, meta any) error {
	c, err := l.loadContracts()
	if err != nil {
		return fmt.Errorf("loading contracts: %w", err)
	}

	c[address] = meta

	if err := l.saveContracts(c); err != nil {
		return fmt.Errorf("saving contracts: %w", err)
	}

	return nil
}

// loadContracts reads the contracts.json file and returns the contracts map
func (l *Local) loadContracts() (contracts, error) {
	c := contracts{}
	path := filepath.Join(l.path, l.contractFile)

	fh, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}

		return nil, fmt.Errorf("opening contracts file: %w", err)
	}
	defer fh.Close()

	if err := json.NewDecoder(fh).Decode(&c); err != nil {
		return c, fmt.Errorf("reading contracts file: %w", err)
	}

	return c, err
}

// saveContracts writes the contracts map to the contracts.json file
func (l *Local) saveContracts(c contracts) error {
	path := filepath.Join(l.path, l.contractFile)

	fh, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening contracts file: %w", err)
	}
	defer fh.Close()

	if err := json.NewEncoder(fh).Encode(c); err != nil {
		return fmt.Errorf("writing contracts file: %w", err)
	}

	return nil
}

// saveABI saves the ABI to a file named <contract_address>.json
func (l *Local) saveABI(address, abi string) error {
	path := filepath.Join(l.path, address+".json")

	fh, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening ABI file: %w", err)
	}

	defer fh.Close()

	_, err = fh.WriteString(abi)
	if err != nil {
		return fmt.Errorf("writing ABI file: %w", err)
	}

	return nil
}

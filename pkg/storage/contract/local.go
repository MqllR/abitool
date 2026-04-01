// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
)

var (
	ErrAlreadyExists = errors.New("contract already exists")
	ErrNotFound      = errors.New("contract not found")
)

// Local implements a local file-based contract storage.
// All contracts are stored in a single JSON file which is a map with the address as key and the metadata as value.
type Local struct {
	contractFile string
}

// internal structure of the contract file
type contracts map[string]map[string]any

func NewLocal(storePath string) (*Local, error) {
	storePathWithEnv := os.ExpandEnv(storePath)

	if err := os.MkdirAll(storePathWithEnv, 0755); err != nil {
		return nil, err
	}

	return &Local{
		contractFile: filepath.Join(storePathWithEnv, "contracts.json"),
	}, nil
}

// Add adds a contract
func (l *Local) Add(address string, meta []byte) error {
	c, err := l.getContracts()
	if err != nil {
		return fmt.Errorf("loading contracts: %w", err)
	}

	if c[address] != nil {
		return ErrAlreadyExists
	}

	var metaMap map[string]any
	if err := json.Unmarshal(meta, &metaMap); err != nil {
		return fmt.Errorf("unmarshaling metadata: %w", err)
	}

	c[address] = metaMap

	if err := l.saveContracts(c); err != nil {
		return fmt.Errorf("saving contracts: %w", err)
	}

	return nil
}

// Get returns a contract info
func (l *Local) Get(address string) ([]byte, error) {
	c, err := l.getContracts()
	if err != nil {
		return nil, fmt.Errorf("loading contracts: %w", err)
	}

	if c[address] == nil {
		return nil, ErrNotFound
	}

	jsn, err := json.Marshal(c[address])
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata: %w", err)
	}

	return jsn, nil
}

// Delete deletes a contract
func (l *Local) Delete(address string) error {
	c, err := l.getContracts()
	if err != nil {
		return fmt.Errorf("loading contracts: %w", err)
	}

	if c[address] == nil {
		return ErrNotFound
	}

	delete(c, address)

	if err := l.saveContracts(c); err != nil {
		return fmt.Errorf("saving contracts: %w", err)
	}

	return nil
}

// All return an iterator over all ABI elements.
func (l *Local) List() (iter.Seq[string], error) {
	contracts, err := l.getContracts()
	if err != nil {
		return nil, fmt.Errorf("loading contracts: %w", err)
	}

	return func(yield func(string) bool) {
		for address := range contracts {
			if !yield(address) {
				return
			}
		}
	}, nil
}

// getContracts reads the contracts.json file and returns the contracts map
func (l *Local) getContracts() (contracts, error) {
	c := contracts{}

	fh, err := os.Open(l.contractFile)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}

		return nil, fmt.Errorf("opening contracts file: %w", err)
	}
	defer func() { _ = fh.Close() }()

	if err := json.NewDecoder(fh).Decode(&c); err != nil {
		return c, fmt.Errorf("reading contracts file: %w", err)
	}

	return c, nil
}

// saveContracts writes the contracts map to the contracts.json file
func (l *Local) saveContracts(c contracts) error {
	fh, err := os.OpenFile(l.contractFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("opening contracts file: %w", err)
	}
	if err := json.NewEncoder(fh).Encode(c); err != nil {
		return fmt.Errorf("writing contracts file: %w", err)
	}

	if err := fh.Close(); err != nil {
		return fmt.Errorf("closing contracts file: %w", err)
	}

	return nil
}

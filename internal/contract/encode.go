// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/pkg/abicodec"
	"github.com/MqllR/abitool/pkg/abiparser"
	abistore "github.com/MqllR/abitool/pkg/storage/abi"
	contractstore "github.com/MqllR/abitool/pkg/storage/contract"
)

// EncodeManager encodes function call arguments into ABI calldata using locally
// stored contract ABIs.
type EncodeManager struct {
	log           *log.Logger
	contractStore *contractstore.Local
	abiStore      *abistore.Local
}

// NewEncodeManager creates an EncodeManager using the configured abi-store and
// chain ID.
func NewEncodeManager(logger *log.Logger) (*EncodeManager, error) {
	storePath := viper.GetString("abi-store")
	chainID := viper.GetInt("chainid")
	basePath := filepath.Join(storePath, strconv.Itoa(chainID))

	cs, err := contractstore.NewLocal(basePath)
	if err != nil {
		return nil, err
	}

	as, err := abistore.NewLocal(basePath)
	if err != nil {
		return nil, err
	}

	return &EncodeManager{
		log:           logger,
		contractStore: cs,
		abiStore:      as,
	}, nil
}

// EncodeOptions controls output formatting.
type EncodeOptions struct {
	OutputJSON bool
}

// encodeResult is the JSON shape for --output json.
type encodeResult struct {
	Address   string `json:"address"`
	Function  string `json:"function"`
	Signature string `json:"signature,omitempty"`
	Selector  string `json:"selector,omitempty"`
	Calldata  string `json:"calldata"`
}

// EncodeCalldata ABI-encodes a function call for the given contract address
// and writes the calldata to out.
func (m *EncodeManager) EncodeCalldata(address, functionName string, args []string, opts EncodeOptions, out io.Writer) error {
	parsedABI, err := m.loadABI(address)
	if err != nil {
		return err
	}

	element, err := findByName(parsedABI, functionName)
	if err != nil {
		return err
	}

	method, err := abicodec.ParseMethod(*element)
	if err != nil {
		return fmt.Errorf("parsing method: %w", err)
	}

	calldata, err := abicodec.EncodeInput(method, args)
	if err != nil {
		return fmt.Errorf("encoding calldata: %w", err)
	}

	calldataHex := "0x" + hex.EncodeToString(calldata)

	if opts.OutputJSON {
		res := encodeResult{
			Address:  address,
			Function: element.Name,
			Calldata: calldataHex,
		}
		if sig, err := element.Signature(); err == nil {
			res.Signature = sig
		}
		if sel, err := element.Selector(); err == nil {
			res.Selector = sel
		}
		return json.NewEncoder(out).Encode(res)
	}

	fmt.Fprintln(out, calldataHex)
	return nil
}

// loadABI reads and parses the stored ABI for a contract address.
func (m *EncodeManager) loadABI(address string) (*abiparser.ABI, error) {
	chainID := viper.GetInt("chainid")

	_, err := m.contractStore.Get(address)
	if err != nil {
		return nil, fmt.Errorf(
			"ABI not found for %s — run: abitool abi download %s --chainid %d",
			address, address, chainID,
		)
	}

	rawABI, err := m.abiStore.Read(address)
	if err != nil {
		return nil, fmt.Errorf("reading ABI for %s: %w", address, err)
	}

	parsedABI, err := abiparser.ParseABI(rawABI)
	if err != nil {
		return nil, fmt.Errorf("parsing ABI: %w", err)
	}

	return parsedABI, nil
}

// findByName returns the first function element with the given name.
func findByName(parsedABI *abiparser.ABI, name string) (*abiparser.Element, error) {
	for el := range parsedABI.All() {
		if el.IsFunction() && el.Name == name {
			e := el
			return &e, nil
		}
	}
	return nil, fmt.Errorf("function %q not found in ABI", name)
}

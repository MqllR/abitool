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
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/pkg/abicodec"
	"github.com/MqllR/abitool/pkg/abiparser"
	abistore "github.com/MqllR/abitool/pkg/storage/abi"
	contractstore "github.com/MqllR/abitool/pkg/storage/contract"
)

// DecodeManager decodes ABI-encoded calldata and return data using locally
// stored contract ABIs.
type DecodeManager struct {
	log           *log.Logger
	contractStore *contractstore.Local
	abiStore      *abistore.Local
}

// NewDecodeManager creates a DecodeManager using the configured abi-store and
// chain ID.
func NewDecodeManager(logger *log.Logger) (*DecodeManager, error) {
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

	return &DecodeManager{
		log:           logger,
		contractStore: cs,
		abiStore:      as,
	}, nil
}

// DecodeOptions controls output formatting.
type DecodeOptions struct {
	OutputJSON bool
}

// DecodeCalldata decodes raw hex calldata for the given contract address,
// auto-detecting the called function from the 4-byte selector.
func (m *DecodeManager) DecodeCalldata(address, calldataHex string, opts DecodeOptions, out io.Writer) error {
	calldata, err := hexDecodeString(calldataHex)
	if err != nil {
		return fmt.Errorf("invalid calldata hex: %w", err)
	}

	if len(calldata) < 4 {
		return fmt.Errorf("calldata must be at least 4 bytes (got %d) — the first 4 bytes are the function selector", len(calldata))
	}

	selector := "0x" + hex.EncodeToString(calldata[:4])

	parsedABI, err := m.loadABI(address)
	if err != nil {
		return err
	}

	element, err := findBySelector(parsedABI, selector)
	if err != nil {
		return err
	}

	method, err := abicodec.ParseMethod(*element)
	if err != nil {
		return fmt.Errorf("parsing method: %w", err)
	}

	values, err := abicodec.DecodeInput(method, calldata)
	if err != nil {
		return fmt.Errorf("decoding calldata: %w", err)
	}

	fmt.Fprintf(out, "Function: %s\n", element.Name)

	sig, _ := element.Signature()
	if sig != "" {
		fmt.Fprintf(out, "Signature: %s\n", sig)
	}

	fmt.Fprintf(out, "Selector: %s\n\n", selector)

	return writeResult(out, toOutputs(element.Inputs), values, opts.OutputJSON)
}

// ethCallBody is the minimal JSON shape of an eth_call request object.
type ethCallBody struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

// DecodeFromEthCall parses an eth_call JSON request body and decodes its calldata.
// The body should be of the form {"to":"0x...","data":"0x..."}.
func (m *DecodeManager) DecodeFromEthCall(jsonBody string, opts DecodeOptions, out io.Writer) error {
	var body ethCallBody
	if err := json.Unmarshal([]byte(jsonBody), &body); err != nil {
		return fmt.Errorf("parsing eth_call JSON body: %w", err)
	}

	if body.To == "" {
		return fmt.Errorf("eth_call body missing 'to' field")
	}

	if body.Data == "" {
		return fmt.Errorf("eth_call body missing 'data' field")
	}

	return m.DecodeCalldata(body.To, body.Data, opts, out)
}

// DecodeFromRawTx decodes the calldata embedded in a RLP-encoded signed
// transaction (as produced by eth_sendRawTransaction).
func (m *DecodeManager) DecodeFromRawTx(rawTxHex string, opts DecodeOptions, out io.Writer) error {
	b, err := hexDecodeString(rawTxHex)
	if err != nil {
		return fmt.Errorf("invalid transaction hex: %w", err)
	}

	var tx types.Transaction
	if err := tx.UnmarshalBinary(b); err != nil {
		return fmt.Errorf("decoding raw transaction: %w", err)
	}

	to := tx.To()
	if to == nil {
		return fmt.Errorf("transaction is a contract creation (no 'to' address) — nothing to decode")
	}

	address := to.Hex()
	calldataHex := "0x" + hex.EncodeToString(tx.Data())

	fmt.Fprintf(out, "Transaction to: %s\n\n", address)

	return m.DecodeCalldata(address, calldataHex, opts, out)
}

// DecodeReturnData decodes the hex-encoded return data of a function call
// using the function's output types from the stored ABI.
func (m *DecodeManager) DecodeReturnData(address, functionName, returnHex string, opts DecodeOptions, out io.Writer) error {
	rawReturn, err := hexDecodeString(returnHex)
	if err != nil {
		return fmt.Errorf("invalid return data hex: %w", err)
	}

	parsedABI, err := m.loadABI(address)
	if err != nil {
		return err
	}

	var element *abiparser.Element
	for el := range parsedABI.All() {
		if el.IsFunction() && el.Name == functionName {
			e := el
			element = &e
			break
		}
	}

	if element == nil {
		return fmt.Errorf("function %q not found in ABI for contract %s", functionName, address)
	}

	method, err := abicodec.ParseMethod(*element)
	if err != nil {
		return fmt.Errorf("parsing method: %w", err)
	}

	values, err := abicodec.DecodeOutput(method, rawReturn)
	if err != nil {
		return fmt.Errorf("decoding return data: %w", err)
	}

	fmt.Fprintf(out, "Function: %s (return data)\n\n", functionName)

	return writeResult(out, element.Outputs, values, opts.OutputJSON)
}

// loadABI reads and parses the stored ABI for a contract address. Returns a
// helpful error message if the ABI is not found in the local store.
func (m *DecodeManager) loadABI(address string) (*abiparser.ABI, error) {
	cfg := abitool.ConfigInstance()
	chainID := viper.GetInt("chainid")

	_, err := m.contractStore.Get(address)
	if err != nil {
		return nil, fmt.Errorf(
			"ABI not found for %s — run: abitool abi download %s --chainid %d --etherscan-api-key %s",
			address, address, chainID, cfg.EtherScan.APIKey,
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

// findBySelector iterates all ABI elements with selectors and returns the one
// whose computed 4-byte selector matches the given hex selector string.
func findBySelector(parsedABI *abiparser.ABI, selector string) (*abiparser.Element, error) {
	for el := range parsedABI.All() {
		if !el.HasSelector() {
			continue
		}

		computed, err := el.Selector()
		if err != nil {
			continue
		}

		if strings.EqualFold(computed, selector) {
			e := el
			return &e, nil
		}
	}

	return nil, fmt.Errorf("no function found with selector %s — the ABI may be incomplete or this is a proxy contract", selector)
}

// toOutputs converts a slice of Input parameters to Output parameters so that
// decoded input values can be printed with the same writeResult helper.
func toOutputs(inputs []abiparser.Input) []abiparser.Output {
	out := make([]abiparser.Output, len(inputs))
	for i, inp := range inputs {
		out[i] = abiparser.Output{Parameter: inp.Parameter}
	}

	return out
}

// hexDecodeString decodes an optional 0x-prefixed hex string to bytes.
func hexDecodeString(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s)%2 != 0 {
		s = "0" + s
	}

	return hex.DecodeString(s)
}

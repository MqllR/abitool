// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

// Package abicodec bridges abitool's ABI parser types with the go-ethereum ABI
// codec (github.com/ethereum/go-ethereum/accounts/abi) to encode function call
// inputs and decode return values.
// go-ethereum is Copyright 2014 The go-ethereum Authors, licensed under LGPL-3.0.
package abicodec

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/MqllR/abitool/pkg/abiparser"
)

// ParseMethod converts an abiparser.Element into a go-ethereum abi.Method by
// re-serialising the element as a single-entry ABI JSON and parsing it.
func ParseMethod(element abiparser.Element) (abi.Method, error) {
	fragment, err := json.Marshal(element)
	if err != nil {
		return abi.Method{}, fmt.Errorf("marshaling ABI element: %w", err)
	}

	parsed, err := abi.JSON(strings.NewReader("[" + string(fragment) + "]"))
	if err != nil {
		return abi.Method{}, fmt.Errorf("parsing ABI fragment: %w", err)
	}

	method, ok := parsed.Methods[element.Name]
	if !ok {
		return abi.Method{}, fmt.Errorf("method %q not found after parsing", element.Name)
	}

	return method, nil
}

// EncodeInput ABI-encodes the function selector + packed arguments for a method.
// args are raw string values in the same order as method.Inputs.
func EncodeInput(method abi.Method, args []string) ([]byte, error) {
	if len(args) != len(method.Inputs) {
		return nil, fmt.Errorf("expected %d argument(s), got %d", len(method.Inputs), len(args))
	}

	values := make([]interface{}, len(args))
	for i, arg := range args {
		v, err := convertArg(method.Inputs[i].Type, arg)
		if err != nil {
			return nil, fmt.Errorf("argument %d (%s %s): %w", i, method.Inputs[i].Name, method.Inputs[i].Type, err)
		}
		values[i] = v
	}

	packed, err := method.Inputs.Pack(values...)
	if err != nil {
		return nil, fmt.Errorf("packing arguments: %w", err)
	}

	// Prepend 4-byte selector.
	calldata := make([]byte, 4+len(packed))
	copy(calldata[:4], method.ID)
	copy(calldata[4:], packed)

	return calldata, nil
}

// DecodeOutput unpacks the raw return bytes into a slice of Go values using the
// method's output types.
func DecodeOutput(method abi.Method, data []byte) ([]interface{}, error) {
	if len(method.Outputs) == 0 {
		return nil, nil
	}

	values, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("unpacking output: %w", err)
	}

	return values, nil
}

// convertArg converts a string CLI argument to the appropriate Go type for the
// given ABI type. Supports the most common EVM types.
func convertArg(abiType abi.Type, value string) (interface{}, error) {
	switch abiType.T {
	case abi.AddressTy:
		if !common.IsHexAddress(value) {
			return nil, fmt.Errorf("invalid Ethereum address: %q", value)
		}
		return common.HexToAddress(value), nil

	case abi.BoolTy:
		switch strings.ToLower(value) {
		case "true", "1":
			return true, nil
		case "false", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid bool: %q (expected true/false)", value)
		}

	case abi.UintTy, abi.IntTy:
		n := new(big.Int)
		if _, ok := n.SetString(value, 0); !ok {
			return nil, fmt.Errorf("invalid integer: %q", value)
		}
		return bigIntToType(abiType, n)

	case abi.StringTy:
		return value, nil

	case abi.BytesTy:
		b, err := hexDecode(value)
		if err != nil {
			return nil, fmt.Errorf("invalid bytes hex: %w", err)
		}
		return b, nil

	case abi.FixedBytesTy:
		b, err := hexDecode(value)
		if err != nil {
			return nil, fmt.Errorf("invalid bytes hex: %w", err)
		}
		if len(b) != abiType.Size {
			return nil, fmt.Errorf("expected %d bytes, got %d", abiType.Size, len(b))
		}
		return toFixedBytes(b, abiType.Size)

	default:
		return nil, fmt.Errorf("unsupported ABI type %q (complex types require --interactive mode)", abiType.String())
	}
}

// bigIntToType narrows a *big.Int to the concrete int/uint type that
// go-ethereum's abi.Pack expects for a given bit size.
func bigIntToType(t abi.Type, n *big.Int) (interface{}, error) {
	if t.T == abi.UintTy {
		switch t.Size {
		case 8:
			return uint8(n.Uint64()), nil
		case 16:
			return uint16(n.Uint64()), nil
		case 32:
			return uint32(n.Uint64()), nil
		case 64:
			return uint64(n.Uint64()), nil
		default:
			return n, nil // *big.Int for uint128, uint256, …
		}
	}
	// IntTy
	switch t.Size {
	case 8:
		return int8(n.Int64()), nil
	case 16:
		return int16(n.Int64()), nil
	case 32:
		return int32(n.Int64()), nil
	case 64:
		return int64(n.Int64()), nil
	default:
		return n, nil // *big.Int for int128, int256, …
	}
}

// hexDecode decodes an optional 0x-prefixed hex string.
func hexDecode(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s)%2 != 0 {
		s = "0" + s
	}

	b := make([]byte, len(s)/2)
	for i := range b {
		var v byte
		if _, err := fmt.Sscanf(s[2*i:2*i+2], "%02x", &v); err != nil {
			return nil, err
		}
		b[i] = v
	}

	return b, nil
}

// toFixedBytes converts a byte slice into the concrete [N]byte array type that
// go-ethereum expects. Supports all valid Solidity fixed-byte sizes (1–32).
func toFixedBytes(b []byte, size int) (interface{}, error) {
	switch size {
	case 1:
		var a [1]byte; copy(a[:], b); return a, nil
	case 2:
		var a [2]byte; copy(a[:], b); return a, nil
	case 3:
		var a [3]byte; copy(a[:], b); return a, nil
	case 4:
		var a [4]byte; copy(a[:], b); return a, nil
	case 5:
		var a [5]byte; copy(a[:], b); return a, nil
	case 6:
		var a [6]byte; copy(a[:], b); return a, nil
	case 7:
		var a [7]byte; copy(a[:], b); return a, nil
	case 8:
		var a [8]byte; copy(a[:], b); return a, nil
	case 9:
		var a [9]byte; copy(a[:], b); return a, nil
	case 10:
		var a [10]byte; copy(a[:], b); return a, nil
	case 11:
		var a [11]byte; copy(a[:], b); return a, nil
	case 12:
		var a [12]byte; copy(a[:], b); return a, nil
	case 13:
		var a [13]byte; copy(a[:], b); return a, nil
	case 14:
		var a [14]byte; copy(a[:], b); return a, nil
	case 15:
		var a [15]byte; copy(a[:], b); return a, nil
	case 16:
		var a [16]byte; copy(a[:], b); return a, nil
	case 17:
		var a [17]byte; copy(a[:], b); return a, nil
	case 18:
		var a [18]byte; copy(a[:], b); return a, nil
	case 19:
		var a [19]byte; copy(a[:], b); return a, nil
	case 20:
		var a [20]byte; copy(a[:], b); return a, nil
	case 21:
		var a [21]byte; copy(a[:], b); return a, nil
	case 22:
		var a [22]byte; copy(a[:], b); return a, nil
	case 23:
		var a [23]byte; copy(a[:], b); return a, nil
	case 24:
		var a [24]byte; copy(a[:], b); return a, nil
	case 25:
		var a [25]byte; copy(a[:], b); return a, nil
	case 26:
		var a [26]byte; copy(a[:], b); return a, nil
	case 27:
		var a [27]byte; copy(a[:], b); return a, nil
	case 28:
		var a [28]byte; copy(a[:], b); return a, nil
	case 29:
		var a [29]byte; copy(a[:], b); return a, nil
	case 30:
		var a [30]byte; copy(a[:], b); return a, nil
	case 31:
		var a [31]byte; copy(a[:], b); return a, nil
	case 32:
		var a [32]byte; copy(a[:], b); return a, nil
	default:
		return nil, fmt.Errorf("unsupported fixed byte size %d", size)
	}
}

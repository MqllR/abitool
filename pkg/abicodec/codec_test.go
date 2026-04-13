// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abicodec

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/MqllR/abitool/pkg/abiparser"
)

// ---- helpers ----

func transferElement() abiparser.Element {
	return abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "transfer",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "to", Type: "address"}},
			{Parameter: abiparser.Parameter{Name: "amount", Type: "uint256"}},
		},
		Outputs: []abiparser.Output{
			{Parameter: abiparser.Parameter{Name: "", Type: "bool"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
}

func totalSupplyElement() abiparser.Element {
	return abiparser.Element{
		Type:            abiparser.FunctionType,
		Name:            "totalSupply",
		Inputs:          []abiparser.Input{},
		Outputs:         []abiparser.Output{{Parameter: abiparser.Parameter{Name: "", Type: "uint256"}}},
		StateMutability: abiparser.ViewStateMutability,
	}
}

// makeFixedBytesElement builds a single-input function element for a bytesN type.
func makeFixedBytesElement(size int) abiparser.Element {
	return abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "testFunc",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "b", Type: fmt.Sprintf("bytes%d", size)}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
}

// hexOfSize creates a 0x-prefixed hex string of n bytes (values 0x01,0x02,...).
func hexOfSize(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i % 255) + 1)
	}
	return "0x" + hex.EncodeToString(b)
}

// ---- ParseMethod ----

func TestParseMethod_Transfer(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}
	if method.Name != "transfer" {
		t.Errorf("method name: got %q, want %q", method.Name, "transfer")
	}
	if len(method.Inputs) != 2 {
		t.Errorf("inputs count: got %d, want 2", len(method.Inputs))
	}
}

func TestParseMethod_TotalSupply(t *testing.T) {
	method, err := ParseMethod(totalSupplyElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}
	if method.Name != "totalSupply" {
		t.Errorf("method name: got %q, want %q", method.Name, "totalSupply")
	}
	if len(method.Inputs) != 0 {
		t.Errorf("inputs count: got %d, want 0", len(method.Inputs))
	}
}

// ---- EncodeInput ----

func TestEncodeInput_Transfer(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	data, err := EncodeInput(method, []string{
		"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"1000",
	})
	if err != nil {
		t.Fatalf("EncodeInput: %v", err)
	}
	if len(data) < 4 {
		t.Fatalf("encoded data too short: %d bytes", len(data))
	}

	// First 4 bytes = selector for transfer(address,uint256) = 0xa9059cbb
	want := []byte{0xa9, 0x05, 0x9c, 0xbb}
	if string(data[:4]) != string(want) {
		t.Errorf("selector: got %x, want %x", data[:4], want)
	}
}

func TestEncodeInput_WrongArgCount(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{"0x742d35Cc6634C0532925a3b844Bc454e4438f44e"})
	if err == nil {
		t.Fatal("expected error for wrong arg count, got nil")
	}
}

func TestEncodeInput_InvalidAddress(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{"notanaddress", "1000"})
	if err == nil {
		t.Fatal("expected error for invalid address, got nil")
	}
}

func TestEncodeInput_InvalidUint256(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{
		"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"not-a-number",
	})
	if err == nil {
		t.Fatal("expected error for invalid uint256, got nil")
	}
}

func TestEncodeInput_Uint256Hex(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	// 0x3E8 = 1000 decimal — should succeed
	data, err := EncodeInput(method, []string{
		"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"0x3E8",
	})
	if err != nil {
		t.Fatalf("EncodeInput with hex uint256: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty encoded data")
	}
}

// ---- DecodeOutput ----

func TestDecodeOutput_Uint256(t *testing.T) {
	method, err := ParseMethod(totalSupplyElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	// ABI-encode 1000 as uint256 (32-byte big-endian)
	want := big.NewInt(1000)
	encoded := make([]byte, 32)
	want.FillBytes(encoded)

	values, err := DecodeOutput(method, encoded)
	if err != nil {
		t.Fatalf("DecodeOutput: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 output value, got %d", len(values))
	}

	got, ok := values[0].(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int, got %T", values[0])
	}
	if got.Cmp(want) != 0 {
		t.Errorf("decoded value: got %v, want %v", got, want)
	}
}

func TestDecodeOutput_NoOutputs(t *testing.T) {
	el := abiparser.Element{
		Type:            abiparser.FunctionType,
		Name:            "noReturn",
		Inputs:          []abiparser.Input{},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	values, err := DecodeOutput(method, nil)
	if err != nil {
		t.Fatalf("DecodeOutput with no outputs: %v", err)
	}
	if values != nil {
		t.Errorf("expected nil values for no-output method, got %v", values)
	}
}

// ---- convertArg (tested indirectly) ----

func TestConvertArg_Bool(t *testing.T) {
	el := abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "setBool",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "v", Type: "bool"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	for _, valid := range []string{"true", "false", "1", "0"} {
		_, err := EncodeInput(method, []string{valid})
		if err != nil {
			t.Errorf("EncodeInput with bool %q: unexpected error: %v", valid, err)
		}
	}

	_, err = EncodeInput(method, []string{"yes"})
	if err == nil {
		t.Error("EncodeInput with invalid bool 'yes': expected error, got nil")
	}
}

func TestConvertArg_String(t *testing.T) {
	el := abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "setName",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "name", Type: "string"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{"hello world"})
	if err != nil {
		t.Fatalf("EncodeInput with string: %v", err)
	}
}

func TestConvertArg_Bytes(t *testing.T) {
	el := abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "setData",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "data", Type: "bytes"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{"0xdeadbeef"})
	if err != nil {
		t.Fatalf("EncodeInput with valid bytes hex: %v", err)
	}

	_, err = EncodeInput(method, []string{"0xgg"})
	if err == nil {
		t.Error("EncodeInput with invalid bytes hex: expected error, got nil")
	}
}

func TestConvertArg_SliceType(t *testing.T) {
	el := abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "setArr",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "arr", Type: "uint256[]"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	// Valid JSON array of uint256 values.
	_, err = EncodeInput(method, []string{"[1,2,3]"})
	if err != nil {
		t.Errorf("EncodeInput with uint256[] [1,2,3]: unexpected error: %v", err)
	}

	// Empty array is valid.
	_, err = EncodeInput(method, []string{"[]"})
	if err != nil {
		t.Errorf("EncodeInput with uint256[] []: unexpected error: %v", err)
	}

	// Invalid JSON should error.
	_, err = EncodeInput(method, []string{"not-json"})
	if err == nil {
		t.Error("EncodeInput with malformed JSON: expected error, got nil")
	}
}

func TestConvertArg_AddressSlice(t *testing.T) {
	el := abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "setAddrs",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "addrs", Type: "address[]"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{`["0x742d35Cc6634C0532925a3b844Bc454e4438f44e","0xdAC17F958D2ee523a2206206994597C13D831ec7"]`})
	if err != nil {
		t.Errorf("EncodeInput with address[]: unexpected error: %v", err)
	}
}

func TestConvertArg_FixedArray(t *testing.T) {
	el := abiparser.Element{
		Type: abiparser.FunctionType,
		Name: "setFixed",
		Inputs: []abiparser.Input{
			{Parameter: abiparser.Parameter{Name: "arr", Type: "uint256[3]"}},
		},
		StateMutability: abiparser.NonpayableStateMutability,
	}
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = EncodeInput(method, []string{"[10,20,30]"})
	if err != nil {
		t.Errorf("EncodeInput with uint256[3]: unexpected error: %v", err)
	}

	// Wrong element count should error.
	_, err = EncodeInput(method, []string{"[10,20]"})
	if err == nil {
		t.Error("EncodeInput with wrong fixed array size: expected error, got nil")
	}
}

// ---- toFixedBytes (tested indirectly via EncodeInput) ----

func TestFixedBytes_AllSizes(t *testing.T) {
	// Test a representative set of sizes including previously-missing ones.
	sizes := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("bytes%d", size), func(t *testing.T) {
			el := makeFixedBytesElement(size)
			method, err := ParseMethod(el)
			if err != nil {
				t.Fatalf("ParseMethod for bytes%d: %v", size, err)
			}

			hexVal := hexOfSize(size)
			data, err := EncodeInput(method, []string{hexVal})
			if err != nil {
				t.Fatalf("EncodeInput for bytes%d with %q: %v", size, hexVal, err)
			}
			// 4 bytes selector + 32 bytes slot
			if len(data) != 36 {
				t.Errorf("bytes%d: expected 36 encoded bytes, got %d", size, len(data))
			}
		})
	}
}

func TestFixedBytes_WrongSize(t *testing.T) {
	// Providing the wrong byte count should error.
	el := makeFixedBytesElement(4)
	method, err := ParseMethod(el)
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	// Provide 3 bytes instead of 4
	_, err = EncodeInput(method, []string{"0xaabbcc"})
	if err == nil {
		t.Error("EncodeInput with wrong fixed bytes size: expected error, got nil")
	}
}

// ---- hexDecode ----

func TestHexDecode_OptionalPrefix(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"0xdeadbeef", "deadbeef"},
		{"deadbeef", "deadbeef"},
		{"0xDEADBEEF", "deadbeef"},
	}
	for _, tc := range cases {
		b, err := hexDecode(tc.input)
		if err != nil {
			t.Errorf("hexDecode(%q): unexpected error: %v", tc.input, err)
			continue
		}
		got := strings.ToLower(hex.EncodeToString(b))
		if got != tc.want {
			t.Errorf("hexDecode(%q): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ---- DecodeInput ----

func TestDecodeInput_Transfer(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	// Encode first so we have valid calldata to round-trip.
	calldata, err := EncodeInput(method, []string{
		"0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"1000",
	})
	if err != nil {
		t.Fatalf("EncodeInput: %v", err)
	}

	values, err := DecodeInput(method, calldata)
	if err != nil {
		t.Fatalf("DecodeInput: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 decoded values, got %d", len(values))
	}
}

func TestDecodeInput_TooShort(t *testing.T) {
	method, err := ParseMethod(transferElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	_, err = DecodeInput(method, []byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for calldata shorter than 4 bytes, got nil")
	}
}

func TestDecodeInput_NoInputs(t *testing.T) {
	method, err := ParseMethod(totalSupplyElement())
	if err != nil {
		t.Fatalf("ParseMethod: %v", err)
	}

	// Calldata with only selector — no arguments.
	calldata, err := EncodeInput(method, []string{})
	if err != nil {
		t.Fatalf("EncodeInput: %v", err)
	}

	values, err := DecodeInput(method, calldata)
	if err != nil {
		t.Fatalf("DecodeInput: %v", err)
	}
	if values != nil {
		t.Errorf("expected nil for no-input method, got %v", values)
	}
}

// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abiparser_test

import (
	"strings"
	"testing"

	"github.com/MqllR/abitool/pkg/abiparser"
)

const sampleABI = `[
  {"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"},
  {"type":"function","name":"totalSupply","inputs":[],"outputs":[{"name":"","type":"uint256"}],"stateMutability":"view"},
  {"type":"event","name":"Transfer","inputs":[{"name":"from","type":"address","indexed":true},{"name":"to","type":"address","indexed":true},{"name":"value","type":"uint256","indexed":false}],"anonymous":false},
  {"type":"event","name":"AnonEvent","inputs":[],"anonymous":true},
  {"type":"constructor","inputs":[{"name":"name","type":"string"},{"name":"symbol","type":"string"}],"stateMutability":"nonpayable"},
  {"type":"error","name":"InsufficientBalance","inputs":[{"name":"balance","type":"uint256"},{"name":"needed","type":"uint256"}]}
]`

// findElement is a test helper that finds an element by type and name.
func findElement(t *testing.T, abi *abiparser.ABI, typ abiparser.Type, name string) abiparser.Element {
	t.Helper()
	for el := range abi.All() {
		if el.Type == typ && el.Name == name {
			return el
		}
	}
	t.Fatalf("element type=%q name=%q not found in ABI", typ, name)
	return abiparser.Element{}
}

// findElementByType finds the first element of the given type.
func findElementByType(t *testing.T, abi *abiparser.ABI, typ abiparser.Type) abiparser.Element {
	t.Helper()
	for el := range abi.All() {
		if el.Type == typ {
			return el
		}
	}
	t.Fatalf("element type=%q not found in ABI", typ)
	return abiparser.Element{}
}

// ---- ParseABI ----

func TestParseABI_Valid(t *testing.T) {
	abi, err := abiparser.ParseABI(sampleABI)
	if err != nil {
		t.Fatalf("ParseABI returned unexpected error: %v", err)
	}

	count := 0
	for range abi.All() {
		count++
	}
	if count != 6 {
		t.Errorf("expected 6 elements, got %d", count)
	}
}

func TestParseABI_EmptyArray(t *testing.T) {
	abi, err := abiparser.ParseABI(`[]`)
	if err != nil {
		t.Fatalf("ParseABI([]) returned unexpected error: %v", err)
	}

	count := 0
	for range abi.All() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 elements, got %d", count)
	}
}

func TestParseABI_InvalidJSON(t *testing.T) {
	cases := []string{
		``,
		`not json`,
		`{`,
		`{"type":"function"}`, // object, not array
	}
	for _, input := range cases {
		_, err := abiparser.ParseABI(input)
		if err == nil {
			t.Errorf("ParseABI(%q) expected error, got nil", input)
		}
	}
}

func TestParseABI_ElementTypes(t *testing.T) {
	abi, err := abiparser.ParseABI(sampleABI)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}

	want := map[abiparser.Type]bool{
		abiparser.FunctionType:    false,
		abiparser.EventType:       false,
		abiparser.ConstructorType: false,
		abiparser.ErrorType:       false,
	}
	for el := range abi.All() {
		if _, ok := want[el.Type]; ok {
			want[el.Type] = true
		}
	}
	for typ, seen := range want {
		if !seen {
			t.Errorf("element type %q not found in parsed ABI", typ)
		}
	}
}

// ---- Element.Signature ----

func TestSignature_SimpleFunction(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElement(t, abi, abiparser.FunctionType, "transfer")

	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	if sig != "transfer(address,uint256)" {
		t.Errorf("got %q, want %q", sig, "transfer(address,uint256)")
	}
}

func TestSignature_NoInputFunction(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElement(t, abi, abiparser.FunctionType, "totalSupply")

	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	if sig != "totalSupply()" {
		t.Errorf("got %q, want %q", sig, "totalSupply()")
	}
}

func TestSignature_Event(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElement(t, abi, abiparser.EventType, "Transfer")

	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	if sig != "Transfer(address,address,uint256)" {
		t.Errorf("got %q, want %q", sig, "Transfer(address,address,uint256)")
	}
}

func TestSignature_ConstructorErrors(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElementByType(t, abi, abiparser.ConstructorType)

	_, err := el.Signature()
	if err == nil {
		t.Fatal("Signature() on constructor expected error, got nil")
	}
}

func TestSignature_FallbackReceiveError(t *testing.T) {
	for _, raw := range []string{
		`[{"type":"fallback","stateMutability":"nonpayable"}]`,
		`[{"type":"receive","stateMutability":"payable"}]`,
	} {
		abi, err := abiparser.ParseABI(raw)
		if err != nil {
			t.Fatalf("ParseABI: %v", err)
		}
		for el := range abi.All() {
			_, sigErr := el.Signature()
			if sigErr == nil {
				t.Errorf("Signature() on %q expected error, got nil", el.Type)
			}
		}
	}
}

func TestSignature_TupleInput(t *testing.T) {
	raw := `[{"type":"function","name":"foo","inputs":[{"name":"p","type":"tuple","components":[{"name":"x","type":"uint256"},{"name":"y","type":"address"}]}],"stateMutability":"nonpayable"}]`
	abi, err := abiparser.ParseABI(raw)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}
	el := findElement(t, abi, abiparser.FunctionType, "foo")
	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	// tuple should expand to (uint256,address) not "tuple"
	if strings.Contains(sig, "tuple") {
		t.Errorf("Signature() contains literal 'tuple': %q", sig)
	}
	if sig != "foo((uint256,address))" {
		t.Errorf("got %q, want %q", sig, "foo((uint256,address))")
	}
}

func TestSignature_TupleArray(t *testing.T) {
	raw := `[{"type":"function","name":"bar","inputs":[{"name":"p","type":"tuple[]","components":[{"name":"x","type":"uint256"},{"name":"y","type":"address"}]}],"stateMutability":"nonpayable"}]`
	abi, err := abiparser.ParseABI(raw)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}
	el := findElement(t, abi, abiparser.FunctionType, "bar")
	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	if sig != "bar((uint256,address)[])" {
		t.Errorf("got %q, want %q", sig, "bar((uint256,address)[])")
	}
}

func TestSignature_TupleFixedArray(t *testing.T) {
	raw := `[{"type":"function","name":"baz","inputs":[{"name":"p","type":"tuple[3]","components":[{"name":"x","type":"uint256"}]}],"stateMutability":"nonpayable"}]`
	abi, err := abiparser.ParseABI(raw)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}
	el := findElement(t, abi, abiparser.FunctionType, "baz")
	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	if sig != "baz((uint256)[3])" {
		t.Errorf("got %q, want %q", sig, "baz((uint256)[3])")
	}
}

func TestSignature_NestedTuple(t *testing.T) {
	raw := `[{"type":"function","name":"nested","inputs":[{"name":"p","type":"tuple","components":[{"name":"inner","type":"tuple","components":[{"name":"x","type":"uint256"},{"name":"y","type":"address"}]},{"name":"z","type":"bytes"}]}],"stateMutability":"nonpayable"}]`
	abi, err := abiparser.ParseABI(raw)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}
	el := findElement(t, abi, abiparser.FunctionType, "nested")
	sig, err := el.Signature()
	if err != nil {
		t.Fatalf("Signature(): %v", err)
	}
	if sig != "nested(((uint256,address),bytes))" {
		t.Errorf("got %q, want %q", sig, "nested(((uint256,address),bytes))")
	}
}

// ---- Element.Selector ----

func TestSelector_KnownValues(t *testing.T) {
	cases := []struct {
		abiJSON string
		name    string
		want    string
	}{
		{
			abiJSON: `[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"stateMutability":"nonpayable"}]`,
			name:    "transfer",
			want:    "0xa9059cbb",
		},
		{
			abiJSON: `[{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address"}],"stateMutability":"view"}]`,
			name:    "balanceOf",
			want:    "0x70a08231",
		},
		{
			abiJSON: `[{"type":"function","name":"totalSupply","inputs":[],"stateMutability":"view"}]`,
			name:    "totalSupply",
			want:    "0x18160ddd",
		},
		{
			abiJSON: `[{"type":"function","name":"approve","inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"stateMutability":"nonpayable"}]`,
			name:    "approve",
			want:    "0x095ea7b3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			abi, err := abiparser.ParseABI(tc.abiJSON)
			if err != nil {
				t.Fatalf("ParseABI: %v", err)
			}
			el := findElement(t, abi, abiparser.FunctionType, tc.name)
			got, err := el.Selector()
			if err != nil {
				t.Fatalf("Selector(): %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSelector_NonFunctionErrors(t *testing.T) {
	// Only types with no selector AND no topic hash should error:
	// constructors, fallbacks, receives, and anonymous events.
	// Non-anonymous events have a topic hash, so Signature() succeeds,
	// and Selector() returns a (truncated) hash without error.
	for _, raw := range []string{
		`[{"type":"constructor","inputs":[{"name":"x","type":"uint256"}],"stateMutability":"nonpayable"}]`,
		`[{"type":"fallback","stateMutability":"nonpayable"}]`,
		`[{"type":"receive","stateMutability":"payable"}]`,
		`[{"type":"event","name":"AnonEv","inputs":[],"anonymous":true}]`,
	} {
		abi, err := abiparser.ParseABI(raw)
		if err != nil {
			t.Fatalf("ParseABI(%q): %v", raw, err)
		}
		for el := range abi.All() {
			_, sErr := el.Selector()
			if sErr == nil {
				t.Errorf("Selector() on type=%q name=%q expected error, got nil", el.Type, el.Name)
			}
		}
	}
}

func TestSelector_ErrorType(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElement(t, abi, abiparser.ErrorType, "InsufficientBalance")

	_, err := el.Selector()
	if err != nil {
		t.Fatalf("Selector() on error type should succeed: %v", err)
	}
}

// ---- Element.TopicHash ----

func TestTopicHash_Transfer(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElement(t, abi, abiparser.EventType, "Transfer")

	got, err := el.TopicHash()
	if err != nil {
		t.Fatalf("TopicHash(): %v", err)
	}
	const want = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTopicHash_AnonymousEventErrors(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)
	el := findElement(t, abi, abiparser.EventType, "AnonEvent")

	_, err := el.TopicHash()
	if err == nil {
		t.Fatal("TopicHash() on anonymous event expected error, got nil")
	}
}

func TestTopicHash_NonEventErrors(t *testing.T) {
	abi, _ := abiparser.ParseABI(sampleABI)

	for el := range abi.All() {
		if el.Type == abiparser.EventType {
			continue
		}
		_, err := el.TopicHash()
		if err == nil {
			t.Errorf("TopicHash() on type=%q name=%q expected error, got nil", el.Type, el.Name)
		}
	}
}

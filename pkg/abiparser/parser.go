package abiparser

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	crypto "golang.org/x/crypto/sha3"
)

type ABI []Element

type Element struct {
	Inputs          []Input         `json:"inputs"`
	Outputs         []Output        `json:"outputs,omitempty"`
	StateMutability StateMutability `json:"stateMutability"`
	Type            Type            `json:"type"`
	Name            string          `json:"name"`
}

type Input struct {
	Parameter
}

type Output struct {
	Parameter
}

type Parameter struct {
	Name         string      `json:"name"`
	InternalType string      `json:"internalType"`
	Type         string      `json:"type"`
	Components   []Parameter `json:"components,omitempty"`
}

func ParseABI(abiJSON string) (*ABI, error) {
	var abi ABI

	if err := json.Unmarshal([]byte(abiJSON), &abi); err != nil {
		return nil, fmt.Errorf("failed to parse ABI JSON: %w", err)
	}

	return &abi, nil
}

// All return an iterator over all ABI elements.
func (a *ABI) All() iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for _, e := range *a {
			if !yield(e) {
				return
			}
		}
	}
}

func (e *Element) IsFunction() bool {
	return e.Type == FunctionType
}

// canonicalType returns the canonical ABI type string for a parameter, recursively expanding
// tuple types into their component form: e.g. "tuple" → "(uint256,address)".
// Array suffixes on tuples (e.g. "tuple[]", "tuple[3]") are preserved.
func canonicalType(p Parameter) string {
	if len(p.Components) == 0 {
		return p.Type
	}

	var parts []string
	for _, comp := range p.Components {
		parts = append(parts, canonicalType(comp))
	}

	// Strip the "tuple" prefix but keep any array suffix (e.g. "[]", "[3]").
	suffix := strings.TrimPrefix(p.Type, "tuple")
	return "(" + strings.Join(parts, ",") + ")" + suffix
}

// Signature computes the function signature which is defined as "functionName(type1,type2,...)"
func (e *Element) Signature() (string, error) {
	if !e.IsFunction() {
		return "", fmt.Errorf("element is not a function")
	}

	var inputTypes []string
	for _, input := range e.Inputs {
		inputTypes = append(inputTypes, canonicalType(input.Parameter))
	}

	return e.Name + "(" + strings.Join(inputTypes, ",") + ")", nil
}

// Selector computes the function selector which is the first 4 bytes of the Keccak-256 hash of the function signature.
// The function signature is defined as "functionName(type1,type2,...)"
func (e *Element) Selector() (string, error) {
	sign, err := e.Signature()
	if err != nil {
		return "", err
	}

	hash := crypto.NewLegacyKeccak256()
	_, err = hash.Write([]byte(sign))
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(hash.Sum(nil)[:4]), nil
}

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
	StateMutability StateMutability `json:"stateMutability"`
	Type            Type            `json:"type"`
	Name            string          `json:"name"`
}

type Input struct {
	Parameter
	Components []Parameter `json:"components,omitempty"`
}

type Parameter struct {
	Name         string `json:"name"`
	InternalType string `json:"internalType"`
	Type         string `json:"type"`
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

// Signature computes the function signature which is defined as "functionName(type1,type2,...)"
func (e *Element) Signature() (string, error) {
	var signature string
	if !e.IsFunction() {
		return "", fmt.Errorf("element is not a function")
	}

	signature += e.Name + "("
	if len(e.Inputs) > 0 {
		var inputTypes []string
		for _, input := range e.Inputs {
			inputTypes = append(inputTypes, input.Type)
		}
		signature += strings.Join(inputTypes, ",")
	}

	signature += ")"
	return signature, nil
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

package abiparser

import (
	"encoding/json"
	"fmt"
	"iter"
)

type ABI []Element

type Element struct {
	Inputs          []any           `json:"inputs"`
	StateMutability StateMutability `json:"stateMutability"`
	Type            Type            `json:"type"`
	Name            string          `json:"name"`
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

func (e *Element) IsPayable() bool {
	return e.StateMutability == PayableStateMutability
}

func (e *Element) IsFunction() bool {
	return e.Type == FunctionType
}

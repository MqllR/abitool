// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

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
	Anonymous       bool            `json:"anonymous,omitempty"`
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
	Indexed      bool        `json:"indexed,omitempty"`
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

func (e *Element) IsError() bool {
	return e.Type == ErrorType
}

func (e *Element) IsEvent() bool {
	return e.Type == EventType
}

// HasSelector reports whether this element type has a 4-byte selector.
// Both functions and errors use Keccak-256(signature)[0:4].
func (e *Element) HasSelector() bool {
	return e.IsFunction() || e.IsError()
}

// HasTopicHash reports whether this element has an event topic hash (topic[0]).
// Non-anonymous events use the full Keccak-256(signature) as topic[0].
// Anonymous events do not emit a topic[0].
func (e *Element) HasTopicHash() bool {
	return e.IsEvent() && !e.Anonymous
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

// Signature computes the canonical ABI signature for functions, errors and events:
// "name(type1,type2,...)". Used to derive the 4-byte selector or the 32-byte topic hash.
func (e *Element) Signature() (string, error) {
	if !e.HasSelector() && !e.HasTopicHash() {
		return "", fmt.Errorf("element type %q does not have a selector or topic hash", e.Type)
	}

	var inputTypes []string
	for _, input := range e.Inputs {
		inputTypes = append(inputTypes, canonicalType(input.Parameter))
	}

	return e.Name + "(" + strings.Join(inputTypes, ",") + ")", nil
}

// Selector computes the 4-byte selector: Keccak-256(signature)[0:4].
// Applies to functions and errors.
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

// TopicHash computes the full 32-byte event topic hash: Keccak-256(signature).
// This is the value used as topic[0] in eth_getLogs filters.
// Applies to non-anonymous events only.
func (e *Element) TopicHash() (string, error) {
	if !e.HasTopicHash() {
		if e.IsEvent() {
			return "", fmt.Errorf("anonymous events do not have a topic[0] hash")
		}
		return "", fmt.Errorf("element type %q does not have a topic hash", e.Type)
	}

	sign, err := e.Signature()
	if err != nil {
		return "", err
	}

	hash := crypto.NewLegacyKeccak256()
	_, err = hash.Write([]byte(sign))
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(hash.Sum(nil)), nil
}

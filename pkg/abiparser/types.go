// Package abiparser provides types and utilities for parsing and inspecting
// Ethereum contract ABIs, following the ABI JSON specification:
// https://docs.soliditylang.org/en/latest/abi-spec.html#json
package abiparser

type Type string

var (
	FunctionType    Type = "function"
	EventType       Type = "event"
	ErrorType       Type = "error"
	ConstructorType Type = "constructor"
	ReceiveType     Type = "receive"
	FallbackType    Type = "fallback"
)

// Stringer
func (t Type) String() string { return string(t) }

type StateMutability string

var (
	PureStateMutability        StateMutability = "pure"
	ViewStateMutability        StateMutability = "view"
	NonpayableStateMutability  StateMutability = "nonpayable"
	PayableStateMutability     StateMutability = "payable"
)

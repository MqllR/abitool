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

var PayableStateMutability StateMutability = "payable"

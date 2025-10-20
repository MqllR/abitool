package abiparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type PrettyPrinter struct {
	a *ABI
}

func NewPrettyPrinter(abi *ABI) *PrettyPrinter {
	return &PrettyPrinter{a: abi}
}

func (p *PrettyPrinter) Print() (string, error) {
	var buf bytes.Buffer

	b, err := json.Marshal(p.a)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ABI: %w", err)
	}

	if err := json.Indent(&buf, []byte(b), "", "  "); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// TablePrinter prints the ABI in a table format.
// The columns are: Type, Name, Inputs, StateMutability
type TablePrinter struct {
	a *ABI
}

func NewTablePrinter(abi *ABI) *TablePrinter {
	return &TablePrinter{a: abi}
}

func (p *TablePrinter) Print() (string, error) {
	if p.a == nil {
		return "", nil
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "%-10s %-20s %-30s %-15s\n", "Type", "Name", "Inputs", "StateMutability")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 80))

	for element := range p.a.All() {
		// inputs := []string{}
		// for _, in := range element.Inputs {
		// 	inputs = append(inputs, fmt.Sprintf("%s %s", in.Type, in.Name))
		// }
		fmt.Fprintf(
			&b,
			"%-10s %-20s %-30s %-15s\n",
			element.Type,
			element.Name,
			"",
			element.StateMutability,
		)
	}

	return b.String(), nil
}

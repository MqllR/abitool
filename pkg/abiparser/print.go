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
	a              *ABI
	WithInputNames bool
}

type tableOptions func(*TablePrinter)

func WithInputNames() tableOptions {
	return func(tp *TablePrinter) {
		tp.WithInputNames = true
	}
}

func NewTablePrinter(abi *ABI, opts ...tableOptions) *TablePrinter {
	table := &TablePrinter{a: abi}

	for _, opt := range opts {
		opt(table)
	}

	return table
}

func (p *TablePrinter) Print() (string, error) {
	if p.a == nil {
		return "", nil
	}

	var b bytes.Buffer

	// Table headers
	headers := []string{"Type", "Name", "Inputs", "Selector", "StateMutability"}
	colWidths := make([]int, len(headers))
	copy(colWidths, []int{len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3]), len(headers[4])})

	// Collect all rows
	var rows [][]string
	for element := range p.a.All() {
		selector, err := element.Selector()
		if err != nil {
			selector = "N/A"
		}

		row := []string{
			string(element.Type),
			element.Name,
			p.formatInputTypes(element.Inputs),
			selector,
			string(element.StateMutability),
		}
		rows = append(rows, row)
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Fprintf(&b, "%-*s ", colWidths[i], h)
	}
	b.WriteByte('\n')

	// Print separator
	for _, w := range colWidths {
		b.WriteString(strings.Repeat("-", w))
	}
	b.WriteByte('\n')

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Fprintf(&b, "%-*s ", colWidths[i], cell)
		}
		b.WriteByte('\n')
	}

	return b.String(), nil
}

func (p *TablePrinter) formatInputTypes(inputs []Input) string {
	types := []string{}

	for _, in := range inputs {
		if p.WithInputNames {
			types = append(types, fmt.Sprintf("%s %s", in.Name, in.Type))
			continue
		}

		types = append(types, in.Type)
	}

	if len(types) == 0 {
		return ""
	}

	return fmt.Sprintf("(%s)", strings.Join(types, ","))
}

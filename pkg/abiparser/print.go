// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abiparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ─── Table styles ─────────────────────────────────────────────────────────────

var (
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2"))
	tableSepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))

	tableTypeFunction    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4BAFED"))
	tableTypeEvent       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F9D449"))
	tableTypeError       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	tableTypeConstructor = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	tableTypeFallback    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))

	tableSelectorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))
	tableSelectorNA     = lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
	tableMutViewStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	tableMutPayStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9D449"))
	tableMutDefaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	tableNameStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	tableInputStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
)

type PrettyPrinter struct {
	a *ABI
}

func NewPrettyPrinter(abi *ABI) *PrettyPrinter {
	return &PrettyPrinter{a: abi}
}

// enrichedElement wraps an Element for JSON output, adding computed fields
// (selector for functions/errors, topicHash for non-anonymous events).
type enrichedElement struct {
	Element
	Selector  *string `json:"selector,omitempty"`
	TopicHash *string `json:"topicHash,omitempty"`
}

func (p *PrettyPrinter) Print() (string, error) {
	enriched := make([]enrichedElement, 0, len(*p.a))
	for el := range p.a.All() {
		e := enrichedElement{Element: el}
		switch {
		case el.HasSelector():
			if sel, err := el.Selector(); err == nil {
				e.Selector = &sel
			}
		case el.HasTopicHash():
			if th, err := el.TopicHash(); err == nil {
				e.TopicHash = &th
			}
		}
		enriched = append(enriched, e)
	}

	var buf bytes.Buffer
	b, err := json.Marshal(enriched)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ABI: %w", err)
	}
	if err := json.Indent(&buf, b, "", "  "); err != nil {
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

	headers := []string{"Type", "Name", "Inputs", "Selector/Topic", "StateMutability"}
	colWidths := []int{len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3]), len(headers[4])}

	// Collect raw (unstyled) rows to measure column widths.
	type rawRow struct {
		typ      string
		name     string
		inputs   string
		selector string
		mut      string
	}
	var raws []rawRow
	for element := range p.a.All() {
		var id string
		switch {
		case element.HasSelector():
			sel, err := element.Selector()
			if err != nil {
				id = "N/A"
			} else {
				id = sel
			}
		case element.HasTopicHash():
			th, err := element.TopicHash()
			if err != nil {
				id = "N/A"
			} else {
				id = th
			}
		default:
			id = "N/A"
		}
		r := rawRow{
			typ:      string(element.Type),
			name:     element.Name,
			inputs:   p.formatInputTypes(element.Inputs),
			selector: id,
			mut:      string(element.StateMutability),
		}
		cells := []string{r.typ, r.name, r.inputs, r.selector, r.mut}
		for i, c := range cells {
			if len(c) > colWidths[i] {
				colWidths[i] = len(c)
			}
		}
		raws = append(raws, r)
	}

	// Spacing between columns.
	const gap = 2

	// Helper: pad a styled string to colWidth visible chars.
	cell := func(styled string, colW int) string {
		return lipgloss.NewStyle().Width(colW + gap).Render(styled)
	}

	// ── Header ────────────────────────────────────────────────────────────────
	for i, h := range headers {
		b.WriteString(cell(tableHeaderStyle.Render(h), colWidths[i]))
	}
	b.WriteByte('\n')

	// ── Separator ─────────────────────────────────────────────────────────────
	totalWidth := gap * len(colWidths)
	for _, w := range colWidths {
		totalWidth += w
	}
	b.WriteString(tableSepStyle.Render(strings.Repeat("─", totalWidth)))
	b.WriteByte('\n')

	// ── Rows ──────────────────────────────────────────────────────────────────
	for _, r := range raws {
		b.WriteString(cell(styledType(r.typ), colWidths[0]))
		b.WriteString(cell(tableNameStyle.Render(r.name), colWidths[1]))
		b.WriteString(cell(tableInputStyle.Render(r.inputs), colWidths[2]))
		b.WriteString(cell(styledSelector(r.selector), colWidths[3]))
		b.WriteString(cell(styledMutability(r.mut), colWidths[4]))
		b.WriteByte('\n')
	}

	return b.String(), nil
}

func styledType(t string) string {
	switch Type(t) {
	case FunctionType:
		return tableTypeFunction.Render(t)
	case EventType:
		return tableTypeEvent.Render(t)
	case ErrorType:
		return tableTypeError.Render(t)
	case ConstructorType:
		return tableTypeConstructor.Render(t)
	default:
		return tableTypeFallback.Render(t)
	}
}

func styledSelector(s string) string {
	if s == "N/A" {
		return tableSelectorNA.Render(s)
	}
	return tableSelectorStyle.Render(s)
}

func styledMutability(m string) string {
	switch m {
	case "view", "pure":
		return tableMutViewStyle.Render(m)
	case "payable":
		return tableMutPayStyle.Render(m)
	default:
		return tableMutDefaultStyle.Render(m)
	}
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

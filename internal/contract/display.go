// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/pkg/abiparser"
)

// ─── List styles ──────────────────────────────────────────────────────────────

var (
	listHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2"))
	listSepStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
	listAddrStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))
	listNameStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	listABITrueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	listABIFalseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
)

type Printer interface {
	Print() (string, error)
}
func Print(abi *abiparser.ABI) (string, error) {
	t := viper.GetString("abi-view-type")

	var filtered abiparser.ABI

	switch t {
	case "all":
		filtered = *abi
	case "function":
		for el := range abi.All() {
			if el.IsFunction() {
				filtered = append(filtered, el)
			}
		}
	case "event":
		for el := range abi.All() {
			if el.Type == abiparser.EventType {
				filtered = append(filtered, el)
			}
		}
	case "constructor":
		for el := range abi.All() {
			if el.Type == abiparser.ConstructorType {
				filtered = append(filtered, el)
			}
		}
	case "fallback":
		for el := range abi.All() {
			if el.Type == abiparser.FallbackType {
				filtered = append(filtered, el)
			}
		}
	case "receive":
		for el := range abi.All() {
			if el.Type == abiparser.ReceiveType {
				filtered = append(filtered, el)
			}
		}
	default:
		return "", fmt.Errorf("unsupported ABI type filter: %s", t)
	}

	var abiPrinter Printer

	switch viper.GetString("abi-view-output") {
	case "json":
		abiPrinter = abiparser.NewPrettyPrinter(&filtered)
	case "table":
		var opts []abiparser.TableOption
		if viper.GetBool("abi-view-with-intput-name") {
			opts = append(opts, abiparser.WithInputNames())
		}
		if viper.GetBool("abi-view-with-output-name") {
			opts = append(opts, abiparser.WithOutputNames())
		}
		abiPrinter = abiparser.NewTablePrinter(&filtered, opts...)
	default:
		return "", fmt.Errorf("unsupported ABI print format: %s", viper.GetString("abi-print-format"))
	}

	return abiPrinter.Print()
}

func PrintContractList(contracts []*Contract, chainID int) string {
	headers := []string{"Address", "Contract Name", "ABI"}

	const gap = 2

	// Build duplicate-display-name set for ⚠ annotation.
	nameCounts := make(map[string]int, len(contracts))
	for _, c := range contracts {
		nameCounts[c.DisplayName()]++
	}

	// Measure column widths from raw (unstyled) content.
	colWidths := []int{len(headers[0]), len(headers[1]), len(headers[2])}
	type row struct {
		addr, name, hasABI string
		rawName            string // unstyled, for width measurement
	}
	rows := make([]row, len(contracts))
	for i, c := range contracts {
		abiStr := fmt.Sprintf("%t", c.HasABI())

		// Build display name: label [EtherscanName] or just the name.
		displayName := c.DisplayName()
		rawName := displayName
		if nameCounts[displayName] > 1 {
			rawName = "⚠ " + displayName
		}
		if c.Metadata.Label != "" && c.Metadata.Label != c.Metadata.ContractName {
			rawName = rawName + " [" + c.Metadata.ContractName + "]"
		}

		rows[i] = row{c.Address, rawName, abiStr, rawName}
		if l := len(c.Address); l > colWidths[0] {
			colWidths[0] = l
		}
		if l := len(rawName); l > colWidths[1] {
			colWidths[1] = l
		}
		if l := len(abiStr); l > colWidths[2] {
			colWidths[2] = l
		}
	}

	cell := func(styled string, colW int) string {
		return lipgloss.NewStyle().Width(colW + gap).Render(styled)
	}

	totalWidth := gap * len(colWidths)
	for _, w := range colWidths {
		totalWidth += w
	}

	var sb strings.Builder

	// Chain header
	chainLabel := fmt.Sprintf("Chain: %s (%d)", ChainName(chainID), chainID)
	sb.WriteString(listSepStyle.Render(chainLabel))
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	// Header
	for i, h := range headers {
		sb.WriteString(cell(listHeaderStyle.Render(h), colWidths[i]))
	}
	sb.WriteByte('\n')
	sb.WriteString(listSepStyle.Render(strings.Repeat("─", totalWidth)))
	sb.WriteByte('\n')

	// Rows
	for _, r := range rows {
		abiStyled := listABIFalseStyle.Render(r.hasABI)
		if r.hasABI == "true" {
			abiStyled = listABITrueStyle.Render(r.hasABI)
		}
		sb.WriteString(cell(listAddrStyle.Render(r.addr), colWidths[0]))
		sb.WriteString(cell(listNameStyle.Render(r.name), colWidths[1]))
		sb.WriteString(cell(abiStyled, colWidths[2]))
		sb.WriteByte('\n')
	}

	return sb.String()
}

package contract

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/pkg/abiparser"
)

type Printer interface {
	Print() (string, error)
}

func NoopPrinter(_ string) (string, error) {
	return "", nil
}

// PrettyPrint displays ABI in a pretty format.
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
		abiPrinter = abiparser.NewTablePrinter(&filtered)
		if viper.GetBool("abi-view-with-intput-name") {
			abiPrinter = abiparser.NewTablePrinter(&filtered, abiparser.WithInputNames())
		}
	default:
		return "", fmt.Errorf("unsupported ABI print format: %s", viper.GetString("abi-print-format"))
	}

	return abiPrinter.Print()
}

func PrintContractList(contracts []*Contract) string {
	headers := []string{"Address", "Contract Name", "ABI"}

	// Prepare rows
	rows := make([][]string, len(contracts))
	for i, c := range contracts {
		rows[i] = []string{c.Address, c.Metadata.ContractName, fmt.Sprintf("%t", c.HasABI())}
	}

	// Calculate max width for each column
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if len(col) > colWidths[i] {
				colWidths[i] = len(col)
			}
		}
	}

	// Print header
	var output strings.Builder
	for i, h := range headers {
		fmt.Fprintf(&output, "%-*s", colWidths[i]+2, h)
	}
	output.WriteString("\n")

	// Print rows
	for _, row := range rows {
		for i, col := range row {
			fmt.Fprintf(&output, "%-*s", colWidths[i]+2, col)
		}
		output.WriteString("\n")
	}

	return output.String()
}

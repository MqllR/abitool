package contract

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/pkg/abiparser"
)

// PrettyPrint displays ABI in a pretty format.
func Print(abi *abiparser.ABI) (string, error) {
	t := viper.GetString("abi-view-type")

	var newABI abiparser.ABI

	switch t {
	case "all":
		newABI = *abi
	case "function":
		for el := range abi.All() {
			if el.IsFunction() {
				newABI = append(newABI, el)
			}
		}
	case "event":
		for el := range abi.All() {
			if el.Type == abiparser.EventType {
				newABI = append(newABI, el)
			}
		}
	case "constructor":
		for el := range abi.All() {
			if el.Type == abiparser.ConstructorType {
				newABI = append(newABI, el)
			}
		}
	case "fallback":
		for el := range abi.All() {
			if el.Type == abiparser.FallbackType {
				newABI = append(newABI, el)
			}
		}
	case "receive":
		for el := range abi.All() {
			if el.Type == abiparser.ReceiveType {
				newABI = append(newABI, el)
			}
		}
	default:
		return "", fmt.Errorf("unsupported ABI type filter: %s", t)
	}

	var abiPrinter Printer

	switch viper.GetString("abi-view-output") {
	case "json":
		abiPrinter = abiparser.NewPrettyPrinter(&newABI)
	case "table":
		abiPrinter = abiparser.NewTablePrinter(&newABI)
	default:
		return "", fmt.Errorf("unsupported ABI print format: %s", viper.GetString("abi-print-format"))
	}

	return abiPrinter.Print()
}

package contract

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/pkg/abiparser"
)

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

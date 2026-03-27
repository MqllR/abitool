package contract

import "github.com/MqllR/abitool/pkg/chains"

// KnownChains is an alias for the shared chains.Known map for backward compatibility
// within the contract package.
var KnownChains = chains.Known

// ChainName returns the human-readable name for a chain ID.
// Falls back to "Chain <id>" for unknown chains.
func ChainName(id int) string {
	return chains.Name(id)
}

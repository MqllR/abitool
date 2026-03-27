// Package chains provides display metadata for well-known EVM chains.
package chains

import "fmt"

// Info holds display metadata for a known EVM chain.
type Info struct {
	Name string
}

// Known maps well-known EVM chain IDs to their display names.
// This is used for display purposes only — abitool accepts any chain ID.
var Known = map[int]Info{
	1:        {Name: "Ethereum Mainnet"},
	5:        {Name: "Goerli"},
	10:       {Name: "Optimism"},
	56:       {Name: "BNB Chain"},
	100:      {Name: "Gnosis"},
	137:      {Name: "Polygon"},
	8453:     {Name: "Base"},
	42161:    {Name: "Arbitrum One"},
	43114:    {Name: "Avalanche"},
	59144:    {Name: "Linea"},
	11155111: {Name: "Sepolia"},
	11155420: {Name: "Optimism Sepolia"},
	84532:    {Name: "Base Sepolia"},
	421614:   {Name: "Arbitrum Sepolia"},
	80002:    {Name: "Polygon Amoy"},
}

// Name returns the human-readable name for a chain ID.
// Falls back to "Chain <id>" for unknown chains.
func Name(id int) string {
	if info, ok := Known[id]; ok {
		return info.Name
	}
	return fmt.Sprintf("Chain %d", id)
}

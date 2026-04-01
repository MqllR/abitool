// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

// Package chains provides display metadata for well-known EVM chains.
package chains

import "fmt"

// Info holds display metadata for a known EVM chain.
type Info struct {
	Name          string
	DefaultRPCURL string
}

// Known maps well-known EVM chain IDs to their display names and public RPC endpoints.
// DefaultRPCURL values are public, rate-limited endpoints suitable as fallbacks.
var Known = map[int]Info{
	1:        {Name: "Ethereum Mainnet", DefaultRPCURL: "https://ethereum-rpc.publicnode.com"},
	10:       {Name: "Optimism Mainnet", DefaultRPCURL: "https://mainnet.optimism.io"},
	56:       {Name: "BNB Chain", DefaultRPCURL: "https://bsc-dataseed.binance.org"},
	137:      {Name: "Polygon Mainnet", DefaultRPCURL: "https://polygon-rpc.com"},
	8453:     {Name: "Base Mainnet", DefaultRPCURL: "https://mainnet.base.org"},
	42161:    {Name: "Arbitrum One Mainnet", DefaultRPCURL: "https://arb1.arbitrum.io/rpc"},
	43114:    {Name: "Avalanche Mainnet", DefaultRPCURL: "https://api.avax.network/ext/bc/C/rpc"},
	59144:    {Name: "Linea Mainnet", DefaultRPCURL: "https://rpc.linea.build"},
	11155111: {Name: "Ethereum Sepolia", DefaultRPCURL: "https://ethereum-sepolia-rpc.publicnode.com"},
	11155420: {Name: "Optimism Sepolia", DefaultRPCURL: "https://sepolia.optimism.io"},
	84532:    {Name: "Base Sepolia", DefaultRPCURL: "https://sepolia.base.org"},
	421614:   {Name: "Arbitrum Sepolia", DefaultRPCURL: "https://sepolia-rollup.arbitrum.io/rpc"},
	80002:    {Name: "Polygon Amoy", DefaultRPCURL: "https://rpc-amoy.polygon.technology"},
}

// Name returns the human-readable name for a chain ID.
// Falls back to "Chain <id>" for unknown chains.
func Name(id int) string {
	if info, ok := Known[id]; ok {
		return info.Name
	}
	return fmt.Sprintf("Chain %d", id)
}

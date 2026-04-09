// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

type Metadata struct {
	ContractName string `json:"contract_name"`
	ABIPath      string `json:"abi_path"`
	Label        string `json:"label,omitempty"`
}

type Contract struct {
	Address  string   `json:"address"`
	Metadata Metadata `json:"metadata"`
}

func (c *Contract) HasABI() bool {
	return c.Metadata.ABIPath != ""
}

func (c *Contract) Name() string {
	return c.Metadata.ContractName
}

// DisplayName returns the user-defined label when set, otherwise the Etherscan contract name.
func (c *Contract) DisplayName() string {
	if c.Metadata.Label != "" {
		return c.Metadata.Label
	}
	return c.Metadata.ContractName
}

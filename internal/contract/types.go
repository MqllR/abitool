// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

type Metadata struct {
	ContractName string `json:"contract_name"`
	ABIPath      string `json:"abi_path"`
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

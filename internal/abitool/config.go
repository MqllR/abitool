// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abitool

var cfg Config

func ConfigInstance() *Config {
	return &cfg
}

type Config struct {
	ChainID   int             `mapstructure:"chainid"`
	EtherScan EtherScanConfig `mapstructure:"etherscan"`
	RPC       RPCConfig       `mapstructure:"rpc"`
}

type EtherScanConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type RPCConfig struct {
	URL string `mapstructure:"url"`
}

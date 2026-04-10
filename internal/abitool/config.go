// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abitool

var cfg Config

func ConfigInstance() *Config {
	return &cfg
}

type Config struct {
	ChainID   int             `mapstructure:"chainid" yaml:"chainid,omitempty"`
	EtherScan EtherScanConfig `mapstructure:"etherscan" yaml:"etherscan,omitempty"`
	RPC       RPCConfig       `mapstructure:"rpc" yaml:"rpc,omitempty"`
}

type EtherScanConfig struct {
	APIKey string `mapstructure:"api_key" yaml:"api_key,omitempty"`
}

type RPCConfig struct {
	URL string `mapstructure:"url" yaml:"url,omitempty"`
}

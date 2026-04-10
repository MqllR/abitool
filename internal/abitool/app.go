// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abitool

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

func Load() error {
	var c Config

	configPath := viper.GetString("config")
	cfgWithEnv := os.ExpandEnv(configPath)

	fh, err := os.Open(cfgWithEnv)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	defer func() { _ = fh.Close() }()

	viper.SetConfigType("yaml")

	if err = viper.ReadConfig(fh); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err = viper.Unmarshal(&c); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg = c
	return nil
}

// SaveChainID updates the default chain ID in memory and persists it to the config file.
func SaveChainID(chainID int) error {
	cfg.ChainID = chainID
	return saveConfig()
}

// saveConfig marshals the current Config back to the YAML config file.
func saveConfig() error {
	configPath := os.ExpandEnv(viper.GetString("config"))

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

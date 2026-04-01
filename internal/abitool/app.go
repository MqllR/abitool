// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abitool

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
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

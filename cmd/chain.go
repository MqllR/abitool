// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/cmd/chain"
)

func init() {
	chainCmd.AddCommand(chain.UseCmd)
	rootCmd.AddCommand(chainCmd)
}

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Manage chain settings",
}

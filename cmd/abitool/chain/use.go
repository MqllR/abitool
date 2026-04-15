// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package chain

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/pkg/chains"
)

var UseCmd = &cobra.Command{
	Use:   "use <chainID>",
	Short: "Set the default chain ID",
	Long:  "Persists the given chain ID as the default in the config file.",
	Args:  cobra.ExactArgs(1),
	Run:   use,
}

func use(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil || id <= 0 {
		fmt.Fprintf(os.Stderr, "Error: %q is not a valid chain ID — must be a positive integer\n", args[0])
		os.Exit(1)
	}

	if err := abitool.SaveChainID(id); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving chain ID: %v\n", err)
		os.Exit(1)
	}

	name := chains.Name(id)
	fmt.Printf("✓ Default chain set to %s (%d)\n", name, id)

	if _, known := chains.Known[id]; !known {
		fmt.Println("  (chain ID not in the known list — double-check before using)")
	}
}

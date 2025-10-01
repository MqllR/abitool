package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(abiCmd)
}

var rootCmd = &cobra.Command{
	Use:   "abitool",
	Short: "CLI tool for Ethereum ABI operations",
	Long:  "Download, parse, and interact with Ethereum smart contract ABIs.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/internal/ui"
)

func init() {
	rootCmd.Version = Version

	rootCmd.PersistentFlags().StringP("config", "c", "$HOME/.config/abitool/config.yaml", "Path to configuration file")

	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(abiCmd)
	rootCmd.AddCommand(rpcCmd)
}

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "abitool",
	Short: "CLI tool for Ethereum ABI operations",
	Long:  "Download, parse, and interact with Ethereum smart contract ABIs.",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if err := abitool.Load(); err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}
	},
	// When called with no subcommand, launch the interactive TUI dashboard.
	RunE: func(cmd *cobra.Command, _ []string) error {
		return ui.RunApp()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().StringP("config", "f", "$HOME/.config/abitool/config.yaml", "Path to configuration file")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(abiCmd)
}

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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package abi

import (
	"github.com/spf13/cobra"
)

// deleteCmd deletes a contract by its address.
var DeleteCmd = &cobra.Command{
	Use:     "delete <address>",
	Aliases: []string{"del", "d"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete a contract by its address",
}

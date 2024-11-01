package main

import (
	"github.com/spf13/cobra"
)

var commandProvider = &cobra.Command{
	Use:   "provider",
	Short: "Manage providers",
}

func init() {
	mainCommand.AddCommand(commandProvider)
}

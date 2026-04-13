package main

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "allele",
	Short: "Allele CLI tools",
}

func initCLI() {
	rootCmd.AddCommand(proveCmd)
}

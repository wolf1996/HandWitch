package cmd

import (
	"github.com/spf13/cobra"
)

func BuildAll() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "handwitch",
		Short: "handwitch root comand",
	}
	registerServeBot(rootCmd)
	return rootCmd
}

package cmd

import (
	"github.com/spf13/cobra"
)

func registerServeBot(parentCmd *cobra.Command) {
	comand := cobra.Command{
		Use:   "serve",
		Short: "Starts bot",
		Run: func(cmd *cobra.Command, args []string) {
			return
		},
	}
	parentCmd.AddCommand(&comand)
}

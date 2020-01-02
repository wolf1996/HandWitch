package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func registerServeBot(parentCmd *cobra.Command) {
	comand := cobra.Command{
		Use:   "serve",
		Short: "Starts bot",
		Run: func(cmd *cobra.Command, args []string) {
			path := viper.GetString("path")
			fmt.Printf("path:%v \n", path)
			return
		},
	}
	comand.PersistentFlags().String("token", "info", "log level [info|warn|debug]")
	comand.PersistentFlags().String("whitelist", "", "configuration path file")
	comand.PersistentFlags().String("formating", "", "descriptions file path")
	comand.PersistentFlags().String("tgproxy", "", "proxy to telegram client")
	parentCmd.AddCommand(&comand)
}

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func exec(cmd *cobra.Command, args []string) {
	loglevelStr := viper.GetString("log")
	loglevel, err := log.ParseLevel(loglevelStr)
	if err != nil {
		log.Fatalf("Failed to parse LogLevel %s", err.Error())
	}
	logger := log.StandardLogger()
	logger.SetLevel(loglevel)

	path := viper.GetString("path")
	logger.Infof("Used description path: %s", path)

	whitelist := viper.GetString("telegram.white_list")
	logger.Infof("Used whitelist: %s", whitelist)

	formating := viper.GetString("telegram.formating")
	logger.Infof("Used formating: %s", formating)

	tgproxy := viper.GetString("telegram.proxy")
	logger.Infof("Used proxy for telegram client: %s", tgproxy)

	logger.Infof("Used config: %s", viper.ConfigFileUsed())
}

func registerServeBot(parentCmd *cobra.Command) {
	comand := cobra.Command{
		Use:   "serve",
		Short: "Starts bot",
		Run:   exec,
	}
	comand.PersistentFlags().String("token", "info", "log level [info|warn|debug]")
	comand.PersistentFlags().String("whitelist", "", "configuration path file")
	comand.PersistentFlags().String("formating", "", "descriptions file path")
	comand.PersistentFlags().String("tgproxy", "", "proxy to telegram client")
	bindFlag(&comand, "telegram.white_list", "whitelist")
	bindFlag(&comand, "telegram.formating", "formating")
	bindFlag(&comand, "telegram.proxy", "tgproxy")
	parentCmd.AddCommand(&comand)
}

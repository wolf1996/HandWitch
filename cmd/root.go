package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	fmt.Printf("config path: %s\n", viper.ConfigFileUsed())
	return nil
}

func prerunRoot(cmd *cobra.Command, args []string) error {
	logLevel := cmd.Flag("log").Value.String()

	loglevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Failed to parse LogLevel %s", err.Error())
	}
	log.SetLevel(loglevel)

	// TODO: сделать дефолтный config path?
	configPath := cmd.Flag("config").Value.String()
	fmt.Printf("config: %s\n", configPath)
	if configPath != "" {
		err = initConfig(configPath)
		if err != nil {
			log.Fatalf("Failed to parse config %s", err.Error())
		}
	}

	return nil
}

func BuildAll() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "handwitch",
		PersistentPreRunE: prerunRoot,
		Short:             "handwitch root comand",
	}
	rootCmd.PersistentFlags().String("log", "info", "log level [info|warn|debug]")
	rootCmd.PersistentFlags().String("config", "", "configuration path file")
	rootCmd.PersistentFlags().String("path", "", "descriptions file path")
	err := viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	if err != nil {
		log.Fatalf("failed to bind pflag \"log\" %s", err.Error())
	}
	err = viper.BindPFlag("path", rootCmd.PersistentFlags().Lookup("path"))
	if err != nil {
		log.Fatalf("failed to bind pflag \"path\" %s", err.Error())
	}
	registerServeBot(rootCmd)
	return rootCmd
}

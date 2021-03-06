package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initConfig устанавливаем конфигурационный файл в viper
// TODO: сделать дефолтные конфигурационные файлы
func initConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	fmt.Printf("config path: %s\n", viper.ConfigFileUsed())
	return nil
}

// prerunRoot парсим флаги и выставляем логлевлы и конфигурационные файлы
func prerunRoot(cmd *cobra.Command, args []string) error {
	logLevel := cmd.Flag("log").Value.String()

	loglevel, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(loglevel)

	// TODO: сделать дефолтный config path?
	configPath := cmd.Flag("config").Value.String()
	fmt.Printf("config: %s\n", configPath)
	if configPath != "" {
		err = initConfig(configPath)
		if err != nil {
			return err
		}
	}

	return nil
}

//buildSystemContext собираем контекст для корректной обработки сигналов системы
func buildSystemContext(log *log.Logger) context.Context {
	// Вешаем обработчики сигналов на контекст
	ctx, cancel := context.WithCancel(context.Background())
	sysSignals := make(chan os.Signal)

	signal.Notify(sysSignals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		signal := <-sysSignals
		log.Infof("Got %s system signal, aborting...", signal)
		cancel()
	}()
	return ctx
}

// bindFlag связываем флаг и поле в конфиге
func bindFlag(cmd *cobra.Command, conf string, name string) error {
	err := viper.BindPFlag(conf, cmd.PersistentFlags().Lookup(name))
	if err != nil {
		return err
	}
	return nil
}

// BuildAll собираем основную команду
func buildRootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:               "handwitch",
		PersistentPreRunE: prerunRoot,
		Short:             "handwitch helps you to handle http request without frontend",
	}
	rootCmd.PersistentFlags().String("log", "info", "log level [info|warn|debug]")
	rootCmd.PersistentFlags().String("config", "", "configuration path file")
	rootCmd.PersistentFlags().String("path", "", "descriptions file path")

	err := rootCmd.MarkPersistentFlagFilename("config")
	if err != nil {
		return nil, err
	}
	err = rootCmd.MarkPersistentFlagRequired("config")
	if err != nil {
		return nil, err
	}
	err = rootCmd.MarkPersistentFlagFilename("path")
	if err != nil {
		return nil, err
	}

	err = bindFlag(rootCmd, "log_level", "log")
	if err != nil {
		return nil, err
	}
	err = bindFlag(rootCmd, "path", "path")
	if err != nil {
		return nil, err
	}
	return rootCmd, nil
}

// BuildAll собираем все команды и припиниваем их к рутовой
func BuildAll() (*cobra.Command, error) {
	rootCmd, err := buildRootCmd()
	if err != nil {
		return nil, err
	}
	_, err = registerServeBot(rootCmd)
	if err != nil {
		return nil, err
	}
	return rootCmd, nil
}

package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bot "github.com/wolf1996/HandWitch/pkg/bot"
	"github.com/wolf1996/HandWitch/pkg/core"
)

func getDescriptionSourceFromFile(path string) (*core.URLProcessor, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	ext := filepath.Ext(path)
	var descriptionSource core.DescriptionsSource
	switch ext {
	case ".yaml":
		descriptionSource, err = core.GetDescriptionSourceFromYAML(reader)
		if err != nil {
			return nil, err
		}

	case ".json":
		descriptionSource, err = core.GetDescriptionSourceFromJSON(reader)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unknown file extension %s", ext)
	}
	if err != nil {
		return nil, err
	}
	processor := core.NewURLProcessor(descriptionSource, http.DefaultClient)
	//TODO: попробовать поправить ссылки и интерфейсы
	return &processor, nil
}

func getAuthSourceFromFile(path string) (bot.Authorisation, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	authSource, err := bot.GetAuthSourceFromJSON(reader)
	if err != nil {
		return nil, err
	}
	return authSource, err
}

func exec(cmd *cobra.Command, args []string) error {
	loglevelStr := viper.GetString("log_level")
	loglevel, err := log.ParseLevel(loglevelStr)
	if err != nil {
		return fmt.Errorf("Failed to parse LogLevel %s", err.Error())
	}
	logger := log.StandardLogger()
	logger.Infof("Used config: %s", viper.ConfigFileUsed())

	logger.SetLevel(loglevel)

	path := viper.GetString("path")
	logger.Infof("Used description path: %s", path)

	whitelist := viper.GetString("telegram.white_list")
	logger.Infof("Used whitelist: %s", whitelist)

	formating := viper.GetString("telegram.formating")
	logger.Infof("Used formating: %s", formating)

	tgproxy := viper.GetString("telegram.proxy")
	logger.Infof("Used proxy for telegram client: %s", tgproxy)

	token := cmd.Flags().Lookup("token").Value.String()

	client := http.DefaultClient
	if tgproxy != "" {
		log.Infof("Got Proxy %s", tgproxy)
		client, err = bot.GetClientWithProxy(tgproxy)
		if err != nil {
			return fmt.Errorf("Failed to create http client with proxy %s", err.Error())
		}
	} else {
		log.Info("Got No Proxy")
	}

	// пытаемся собрать список разрешённых пользователей
	// если такой список не указан - доступны всем ветрам
	var auth bot.Authorisation
	if whitelist != "" {
		auth, err = getAuthSourceFromFile(whitelist)
		if err != nil {
			return fmt.Errorf("Failed to get auth %s, stop", err.Error())
		}
	} else {
		auth = bot.DummyAuthorisation{}
		log.Info("No whitelist found starting with dummy auth")
	}

	// грузим описания ручек
	log.Infof("Description file path used %s", path)
	urlContainer, err := getDescriptionSourceFromFile(path)
	if err != nil {
		return fmt.Errorf("Failed to get description source file %s", err.Error())
	}

	log.Info("Creating telegram bot api client")
	botInstance, err := bot.NewBot(client, token, *urlContainer, auth, formating)
	if err != nil {
		return fmt.Errorf("Failed to create bot %s", err.Error())
	}

	log.Info("Telegram bot api client created")

	ctx := buildSystemContext(logger)

	err = botInstance.Listen(ctx, logger)
	if err != nil {
		log.Infof("Stopping bot %s", err.Error())
	}
	return nil
}

func registerServeBot(parentCmd *cobra.Command) (*cobra.Command, error) {
	comand := cobra.Command{
		Use:   "serve",
		Short: "Starts bot",
		RunE:  exec,
	}
	comand.PersistentFlags().String("token", "info", "log level [info|warn|debug]")
	comand.PersistentFlags().String("whitelist", "", "configuration path file")
	comand.PersistentFlags().String("formating", "", "descriptions file path")
	comand.PersistentFlags().String("tgproxy", "", "proxy to telegram client")
	err := bindFlag(&comand, "telegram.white_list", "whitelist")
	if err != nil {
		return &comand, err
	}
	err = bindFlag(&comand, "telegram.formating", "formating")
	if err != nil {
		return &comand, err
	}
	err = bindFlag(&comand, "telegram.proxy", "tgproxy")
	if err != nil {
		return &comand, err
	}
	parentCmd.AddCommand(&comand)
	return &comand, nil
}
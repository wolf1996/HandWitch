package telegram

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
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

func buildHttpClient(proxyUrl string, logger *log.Logger) (*http.Client, error) {
	if proxyUrl != "" {
		logger.Infof("Got Proxy %s", proxyUrl)
		client, err := bot.GetClientWithProxy(proxyUrl)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	logger.Info("Got No Proxy")
	return http.DefaultClient, nil
}

func buildTelegramAuth(authPath string, logger *log.Logger) (bot.Authorisation, error) {

	// пытаемся собрать список разрешённых пользователей
	// если такой список не указан - открыты всем ветрам
	if authPath != "" {
		auth, err := getAuthSourceFromFile(authPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to get auth %w, stop", err)
		}
		return auth, nil
	}
	logger.Info("No whitelist found starting with dummy auth")
	return bot.DummyAuthorisation{}, nil
}

func tryExtractHookInfo() (*bot.HookConfig, error) {
	// TODO: Переделать на нормальный парсинг конфига

	hookMap := viper.GetStringMap("hook")

	if len(hookMap) == 0 {
		return nil, fmt.Errorf("no hook info")
	}

	getHookField := func(fieldName string) (string, error) {
		value, ok := hookMap[fieldName]
		if !ok {
			return "", fmt.Errorf("no field with such name")
		}
		strValue, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("field type is not string")
		}
		return strValue, nil
	}

	urlStr, err := getHookField("url")
	if err != nil {
		return nil, fmt.Errorf("Can't get url info %w", err)
	}

	cert, err := getHookField("cert")
	if err != nil {
		return nil, fmt.Errorf("Can't get cert info %w", err)
	}

	key, err := getHookField("key")
	if err != nil {
		return nil, fmt.Errorf("Can't get key info %w", err)
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse url from config: %w", err)
	}

	var bot bot.HookConfig
	bot.Cert = cert
	bot.Key = key
	bot.Port = url.Port()
	bot.Host = url.Host
	bot.URLPath = url.Path

	return &bot, nil
}

func Exec(ctx context.Context, cmd *cobra.Command, args []string, logger *log.Logger) error {
	path := viper.GetString("path")
	logger.Infof("Used description path: %s", path)

	whitelist := viper.GetString("telegram.white_list")
	logger.Infof("Used whitelist: %s", whitelist)

	formating := viper.GetString("telegram.formatting")
	logger.Infof("Used formatting: %s", formating)

	telegramProxy := viper.GetString("telegram.proxy")
	logger.Infof("Used proxy for telegram client: %s", telegramProxy)

	token := cmd.Flags().Lookup("token").Value.String()

	httpClient, err := buildHttpClient(telegramProxy, logger)
	if err != nil {
		logger.Errorf("Failed to build http client with proxy %s error: %s", telegramProxy, err.Error())
		return nil
	}

	auth, err := buildTelegramAuth(whitelist, logger)
	if err != nil {
		logger.Errorf("Failed to build telegram auth %s", err)
	}

	// грузим описания ручек
	logger.Infof("Description file path used %s", path)

	urlContainer, err := getDescriptionSourceFromFile(path)
	if err != nil {
		logger.Errorf("Failed to get description source file %s", err.Error())
		return nil
	}

	logger.Info("Creating telegram bot api client")

	hookConfig, err := tryExtractHookInfo()

	if err != nil {
		logger.Infof("Failed to get hook info %s", err.Error())
		hookConfig = nil
	}

	botInstance, err := bot.NewBot(httpClient, token, *urlContainer, auth, formating, hookConfig)
	if err != nil {
		logger.Errorf("Failed to create bot %s", err.Error())
		return nil
	}

	logger.Info("Telegram bot api client created")

	err = botInstance.Listen(ctx, logger)
	if err != nil {
		logger.Errorf("Stopping bot %s", err.Error())
		return nil
	}
	return nil
}

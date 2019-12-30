package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"
	bot "github.com/wolf1996/HandWitch/pkg/bot"
	"github.com/wolf1996/HandWitch/pkg/core"
	"path/filepath"
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

func getConfigFromPath(path string) (*bot.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	botConfig, err := bot.GetConfigFromJSON(reader)
	if err != nil {
		return nil, err
	}
	return botConfig, nil
}

func main() {
	token := flag.String("token", "", "telegramm token")
	proxy := flag.String("proxy", "", "proxy to telegram")
	path := flag.String("path", "", "path to descriptions")
	logLevel := flag.String("log", "info", "log level")
	whiteListPath := flag.String("whitelist", "", "path to list of allowed users")
	formating := flag.String("formating", "", "formating [markdown/html] for message")
	configPath := flag.String("config", "", "bot configuration path")

	flag.Parse()
	var err error
	config := bot.GetDefaultConfig()
	if *configPath != "" {
		config, err = getConfigFromPath(*configPath)
		if err != nil {
			log.Fatalf("Failed to parse config: %s", err.Error())
		}
	}

	if *formating == "" {
		*formating = config.Formatting
	}

	if *logLevel == "" {
		*logLevel = config.LogLevel
	}

	if *whiteListPath == "" {
		*whiteListPath = config.WhiteList
	}

	if *path == "" {
		*path = config.Path
	}

	if *proxy == "" {
		*proxy = config.Proxy
	}

	loglevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Failed to parse LogLevel %s", err.Error())
	}
	logger := log.StandardLogger()
	logger.SetLevel(loglevel)

	// из-за блокировок телеграмм даём возможность работы через прокси
	client := http.DefaultClient
	if *proxy != "" {
		log.Infof("Got Proxy %s", *proxy)
		client, err = bot.GetClientWithProxy(*proxy)
		if err != nil {
			log.Fatalf("Failed to create http client with proxy %s", err.Error())
		}
	} else {
		log.Info("Got No Proxy")
	}

	// пытаемся собрать список разрешённых пользователей
	// если такой список не указан - доступны всем ветрам
	var auth bot.Authorisation
	if *whiteListPath != "" {
		auth, err = getAuthSourceFromFile(*whiteListPath)
		if err != nil {
			log.Fatalf("Failed to get auth %s, stop", err.Error())
		}
	} else {
		auth = bot.DummyAuthorisation{}
		log.Info("No whitelist found starting with dummy auth")
	}

	// грузим описания ручек
	log.Printf("Description file path used %s", *path)
	urlContainer, err := getDescriptionSourceFromFile(*path)
	if err != nil {
		log.Fatalf("Failed to get description source file %s", err.Error())
	}

	// собираем бота
	botInstance, err := bot.NewBot(client, *token, *urlContainer, auth, *formating)
	if err != nil {
		log.Fatalf("Failed tto create bot %s", err.Error())
	}

	// Вешаем обработчики сигналов на контекст
	ctx, cancel := context.WithCancel(context.Background())
	sysSignals := make(chan os.Signal, 1)

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

	err = botInstance.Listen(ctx, logger)
	if err != nil {
		log.Infof("Stopping bot %s", err.Error())
	}
}

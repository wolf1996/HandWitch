package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"flag"
	bot "github.com/wolf1996/HandWitch/pkg/bot"
	"github.com/wolf1996/HandWitch/pkg/core"
	"net/http"
)

func getDescriptionSourceFromFile(path string) (*core.URLProcessor, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	descriptionSource, err := core.GetDescriptionSourceFromJSON(reader)
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

func main() {
	token := flag.String("token", "", "telegramm token")
	proxy := flag.String("proxy", "", "proxy to telegram")
	path := flag.String("path", "", "path to descriptions")
	whiteListPath := flag.String("whitelist", "", "path to list of allowed users")
	flag.Parse()
	var err error
	client := http.DefaultClient
	if *proxy != "" {
		log.Printf("Got Proxy %s", *proxy)
		client, err = bot.GetClientWithProxy(*proxy)
	} else {
		log.Print("Got No Proxy")
	}
	var auth bot.Authorisation
	if *whiteListPath != "" {
		auth, err = getAuthSourceFromFile(*whiteListPath)
		if err != nil {
			log.Fatalf("Failed to get auth %s, stop", err.Error())
		}
	} else {
		auth = bot.DummyAuthorisation{}
		log.Printf("No whitelist found starting with dummy auth")
	}
	if err != nil {
		log.Fatalf("Failed %s", err.Error())
	}
	log.Printf("Description file path used %s", *path)
	urlContainer, err := getDescriptionSourceFromFile(*path)
	if err != nil {
		log.Fatalf("Failed to get description source file %s", err.Error())
	}
	botInstance, err := bot.NewBot(client, *token, *urlContainer, auth)
	if err != nil {
		log.Fatalf("Failed tto create bot %s", err.Error())
	}
	err = botInstance.Listen(context.Background())
	if err != nil {
		log.Fatalf("Failed on listen %s", err.Error())
	}
}

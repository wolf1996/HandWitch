package main

import (
	"bufio"
	"log"
	"os"

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

func main() {
	token := os.Args[1]
	proxy := os.Args[2]
	path := os.Args[3]

	client, err := bot.GetClientWithProxy(proxy)
	if err != nil {
		log.Fatalf("Failed %s", err.Error())
	}
	urlContainer, err := getDescriptionSourceFromFile(path)
	if err != nil {
		log.Fatalf("Failed to get description source file %s", err.Error())
	}
	botInstance, err := bot.NewBot(client, token, *urlContainer)
	if err != nil {
		log.Fatalf("Failed tto create bot %s", err.Error())
	}
	err = botInstance.Listen()
	if err != nil {
		log.Fatalf("Failed on listen %s", err.Error())
	}
}

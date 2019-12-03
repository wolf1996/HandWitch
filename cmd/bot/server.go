package main

import (
	"log"
	"os"

	bot "github.com/wolf1996/HandWitch/pkg/bot"
)

func main() {
	token := os.Args[1]
	proxy := os.Args[2]

	client, err := bot.GetClientWithProxy(proxy)
	if err != nil {
		log.Fatalf("Failed %s", err.Error())
	}
	botInstance, err := bot.NewBot(client, token, nil)
	if err != nil {
		log.Fatalf("Failed tto create bot %s", err.Error())
	}
	err = botInstance.Listen()
	if err != nil {
		log.Fatalf("Failed on listen %s", err.Error())
	}
}

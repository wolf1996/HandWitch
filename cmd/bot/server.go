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
	botInstance.Listen()
}

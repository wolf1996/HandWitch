package bot

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wolf1996/HandWitch/pkg/core"
)

// Bot создаёт общий интерфейс для бота
type Bot struct {
	api *tgbotapi.BotAPI
	app core.URLProcessor
}

// NewBot создаёт новый инстанс бота
func NewBot(client *http.Client, token string, app core.URLProcessor) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPIWithClient(token, client)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return &Bot{
		api: bot,
		app: app,
	}, nil
}

func (b *Bot) executeMessageText(ctx context.Context, writer io.Writer, messageText string) error {
	if messageText == "" {
		return fmt.Errorf("Empty command")
	}
	rows := strings.Split(messageText, "\n")
	handName := rows[0]
	params := rows[1:]
	praramsMap := make(map[string]interface{})
	handProcessor, err := b.app.GetHand(handName)
	if err != nil {
		return err
	}
	for _, row := range params {
		splited := strings.Split(row, " ")
		paramName := splited[0]
		paramValueStr := splited[1]
		paramProcessor, err := handProcessor.GetParam(paramName)
		if err != nil {
			return err
		}
		value, err := paramProcessor.ParseFromString(paramValueStr)
		if err != nil {
			return err
		}
		praramsMap[paramName] = value
	}
	return handProcessor.Process(writer, praramsMap)
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	var resp bytes.Buffer
	err := b.executeMessageText(ctx, &resp, message.Text)
	if err != nil {
		errmsg := fmt.Sprintf("Error on processing message %s: %s", message.Text, err.Error())
		msg := tgbotapi.NewMessage(message.Chat.ID, errmsg)
		_, err = b.api.Send(msg)
		if err != nil {
			log.Printf("Error on sending message %s", err.Error())
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, resp.String())
	_, err = b.api.Send(msg)
	if err != nil {
		log.Printf("Error on sending message %s", err.Error())
	}
}

// Listen слушаем сообщения и отправляем ответ
func (b *Bot) Listen() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		log.Printf("Got message [%s] %s", update.Message.From.UserName, update.Message.Text)
		go b.handleMessage(context.Background(), update.Message)
	}
	return nil
}

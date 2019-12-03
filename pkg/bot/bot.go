package bot

import (
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wolf1996/HandWitch/pkg/core"
)

// Bot создаёт общий интерфейс для бота
type Bot struct {
	api *tgbotapi.BotAPI
	app *core.URLContrainer
}

// NewBot создаёт новый инстанс бота
func NewBot(client *http.Client, token string, app *core.URLContrainer) (*Bot, error) {
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

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		_, err = b.api.Send(msg)
		if err != nil {
			log.Printf("Error on sending message %s", err.Error())
		}
	}
	return nil
}

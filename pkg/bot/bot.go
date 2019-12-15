package bot

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

// Bot создаёт общий интерфейс для бота
type Bot struct {
	api  *tgbotapi.BotAPI
	app  core.URLProcessor
	auth Authorisation
}

//TODO: проверить каноничность

// NewBot создаёт новый инстанс бота
func NewBot(client *http.Client, token string, app core.URLProcessor, auth Authorisation) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPIWithClient(token, client)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Info("Authorized on account %s", bot.Self.UserName)
	return &Bot{
		api:  bot,
		app:  app,
		auth: auth,
	}, nil
}

func (b *Bot) processHand(ctx context.Context, writer io.Writer, messageArguments string) error {
	if messageArguments == "" {
		return fmt.Errorf("Empty arguments")
	}
	rows := strings.Split(messageArguments, "\n")
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
	return handProcessor.Process(ctx, writer, praramsMap)
}

func (b *Bot) helpHand(ctx context.Context, writer io.Writer, messageArguments string) error {
	if messageArguments == "" {
		return fmt.Errorf("Empty arguments")
	}
	handProcessor, err := b.app.GetHand(messageArguments)
	if err != nil {
		return err
	}
	return handProcessor.WriteHelp(writer)
}

func (b *Bot) executeMessage(ctx context.Context, writer io.Writer, message *tgbotapi.Message) error {
	switch message.Command() {
	case "process":
		return b.processHand(ctx, writer, message.CommandArguments())
	case "help":
		return b.helpHand(ctx, writer, message.CommandArguments())
	}
	return fmt.Errorf("Wrong comand %s", message.Command())
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	var resp bytes.Buffer
	err := b.executeMessage(ctx, &resp, message)
	if err != nil {
		errmsg := fmt.Sprintf("Error on processing message %s: %s", message.Text, err.Error())
		msg := tgbotapi.NewMessage(message.Chat.ID, errmsg)
		_, err = b.api.Send(msg)
		if err != nil {
			log.Errorf("Error on sending message %s", err.Error())
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, resp.String())
	_, err = b.api.Send(msg)
	if err != nil {
		log.Errorf("Error on sending message %s", err.Error())
	}
}

func (b *Bot) checkMessageAuth(message *tgbotapi.Message) (bool, error) {
	role, err := b.auth.GetRoleByLogin(message.From.UserName)
	if err != nil {
		return false, err
	}
	return role == User, nil
}

func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil { // ignore any non-Message Updates
		return
	}
	if !update.Message.IsCommand() {
		// ignore non-Command  Updates
		return
	}
	allowed, err := b.checkMessageAuth(update.Message)
	if err != nil {
		log.Errorf("Failed to check user role %s", err.Error())
		return
	}
	if !allowed {
		log.Warn("User %s has a \"Guest\" role, ignore", update.Message.From.UserName)
		return
	}
	log.Debug("Got message [%s] %s", update.Message.From.UserName, update.Message.Text)
	go b.handleMessage(ctx, update.Message)
}

// Listen слушаем сообщения и отправляем ответ
func (b *Bot) Listen(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	for {
		select {
		case up := <-updates:
			b.processUpdate(ctx, up)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

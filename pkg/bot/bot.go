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
	log.Infof("Authorized on account %s", bot.Self.UserName)
	return &Bot{
		api:  bot,
		app:  app,
		auth: auth,
	}, nil
}

func (b *Bot) processHand(ctx context.Context, writer io.Writer, messageArguments string, logger *log.Entry) error {
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
	return handProcessor.Process(ctx, writer, praramsMap, logger)
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

func (b *Bot) executeMessage(ctx context.Context, writer io.Writer, message *tgbotapi.Message, logger *log.Entry) error {
	switch message.Command() {
	case "process":
		logger.Debugf("found \"process\" command")
		return b.processHand(ctx, writer, message.CommandArguments(), logger)
	case "help":
		logger.Debugf("found \"help\" command")
		return b.helpHand(ctx, writer, message.CommandArguments())
	}
	return fmt.Errorf("Wrong comand %s", message.Command())
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	var resp bytes.Buffer
	err := b.executeMessage(ctx, &resp, message, logger)
	if err != nil {
		errmsg := fmt.Sprintf("Error on processing message %s: %s", message.Text, err.Error())
		msg := tgbotapi.NewMessage(message.Chat.ID, errmsg)
		_, err = b.api.Send(msg)
		if err != nil {
			logger.Errorf("Error on sending message %s", err.Error())
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, resp.String())
	_, err = b.api.Send(msg)
	if err != nil {
		logger.Errorf("Error on sending message %s", err.Error())
	}
}

func (b *Bot) checkMessageAuth(message *tgbotapi.Message) (bool, error) {
	role, err := b.auth.GetRoleByLogin(message.From.UserName)
	if err != nil {
		return false, err
	}
	return role == User, nil
}

func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update, logger *log.Entry) {
	if update.Message == nil { // ignore any non-Message Updates
		logger.Debugf("No message in update skipping")
		return
	}
	if !update.Message.IsCommand() {
		// ignore non-Command  Updates
		logger.Debugf("No command in message update skipping")
		return
	}
	allowed, err := b.checkMessageAuth(update.Message)
	if err != nil {
		logger.Errorf("Failed to check user role %s", err.Error())
		return
	}
	if !allowed {
		logger.Warnf("User %s has a \"Guest\" role, ignore", update.Message.From.UserName)
		return
	}
	logger.Debugf("Got message [%s] %s", update.Message.From.UserName, update.Message.Text)
	messageLogger := logger.WithFields(log.Fields{
		"user_login":   update.Message.From.UserName,
		"message_text": update.Message.Text,
	})
	go b.handleMessage(ctx, update.Message, messageLogger)
}

// Listen слушаем сообщения и отправляем ответ
func (b *Bot) Listen(ctx context.Context, logger *log.Logger) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	for {
		select {
		case up := <-updates:
			b.processUpdate(ctx, up, log.NewEntry(logger))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

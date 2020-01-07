package bot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type (
	taskKey struct {
		ChatId int64
		UserId string
	}

	inProgresTask = map[taskKey]chan *tgbotapi.Message
)

func getTaskKeyFromMessage(message *tgbotapi.Message) (taskKey, error) {
	return taskKey{
		ChatId: message.Chat.ID,
		UserId: message.From.UserName,
	}, nil
}

// Bot создаёт общий интерфейс для бота
type Bot struct {
	api        *tgbotapi.BotAPI
	app        core.URLProcessor
	auth       Authorisation
	formating  string
	processing inProgresTask
}

//TODO: проверить каноничность

// NewBot создаёт новый инстанс бота
func NewBot(client *http.Client, token string, app core.URLProcessor, auth Authorisation, formating string) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPIWithClient(token, client)
	if err != nil {
		return nil, fmt.Errorf("failed create new bot api with client %w", err)
	}
	log.Infof("Authorized on account %s", bot.Self.UserName)
	normalizedMessageMode, err := normilizeMessageMode(formating)
	if err != nil {
		return nil, fmt.Errorf("Invalid formating %w", err)
	}
	return &Bot{
		api:        bot,
		app:        app,
		auth:       auth,
		formating:  normalizedMessageMode,
		processing: make(inProgresTask),
	}, nil
}

func (b *Bot) getHandName(messageArguments string) (string, error) {
	rows := strings.Split(messageArguments, "\n")
	if len(rows) < 1 {
		return "", fmt.Errorf("Failed to get hand name: empty arguments")
	}
	handName := rows[0]
	return strings.TrimSpace(handName), nil
}

// TODO: подумать о каноничности такого подхода
// для разных операций за формирование конечного сообщения отвечают различные уровни архитектуры
func (b *Bot) getHandParams(handProcessor core.HandProcessor, messageArguments string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	rows := strings.Split(messageArguments, "\n")
	if len(rows) < 1 {
		return result, fmt.Errorf("Failed to get hand name: empty arguments")
	}
	hands := rows[1:]
	for _, row := range hands {
		splited := strings.Fields(row)
		//TODO: сделать более адекватный парсинг, с возможностью пробелов в значениях
		if len(splited) != 2 {
			return result, fmt.Errorf("Failed to parse param row %s, splited on %d args instead 2", row, len(splited))
		}
		paramName := splited[0]
		paramValueStr := splited[1]
		paramProcessor, err := handProcessor.GetParam(paramName)
		if err != nil {
			return result, err
		}
		value, err := paramProcessor.ParseFromString(paramValueStr)
		if err != nil {
			return result, err
		}
		result[paramName] = value
	}
	return result, nil
}

func (b *Bot) processHand(ctx context.Context, writer io.Writer, messageArguments string, message *tgbotapi.Message, input chan *tgbotapi.Message, logger *log.Entry) error {
	if messageArguments == "" {
		return errors.New("Empty arguments")
	}
	handName, err := b.getHandName(messageArguments)
	if err != nil {
		return err
	}
	handProcessor, err := b.app.GetHand(handName)
	if err != nil {
		return err
	}
	params, err := b.getHandParams(handProcessor, messageArguments)
	if err != nil {
		return err
	}
	return handProcessor.Process(ctx, writer, params, logger)
}

func (b *Bot) helpHand(ctx context.Context, writer io.Writer, messageArguments string, logger *log.Entry) error {
	if messageArguments == "" {
		return fmt.Errorf("Empty arguments")
	}
	handProcessor, err := b.app.GetHand(messageArguments)
	if err != nil {
		return err
	}
	return handProcessor.WriteHelp(writer)
}

func (b *Bot) executeMessage(ctx context.Context, writer io.Writer, message *tgbotapi.Message, input chan *tgbotapi.Message, logger *log.Entry) error {
	switch message.Command() {
	case "process":
		logger.Debug("found \"process\" command")
		return b.processHand(ctx, writer, message.CommandArguments(), message, input, logger)
	case "help":
		logger.Debug("found \"help\" command")
		return b.helpHand(ctx, writer, message.CommandArguments(), logger)
	}
	return fmt.Errorf("Wrong comand %s", message.Command())
}

func normilizeMessageMode(raw string) (string, error) {
	switch strings.ToLower(raw) {
	case "markdown":
		return tgbotapi.ModeMarkdown, nil
	case "html":
		return tgbotapi.ModeHTML, nil
	}
	return "", fmt.Errorf("Invalid message mode %s", raw)
}

func (b *Bot) newHandleMessage(ctx context.Context, message *tgbotapi.Message, input chan *tgbotapi.Message, logger *log.Entry) {
	var resp bytes.Buffer
	err := b.executeMessage(ctx, &resp, message, input, logger)
	if err != nil {
		errmsg := fmt.Sprintf("Error on processing message %s: %s", message.Text, err.Error())
		msg := tgbotapi.NewMessage(message.Chat.ID, errmsg)
		_, err = b.api.Send(msg)
		if err != nil {
			logger.Errorf("Error on sending message %s", err.Error())
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, resp.String())
	if b.formating != "" {
		logger.Debugf("setting formating: %s", b.formating)
		msg.ParseMode = b.formating
	}
	_, err = b.api.Send(msg)
	if err != nil {
		logger.Errorf("Error on sending message %s:\n message text:\n %s", err.Error(), msg.Text)
	}
}

func (b *Bot) checkMessageAuth(message *tgbotapi.Message) (bool, error) {
	role, err := b.auth.GetRoleByLogin(message.From.UserName)
	if err != nil {
		return false, err
	}
	return role == User, nil
}

func (b *Bot) initMessageHandle(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) chan *tgbotapi.Message {
	input := make(chan *tgbotapi.Message)
	go b.newHandleMessage(ctx, message, input, logger)
	return input
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) error {
	key, err := getTaskKeyFromMessage(message)
	if err != nil {
		return fmt.Errorf("Failed to get task key %w", err)
	}
	taskChan, ok := b.processing[key]
	if !ok {
		// создаём хэндлер этого задания
		input := b.initMessageHandle(ctx, message, logger)
		b.processing[key] = input
	} else {
		go func() {
			select {
			case taskChan <- message:
				{
				}
			case <-ctx.Done():
				{
					logger.Errorf("Failed to send message to task, canceled")
				}
			}
		}()
	}
	return nil
}

// TODO: думаю таки будет иметь смысл сделать тут возврат ошибки
func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update, logger *log.Entry) error {
	if update.Message == nil { // ignore any non-Message Updates
		logger.Debugf("No message in update skipping")
		return nil
	}
	if !update.Message.IsCommand() {
		// ignore non-Command  Updates
		logger.Debugf("No command in message update skipping")
		return nil
	}
	allowed, err := b.checkMessageAuth(update.Message)
	if err != nil {
		logger.Errorf("Failed to check user role %s", err.Error())
		return nil
	}
	if !allowed {
		logger.Warnf("User %s has a \"Guest\" role, ignore", update.Message.From.UserName)
		return nil
	}
	logger.Debugf("Got message [%s] %s", update.Message.From.UserName, update.Message.Text)
	messageLogger := logger.WithFields(log.Fields{
		"user_login":   update.Message.From.UserName,
		"message_text": update.Message.Text,
	})
	return b.handleMessage(ctx, update.Message, messageLogger)
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

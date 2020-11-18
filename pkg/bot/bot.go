package bot

import (
	"fmt"
	"net/http"
	"strings"

	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type (
	taskKey struct {
		ChatID int64  // Идентификатор чата
		UserID string // Идентификатор пользователя
	}
	inProgresTask = map[taskKey]messagesChan
)

type comand interface {
	Process(string) error
}

type comandFabric = func(ctx context.Context, urlProcessor core.URLProcessor, tg telegram, log *log.Entry) comand

func getTaskKeyFromMessage(message *tgbotapi.Message) (taskKey, error) {
	return taskKey{
		ChatID: message.Chat.ID,
		UserID: message.From.UserName,
	}, nil
}

type HookConfig struct {
	URLPath string
	Host    string
	Port    string
	Cert    string
	Key     string
}

// Bot создаёт общий интерфейс для бота
type Bot struct {
	api        *tgbotapi.BotAPI
	app        core.URLProcessor
	auth       Authorisation
	formating  string
	processing inProgresTask
	cmds       map[string]comandFabric
	hookCfg    *HookConfig
}

// NewBot создаёт новый инстанс бота
func NewBot(client *http.Client, token string, app core.URLProcessor, auth Authorisation, formating string, hookCfg *HookConfig) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPIWithClient(token, client)
	if err != nil {
		return nil, fmt.Errorf("failed create new bot api with client %w", err)
	}
	log.Infof("Authorized on account %s", bot.Self.UserName)
	normalizedMessageMode, err := normilizeMessageMode(formating)
	if err != nil {
		return nil, fmt.Errorf("Invalid formating %w", err)
	}
	cmds := make(map[string]comandFabric)
	cmds["process"] = newProcessCommand
	cmds["help"] = newHelpCommand
	cmds["start"] = newStartCommand
	return &Bot{
		api:        bot,
		app:        app,
		auth:       auth,
		formating:  normalizedMessageMode,
		processing: make(inProgresTask),
		cmds:       cmds,
		hookCfg:    hookCfg,
	}, nil
}

func (b *Bot) processCmd(ctx context.Context, messageArguments string, message *tgbotapi.Message, input messagesChan, fabric comandFabric, logger *log.Entry) error {
	tg := newWrapper(input, b.api, message, b.formating, logger)
	command := fabric(ctx, b.app, tg, logger)
	return command.Process(messageArguments)
}

func (b *Bot) executeMessage(ctx context.Context, message *tgbotapi.Message, input messagesChan, logger *log.Entry) error {
	fabric, ok := b.cmds[message.Command()]
	if !ok {
		return fmt.Errorf("Wrong comand %s", message.Command())
	}
	return b.processCmd(ctx, message.CommandArguments(), message, input, fabric, logger)
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

func (b *Bot) newHandleMessage(ctx context.Context, message *tgbotapi.Message, input messagesChan, logger *log.Entry) {
	defer func() {
		key, _ := getTaskKeyFromMessage(message)
		delete(b.processing, key)
	}()
	err := b.executeMessage(ctx, message, input, logger)
	if err != nil {
		errmsg := fmt.Sprintf("Error on processing message %s: %s", message.Text, err.Error())
		msg := tgbotapi.NewMessage(message.Chat.ID, errmsg)
		_, err = b.api.Send(msg)
		if err != nil {
			logger.Errorf("Error on sending message %s", err.Error())
		}
	}
}

func (b *Bot) checkMessageAuth(message *tgbotapi.Message) (bool, error) {
	role, err := b.auth.GetRoleByLogin(message.From.UserName)
	if err != nil {
		return false, err
	}
	return role == User, nil
}

func (b *Bot) initMessageHandle(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) messagesChan {
	proxyInput := make(messagesChan)
	input := make(messagesChan)
	go func() {
		buffer := make([]*tgbotapi.Message, 0)
		getChan := func() messagesChan {
			if len(buffer) == 0 {
				return nil
			}
			return input
		}
		getVal := func() *tgbotapi.Message {
			if len(buffer) == 0 {
				return nil
			}
			return buffer[0]
		}
	loop:
		for {
			select {
			case msg, ok := <-proxyInput:
				{
					if !ok {
						break loop
					}
					logger.Debug("Got message to proxy")
					buffer = append(buffer, msg)
				}
			case <-ctx.Done():
				{
					logger.Errorf("Failed to send message to task, canceled")
					break loop
				}
			case getChan() <- getVal():
				{
					buffer = buffer[:len(buffer)-1]
					logger.Debug("Send message from proxy")
				}
			}
		}
	}()
	go b.newHandleMessage(ctx, message, input, logger)
	return proxyInput
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
		select {
		case taskChan <- message:
			{
			}
		case <-ctx.Done():
			{
				logger.Errorf("Failed to send message to task, canceled")
			}
		}
	}
	return nil
}

// TODO: думаю таки будет иметь смысл сделать тут возврат ошибки
func (b *Bot) processUpdate(ctx context.Context, update tgbotapi.Update, logger *log.Entry) error {
	if update.Message == nil { // ignore any non-Message Updates
		logger.Debugf("No message in update skipping")
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

func (b *Bot) activeModUpdatesChan() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return b.api.GetUpdatesChan(u)
}

func (b *Bot) hookModUpdatesChan(logger *log.Logger) (tgbotapi.UpdatesChannel, error) {
	if b.hookCfg == nil {
		return nil, fmt.Errorf("No config for hook provided")
	}
	_, err := b.api.SetWebhook(tgbotapi.NewWebhookWithCert(b.hookCfg.URLPath, b.hookCfg.Cert))
	if err != nil {
		return nil, fmt.Errorf("Failed to create webhook: %w", err)
	}
	info, err := b.api.GetWebhookInfo()
	if err != nil {
		return nil, fmt.Errorf("Failed to get webhook info: %w", err)
	}
	if info.LastErrorDate != 0 {
		return nil, fmt.Errorf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	updates := b.api.ListenForWebhook("/" + b.hookCfg.URLPath)
	serveFunc := func() {
		listenHost := "0.0.0.0"
		if b.hookCfg.Port != "" {
			listenHost += ":" + b.hookCfg.Port
		}
		err := http.ListenAndServeTLS(listenHost, b.hookCfg.Cert, b.hookCfg.Key, nil)
		logger.Fatalf("failed to start bot: %w", err)
	}
	go serveFunc()
	return updates, nil
}

func (b *Bot) getUpdatesChan(logger *log.Logger) (tgbotapi.UpdatesChannel, error) {
	if b.hookCfg != nil {
		return b.hookModUpdatesChan(logger)
	}
	return b.activeModUpdatesChan()
}

// Listen слушаем сообщения и отправляем ответ
func (b *Bot) Listen(ctx context.Context, logger *log.Logger) error {
	updates, err := b.getUpdatesChan(logger)
	if err != nil {
		return err
	}
	for {
		select {
		case up := <-updates:
			err = b.processUpdate(ctx, up, log.NewEntry(logger))
			if err != nil {
				logger.Errorf("Failed to process update %s", err.Error())
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

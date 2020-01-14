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

	messages = chan *tgbotapi.Message

	inProgresTask = map[taskKey]messages
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

func (b *Bot) parseParamRow(handProcessor core.HandProcessor, messageRow string) (string, interface{}, error) {
	splited := strings.Fields(messageRow)
	//TODO: сделать более адекватный парсинг, с возможностью пробелов в значениях
	if len(splited) != 2 {
		return "", nil, fmt.Errorf("Failed to parse param row %s, splited on %d args instead 2", messageRow, len(splited))
	}
	paramName := splited[0]
	paramValueStr := splited[1]
	paramProcessor, err := handProcessor.GetParam(paramName)
	if err != nil {
		return "", nil, err
	}
	value, err := paramProcessor.ParseFromString(paramValueStr)
	if err != nil {
		return "", nil, err
	}
	return paramName, value, nil
}

func (b *Bot) buildKeyboard(missingParams map[string]core.ParamProcessor) tgbotapi.ReplyKeyboardMarkup {
	buttons := make([]tgbotapi.KeyboardButton, 0)
	for paramName := range missingParams {
		buttons = append(buttons, tgbotapi.NewKeyboardButton(paramName))
	}
	return tgbotapi.NewReplyKeyboard(buttons)
}

func (b *Bot) parseAll(input *tgbotapi.Message, handProcessor core.HandProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor) error {
	var err error
PARSE_PARAMS:
	for _, row := range strings.Split(input.Text, "\n") {
		name, val, err := b.parseParamRow(handProcessor, row)
		if err != nil {
			msg := tgbotapi.NewMessage(input.Chat.ID, fmt.Sprintf("Failed to parse param: \"%s\" %s", name, err.Error()))
			_, err = b.api.Send(msg)
			if err != nil {
				return fmt.Errorf("Failed to send error message to user %w", err)
			}
			continue PARSE_PARAMS
		}
		delete(missingParams, name)
		params[name] = val
	}
	if err != nil {
		msg := tgbotapi.NewMessage(input.Chat.ID, fmt.Sprintf("Failed to parse param: \"%s\"", err.Error()))
		_, err := b.api.Send(msg)
		if err != nil {
			return fmt.Errorf("Failed to send error message to user %w", err)
		}
	}
	return nil
}

func (b *Bot) handleSingleParam(ctx context.Context, paramProcessor core.ParamProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor, message *tgbotapi.Message, input messages) error {
	// TODO: сделать более подробное описание в сообщении, возможно - хэлп
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Input value for param: \"%s\"", paramProcessor.GetInfo().Name))
	_, err := b.api.Send(msg)
	if err != nil {
		//TODO проверить обработку ошибок и ретраи
		return fmt.Errorf("failed request missing parameters from user %w", err)
	}
LOOP:
	for {
		select {
		case inp := <-input:
			value, err := paramProcessor.ParseFromString(inp.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(inp.Chat.ID, fmt.Sprintf("Failed to parse param:  %s", err.Error()))
				_, err = b.api.Send(msg)
				if err != nil {
					return fmt.Errorf("Failed to send error message to user %w", err)
				}
				continue LOOP
			}
			delete(missingParams, paramProcessor.GetInfo().Name)
			params[paramProcessor.GetInfo().Name] = value
			//TODO: Переделать, обязательно! выглядит и читается ужасно
			break LOOP
		case <-ctx.Done():
			{
				return fmt.Errorf("Context canceled %w", err)
			}
		}
	}
	return nil
}

func (b *Bot) inqueryParams(ctx context.Context, handProcessor core.HandProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor, message *tgbotapi.Message, input messages) error {
	for len(missingParams) != 0 {
		var paramsNames []string
		for _, param := range missingParams {
			paramsNames = append(paramsNames, param.GetInfo().Name)
		}
		missingParamsList := strings.Join(paramsNames, "\", \"")
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Missed params: \"%s\"", missingParamsList))
		keyboard := b.buildKeyboard(missingParams)
		msg.ReplyMarkup = keyboard
		_, err := b.api.Send(msg)
		if err != nil {
			//TODO проверить обработку ошибок и ретраи
			return fmt.Errorf("failed request missing parameters from user %w", err)
		}
		select {
		case inp := <-input:
			{
				// TODO Переделать при рефакторинге; добавить обработку дополнительных и некорректных вариантов
				if handle, ok := missingParams[inp.Text]; ok {
					err := b.handleSingleParam(ctx, handle, params, missingParams, message, input)
					if err != nil {
						return err
					}
				} else {
					err := b.parseAll(inp, handProcessor, params, missingParams)
					if err != nil {
						return err
					}
				}
			}
		case <-ctx.Done():
			{
				return fmt.Errorf("Context canceled %w", err)
			}
		}
	}
	return nil
}

// TODO: подумать о каноничности такого подхода
// для разных операций за формирование конечного сообщения отвечают различные уровни архитектуры
func (b *Bot) getHandParams(ctx context.Context, handProcessor core.HandProcessor, messageArguments string, message *tgbotapi.Message, input messages) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// TODO: переделать это на reader и построчное чтение?
	for _, row := range strings.Split(message.Text, "\n")[1:] {
		name, val, err := b.parseParamRow(handProcessor, row)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Failed to parse param: \"%s\" %s", name, err.Error()))
			_, err = b.api.Send(msg)
			if err != nil {
				return params, fmt.Errorf("Failed to send error message to user %w", err)
			}
		}
		params[name] = val
	}

	requiredParams, err := handProcessor.GetRequiredParams()
	if err != nil {
		return params, fmt.Errorf("Failed to get hand required parameters: %w", err)
	}
	if err != nil {
		return params, err
	}
	getMissingParams := func() map[string]core.ParamProcessor {
		missingParams := make(map[string]core.ParamProcessor)
		for _, param := range requiredParams {
			if _, ok := params[param.GetInfo().Name]; !ok {
				missingParams[param.GetInfo().Name] = param
			}
		}
		return missingParams
	}

	missingParams := getMissingParams()
	err = b.inqueryParams(ctx, handProcessor, params, missingParams, message, input)
	if err != nil {
		return params, fmt.Errorf("Failed to inquery params %w", err)
	}

	return params, nil
}

func (b *Bot) processHand(ctx context.Context, writer io.Writer, messageArguments string, message *tgbotapi.Message, input messages, logger *log.Entry) error {
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
	params, err := b.getHandParams(ctx, handProcessor, messageArguments, message, input)
	if err != nil {
		return err
	}
	logger.Debugf("Got parameters %v", params)
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

func (b *Bot) executeMessage(ctx context.Context, writer io.Writer, message *tgbotapi.Message, input messages, logger *log.Entry) error {
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

func (b *Bot) newHandleMessage(ctx context.Context, message *tgbotapi.Message, input messages, logger *log.Entry) {
	defer func() {
		// Todo: поправить обработку возможной ошибки
		key, _ := getTaskKeyFromMessage(message)
		delete(b.processing, key)
	}()
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

func (b *Bot) initMessageHandle(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) messages {
	proxyInput := make(messages)
	input := make(messages)
	go func() {
		buffer := make([]*tgbotapi.Message, 0)
		getChan := func() messages {
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
			err = b.processUpdate(ctx, up, log.NewEntry(logger))
			if err != nil {
				logger.Errorf("Failed to process update %s", err.Error())
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

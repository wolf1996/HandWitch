package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type (
	messagesChan = chan *tgbotapi.Message
	message      = string
)

type telegram interface {
	Get(ctx context.Context) (message, error)
	Send(ctx context.Context, msg string) error
	RequestParams(map[string]core.ParamProcessor) error
}

type wrapper struct {
	input     messagesChan
	api       *tgbotapi.BotAPI
	chat      *tgbotapi.Chat
	formating string
	logger    *log.Entry
}

func newWrapper(input messagesChan, api *tgbotapi.BotAPI, msg *tgbotapi.Message, formating string, logger *log.Entry) telegram {
	return &wrapper{
		input:     input,
		api:       api,
		chat:      msg.Chat,
		formating: formating,
		logger:    logger,
	}
}

func (wp *wrapper) Get(ctx context.Context) (message, error) {
	wp.logger.Debug("Waiting for message")
	select {
	case inp := <-wp.input:
		{
			return inp.Text, nil
		}
	case <-ctx.Done():
		{
			return "", fmt.Errorf("Context canceled %w", ctx.Err())
		}
	}

}

func (wp *wrapper) Send(ctx context.Context, msgTxt string) error {
	msg := tgbotapi.NewMessage(wp.chat.ID, msgTxt)
	log.Debugf("Sending message %s", msgTxt)
	if wp.formating != "" {
		wp.logger.Debugf("setting formating: %s", wp.formating)
		msg.ParseMode = wp.formating
	}

	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)

	_, err := wp.api.Send(msg)
	if err != nil {
		wp.logger.Errorf("Error on sending message %s:\n message text:\n %s", err.Error(), msg.Text)
		return err
	}
	return nil
}

func buildKeyboard(missingParams map[string]core.ParamProcessor) tgbotapi.ReplyKeyboardMarkup {
	buttons := make([][]tgbotapi.KeyboardButton, 0)
	for paramName := range missingParams {
		paramButton := tgbotapi.NewKeyboardButton(paramName)
		helpButton := tgbotapi.NewKeyboardButton(fmt.Sprintf("🤖help %s", paramName))
		buttons = append(buttons, []tgbotapi.KeyboardButton{paramButton, helpButton})
	}
	return tgbotapi.NewReplyKeyboard(buttons...)
}

func (wp *wrapper) RequestParams(missingParams map[string]core.ParamProcessor) error {
	var paramsNames []string
	for _, param := range missingParams {
		paramsNames = append(paramsNames, param.GetInfo().Name)
	}
	missingParamsList := strings.Join(paramsNames, "\", \"")
	msg := tgbotapi.NewMessage(wp.chat.ID, fmt.Sprintf("Missed params: \"%s\"", missingParamsList))
	keyboard := buildKeyboard(missingParams)
	msg.ReplyMarkup = keyboard
	_, err := wp.api.Send(msg)
	if err != nil {
		//TODO проверить обработку ошибок и ретраи
		return fmt.Errorf("failed request missing parameters from user %w", err)
	}
	return nil
}

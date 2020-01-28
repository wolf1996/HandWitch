package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

// ExtraButton comands to add special buttons to messages
type ExtraButton int

const (
	// CancelButton button
	CancelButton = iota
	// OkButton button
	OkButton
	// HelpButton button
	HelpButton
)

const (
	// ParamHelpButtonContent prefix of string in parameter help button
	ParamHelpButtonContent = "🤖 help"
	// HandHelpButtonContent data of string in hand help button
	HandHelpButtonContent = "🤖 hand help"
	// OkButtonContent data of string in Ok button
	OkButtonContent = "🤖 Start!"
	// CancelButtonContent data of string in Cancel button
	CancelButtonContent = "🤖 cancel"
)

type (
	messagesChan = chan *tgbotapi.Message
	message      = string
)

type telegram interface {
	Get(ctx context.Context) (message, error)
	Send(ctx context.Context, msg string) error
	RequestParams(missingParams map[string]core.ParamProcessor, params map[string]core.ParamProcessor, values map[string]interface{}, buttons []ExtraButton) error
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

func buildKeyboard(missingParams map[string]core.ParamProcessor, buttonsDescriptions []ExtraButton) ([][]tgbotapi.KeyboardButton, error) {
	buttons := make([][]tgbotapi.KeyboardButton, 0)
	for paramName := range missingParams {
		paramButton := tgbotapi.NewKeyboardButton(paramName)
		helpButton := tgbotapi.NewKeyboardButton(fmt.Sprintf("%s %s", ParamHelpButtonContent, paramName))
		buttons = append(buttons, []tgbotapi.KeyboardButton{paramButton, helpButton})
	}
	additionalButtons, err := getCustomButtons(buttonsDescriptions)
	if err != nil {
		return buttons, err
	}
	buttons = append(buttons, additionalButtons)
	return buttons, nil
}

func getCustomButton(buttonDescription ExtraButton) (*tgbotapi.KeyboardButton, error) {
	switch buttonDescription {
	case CancelButton:
		{
			button := tgbotapi.NewKeyboardButton(CancelButtonContent)
			return &button, nil
		}
	case OkButton:
		{
			button := tgbotapi.NewKeyboardButton(OkButtonContent)
			return &button, nil
		}
	case HelpButton:
		{
			button := tgbotapi.NewKeyboardButton(HandHelpButtonContent)
			return &button, nil
		}
	}
	return nil, fmt.Errorf("Failed to get custom button %d", buttonDescription)
}

func getCustomButtons(buttons []ExtraButton) ([]tgbotapi.KeyboardButton, error) {
	result := make([]tgbotapi.KeyboardButton, 0)

	for _, buttonDescr := range buttons {
		button, err := getCustomButton(buttonDescr)
		if err != nil {
			return result, err
		}
		result = append(result, *button)
	}
	return result, nil
}

func addMissing(writer *strings.Builder, missingParams map[string]core.ParamProcessor) {
	if len(missingParams) == 0 {
		return
	}
	var paramsNames []string
	for _, param := range missingParams {
		paramsNames = append(paramsNames, param.GetInfo().Name)
	}
	missingParamsList := strings.Join(paramsNames, "\", \"")
	writer.WriteString(fmt.Sprintf("Missed params: \"%s\" \n", missingParamsList))
}

func addValues(writer *strings.Builder, params map[string]core.ParamProcessor, values map[string]interface{}) {
	if len(params) == 0 {
		return
	}
	writer.WriteString("Current values: \n")
	for name, val := range values {
		writer.WriteString(fmt.Sprintf("%s %v \n", name, val))
	}
}

func (wp *wrapper) RequestParams(missingParams map[string]core.ParamProcessor, params map[string]core.ParamProcessor, values map[string]interface{}, buttons []ExtraButton) error {
	var rspBuilder strings.Builder
	addValues(&rspBuilder, params, values)
	addMissing(&rspBuilder, missingParams)
	msg := tgbotapi.NewMessage(wp.chat.ID, rspBuilder.String())
	keyboardRows, err := buildKeyboard(params, buttons)
	if err != nil {
		return fmt.Errorf("Failed while keyboard build %w", err)
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboardRows...)
	_, err = wp.api.Send(msg)
	if err != nil {
		//TODO проверить обработку ошибок и ретраи
		return fmt.Errorf("failed request missing parameters from user %w", err)
	}
	return nil
}

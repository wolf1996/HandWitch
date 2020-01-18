package bot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type processCommand struct {
	ctx      context.Context
	input    messagesChan
	api      *tgbotapi.BotAPI
	handProc core.HandProcessor
	message  *tgbotapi.Message
	log      *log.Entry
}

func NewProcessCommand(ctx context.Context, input messagesChan, api *tgbotapi.BotAPI, handProc core.HandProcessor, message *tgbotapi.Message, log *log.Entry) processCommand {
	return processCommand{
		ctx:      ctx,
		input:    input,
		api:      api,
		handProc: handProc,
		message:  message,
		log:      log,
	}
}

func buildKeyboard(missingParams map[string]core.ParamProcessor) tgbotapi.ReplyKeyboardMarkup {
	buttons := make([]tgbotapi.KeyboardButton, 0)
	for paramName := range missingParams {
		buttons = append(buttons, tgbotapi.NewKeyboardButton(paramName))
	}
	return tgbotapi.NewReplyKeyboard(buttons)
}

func parseParamRow(handProcessor core.HandProcessor, messageRow string) (string, interface{}, error) {
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

func (b *processCommand) handleSingleParam(ctx context.Context, paramProcessor core.ParamProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor, message *tgbotapi.Message, input messagesChan) error {
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

// TODO: подумать о каноничности такого подхода
// для разных операций за формирование конечного сообщения отвечают различные уровни архитектуры
func (b *processCommand) getHandParams(ctx context.Context, handProcessor core.HandProcessor, messageArguments string, message *tgbotapi.Message, input messagesChan) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// TODO: переделать это на reader и построчное чтение?
	for _, row := range strings.Split(message.Text, "\n")[1:] {
		name, val, err := parseParamRow(handProcessor, row)
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

func (b *processCommand) inqueryParams(ctx context.Context, handProcessor core.HandProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor, message *tgbotapi.Message, input messagesChan) error {
	for len(missingParams) != 0 {
		var paramsNames []string
		for _, param := range missingParams {
			paramsNames = append(paramsNames, param.GetInfo().Name)
		}
		missingParamsList := strings.Join(paramsNames, "\", \"")
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Missed params: \"%s\"", missingParamsList))
		keyboard := buildKeyboard(missingParams)
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

func (b *processCommand) parseAll(input *tgbotapi.Message, handProcessor core.HandProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor) error {
	var err error
PARSE_PARAMS:
	for _, row := range strings.Split(input.Text, "\n") {
		name, val, err := parseParamRow(handProcessor, row)
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

func (proc *processCommand) Process(messageArguments string, writer io.Writer) error {
	if messageArguments == "" {
		return errors.New("Empty arguments")
	}
	params, err := proc.getHandParams(proc.ctx, proc.handProc, messageArguments, proc.message, proc.input)
	if err != nil {
		return err
	}
	proc.log.Debugf("Got parameters %v", params)
	proc.handProc.Process(proc.ctx, writer, params, proc.log)
	return nil
}

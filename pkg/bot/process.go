package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wolf1996/HandWitch/pkg/core"
)

type processCommand struct {
	ctx      context.Context
	tg       telegram
	handProc core.HandProcessor
	log      *log.Entry
}

func NewProcessCommand(ctx context.Context, handProc core.HandProcessor, tg telegram, log *log.Entry) comand {
	return &processCommand{
		ctx:      ctx,
		tg:       tg,
		handProc: handProc,
		log:      log,
	}
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

func (b *processCommand) handleSingleParam(ctx context.Context, paramProcessor core.ParamProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor) error {
	// TODO: сделать более подробное описание в сообщении, возможно - хэлп
	err := b.tg.Send(ctx, fmt.Sprintf("Input value for param: \"%s\"", paramProcessor.GetInfo().Name))
	if err != nil {
		//TODO проверить обработку ошибок и ретраи
		return fmt.Errorf("failed request missing parameters from user %w", err)
	}
LOOP:
	for {
		inp, err := b.tg.Get(ctx)
		if err != nil {
			continue LOOP
		}
		value, err := paramProcessor.ParseFromString(inp)
		if err != nil {
			err = b.tg.Send(ctx, fmt.Sprintf("Failed to parse param:  %s", err.Error()))
			if err != nil {
				return fmt.Errorf("Failed to send error message to user %w", err)
			}
			continue LOOP
		}
		delete(missingParams, paramProcessor.GetInfo().Name)
		params[paramProcessor.GetInfo().Name] = value
		break
	}
	return nil
}

// TODO: подумать о каноничности такого подхода
// для разных операций за формирование конечного сообщения отвечают различные уровни архитектуры
func (b *processCommand) getHandParams(ctx context.Context, handProcessor core.HandProcessor, messageArguments string) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// TODO: переделать это на reader и построчное чтение?
	for _, row := range strings.Split(messageArguments, "\n")[1:] {
		name, val, err := parseParamRow(handProcessor, row)
		if err != nil {
			err = b.tg.Send(ctx, fmt.Sprintf("Failed to parse param: \"%s\" %s", name, err.Error()))
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
	err = b.inqueryParams(ctx, handProcessor, params, missingParams)
	if err != nil {
		return params, fmt.Errorf("Failed to inquery params %w", err)
	}

	return params, nil
}

func (b *processCommand) inqueryParams(ctx context.Context, handProcessor core.HandProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor) error {
	for len(missingParams) != 0 {
		err := b.tg.RequestParams(missingParams)
		if err != nil {
			//TODO проверить обработку ошибок и ретраи
			return fmt.Errorf("failed request missing parameters from user %w", err)
		}
		txt, err := b.tg.Get(ctx)
		if err != nil {
			return err
		}
		if handle, ok := missingParams[txt]; ok {
			err := b.handleSingleParam(ctx, handle, params, missingParams)
			if err != nil {
				return err
			}
		} else {
			err := b.parseAll(txt, handProcessor, params, missingParams)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *processCommand) parseAll(input string, handProcessor core.HandProcessor, params map[string]interface{}, missingParams map[string]core.ParamProcessor) error {
	var err error
PARSE_PARAMS:
	for _, row := range strings.Split(input, "\n") {
		name, val, err := parseParamRow(handProcessor, row)
		if err != nil {
			err = b.tg.Send(b.ctx, fmt.Sprintf("Failed to parse param: \"%s\" %s", name, err.Error()))
			if err != nil {
				return fmt.Errorf("Failed to send error message to user %w", err)
			}
			continue PARSE_PARAMS
		}
		delete(missingParams, name)
		params[name] = val
	}
	if err != nil {
		err = b.tg.Send(b.ctx, fmt.Sprintf("Failed to parse param: \"%s\"", err.Error()))
		if err != nil {
			return fmt.Errorf("Failed to send error message to user %w", err)
		}
	}
	return nil
}

func (proc *processCommand) Process(messageArguments string) error {
	if messageArguments == "" {
		return errors.New("Empty arguments")
	}
	params, err := proc.getHandParams(proc.ctx, proc.handProc, messageArguments)
	if err != nil {
		return err
	}
	proc.log.Debugf("Got parameters %v", params)
	var respWriter strings.Builder
	err = proc.handProc.Process(proc.ctx, &respWriter, params, proc.log)
	if err != nil {
		return err
	}
	return proc.tg.Send(proc.ctx, respWriter.String())
}

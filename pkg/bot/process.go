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

// processingState базовый интерфейс для процессинга
type processingState interface {
	Do() (processingState, error)
}

// baseState базовое состояние хранящее всю информацию необходимую для работы в общем случае
type baseState struct {
	logger        *log.Entry
	ctx           context.Context
	handProcessor core.HandProcessor
	tg            telegram
}

// startState начальное состояние разбирающее стартовые аргументы
type startState struct {
	baseState
	arguments string
}

// inqueryParamsState состояние дозапроса аргументов (если нет пропущенных - идём дальше)
type inqueryParamsState struct {
	baseState
	params map[string]interface{}
}

// finishState пишем результат
type finishState struct {
	baseState
	params map[string]interface{}
}

// finishState пишем результат
type queryParam struct {
	baseState
	paramProcessor core.ParamProcessor
	params         map[string]interface{}
	missingParams  map[string]core.ParamProcessor
}

//--------------------------------------------- start states methods -------------------------------------------------------

func (st *startState) Do() (processingState, error) {
	params := make(map[string]interface{})

	// TODO: переделать это на reader и построчное чтение?
	for _, row := range strings.Split(st.arguments, "\n")[1:] {
		name, val, err := parseParamRow(st.handProcessor, row)
		if err != nil {
			err = st.tg.Send(st.ctx, fmt.Sprintf("Failed to parse param: \"%s\" %s", name, err.Error()))
			if err != nil {
				return nil, fmt.Errorf("Failed to send error message to user %w", err)
			}
			continue
		}
		params[name] = val
	}
	return &inqueryParamsState{
		st.baseState,
		params,
	}, nil
}

//--------------------------------------------- inquery states methods -------------------------------------------------------

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

func (st *inqueryParamsState) parseAll(input string, missingParams map[string]core.ParamProcessor) error {
	var err error
PARSE_PARAMS:
	for _, row := range strings.Split(input, "\n") {
		name, val, err := parseParamRow(st.handProcessor, row)
		if err != nil {
			err = st.tg.Send(st.ctx, fmt.Sprintf("Failed to parse param: \"%s\" %s", name, err.Error()))
			if err != nil {
				return fmt.Errorf("Failed to send error message to user %w", err)
			}
			continue PARSE_PARAMS
		}
		delete(missingParams, name)
		st.params[name] = val
	}
	if err != nil {
		err = st.tg.Send(st.ctx, fmt.Sprintf("Failed to parse param: \"%s\"", err.Error()))
		if err != nil {
			return fmt.Errorf("Failed to send error message to user %w", err)
		}
	}
	return nil
}

func (st *inqueryParamsState) helpCommand(msg string) (processingState, error) {

	if !strings.HasPrefix(msg, "🤖 help") {
		return nil, fmt.Errorf("Invalid comand %s", msg)
	}
	var param string
	// TODO: унифицировать работу с кнопками
	// сделать генерацию кнопок и их парсинг через один формат
	fmt.Sscanf(msg, "🤖 help %s", &param)

	var respWriter strings.Builder
	paramProc, err := st.handProcessor.GetParam(param)
	if err != nil {
		return nil, err
	}
	err = paramProc.WriteHelp(&respWriter)
	if err != nil {
		return nil, err
	}
	err = st.tg.Send(st.ctx, respWriter.String())
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (st *inqueryParamsState) Do() (processingState, error) {
	requiredParams, err := st.handProcessor.GetRequiredParams()
	if err != nil {
		return nil, fmt.Errorf("Failed to get hand required parameters: %w", err)
	}

	getMissingParams := func() map[string]core.ParamProcessor {
		missingParams := make(map[string]core.ParamProcessor)
		for _, param := range requiredParams {
			if _, ok := st.params[param.GetInfo().Name]; !ok {
				missingParams[param.GetInfo().Name] = param
			}
		}
		return missingParams
	}

	missingParams := getMissingParams()

	cmdProcessors := buttonRouters{
		func(msg string) (processingState, error) {
			state, err := st.helpCommand(msg)
			if err != nil {
				st.logger.Debug("Failed to apply help %s", err.Error())
			} else {
				st.logger.Debug("Ok to apply help!")
			}
			return state, err
		},
		func(msg string) (processingState, error) {
			if handle, ok := missingParams[msg]; ok {
				return &queryParam{
					st.baseState,
					handle,
					st.params,
					missingParams,
				}, nil
			}
			st.logger.Debug("Failed to move to queryParam %s no such param", msg)
			return nil, fmt.Errorf("No such param %s", msg)
		},
	}

	for len(missingParams) != 0 {
		err := st.tg.RequestParams(missingParams)
		if err != nil {
			//TODO проверить обработку ошибок и ретраи
			return nil, fmt.Errorf("failed request missing parameters from user %w", err)
		}
		txt, err := st.tg.Get(st.ctx)
		if err != nil {
			return nil, err
		}
		state, err := applyMessageRouters(txt, cmdProcessors)
		if err == nil {
			if state != nil {
				return state, nil
			}
			continue
		}
		st.logger.Debugf("Error on apply routers %s", err.Error())
		_ = st.tg.Send(st.ctx, fmt.Sprintf("I don't know what is: \"%s\"", txt))
	}

	return &finishState{
		st.baseState,
		st.params,
	}, nil
}

//-------------------------------------------- queryParam states methods -------------------------------------------------------

func (st *queryParam) Do() (processingState, error) {
	err := st.tg.Send(st.ctx, fmt.Sprintf("Input value for param: \"%s\"", st.paramProcessor.GetInfo().Name))
	if err != nil {
		//TODO проверить обработку ошибок и ретраи
		return nil, fmt.Errorf("failed request missing parameters from user %w", err)
	}
LOOP:
	for {
		inp, err := st.tg.Get(st.ctx)
		if err != nil {
			continue LOOP
		}
		value, err := st.paramProcessor.ParseFromString(inp)
		if err != nil {
			err = st.tg.Send(st.ctx, fmt.Sprintf("Failed to parse param:  %s", err.Error()))
			if err != nil {
				return nil, fmt.Errorf("Failed to send error message to user %w", err)
			}
			continue LOOP
		}
		if _, ok := st.missingParams[st.paramProcessor.GetInfo().Name]; ok {
			delete(st.missingParams, st.paramProcessor.GetInfo().Name)

		}
		st.params[st.paramProcessor.GetInfo().Name] = value
		break
	}

	return &inqueryParamsState{
		st.baseState,
		st.params,
	}, nil
}

//--------------------------------------------- finish states methods -------------------------------------------------------

func (st *finishState) Do() (processingState, error) {
	var respWriter strings.Builder
	err := st.handProcessor.Process(st.ctx, &respWriter, st.params, st.logger)
	if err != nil {
		return nil, err
	}
	err = st.tg.Send(st.ctx, respWriter.String())
	return nil, err
}

//-------------------------------------------------- base methods -----------------------------------------------------------

func newProcessCommand(ctx context.Context, handProc core.HandProcessor, tg telegram, log *log.Entry) comand {
	return &processCommand{
		ctx:      ctx,
		tg:       tg,
		handProc: handProc,
		log:      log,
	}
}

func (proc *processCommand) Process(messageArguments string) error {
	if messageArguments == "" {
		return errors.New("Empty arguments")
	}
	var currentState processingState = &startState{
		baseState: baseState{
			ctx:           proc.ctx,
			logger:        proc.log,
			handProcessor: proc.handProc,
			tg:            proc.tg,
		},
		arguments: messageArguments,
	}
	var err error
	for currentState != nil {
		currentState, err = currentState.Do()
	}
	return err
}

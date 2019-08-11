package core

import (
	"fmt"
	"io"
	"sort"
)

type HandProcessorImp struct {
	*UrlRecord
}

func NewHandProcessor(rec *UrlRecord) HandProcessor {
	imp := HandProcessorImp{rec}
	return &imp
}

func (processor *HandProcessorImp) WriteHelp(writer io.Writer) error {
	_, err := io.WriteString(writer, fmt.Sprintf("Name: %s\n", processor.UrlName))
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, fmt.Sprintf("URL template: %s\n", string(processor.UrlTemplate)))
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, fmt.Sprintf("Parameters:\n"))
	if err != nil {
		return err
	}
	var keys []string
	for key := range processor.UrlRecord.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		proc, _ := processor.GetParam(key)
		if err != nil {
			// TODO: Предположим что ошибок тут быть не может, но если есть
			// падаем
			return err
		}
		err = proc.WriteHelp(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*HandProcessorImp) Process(writer io.Writer) error {
	return nil
}

func (*HandProcessorImp) GetInfo() *UrlRecord {
	return nil
}

func (imp *HandProcessorImp) GetParam(paramName string) (ParamProcessor, error) {
	paramValue, ok := imp.Parameters[paramName]
	if !ok {
		return nil, NonExistentParamError
	}
	param := NewParamProcessor(paramValue)
	return &param, nil
}

type ParamProcessorImp struct {
	ParamInfo
}

func NewParamProcessor(info ParamInfo) ParamProcessorImp {
	return ParamProcessorImp{info}
}

func (p *ParamProcessorImp) WriteHelp(writer io.Writer) error {
	destination, err := p.Destination.ToString()
	if err != nil {
		return err
	}
	typeStr, err := p.Type.ToString()
	if err != nil {
		return err
	}
	res := fmt.Sprintf(
		"%s(%s)\t%s\n\t%s\n",
		p.Name,
		typeStr,
		destination,
		p.Help,
	)
	_, err = io.WriteString(writer, res)
	return err
}

func (imp *ParamProcessorImp) GetInfo() ParamInfo {
	return imp.ParamInfo
}

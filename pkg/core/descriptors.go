package core

import (
	"fmt"
	"io"
)

type HandProcessorImp struct {
	*UrlRecord
}

func NewHandProcessor(rec *UrlRecord) HandProcessor {
	imp := HandProcessorImp{rec}
	return &imp
}

func (*HandProcessorImp) WriteHelp(writer io.Writer) error {
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
		"%s(%s)\t%s\n%s",
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

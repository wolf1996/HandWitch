package core

import (
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
	paramValue, ok := imp.UrlParameters[paramName]
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

func (*ParamProcessorImp) WriteHelp(writer io.Writer) error {
	return nil
}

func (imp *ParamProcessorImp) GetInfo() ParamInfo {
	return imp.ParamInfo
}

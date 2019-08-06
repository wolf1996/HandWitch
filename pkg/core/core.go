package core

import (
	"errors"
	"html/template"
	"io"
	"net/http"
)

type ParamInfo = string
type ParamsDescription = map[string]ParamInfo

var (
	NonExistentHandError  = errors.New("Can't Find key")
	NonExistentParamError = errors.New("Can't Find param")
)

type UrlRecord struct {
	UrlTemplate     template.URL
	UrlParameters   ParamsDescription
	QueryParameters ParamsDescription
	Body            string
	UrlName         string
	Help            string
}

type UrlContrainer = map[string]UrlRecord

type UrlProcessor struct {
	container  UrlContrainer
	httpClient *http.Client
}

type HandProcessor interface {
	WriteHelp(writer io.Writer) error
	Process(writer io.Writer) error
	GetInfo() *UrlRecord
	GetParam(string) (ParamProcessor, error)
}

type ParamProcessor interface {
	WriteHelp(writer io.Writer) error
	GetInfo() ParamInfo
}

func NewUrlProcessor(container UrlContrainer, httpClient *http.Client) UrlProcessor {
	return UrlProcessor{
		container:  container,
		httpClient: httpClient,
	}
}

func (processor *UrlProcessor) GetHand(name string) (HandProcessor, error) {
	urlInfo, ok := processor.container[name]
	if !ok {
		return nil, NonExistentHandError
	}
	handProc := NewHandProcessor(&urlInfo)
	return handProc, nil
}

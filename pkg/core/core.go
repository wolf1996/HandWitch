package core

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type ParamDestination int

const (
	URL_PARAM ParamDestination = iota
	QUERY_PARAM
)

func (dst ParamDestination) ToString() (string, error) {
	switch dst {
	case URL_PARAM:
		{
			return "URL Param", nil
		}
	case QUERY_PARAM:
		{
			return "Query Param", nil
		}
	}
	return "", fmt.Errorf("Wrong parameter destination %d", dst)
}

type ParamType int

const (
	INTEGER ParamType = iota
	STRING
)

func (tp ParamType) ToString() (string, error) {
	switch tp {
	case INTEGER:
		{
			return "Integer", nil
		}
	case STRING:
		{
			return "String", nil
		}
	}
	return "", fmt.Errorf("Wrong parameter type %d", tp)
}

type ParamInfo struct {
	Help        string
	Name        string
	Destination ParamDestination
	Type        ParamType
}

type ParamsDescription map[string]ParamInfo

var (
	NonExistentHandError  = errors.New("Can't Find key")
	NonExistentParamError = errors.New("Can't Find param")
)

type UrlRecord struct {
	UrlTemplate template.URL
	Parameters  ParamsDescription
	Body        string
	UrlName     string
	Help        string
}

type UrlContrainer map[string]UrlRecord

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

// Package core provides base types for library usage
package core

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ParamDestination - type to represent where, the parameter
// should be placed in query
// In url - URL_PARAM
// Like a query param - QUERY_PARAM
type ParamDestination int

const (
	UrlPlaced ParamDestination = iota
	QueryPlaced
)

// Get String representation of ParamDestination
func (dst ParamDestination) ToString() (string, error) {
	switch dst {
	case UrlPlaced:
		{
			return "URL Param", nil
		}
	case QueryPlaced:
		{
			return "Query Param", nil
		}
	}
	return "", fmt.Errorf("Wrong parameter destination %d", dst)
}

// Param type representations
type ParamType int

const (
	IntegerType ParamType = iota
	StringType
)

func (tp ParamType) ToString() (string, error) {
	switch tp {
	case IntegerType:
		{
			return "Integer", nil
		}
	case StringType:
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
	ErrNonExistentParam = errors.New("Can't Find param")
)

type UrlRecord struct {
	UrlTemplate string
	Parameters  ParamsDescription
	Body        string
	UrlName     string
	Help        string
}

type UrlContrainer map[string]UrlRecord

type UrlProcessor struct {
	container  DescriptionsSource
	httpClient *http.Client
}

type HandProcessor interface {
	WriteHelp(writer io.Writer) error
	Process(writer io.Writer, params map[string]interface{}) error
	GetInfo() *UrlRecord
	GetParam(string) (ParamProcessor, error)
}

type ParamProcessor interface {
	WriteHelp(writer io.Writer) error
	GetInfo() ParamInfo
}

func NewUrlProcessor(container DescriptionsSource, httpClient *http.Client) UrlProcessor {
	return UrlProcessor{
		container:  container,
		httpClient: httpClient,
	}
}

func (processor *UrlProcessor) GetHand(name string) (HandProcessor, error) {
	urlInfo, err := processor.container.GetByName(name)
	if err != nil {
		return nil, err
	}
	return NewHandProcessor(urlInfo, processor.httpClient)
}

// Package core provides base types for library usage
package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ParamDestination - type to represent where, the parameter
// should be placed in query
// In URL - URL_PARAM
// Like a query param - QUERY_PARAM
type ParamDestination string

const (
	//URLPlaced hand is part of url path
	URLPlaced ParamDestination = "URL"
	//QueryPlaced query parameter
	QueryPlaced ParamDestination = "query"
)

// ToString Get String representation of ParamDestination
func (dst ParamDestination) ToString() (string, error) {
	switch dst {
	case URLPlaced:
		{
			return "URL Param", nil
		}
	case QueryPlaced:
		{
			return "Query Param", nil
		}
	}
	return "", fmt.Errorf("Wrong parameter destination %s", dst)
}

// ParamType Param type representations
type ParamType string

const (
	//IntegerType param represented as an integer
	IntegerType ParamType = "integer"
	//StringType param represented as a string
	StringType ParamType = "string"
)

//ToString Get human-readable String representation
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
	return "", fmt.Errorf("Wrong parameter type %s", tp)
}

//ParamInfo Config parameter description
type ParamInfo struct {
	Help        string           `json:"help"`
	Name        string           `json:"name"`
	Destination ParamDestination `json:"destination"`
	Type        ParamType        `json:"type"`
}

//ParamsDescription Container for param
type ParamsDescription map[string]ParamInfo

var (
	// ErrNonExistentParam don't requested param
	ErrNonExistentParam = errors.New("Can't Find param")
)

//URLRecord Full Hand description in configuration file
type URLRecord struct {
	URLTemplate string            `json:"URL_template"`
	Parameters  ParamsDescription `json:"params"`
	Body        string            `json:"body"`
	URLName     string            `json:"name"`
	Help        string            `json:"help"`
}

//URLContrainer Container of all URLs
type URLContrainer map[string]URLRecord

//URLProcessor url processor
type URLProcessor struct {
	container  DescriptionsSource
	httpClient *http.Client
}

//HandProcessor hand processor
type HandProcessor interface {
	WriteHelp(writer io.Writer) error
	Process(ctx context.Context, writer io.Writer, params map[string]interface{}) error
	GetInfo() *URLRecord
	GetParam(string) (ParamProcessor, error)
}

//ParamProcessor param processor handles param descriptions
type ParamProcessor interface {
	ParseFromString(str string) (interface{}, error)
	WriteHelp(writer io.Writer) error
	GetInfo() ParamInfo
}

//NewURLProcessor creates new url processor, using data source and http client
func NewURLProcessor(container DescriptionsSource, httpClient *http.Client) URLProcessor {
	return URLProcessor{
		container:  container,
		httpClient: httpClient,
	}
}

//GetHand build hand processor object by name
func (processor *URLProcessor) GetHand(name string) (HandProcessor, error) {
	URLInfo, err := processor.container.GetByName(name)
	if err != nil {
		return nil, err
	}
	return NewHandProcessor(URLInfo, processor.httpClient)
}

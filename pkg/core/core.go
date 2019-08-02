package core

import (
	"html/template"
	"io"
	"net/http"
)

type ParamsDescription = map[string]string

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
	container  *UrlContrainer
	httpClient *http.Client
}

func NewUrlProcessor(container *UrlContrainer, httpClient *http.Client) UrlProcessor {
	return UrlProcessor{
		container:  container,
		httpClient: httpClient,
	}
}

func (processor *UrlProcessor) ProcessHand(name string, writer io.Writer) error {
	return nil
}

func (processor *UrlProcessor) GetParamDescription(name string, writer io.Writer) error {
	return nil
}

func (processor *UrlProcessor) GetHelp(writer io.Writer, handName string) error {
	return nil
}

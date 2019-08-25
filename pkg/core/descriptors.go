package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"text/template"
)

type HandProcessorImp struct {
	*UrlRecord
	client *http.Client
}

func (processor *HandProcessorImp) compileTemplate(rec *UrlRecord, _ map[string]interface{}) (*template.Template, error) {

	getValue := func(name string) string {
		return ""
	}
	return template.New(rec.UrlName).Funcs(
		template.FuncMap{
			"GetValue": getValue,
		},
	).Parse(rec.Body)
}

func NewHandProcessor(rec *UrlRecord, client *http.Client) (HandProcessor, error) {
	imp := HandProcessorImp{
		UrlRecord: rec,
		client:    client,
	}
	return &imp, nil
}

func (processor *HandProcessorImp) WriteHelp(writer io.Writer) error {
	_, err := io.WriteString(writer, fmt.Sprintf("Name: %s\n", processor.UrlName))
	if err != nil {
		return fmt.Errorf("Error while writing name %s", err.Error())
	}
	_, err = io.WriteString(writer, fmt.Sprintf("URL template: %s\n", processor.UrlTemplate))
	if err != nil {
		return fmt.Errorf("Error while writing url template %s", err.Error())
	}
	_, err = io.WriteString(writer, fmt.Sprintf("Parameters:\n"))
	if err != nil {
		return fmt.Errorf("Error while writing url parameters header %s", err.Error())
	}
	var keys []string
	for key := range processor.UrlRecord.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		proc, _ := processor.GetParam(key)
		if err != nil {
			return fmt.Errorf("Error while writing get parameter help for key %s: %s", key, err.Error())
		}
		err = proc.WriteHelp(writer)
		if err != nil {
			return fmt.Errorf("Error while writing parameters help for key %s: %s", key, err.Error())
		}
	}
	return nil
}

func (processor *HandProcessorImp) Process(writer io.Writer, params map[string]interface{}) error {
	buf := new(bytes.Buffer)
	tmp, err := template.New(processor.UrlName).Parse(processor.UrlTemplate)
	if err != nil {
		return fmt.Errorf("Failed to build url template %s", err.Error())
	}
	err = tmp.Execute(buf, params)
	if err != nil {
		return fmt.Errorf("Failed to build url %s", err.Error())
	}
	req, err := http.NewRequest("GET", buf.String(), nil)
	if err != nil {
		return fmt.Errorf("Failed to build request %s", err.Error())
	}
	responce, err := processor.client.Do(req)
	data := make(map[string]interface{})
	if err != nil {
		return fmt.Errorf("Failed to read result %s", err.Error())
	}
	err = json.NewDecoder(responce.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("Failed to decode json result %s", err.Error())
	}
	template, err := processor.compileTemplate(processor.UrlRecord, params)
	if err != nil {
		return fmt.Errorf("Failed to build request %s", err.Error())
	}
	err = template.Lookup(processor.UrlRecord.UrlName).Execute(writer, data)
	if err != nil {
		return fmt.Errorf("Failed to execute %s", err.Error())
	}
	return err
}

func (*HandProcessorImp) GetInfo() *UrlRecord {
	return nil
}

func (imp *HandProcessorImp) GetParam(paramName string) (ParamProcessor, error) {
	paramValue, ok := imp.Parameters[paramName]
	if !ok {
		return nil, ErrNonExistentParam
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
		return fmt.Errorf("Error while destination help converting %s", err.Error())
	}
	typeStr, err := p.Type.ToString()
	if err != nil {
		return fmt.Errorf("Error while Type help converting %s", err.Error())
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

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sort"
	"strconv"
	"text/template"

	log "github.com/sirupsen/logrus"
)

//HandProcessorImp hand processor implementation
type HandProcessorImp struct {
	*URLRecord
	client *http.Client
}

func (processor *HandProcessorImp) compileTemplate(rec *URLRecord, _ map[string]interface{}) (*template.Template, error) {

	getValue := func(name string) string {
		return ""
	}
	return template.New(rec.URLName).Funcs(
		template.FuncMap{
			"GetValue": getValue,
		},
	).Parse(rec.Body)
}

//NewHandProcessor build hand processor to current data
func NewHandProcessor(rec *URLRecord, client *http.Client) (HandProcessor, error) {
	imp := HandProcessorImp{
		URLRecord: rec,
		client:    client,
	}
	return &imp, nil
}

//WriteHelp write help to current data
func (processor *HandProcessorImp) WriteHelp(writer io.Writer) error {
	_, err := io.WriteString(writer, fmt.Sprintf("Name: %s\n", processor.URLName))
	if err != nil {
		return fmt.Errorf("Error while writing name %w", err)
	}
	_, err = io.WriteString(writer, fmt.Sprintf("URL template: %s\n", processor.URLTemplate))
	if err != nil {
		return fmt.Errorf("Error while writing URL template %w", err)
	}
	_, err = io.WriteString(writer, fmt.Sprintf("Parameters:\n"))
	if err != nil {
		return fmt.Errorf("Error while writing URL parameters header %w", err)
	}
	var keys []string
	for key := range processor.URLRecord.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		proc, _ := processor.GetParam(key)
		if err != nil {
			return fmt.Errorf("Error while writing get parameter help for key %s: %w", key, err)
		}
		err = proc.WriteHelp(writer)
		if err != nil {
			return fmt.Errorf("Error while writing parameters help for key %s: %w", key, err)
		}
	}
	return nil
}

func (processor *HandProcessorImp) addQueryParams(req *http.Request, params map[string]interface{}) {
	qry := req.URL.Query()
	for name, description := range processor.URLRecord.Parameters {
		if description.Destination == QueryPlaced {
			//TODO: add reflection?
			val, ok := params[name]
			if ok {
				qry.Add(name, fmt.Sprintf("%v", val))
			}
		}
	}
	req.URL.RawQuery = qry.Encode()
}

//Process load data from hand url and
//execute template with it
func (processor *HandProcessorImp) Process(ctx context.Context, writer io.Writer, params map[string]interface{}, logger *log.Entry) error {
	url := new(bytes.Buffer)
	tmp, err := template.New(processor.URLName).Parse(processor.URLTemplate)
	if err != nil {
		return fmt.Errorf("Failed to build URL template %w", err)
	}
	err = tmp.Execute(url, params)
	if err != nil {
		return fmt.Errorf("Failed to build URL %w", err)
	}
	logger.Debugf("Got URL %s", url.String())
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return fmt.Errorf("Failed to build request %w", err)
	}
	processor.addQueryParams(req, params)
	logger.Debugf("Got request %s", func() string {
		bytes, err := httputil.DumpRequest(req, true)
		if err != nil {
			return err.Error()
		}
		return string(bytes)
	}())
	responce, err := processor.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to read result %w", err)
	}
	logger.Debugf("Got responce %s", func() string {
		bytes, err := httputil.DumpResponse(responce, true)
		if err != nil {
			return err.Error()
		}
		return string(bytes)
	}())
	responceData := make(map[string]interface{})
	err = json.NewDecoder(responce.Body).Decode(&responceData)
	if err != nil {
		return fmt.Errorf("Failed to decode json result %w", err)
	}
	template, err := processor.compileTemplate(processor.URLRecord, params)
	if err != nil {
		return fmt.Errorf("Failed to build request %w", err)
	}
	templateData := map[string]interface{}{
		"responce": responceData,
		"meta": map[string]interface{}{
			"url":    req.URL.String(),
			"params": params,
		},
	}
	err = template.Lookup(processor.URLRecord.URLName).Execute(writer, templateData)
	if err != nil {
		return fmt.Errorf("Failed to execute %w", err)
	}
	return nil
}

//GetInfo get row data
func (processor *HandProcessorImp) GetInfo() *URLRecord {
	return processor.URLRecord
}

//GetParam get parametere handler
func (processor *HandProcessorImp) GetParam(paramName string) (ParamProcessor, error) {
	paramValue, ok := processor.Parameters[paramName]
	if !ok {
		return nil, ErrNonExistentParam
	}
	param := NewParamProcessor(paramValue)
	return &param, nil
}

//GetRequiredParams get all required parameters
func (processor *HandProcessorImp) GetRequiredParams() ([]ParamProcessor, error) {
	result := make([]ParamProcessor, 0)
	for paramName, _ := range processor.Parameters {
		param, err := processor.GetParam(paramName)
		if err != nil {
			return result, err
		}
		if param.IsRequired() {
			result = append(result, param)
		}
	}
	return result, nil
}

//ParamProcessorImp implemet
type ParamProcessorImp struct {
	ParamInfo
}

//NewParamProcessor build new parameter processor by info
func NewParamProcessor(info ParamInfo) ParamProcessorImp {
	return ParamProcessorImp{info}
}

//WriteHelp writs help
func (p *ParamProcessorImp) WriteHelp(writer io.Writer) error {
	destination, err := p.Destination.ToString()
	if err != nil {
		return fmt.Errorf("Error while destination help converting %w", err)
	}
	typeStr, err := p.Type.ToString()
	if err != nil {
		return fmt.Errorf("Error while Type help converting %w", err)
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

//GetInfo get raw info
func (p *ParamProcessorImp) GetInfo() ParamInfo {
	return p.ParamInfo
}

//IsRequired check if parameter is required
func (p *ParamProcessorImp) IsRequired() bool {
	return p.Destination == URLPlaced
}

//ParseFromString get param value from string
func (p *ParamProcessorImp) ParseFromString(str string) (interface{}, error) {
	switch p.Type {
	case StringType:
		return str, nil
	case IntegerType:
		return strconv.Atoi(str)
	}
	//TODO: make a new good errors
	return nil, fmt.Errorf("Unknown type %s", p.Type)
}

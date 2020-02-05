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

func (processor *HandProcessorImp) mergeWithDefault(params map[string]interface{}, logger *log.Entry) {
	defaultValues := processor.GetParamsDefaultValues()
	for paramName, defaultValue := range defaultValues {
		_, hasValue := params[paramName]
		if !hasValue {
			params[paramName] = defaultValue
			logger.Debugf("using default value %v for param %s", defaultValue, paramName)
		}
	}
}

//Process load data from hand url and
//execute template with it
func (processor *HandProcessorImp) Process(ctx context.Context, writer io.Writer, params map[string]interface{}, logger *log.Entry) error {
	processor.mergeWithDefault(params, logger)
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

	defer responce.Body.Close()
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
// TODO: поправить на мап
func (processor *HandProcessorImp) GetRequiredParams() ([]ParamProcessor, error) {
	result := make([]ParamProcessor, 0)
	for paramName := range processor.Parameters {
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

//GetParams get all parameters
func (processor *HandProcessorImp) GetParams() (map[string]ParamProcessor, error) {
	result := make(map[string]ParamProcessor)
	for paramName := range processor.Parameters {
		param, err := processor.GetParam(paramName)
		if err != nil {
			return result, err
		}
		result[paramName] = param
	}
	return result, nil
}

//GetParamsDefaultValues get all default parameters values
func (processor *HandProcessorImp) GetParamsDefaultValues() map[string]interface{} {
	result := make(map[string]interface{})
	for paramName, param := range processor.Parameters {
		if param.DefaultValue != nil {
			result[paramName] = param.DefaultValue
		}
	}
	return result
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

	_, err = io.WriteString(writer, p.Name)
	if err != nil {
		return fmt.Errorf("Failed to write help for %s: %w", p.Name, err)
	}
	_, err = io.WriteString(writer, fmt.Sprintf("(%s)", typeStr))
	if err != nil {
		return fmt.Errorf("Failed to write help for %s: %w", p.Name, err)
	}
	_, err = io.WriteString(writer, "\t"+destination)
	if err != nil {
		return fmt.Errorf("Failed to write help for %s: %w", p.Name, err)
	}
	if !p.IsRequired() {
		_, err = io.WriteString(writer, "\t[Optional]")
		if err != nil {
			return fmt.Errorf("Failed to write help for %s: %w", p.Name, err)
		}
	}
	io.WriteString(writer, "\n")
	if p.DefaultValue != nil {
		_, err = io.WriteString(writer, fmt.Sprintf("\tDefault: %v\n", p.DefaultValue))
		if err != nil {
			return fmt.Errorf("Failed to write help for %s: %w", p.Name, err)
		}
	}
	_, err = io.WriteString(writer, "\t"+p.Help+"\n")
	if err != nil {
		return fmt.Errorf("Failed to write help for %s: %w", p.Name, err)
	}
	return nil
}

//GetInfo get raw info
func (p *ParamProcessorImp) GetInfo() ParamInfo {
	return p.ParamInfo
}

//IsRequired check if parameter is required
func (p *ParamProcessorImp) IsRequired() bool {
	return (p.Destination == URLPlaced && p.DefaultValue == nil) || (p.Destination == QueryPlaced && (!p.Optional && p.DefaultValue == nil))
}

func parseString(str string) (interface{}, error) {
	return str, nil
}

func parseInt(str string) (interface{}, error) {
	return strconv.Atoi(str)
}

func parseFromInterface(tp ParamType, val interface{}) (interface{}, error) {
	switch valType := val.(type) {
	case int:
		{
			return val, nil
		}
	case string:
		{
			return parseValue(tp, val.(string))
		}
	default:
		return nil, fmt.Errorf("failed to get %v from param %v with type %v", tp, val, valType)
	}
}

func parseValue(tp ParamType, str string) (interface{}, error) {
	switch tp {
	case StringType:
		return parseString(str)
	case IntegerType:
		return parseInt(str)
	}
	//TODO: make a new good errors
	return nil, fmt.Errorf("Unknown type %s", tp)
}

//ParseFromString get param value from string
func (p *ParamProcessorImp) ParseFromString(str string) (interface{}, error) {
	return parseValue(p.Type, str)
}

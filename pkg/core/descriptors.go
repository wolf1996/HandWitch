package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"text/template"
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
		return fmt.Errorf("Error while writing name %s", err.Error())
	}
	_, err = io.WriteString(writer, fmt.Sprintf("URL template: %s\n", processor.URLTemplate))
	if err != nil {
		return fmt.Errorf("Error while writing URL template %s", err.Error())
	}
	_, err = io.WriteString(writer, fmt.Sprintf("Parameters:\n"))
	if err != nil {
		return fmt.Errorf("Error while writing URL parameters header %s", err.Error())
	}
	var keys []string
	for key := range processor.URLRecord.Parameters {
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
func (processor *HandProcessorImp) Process(writer io.Writer, params map[string]interface{}) error {
	buf := new(bytes.Buffer)
	tmp, err := template.New(processor.URLName).Parse(processor.URLTemplate)
	if err != nil {
		return fmt.Errorf("Failed to build URL template %s", err.Error())
	}
	err = tmp.Execute(buf, params)
	if err != nil {
		return fmt.Errorf("Failed to build URL %s", err.Error())
	}
	req, err := http.NewRequest("GET", buf.String(), nil)
	if err != nil {
		return fmt.Errorf("Failed to build request %s", err.Error())
	}
	processor.addQueryParams(req, params)
	responce, err := processor.client.Do(req)
	data := make(map[string]interface{})
	if err != nil {
		return fmt.Errorf("Failed to read result %s", err.Error())
	}
	err = json.NewDecoder(responce.Body).Decode(&data)
	if err != nil {
		return fmt.Errorf("Failed to decode json result %s", err.Error())
	}
	template, err := processor.compileTemplate(processor.URLRecord, params)
	if err != nil {
		return fmt.Errorf("Failed to build request %s", err.Error())
	}
	err = template.Lookup(processor.URLRecord.URLName).Execute(writer, data)
	if err != nil {
		return fmt.Errorf("Failed to execute %s", err.Error())
	}
	return err
}

//GetInfo get row data
func (*HandProcessorImp) GetInfo() *URLRecord {
	return nil
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

//GetInfo get raw info
func (p *ParamProcessorImp) GetInfo() ParamInfo {
	return p.ParamInfo
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

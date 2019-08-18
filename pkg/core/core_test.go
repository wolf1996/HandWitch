package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParametersList(t *testing.T) {
	testCases := []struct {
		inp UrlContrainer
		out map[string]map[string]string
	}{
		{
			UrlContrainer{
				"hand1": {
					UrlTemplate: "",
					Parameters: ParamsDescription{
						"Name1": ParamInfo{
							Name:        "Name1",
							Help:        "Help to Name1",
							Type:        INTEGER,
							Destination: QUERY_PARAM,
						},
						"Name2": ParamInfo{
							Name:        "Name2",
							Help:        "Help to Name2",
							Type:        STRING,
							Destination: URL_PARAM,
						},
					},
					Body:    "",
					UrlName: "name",
				},
			},
			map[string]map[string]string{
				"hand1": {
					"Name1": "Name1(Integer)\tQuery Param\n\tHelp to Name1\n",
					"Name2": "Name2(String)\tURL Param\n\tHelp to Name2\n",
				},
			},
		},
	}

	for _, testCase := range testCases {
		input := testCase.inp
		output := testCase.out
		processor := NewUrlProcessor(input, nil)
		for key, params := range output {
			handDescr, err := processor.GetHand(key)
			if err != nil {
				t.Errorf("Failed to get hand handler %s for hand %s", err.Error(), key)
				continue
			}
			for paramName, expect := range params {
				buf := new(bytes.Buffer)
				param, err := handDescr.GetParam(paramName)
				if err != nil {
					t.Errorf("Failed to get param handler %s for param %s for hand %s", err.Error(), paramName, key)
					continue
				}
				err = param.WriteHelp(buf)
				if err != nil {
					t.Errorf("Failed to write param handler %s for param %s for hand %s", err.Error(), paramName, key)
					continue
				}
				got := buf.String()
				if got != expect {
					t.Errorf("Wrong parameter help %s param name %s expected \n%s\n got\n%s", key, paramName, expect, got)
				}
			}
		}
	}
}

func TestHelp(t *testing.T) {
	testCases := []struct {
		inp UrlContrainer
		out map[string]string
	}{
		{
			UrlContrainer{
				"hand1": {
					UrlTemplate: "http://example.com/entity/{entity_id}/v/{v}",
					Parameters: ParamsDescription{
						"entity_id": ParamInfo{
							Name:        "entity_id",
							Help:        "Help to entity_id",
							Type:        INTEGER,
							Destination: URL_PARAM,
						},
						"v": ParamInfo{
							Name:        "v",
							Help:        "Help to v",
							Type:        STRING,
							Destination: URL_PARAM,
						},
						"QueryParam1": ParamInfo{
							Name:        "QueryParam1",
							Help:        "Help to QueryParam1",
							Type:        INTEGER,
							Destination: QUERY_PARAM,
						},
						"QueryParam2": ParamInfo{
							Name:        "QueryParam2",
							Help:        "Help to QueryParam2",
							Type:        STRING,
							Destination: QUERY_PARAM,
						},
					},
					Body:    "",
					UrlName: "hand1",
				},
				"handNoUrlParams": {
					UrlTemplate: "http://example.com/entity",
					Parameters: ParamsDescription{
						"QueryParam1": ParamInfo{
							Name:        "QueryParam1",
							Help:        "Help to QueryParam1",
							Type:        INTEGER,
							Destination: QUERY_PARAM,
						},
						"QueryParam2": ParamInfo{
							Name:        "QueryParam2",
							Help:        "Help to QueryParam2",
							Type:        STRING,
							Destination: QUERY_PARAM,
						},
					},
					Body:    "",
					UrlName: "handNoUrlParams",
				},
				"handNoQueryParams": {
					UrlTemplate: "http://example.com/entity/{entity_id}/v/{v}",
					Parameters: ParamsDescription{
						"entity_id": ParamInfo{
							Name:        "entity_id",
							Help:        "Help to entity_id",
							Type:        INTEGER,
							Destination: URL_PARAM,
						},
						"v": ParamInfo{
							Name:        "v",
							Help:        "Help to v",
							Type:        STRING,
							Destination: URL_PARAM,
						},
					},
					Body:    "",
					UrlName: "handNoQueryParams",
				},
				"handNoParams": {
					UrlTemplate: "http://example.com/entity/",
					Parameters:  ParamsDescription{},
					Body:        "",
					UrlName:     "handNoParams",
				},
			},
			map[string]string{
				"hand1": "Name: hand1\n" +
					"URL template: http://example.com/entity/{entity_id}/v/{v}\n" +
					"Parameters:\n" +
					"QueryParam1(Integer)\tQuery Param\n\tHelp to QueryParam1\n" +
					"QueryParam2(String)\tQuery Param\n\tHelp to QueryParam2\n" +
					"entity_id(Integer)\tURL Param\n\tHelp to entity_id\n" +
					"v(String)\tURL Param\n\tHelp to v\n",
				"handNoUrlParams": "Name: handNoUrlParams\n" +
					"URL template: http://example.com/entity\n" +
					"Parameters:\n" +
					"QueryParam1(Integer)\tQuery Param\n\tHelp to QueryParam1\n" +
					"QueryParam2(String)\tQuery Param\n\tHelp to QueryParam2\n",
				"handNoQueryParams": "Name: handNoQueryParams\n" +
					"URL template: http://example.com/entity/{entity_id}/v/{v}\n" +
					"Parameters:\n" +
					"entity_id(Integer)\tURL Param\n\tHelp to entity_id\n" +
					"v(String)\tURL Param\n\tHelp to v\n",
			},
		},
	}

	for _, testCase := range testCases {
		input := testCase.inp
		output := testCase.out
		processor := NewUrlProcessor(input, nil)
		for key, val := range output {
			buf := new(bytes.Buffer)
			handProcessor, err := processor.GetHand(key)
			if err != nil {
				t.Errorf("Failed to get handler %s for hand %s", err.Error(), key)
			}
			err = handProcessor.WriteHelp(buf)
			if err != nil {
				t.Errorf("Failed to process help %s for hand %s", err.Error(), key)
			}
			got := buf.String()
			if got != val {
				t.Errorf("Failed to get parameter help %s expected %s got %s", key, val, got)
			}
		}
	}
}

func TestRender(t *testing.T) {
	mustBuildRequest := func(method string, url string) *http.Request {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			panic(fmt.Sprintf("Failed to build request %s", err.Error()))
		}
		return req
	}

	var requests []*http.Request

	handler := func(rw http.ResponseWriter, req *http.Request) {
	}

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requests = append(requests, req)
		handler(rw, req)
	}))
	defer serv.Close()

	testCases := []struct {
		inp struct {
			UrlS    UrlContrainer
			Handler func(rw http.ResponseWriter, req *http.Request)
		}
		out map[string]struct {
			Inp      map[string]interface{}
			Output   string
			Requests []*http.Request
			Err      error
		}
	}{
		{
			struct {
				UrlS    UrlContrainer
				Handler func(rw http.ResponseWriter, req *http.Request)
			}{
				UrlContrainer{
					"hand1": {
						UrlTemplate: fmt.Sprintf("%s/entity/{entity_id}/v/{v}", serv.URL),
						Parameters: ParamsDescription{
							"entity_id": ParamInfo{
								Name:        "entity_id",
								Help:        "Help to entity_id",
								Type:        INTEGER,
								Destination: URL_PARAM,
							},
							"v": ParamInfo{
								Name:        "v",
								Help:        "Help to v",
								Type:        STRING,
								Destination: URL_PARAM,
							},
							"QueryParam1": ParamInfo{
								Name:        "QueryParam1",
								Help:        "Help to QueryParam1",
								Type:        INTEGER,
								Destination: QUERY_PARAM,
							},
							"QueryParam2": ParamInfo{
								Name:        "QueryParam2",
								Help:        "Help to QueryParam2",
								Type:        STRING,
								Destination: QUERY_PARAM,
							},
						},
						Body:    "Value of Value is {{ .value }}",
						UrlName: "ValuableName",
					},
				},
				func(rw http.ResponseWriter, req *http.Request) {
					err := json.NewEncoder(rw).Encode(map[string]interface{}{
						"value": "ValueForValue",
					})
					if err != nil {
						panic(err.Error())
					}
				},
			},
			map[string]struct {
				Inp      map[string]interface{}
				Output   string
				Requests []*http.Request
				Err      error
			}{
				"hand1": {
					map[string]interface{}{
						"entity_id":   1,
						"v":           "a",
						"QueryParam1": 2,
						"QueryParam2": "b",
					},
					"Value of Value is ValueForValue",
					[]*http.Request{
						mustBuildRequest("GET", fmt.Sprintf("%s/entity/{entity_id}/v/{v}", serv.URL)),
					},
					nil,
				},
			},
		},
	}
	for _, testCase := range testCases {
		input := testCase.inp
		output := testCase.out
		handler = input.Handler
		processor := NewUrlProcessor(input.UrlS, serv.Client())
	KEYLOOP:
		for key, expect := range output {
			buf := new(bytes.Buffer)
			handProcessor, err := processor.GetHand(key)
			if err != nil {
				if err != expect.Err {
					continue KEYLOOP
				}
				t.Errorf("Failed to get param handler %s for hand %s", err.Error(), key)
				continue
			}
			err = handProcessor.Process(buf, expect.Inp)
			if err != nil {
				if err != expect.Err {
					continue KEYLOOP
				}
				t.Errorf("Failed to process param handler %s for hand %s", err.Error(), key)
				continue
			}
			got := buf.String()
			if got != expect.Output {
				t.Errorf("Failed to get parameter help %s expected %s got %s", key, expect.Output, got)
				continue KEYLOOP
			}
			// TODO: Сделать проверку урлов
			if err != expect.Err {
				t.Errorf("No error in test expected %v got %v", expect.Err, err)
				continue KEYLOOP
			}
		}
	}
}

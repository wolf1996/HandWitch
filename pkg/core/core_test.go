package core

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

type Input struct {
}

func TestParametersList(t *testing.T) {
	testCases := []struct {
		inp UrlContrainer
		out map[string]map[string]string
	}{
		{
			UrlContrainer{
				"hand1": {
					UrlTemplate: "",
					UrlParameters: ParamsDescription{
						"a": "b",
					},
					QueryParameters: ParamsDescription{
						"c": "d",
					},
					Body:    "",
					UrlName: "name",
				},
			},
			map[string]map[string]string{
				"hand1": {
					"a": "b",
					"c": "d",
				},
			},
		},
	}

	for _, testCase := range testCases {
		input := testCase.inp
		output := testCase.out
		processor := NewUrlProcessor(input, nil)
		for key, params := range output {
			buf := new(bytes.Buffer)
			handDescr, err := processor.GetHand(key)
			if err != nil {
				t.Errorf("Failed to get hand handler %s for hand %s", err.Error(), key)
			}
			for paramName, expect := range params {
				param, err := handDescr.GetParam(paramName)
				if err != nil {
					t.Errorf("Failed to get param handler %s for param %s for hand %s", err.Error(), paramName, key)
				}
				param.WriteHelp(buf)
				got := buf.String()
				if got != expect {
					t.Errorf("Failed to get parameter help %s expected %s got %s", key, expect, got)
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
					UrlParameters: ParamsDescription{
						"entity_id": "Entity id as an integer",
						"v":         "V value",
					},
					QueryParameters: ParamsDescription{
						"QueryParam1": "Query Parameter help 1 ",
						"QueryParam2": "Query Parameter help 2 ",
					},
					Body:    "",
					UrlName: "hand1",
				},
				"handNoUrlParams": {
					UrlTemplate:   "http://example.com/entity",
					UrlParameters: ParamsDescription{},
					QueryParameters: ParamsDescription{
						"QueryParam1": "Query Parameter help 1 ",
						"QueryParam2": "Query Parameter help 2 ",
					},
					Body:    "",
					UrlName: "handNoUrlParams",
				},
				"handNoQueryParams": {
					UrlTemplate: "http://example.com/entity/{entity_id}/v/{v}",
					UrlParameters: ParamsDescription{
						"entity_id": "Entity id as an integer",
						"v":         "V value",
					},
					Body:    "",
					UrlName: "handNoQueryParams",
				},
				"handNoParams": {
					UrlTemplate:   "http://example.com/entity/",
					UrlParameters: ParamsDescription{},
					Body:          "",
					UrlName:       "handNoParams",
				},
			},
			map[string]string{
				"hand1": "Name: hand1\n" +
					"URL template: http://example.com/entity/{entity_id}/v/{v}\n" +
					"Parameters:\n" +
					"UrlPart Parameters: \n" +
					"	entity_id: Entity id as an integer\n" +
					"	v: V value\n" +
					"QueryParameters:\n" +
					"	QueryParam1: Query Parameter help 1\n" +
					"	QueryParam2: Query Parameter help 2\n",
				"handNoUrlParams": "Name: handNoUrlParams\n" +
					"URL template: http://example.com/entity/\n" +
					"Parameters:\n" +
					"QueryParameters:\n" +
					"	QueryParam1: Query Parameter help 1\n" +
					"	QueryParam2: Query Parameter help 2\n",
				"handNoQueryParams": "Name: handNoQueryParams\n" +
					"URL template: http://example.com/entity/\n" +
					"Parameters:\n" +
					"UrlPart Parameters: \n" +
					"	entity_id: Entity id as an integer\n" +
					"	v: V value\n",
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
				t.Errorf("Failed to get param handler %s for hand %s", err.Error(), key)
			}
			handProcessor.WriteHelp(buf)
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
	var reqs []*http.Request
	testCases := []struct {
		inp struct {
			UrlS    UrlContrainer
			Handler http.HandlerFunc
		}
		out map[string]struct {
			Output   string
			Requests []*http.Request
		}
	}{
		{
			struct {
				UrlS    UrlContrainer
				Handler http.HandlerFunc
			}{
				UrlContrainer{
					"hand1": {
						UrlTemplate: "http://example.com/entity/{entity_id}/v/{v}",
						UrlParameters: ParamsDescription{
							"entity_id": "Entity id as an integer",
							"v":         "V value",
						},
						QueryParameters: ParamsDescription{
							"QueryParam1": "Query Parameter help 1 ",
							"QueryParam2": "Query Parameter help 2 ",
						},
						Body:    "{{GetValue(\"value\")}}",
						UrlName: "ValuableName",
					},
				},
				http.HandlerFunc(
					func(rw http.ResponseWriter, req *http.Request) {
						reqs = append(reqs, req)
					}),
			},
			map[string]struct {
				Output   string
				Requests []*http.Request
			}{
				"hand1": {
					"valueValue1",
					[]*http.Request{
						mustBuildRequest("GET", "http://example.com/entity/{entity_id}/v/{v}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		input := testCase.inp
		output := testCase.out
		processor := NewUrlProcessor(input.UrlS, nil)
		for key, expect := range output {
			buf := new(bytes.Buffer)
			handProcessor, err := processor.GetHand(key)
			if err != nil {
				t.Errorf("Failed to get param handler %s for hand %s", err.Error(), key)
			}
			handProcessor.Process(buf)
			got := buf.String()
			if got != expect.Output {
				t.Errorf("Failed to get parameter help %s expected %s got %s", key, expect.Output, got)
			}
			// TODO: Сделать проверку урлов
		}
	}
}

package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestInput struct {
	Descriptions DescriptionsSource
	Handler      func(rw http.ResponseWriter, req *http.Request)
}

type TestOutput struct {
	HandName string
	Inp      map[string]interface{}
	Output   string
	Requests []*http.Request
	Err      error
}

type TestOutputParamHelp struct {
	HandName string
	Helps    map[string]string
	Err      error
}

type TestCases = []interface{}

type TestDescription struct {
	Inp TestInput
	Out TestCases
}

type TestDescriptions = []TestDescription

func TestParametersList(t *testing.T) {
	// Проверяем как обрабатывается запрос справки для отдельных параметров
	testCases := TestDescriptions{
		TestDescription{
			TestInput{
				NewDescriptionSourceFromDict(URLContrainer{
					"hand1": {
						URLTemplate: "",
						Parameters: ParamsDescription{
							"Name1": ParamInfo{
								Name:        "Name1",
								Help:        "Help to Name1",
								Type:        IntegerType,
								Destination: QueryPlaced,
							},
							"Name2": ParamInfo{
								Name:        "Name2",
								Help:        "Help to Name2",
								Type:        StringType,
								Destination: URLPlaced,
							},
						},
						Body:    "",
						URLName: "name",
					},
				}),
				nil,
			},
			TestCases{
				TestOutputParamHelp{
					HandName: "hand1",
					Helps: map[string]string{
						"Name1": "Name1(Integer)\tQuery Param\n\tHelp to Name1\n",
						"Name2": "Name2(String)\tURL Param\n\tHelp to Name2\n",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		input := testCase.Inp
		output := testCase.Out
		processor := NewURLProcessor(input.Descriptions, nil)
		for _, outCase := range output {
			out := outCase.(TestOutputParamHelp)
			handDescr, err := processor.GetHand(out.HandName)
			if err != nil {
				t.Errorf("Failed to get hand handler %s for hand %s", err.Error(), out.HandName)
				continue
			}
			for paramName, expect := range out.Helps {
				buf := new(bytes.Buffer)
				param, err := handDescr.GetParam(paramName)
				if err != nil {
					t.Errorf("Failed to get param handler %s for param %s for hand %s", err.Error(), paramName, out.HandName)
					continue
				}
				err = param.WriteHelp(buf)
				if err != nil {
					t.Errorf("Failed to write param handler %s for param %s for hand %s", err.Error(), paramName, out.HandName)
					continue
				}
				got := buf.String()
				if got != expect {
					t.Errorf("Wrong parameter help %s param name %s expected \n%s\n got\n%s", out.HandName, paramName, expect, got)
				}
			}
		}
	}
}

func TestHelp(t *testing.T) {
	// Проверяем справку для ручки полностью
	testCases := TestDescriptions{
		TestDescription{
			TestInput{
				NewDescriptionSourceFromDict(
					URLContrainer{
						"hand1": { // Присутствуют параметры всех типов
							URLTemplate: "http://example.com/entity/{entity_id}/v/{v}",
							Parameters: ParamsDescription{
								"entity_id": ParamInfo{
									Name:        "entity_id",
									Help:        "Help to entity_id",
									Type:        IntegerType,
									Destination: URLPlaced,
								},
								"v": ParamInfo{
									Name:        "v",
									Help:        "Help to v",
									Type:        StringType,
									Destination: URLPlaced,
								},
								"QueryParam1": ParamInfo{
									Name:        "QueryParam1",
									Help:        "Help to QueryParam1",
									Type:        IntegerType,
									Destination: QueryPlaced,
								},
								"QueryParam2": ParamInfo{
									Name:        "QueryParam2",
									Help:        "Help to QueryParam2",
									Type:        StringType,
									Destination: QueryPlaced,
								},
							},
							Body:    "",
							URLName: "hand1",
						},
						"handNoURLParams": { // Только параметры подставляемые как query
							URLTemplate: "http://example.com/entity",
							Parameters: ParamsDescription{
								"QueryParam1": ParamInfo{
									Name:        "QueryParam1",
									Help:        "Help to QueryParam1",
									Type:        IntegerType,
									Destination: QueryPlaced,
								},
								"QueryParam2": ParamInfo{
									Name:        "QueryParam2",
									Help:        "Help to QueryParam2",
									Type:        StringType,
									Destination: QueryPlaced,
								},
							},
							Body:    "",
							URLName: "handNoURLParams",
						},
						"handNoQueryParams": { // Только параметры подставляемые в URL
							URLTemplate: "http://example.com/entity/{entity_id}/v/{v}",
							Parameters: ParamsDescription{
								"entity_id": ParamInfo{
									Name:        "entity_id",
									Help:        "Help to entity_id",
									Type:        IntegerType,
									Destination: URLPlaced,
								},
								"v": ParamInfo{
									Name:        "v",
									Help:        "Help to v",
									Type:        StringType,
									Destination: URLPlaced,
								},
							},
							Body:    "",
							URLName: "handNoQueryParams",
						},
						"handNoParams": {
							URLTemplate: "http://example.com/entity/",
							Parameters:  ParamsDescription{},
							Body:        "",
							URLName:     "handNoParams",
						},
					},
				),
				nil,
			},
			TestCases{
				TestOutput{
					HandName: "hand1",
					Output: "Name: hand1\n" +
						"URL template: http://example.com/entity/{entity_id}/v/{v}\n" +
						"Parameters:\n" +
						"QueryParam1(Integer)\tQuery Param\n\tHelp to QueryParam1\n" +
						"QueryParam2(String)\tQuery Param\n\tHelp to QueryParam2\n" +
						"entity_id(Integer)\tURL Param\n\tHelp to entity_id\n" +
						"v(String)\tURL Param\n\tHelp to v\n",
				},
				TestOutput{
					HandName: "handNoURLParams",
					Output: "Name: handNoURLParams\n" +
						"URL template: http://example.com/entity\n" +
						"Parameters:\n" +
						"QueryParam1(Integer)\tQuery Param\n\tHelp to QueryParam1\n" +
						"QueryParam2(String)\tQuery Param\n\tHelp to QueryParam2\n",
				},
				TestOutput{
					HandName: "handNoQueryParams",
					Output: "Name: handNoQueryParams\n" +
						"URL template: http://example.com/entity/{entity_id}/v/{v}\n" +
						"Parameters:\n" +
						"entity_id(Integer)\tURL Param\n\tHelp to entity_id\n" +
						"v(String)\tURL Param\n\tHelp to v\n",
				},
			},
		},
	}
	for _, testCase := range testCases {
		input := testCase.Inp
		output := testCase.Out
		processor := NewURLProcessor(input.Descriptions, nil)
		for _, valI := range output {
			val := valI.(TestOutput)
			buf := new(bytes.Buffer)
			handProcessor, err := processor.GetHand(val.HandName)
			if err != nil {
				t.Errorf("Failed to get handler %s for hand %s", err.Error(), val.HandName)
			}
			err = handProcessor.WriteHelp(buf)
			if err != nil {
				t.Errorf("Failed to process help %s for hand %s", err.Error(), val.HandName)
			}
			got := buf.String()
			if got != val.Output {
				t.Errorf("Failed to get parameter help %s expected %s got %s", val.HandName, val.Output, got)
			}
		}
	}
}

func TestRender(t *testing.T) {
	// Проверяем как строится результат по запросу данных с ручки
	mustBuildRequest := func(method string, URL string) *http.Request {
		req, err := http.NewRequest(method, URL, nil)
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

	testCases := TestDescriptions{
		TestDescription{
			TestInput{
				NewDescriptionSourceFromDict(
					URLContrainer{
						"hand1": {
							URLTemplate: fmt.Sprintf("%s/entity/{entity_id}/v/{v}", serv.URL),
							Parameters: ParamsDescription{
								"entity_id": ParamInfo{
									Name:        "entity_id",
									Help:        "Help to entity_id",
									Type:        IntegerType,
									Destination: URLPlaced,
								},
								"v": ParamInfo{
									Name:        "v",
									Help:        "Help to v",
									Type:        StringType,
									Destination: URLPlaced,
								},
								"QueryParam1": ParamInfo{
									Name:        "QueryParam1",
									Help:        "Help to QueryParam1",
									Type:        IntegerType,
									Destination: QueryPlaced,
								},
								"QueryParam2": ParamInfo{
									Name:        "QueryParam2",
									Help:        "Help to QueryParam2",
									Type:        StringType,
									Destination: QueryPlaced,
								},
							},
							Body:    "Value of Value is {{ .value }}",
							URLName: "ValuableName",
						},
					},
				),
				func(rw http.ResponseWriter, req *http.Request) {
					err := json.NewEncoder(rw).Encode(map[string]interface{}{
						"value": "ValueForValue",
					})
					if err != nil {
						panic(err.Error())
					}
				},
			},
			TestCases{
				TestOutput{
					HandName: "hand1",
					Inp: map[string]interface{}{
						"entity_id":   1,
						"v":           "a",
						"QueryParam1": 2,
						"QueryParam2": "b",
					},
					Output: "Value of Value is ValueForValue",
					Requests: []*http.Request{
						mustBuildRequest("GET", fmt.Sprintf("%s/entity/{entity_id}/v/{v}", serv.URL)),
					},
					Err: nil,
				},
			},
		},
	}
	for _, testCase := range testCases {
		input := testCase.Inp
		output := testCase.Out
		handler = input.Handler
		processor := NewURLProcessor(input.Descriptions, serv.Client())
	KEYLOOP:
		for _, expectI := range output {
			expect := expectI.(TestOutput)
			buf := new(bytes.Buffer)
			handProcessor, err := processor.GetHand(expect.HandName)
			if err != nil {
				if err != expect.Err {
					continue KEYLOOP
				}
				t.Errorf("Failed to get param handler %s for hand %s", err.Error(), expect.HandName)
				continue
			}
			err = handProcessor.Process(buf, expect.Inp)
			if err != nil {
				if err != expect.Err {
					continue KEYLOOP
				}
				t.Errorf("Failed to process param handler %s for hand %s", err.Error(), expect.HandName)
				continue
			}
			got := buf.String()
			if got != expect.Output {
				t.Errorf("Failed to get parameter help %s expected %s got %s", expect.HandName, expect.Output, got)
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

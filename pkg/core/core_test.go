package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
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
						"handWithOptionalParam": { // Есть опциональный параметр
							URLTemplate: "http://example.com/entity",
							Parameters: ParamsDescription{
								"QueryParam1": ParamInfo{
									Name:        "QueryParam1",
									Help:        "Help to QueryParam1",
									Type:        IntegerType,
									Destination: QueryPlaced,
									Optional:    true,
								},
								"QueryParam2": ParamInfo{
									Name:        "QueryParam2",
									Help:        "Help to QueryParam2",
									Type:        StringType,
									Destination: QueryPlaced,
								},
							},
							Body:    "",
							URLName: "handWithOptionalParam",
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
				TestOutput{
					HandName: "handWithOptionalParam",
					Output: "Name: handWithOptionalParam\n" +
						"URL template: http://example.com/entity\n" +
						"Parameters:\n" +
						"QueryParam1(Integer)\tQuery Param\t[Optional]\n\tHelp to QueryParam1\n" +
						"QueryParam2(String)\tQuery Param\n\tHelp to QueryParam2\n",
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
							URLTemplate: fmt.Sprintf("%s/entity/{{.entity_id}}/v/{{.v}}", serv.URL),
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
							// TODO: сделать проверку для метаинформации
							Body: `Value of Value is {{ .responce.value }}
Debug url is {{ .meta.url }}
Params parm a is {{ .meta.params.entity_id }}`,
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
					Output: fmt.Sprintf(`Value of Value is ValueForValue
Debug url is %s/entity/1/v/a?QueryParam1=2&QueryParam2=b
Params parm a is 1`, serv.URL),
					Requests: []*http.Request{
						mustBuildRequest("GET", fmt.Sprintf("%s/entity/1/v/a?QueryParam1=2&QueryParam2=b", serv.URL)),
					},
					Err: nil,
				},
			},
		},

		TestDescription{
			TestInput{
				NewDescriptionSourceFromDict(
					URLContrainer{
						"hand1": {
							URLTemplate: fmt.Sprintf("%s/entity/{{.entity_id}}", serv.URL),
							Parameters: ParamsDescription{
								"entity_id": ParamInfo{
									Name:        "entity_id",
									Help:        "Help to entity_id",
									Type:        IntegerType,
									Destination: URLPlaced,
								},
								"QueryParam2": ParamInfo{
									Name:         "QueryParam2",
									Help:         "Help to QueryParam2",
									Type:         StringType,
									Destination:  QueryPlaced,
									DefaultValue: "queryparam2val",
								},
							},
							// TODO: сделать проверку для метаинформации
							Body: `Value of Value is {{ .responce.value }}
Debug url is {{ .meta.url }}
Params parm a is {{ .meta.params.entity_id }}`,
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
						"entity_id": 1,
					},
					Output: fmt.Sprintf(`Value of Value is ValueForValue
Debug url is %s/entity/1?QueryParam2=queryparam2val
Params parm a is 1`, serv.URL),
					Requests: []*http.Request{
						mustBuildRequest("GET", fmt.Sprintf("%s/entity/1?QueryParam2=queryparam2val", serv.URL)),
					},
					Err: nil,
				},
			},
		},
		TestDescription{
			TestInput{
				NewDescriptionSourceFromDict(
					URLContrainer{
						"hand1": {
							URLTemplate: fmt.Sprintf("%s/entity/{{.entity_id}}", serv.URL),
							Parameters: ParamsDescription{
								"entity_id": ParamInfo{
									Name:        "entity_id",
									Help:        "Help to entity_id",
									Type:        IntegerType,
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
									Optional:    true,
								},
							},
							// TODO: сделать проверку для метаинформации
							Body: `Value of Value is {{ .responce.value }}
Debug url is {{ .meta.url }}
Params parm a is {{ .meta.params.entity_id }}`,
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
						"QueryParam1": 2,
					},
					Output: fmt.Sprintf(`Value of Value is ValueForValue
Debug url is %s/entity/1?QueryParam1=2
Params parm a is 1`, serv.URL),
					Requests: []*http.Request{
						mustBuildRequest("GET", fmt.Sprintf("%s/entity/1?QueryParam1=2", serv.URL)),
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
			err = handProcessor.Process(context.Background(), buf, expect.Inp, log.NewEntry(&log.Logger{}))
			if err != nil {
				if err != expect.Err {
					continue KEYLOOP
				}
				t.Errorf("Failed to process param handler %s for hand %s", err.Error(), expect.HandName)
				continue
			}
			got := buf.String()
			if got != expect.Output {
				t.Errorf("Wrong hand output %s expected:\n[%s]\n got:\n[%s]\n", expect.HandName, expect.Output, got)
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

func TestRequireMentCheck(t *testing.T) {
	// проверяем логику вычисления опциональности параметров
	// сейчас все url параметры - обязательные
	// все query параметры - обязательные, но может быть проставлен параметр optional

	testCases := []struct {
		Param    ParamInfo
		Required bool
	}{
		{
			Param: ParamInfo{
				Name:        "optional_query_param",
				Help:        "",
				Type:        StringType,
				Destination: QueryPlaced,
				Optional:    true,
			},
			Required: false,
		},
		{
			Param: ParamInfo{
				Name:        "required_query_param",
				Help:        "",
				Type:        StringType,
				Destination: QueryPlaced,
			},
			Required: true,
		},
		{
			Param: ParamInfo{
				Name:        "required_url_placed_param",
				Help:        "",
				Type:        IntegerType,
				Destination: URLPlaced,
			},
			Required: true,
		},
		{
			Param: ParamInfo{
				Name:        "required_url_placed_param_with_ignored_optional",
				Help:        "",
				Type:        IntegerType,
				Destination: URLPlaced,
				Optional:    true,
			},
			Required: true,
		},
		{
			Param: ParamInfo{
				Name:         "required_url_placed_param_with_default_value",
				Help:         "",
				Type:         IntegerType,
				Destination:  URLPlaced,
				DefaultValue: "1",
			},
			Required: false,
		},
	}
	for _, testCase := range testCases {
		paramInfo := testCase.Param
		paramProc := NewParamProcessor(paramInfo)
		required := paramProc.IsRequired()
		if required != testCase.Required {
			t.Errorf("required constrained failedfor %s expected %v got %v", paramInfo.Name, testCase.Required, required)
		}
	}
}

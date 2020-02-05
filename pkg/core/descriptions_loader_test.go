package core

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type DescriptionParsingResults struct {
	Container URLContrainer
	Err       error
}

func cmpHandlers(container URLContrainer, dataSource DescriptionsSource) (bool, string) {
	for name, handler := range container {
		handlerRes, err := dataSource.GetByName(name)
		if err != nil {
			return false, fmt.Sprintf("On name %s: Got an unexpected error %s", name, err.Error())
		}
		if !reflect.DeepEqual(*handlerRes, handler) {
			return false, fmt.Sprintf("On name %s: \n Expected: \n %v \n\n Got:\n %v", name, handler, *handlerRes)
		}
	}
	return true, ""
}

func TestDescriptionsLoader(t *testing.T) {
	testCases := []struct {
		Input  string
		Name   string
		Output DescriptionParsingResults
	}{
		{
			Name: "simple result",
			Input: `{
				"ValuableName":{
				   "URL_template":"https://bash.im/entity/{entity_id}/v/{v}",
				   "params":{
					  "QueryParam1":{
						 "help":"Help to QueryParam1",
						 "name":"QueryParam1",
						 "destination":"query",
						 "type":"integer",
						 "optional":true
					  },
					  "QueryParam2":{
						 "help":"Help to QueryParam2",
						 "name":"QueryParam2",
						 "destination":"query",
						 "type":"string"
					  },
					  "entity_id":{
						 "help":"Help to entity_id",
						 "name":"entity_id",
						 "destination":"URL",
						 "type":"integer",
						 "default_value": "1"
					  },
					  "v":{
						 "help":"Help to v",
						 "name":"v",
						 "destination":"URL",
						 "type":"string"
					  }
				   },
				   "body":"Value of Value is {{ .value }}",
				   "name":"ValuableName",
				   "help":""
				}
			 }`,
			Output: DescriptionParsingResults{
				Container: URLContrainer{
					"ValuableName": {
						URLTemplate: fmt.Sprintf("%s/entity/{entity_id}/v/{v}", "https://bash.im"),
						Parameters: ParamsDescription{
							"entity_id": ParamInfo{
								Name:         "entity_id",
								Help:         "Help to entity_id",
								Type:         IntegerType,
								Destination:  URLPlaced,
								DefaultValue: "1",
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
								Optional:    true,
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
				Err: nil,
			},
		},
	}
	for _, testCase := range testCases {
		reader := strings.NewReader(testCase.Input)
		result, errResult := GetDescriptionSourceFromJSON(reader)
		if (errResult != nil) != (testCase.Output.Err != nil) {
			safeErrorPrint := func(errOut error) string {
				if errOut == nil {
					return "nil"
				}
				return errOut.Error()
			}
			t.Errorf("%s: Not equal errors, got %s, expected %s", testCase.Name, safeErrorPrint(errResult), safeErrorPrint(testCase.Output.Err))
		}

		if ok, msg := cmpHandlers(testCase.Output.Container, result); !ok {
			t.Errorf("%s: error on results comparision %s", testCase.Name, msg)
		}
	}
}

func TestDescriptionsLoaderYAML(t *testing.T) {
	testCases := []struct {
		Input  string
		Name   string
		Output DescriptionParsingResults
	}{
		{
			Name: "simple result",
			Input: `ValuableName:
  url_template: https://bash.im/entity/{entity_id}/v/{v}
  parameters:
    entity_id:
      help: Help to entity_id
      name: entity_id
      destination: URL
      type: integer
      default_value: 1
    query_param_1:
      help: Help to query_param_1
      name: query_param_1
      destination: query
      type: integer
      optional: true
    query_param_2:
      help: Help to query_param_2
      name: query_param_2
      destination: query
      type: string
    v:
      help: Help to v
      name: v
      destination: URL
      type: string
  body: Value of Value is {{ .value }}
  url_name: ValuableName
  help: ""`,
			Output: DescriptionParsingResults{
				Container: URLContrainer{
					"ValuableName": {
						URLTemplate: fmt.Sprintf("%s/entity/{entity_id}/v/{v}", "https://bash.im"),
						Parameters: ParamsDescription{
							"entity_id": ParamInfo{
								Name:         "entity_id",
								Help:         "Help to entity_id",
								Type:         IntegerType,
								Destination:  URLPlaced,
								DefaultValue: 1,
							},
							"v": ParamInfo{
								Name:        "v",
								Help:        "Help to v",
								Type:        StringType,
								Destination: URLPlaced,
							},
							"query_param_1": ParamInfo{
								Name:        "query_param_1",
								Help:        "Help to query_param_1",
								Type:        IntegerType,
								Destination: QueryPlaced,
								Optional:    true,
							},
							"query_param_2": ParamInfo{
								Name:        "query_param_2",
								Help:        "Help to query_param_2",
								Type:        StringType,
								Destination: QueryPlaced,
							},
						},
						Body:    "Value of Value is {{ .value }}",
						URLName: "ValuableName",
					},
				},
				Err: nil,
			},
		},
		{
			Name: "bad default value",
			Input: `ValuableName:
  url_template: https://bash.im/entity/{entity_id}/v/{v}
  parameters:
    entity_id:
      help: Help to entity_id
      name: entity_id
      destination: URL
      type: integer
      default_value: a
    query_param_1:
      help: Help to query_param_1
      name: query_param_1
      destination: query
      type: integer
      optional: true
  body: Value of Value is {{ .value }}
  url_name: ValuableName
  help: ""`,
			Output: DescriptionParsingResults{
				Container: URLContrainer{},
				Err: &ValidationError{
					Field: "",
					WrappedError: []error{
						&ValidationError{
							Field: "ValuableName",
							WrappedError: []error{
								&ValidationError{
									Field: "entity_id",
									WrappedError: []error{
										fmt.Errorf("Error on default value strconv.Atoi: parsing \"a\": invalid syntax"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	safeErrorPrint := func(errOut error) string {
		if errOut == nil {
			return "nil"
		}
		return errOut.Error()
	}
	for _, testCase := range testCases {
		reader := strings.NewReader(testCase.Input)
		result, errResult := GetDescriptionSourceFromYAML(reader)
		if (errResult != nil) != (testCase.Output.Err != nil) {
			t.Errorf("%s: Not equal errors, got %s, expected %s", testCase.Name, safeErrorPrint(errResult), safeErrorPrint(testCase.Output.Err))
			t.FailNow()
		}
		if (errResult != nil) && (testCase.Output.Err != nil) {
			// TODO: возможно придумать лучший способ сравнения
			if errResult.Error() != testCase.Output.Err.Error() {
				t.Errorf("%s: Not equal errors, got %s, expected %s", testCase.Name, safeErrorPrint(errResult), safeErrorPrint(testCase.Output.Err))
			}
		}

		if ok, msg := cmpHandlers(testCase.Output.Container, result); !ok {
			t.Errorf("%s: error on results comparision %s", testCase.Name, msg)
		}
	}
}

func TestParamsValidation(t *testing.T) {
	// проверяем проверку параметров
	testCases := []struct {
		Param  ParamInfo
		Errors string
	}{
		{
			Param: ParamInfo{
				Name:        "good_param",
				Help:        "",
				Type:        StringType,
				Destination: QueryPlaced,
				Optional:    true,
			},
			Errors: "",
		},
		{
			Param: ParamInfo{
				Name:        "optional_url_placed",
				Help:        "",
				Type:        StringType,
				Destination: URLPlaced,
				Optional:    true,
			},
			Errors: "Error(s) on processing entity optional_url_placed: UrlPlaced param can't be marked as optional\n",
		},
		{
			Param: ParamInfo{
				Name:         "invalid_value",
				Help:         "",
				Type:         IntegerType,
				Destination:  URLPlaced,
				DefaultValue: "a",
			},
			Errors: "Error(s) on processing entity invalid_value: Error on default value strconv.Atoi: parsing \"a\": invalid syntax\n",
		},
	}

	for _, testCase := range testCases {
		paramInfo := testCase.Param
		errs := validateParam(&paramInfo)
		err := newValidationError(paramInfo.Name, errs)
		if err == nil {
			if testCase.Errors == "" {
				continue
			}
			t.Errorf("required constrained failedfor %s expected %v got nil", paramInfo.Name, testCase.Errors)
		}
		if err.Error() != testCase.Errors {
			t.Errorf("required constrained failedfor %s expected %v got %v", paramInfo.Name, testCase.Errors, err.Error())
		}
	}
}

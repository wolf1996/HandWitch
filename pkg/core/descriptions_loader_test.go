package core

import (
	"strings"
	"testing"
)

type DescriptionParsingResults struct {
	Container URLContrainer
	Err       error
}

func TestDescriptionsLoader(t *testing.T) {
	testCases := []struct {
		Input  string
		Output DescriptionParsingResults
	}{
		{
			Input: `{
				"hand1":{
				   "URL_template":"https://bash.im/entity/{entity_id}/v/{v}",
				   "params":{
					  "QueryParam1":{
						 "help":"Help to QueryParam1",
						 "name":"QueryParam1",
						 "destination":"query",
						 "type":"integer"
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
						 "type":"integer"
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
				Container: URLContrainer{},
				Err:       nil,
			},
		},
	}
	for _, testCase := range testCases {
		reader := strings.NewReader(testCase.Input)
		_, errResult := ReadJSON(reader)
		if (errResult != nil) != (testCase.Output.Err != nil) {
			safeErrorPrint := func(errOut error) string {
				if errOut == nil {
					return "nil"
				}
				return errOut.Error()
			}
			t.Errorf("Not equal errors, got %s, expected %s", safeErrorPrint(errResult), safeErrorPrint(testCase.Output.Err))
		}
	}
}

package bot

import (
	"strings"
	"testing"
)

type DescriptionParsingResults struct {
	Output Config
	Err    error
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
				"formatting": "Markdown",
				"log_level": "Debug",
				"path": "/path/descriptions.json",
				"white_list": "/path/whitelist.json"
			 }`,
			Output: DescriptionParsingResults{
				Output: Config{
					Formatting: "Markdown",
					LogLevel:   "Debug",
					Path:       "/path/descriptions.json",
					WhiteList:  "/path/whitelist.json",
				},
				Err: nil,
			},
		},
	}
	for _, testCase := range testCases {
		reader := strings.NewReader(testCase.Input)
		result, errResult := GetConfigFromJSON(reader)
		if (errResult != nil) != (testCase.Output.Err != nil) {
			safeErrorPrint := func(errOut error) string {
				if errOut == nil {
					return "nil"
				}
				return errOut.Error()
			}
			t.Errorf("%s: Not equal errors, got %s, expected %s", testCase.Name, safeErrorPrint(errResult), safeErrorPrint(testCase.Output.Err))
		}

		if testCase.Output.Output == result {
			t.Errorf("%s: error on results comparision %v  %v", testCase.Name, testCase.Output.Output, result)
		}
	}
}

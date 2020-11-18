package bot

import (
	"fmt"
	"reflect"
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
				"white_list": "/path/whitelist.json",
				"proxy": "http://163.172.152.52:8811"
			 }`,
			Output: DescriptionParsingResults{
				Output: Config{
					Formatting: "Markdown",
					LogLevel:   "Debug",
					Path:       "/path/descriptions.json",
					WhiteList:  "/path/whitelist.json",
					Proxy:      "http://163.172.152.52:8811",
				},
				Err: nil,
			},
		},
		{
			Name: "with hook",
			Input: `{
				"formatting": "Markdown",
				"log_level": "Debug",
				"path": "/path/descriptions.json",
				"white_list": "/path/whitelist.json",
				"proxy": "http://163.172.152.52:8811",
				"hook": {
					"url_path": "https://url.com/secretpath",
					"cert": "certpath.pem",
					"key": "keypath.pem"
				}
			 }`,
			Output: DescriptionParsingResults{
				Output: Config{
					Formatting: "Markdown",
					LogLevel:   "Debug",
					Path:       "/path/descriptions.json",
					WhiteList:  "/path/whitelist.json",
					Proxy:      "http://163.172.152.52:8811",
					Hook: &HookInfo{
						URLPath: "https://url.com/secretpath",
						Cert:    "certpath.pem",
						Key:     "keypath.pem",
					},
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

		if reflect.DeepEqual(testCase.Output.Output, result) {
			t.Errorf("%s: error on results comparision %v  %v", testCase.Name, testCase.Output.Output, result)
			getHookStr := func(hook *HookInfo) string {
				repr := "nil"
				if hook != nil {
					repr = fmt.Sprintf("%v", *hook)
				}
				return repr
			}
			t.Logf("Hook expected: %s", getHookStr(testCase.Output.Output.Hook))
			t.Logf("Hook got: %s", getHookStr(result.Hook))
		}
	}
}

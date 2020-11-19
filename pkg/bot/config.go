package bot

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type HookInfo struct {
	URLPath string `json:"path"`
	Cert    string `json:"cert"`
	Key     string `json:"key"`
}

// Config main bot configuration
type Config struct {
	Formatting string    `json:"formatting"`
	LogLevel   string    `json:"log_level"`
	Path       string    `json:"path"`
	WhiteList  string    `json:"white_list"`
	Proxy      string    `json:"proxy"`
	Hook       *HookInfo `json:"hook"`
}

// GetConfigFromJSON parses config from reader as a JSON
func GetConfigFromJSON(reader io.Reader) (*Config, error) {
	var config Config
	// TODO: возможно стоит переделать на работу парсера, чтобы не вычитывать весь файл
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return &config, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return &config, err
	}
	return &config, nil
}

// GetDefaultConfig return default configuration
func GetDefaultConfig() *Config {
	config := Config{
		Formatting: "MarkDown",
		LogLevel:   "Info",
		Path:       "descriptions.yaml",
		WhiteList:  "whitelist.json",
	}
	return &config
}

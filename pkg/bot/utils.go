package bot

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetClientWithProxy Получает http client работающий через указанный
// proxy
func GetClientWithProxy(proxyString string) (*http.Client, error) {
	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: transport,
	}
	return client, nil
}

func getHandNameFromArguments(messageArguments string) (string, error) {
	rows := strings.Split(messageArguments, "\n")
	if len(rows) < 1 {
		return "", fmt.Errorf("Failed to get hand name: empty arguments")
	}
	handName := rows[0]
	return strings.TrimSpace(handName), nil
}

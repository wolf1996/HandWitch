package bot

import (
	"net/http"
	"net/url"
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

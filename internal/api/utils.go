package api

import (
	"downloader/internal/config"
	"net/http"
)

type ConfigurableTransport struct {
	Base http.RoundTripper
}

func (t *ConfigurableTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cfg := config.Get().Network.HTTP

	req.Header.Set("Authorization", Authorization())
	req.Header.Set("User-Agent", cfg.UserAgent)
	req.Header.Set("Referer", cfg.Referer)
	req.Header.Set("Origin", cfg.Origin)

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

var client = &http.Client{
	Transport: &ConfigurableTransport{
		Base: http.DefaultTransport,
	},
}

func Client() *http.Client {
	return client
}

package rest

import (
	"net/http"
)

func HTTPClientFor(config *Config) (*http.Client, error) {
	transport, err := TransportFor(config)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

func TransportFor(config *Config) (http.RoundTripper, error) {
	var rt = http.DefaultTransport

	rt = NewDebugRoundTripper(rt)
	if config.Username != "" {
		rt = NewBasicAuthRoundTripper(config.Username, config.Password, rt)
	}
	if len(config.UserAgent) > 0 {
		rt = NewUserAgentRoundTripper(config.UserAgent, rt)
	}

	return rt, nil
}

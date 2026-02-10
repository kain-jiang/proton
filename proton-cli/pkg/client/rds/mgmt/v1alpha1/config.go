package v1alpha1

import (
	"net/http"
)

// Config holds the attributes that can be passed to a rds mgmt client on
// initialization.
type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the
	// apiserver. If a URL is given then the (optional) Path of that URL
	// represents a prefix that must be appended to all request URIs used to
	// access the apiserver. This allows a frontend proxy to easily relocate all
	// of the apiserver endpoints.
	Host string

	// Server requires Basic authentication
	Username string
	Password string `datapolicy:"password"`
}

// ClientFor returns a Client that satisfies the requested attributes on a
// client Config object.
func ClientFor(config *Config) (*Client, error) {
	// Validate config.Host before constructing the transport/host so we can
	// fail fast.
	if _, err := DefaultServerURL(config.Host); err != nil {
		return nil, err
	}

	httpClient, err := HTTPClientFor(config)
	if err != nil {
		return nil, err
	}

	return ClientForConfigAndHTTPClient(config, httpClient)
}

// ClientForConfigAndHTTPClient returns a Client that satisfies the requested
// attributes on a client Config object.
//
// Unlike ClientFor, ClientForConfigAndHTTPClient allows to pass an http.Client
// that is shared between all the APIs.
//
// Note that the http client takes precedence over the transport values
// configured. The http client defaults to the `http.DefaultClient` if nil.
func ClientForConfigAndHTTPClient(config *Config, httpClient *http.Client) (*Client, error) {
	baseURL, err := DefaultServerURL(config.Host)
	if err != nil {
		return nil, err
	}
	return NewClient(baseURL, httpClient)
}

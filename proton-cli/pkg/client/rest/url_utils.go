package rest

import (
	"fmt"
	"net/url"
	"path"
)

func DefaultServerURLFor(config *Config) (*url.URL, string, error) {
	host := config.Host
	if host == "" {
		host = "localhost"
	}

	base := host
	hostURL, err := url.Parse(base)
	if err != nil || hostURL.Scheme == "" || hostURL.Host == "" {
		scheme := "http://"
		hostURL, err = url.Parse(scheme + base)
		if err != nil {
			return nil, "", err
		}
		if hostURL.Path != "" && hostURL.Path != "/" {
			return nil, "", fmt.Errorf("host must be a URL or a host:port pair: %q", base)
		}
	}

	var group, version string
	if config.GroupVersion != nil {
		group, version = config.GroupVersion.Group, config.GroupVersion.Version
	}

	versionedAPIPath := path.Join("/", config.APIPath, group, version)

	return hostURL, versionedAPIPath, nil
}

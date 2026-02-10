package rest

import (
	"fmt"
	"runtime"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

// Config holds the common attributes that can be passed to a Kubernetes client on
// initialization.
type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	// If a URL is given then the (optional) Path of that URL represents a prefix that must
	// be appended to all request URIs used to access the apiserver. This allows a frontend
	// proxy to easily relocate all of the apiserver endpoints.
	Host string
	// APIPath is a sub-path that points to an API root.
	APIPath string

	// GroupVersion is the API version to talk to. Must be provided when initializing
	// a RESTClient directly. When initializing a Client, will be set with the default
	// code version.
	GroupVersion *GroupVersion

	// Server requires Basic authentication
	Username string
	Password string `datapolicy:"password"`

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string
}

func RESTClientFor(config *Config) (*RESTClient, error) {
	baseURL, versionedAPIPath, err := DefaultServerURLFor(config)
	if err != nil {
		return nil, err
	}
	client, err := HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	return &RESTClient{
		base:             baseURL,
		versionedAPIPath: versionedAPIPath,
		Client:           client,
	}, nil
}

// adjustCommit returns sufficient significant figures of the commit's git hash.
func adjustCommit(c string) string {
	if len(c) == 0 {
		return "unknown"
	}
	if len(c) > 7 {
		return c[:7]
	}
	return c
}

// adjustVersion strips "alpha", "beta", etc. from version in form
// major.minor.patch-[alpha|beta|etc].
func adjustVersion(v string) string {
	if len(v) == 0 {
		return "unknown"
	}
	seg := strings.SplitN(v, "-", 2)
	return seg[0]
}

// DefaultProtonCLIUserAgent returns a User-Agent string built from static
// global vars.
func DefaultProtonCLIUserAgent() string {
	return fmt.Sprintf("proton-cli/%s (%s/%s) %s", adjustVersion(version.Get().GitVersion), runtime.GOOS, runtime.GOARCH, adjustCommit(version.Get().GitCommit))
}

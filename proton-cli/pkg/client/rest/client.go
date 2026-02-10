package rest

import (
	"net/http"
	"net/url"
	"strings"
)

type Interface interface {
	Verb(verb string) *Request
	Post() *Request
	Put() *Request
	Get() *Request
	Delete() *Request
}

// RESTClient imposes common Kubernetes API conventions on a set of resource paths.
// The baseURL is expected to point to an HTTP or HTTPS path that is the parent
// of one or more resources.  The server should return a decodable API resource
// object, or an api.Status object which contains information about the reason for
// any failure.
//
// Most consumers should use client.New() to get a Kubernetes API client.
type RESTClient struct {
	// base is the root URL ofr all invocations of the client
	base *url.URL
	// versionedAPIPath is a path segment connecting the base URL to the resource root
	versionedAPIPath string

	// Set specific behavior of the client.  If not set http.DefaultClient will be used.
	Client *http.Client
}

// NewRESTClient creates a new RESTClient. This client performs generic REST functions
// such as Get, Put, Post, and Delete on specified paths.
func NewRESTClient(baseURL *url.URL, versionedAPIPath string, client *http.Client) (*RESTClient, error) {
	base := *baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	return &RESTClient{
		base:             &base,
		versionedAPIPath: versionedAPIPath,

		Client: client,
	}, nil
}

// Verb begins a request with a verb (GET, POST, PUT, DELETE).
//
// Example usage of RESTClient's request building interface:
// c, err := NewRESTClient(...)
// if err != nil { ... }
// resp, err := c.Verb("GET").
//
//	Path("pods").
//	SelectorParam("labels", "area=staging").
//	Timeout(10*time.Second).
//	Do()
//
// if err != nil { ... }
// list, ok := resp.(*api.PodList)
func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb)
}

// Delete implements Interface.
func (c *RESTClient) Delete() *Request {
	return c.Verb(http.MethodDelete)
}

// Get begins a GET request. Short for c.Verb("GET").
func (c *RESTClient) Get() *Request {
	return c.Verb(http.MethodGet)
}

// Post implements Interface.
func (c *RESTClient) Post() *Request {
	return c.Verb(http.MethodPost)
}

// Put implements Interface.
func (c *RESTClient) Put() *Request {
	return c.Verb(http.MethodPut)
}

var _ Interface = (*RESTClient)(nil)

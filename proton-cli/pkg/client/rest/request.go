package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

// Request allows for building up a request to a server in a chained fashion.
// Any errors are stored until the end of your call, so you only have to check
// once.
type Request struct {
	c *RESTClient

	timeout time.Duration

	// generic components accessible via method setters
	verb       string
	pathPrefix string
	params     url.Values

	// structural elements of the request that are part of the REST API conventions
	resource     string
	resourceName string

	// output
	err error

	// only on of body / bodyBytes may be set. requests using body are not retriable.
	body      io.Reader
	bodyBytes []byte
}

// NewRequest creates a new request helper object for accessing runtime.Objects on a server.
func NewRequest(c *RESTClient) *Request {
	var pathPrefix string
	if c.base != nil {
		pathPrefix = path.Join("/", c.base.Path, c.versionedAPIPath)
	} else {
		pathPrefix = path.Join("/", c.versionedAPIPath)
	}

	var timeout time.Duration
	if c.Client != nil {
		timeout = c.Client.Timeout
	}

	r := &Request{
		c:          c,
		timeout:    timeout,
		pathPrefix: pathPrefix,
	}

	return r
}

// Verb sets the verb this request will use.
func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

// Resource sets the resource to access (<resource>/[ns/<namespace>/]<name>)
func (r *Request) Resource(resource string) *Request {
	if r.err != nil {
		return r
	}
	if len(r.resource) != 0 {
		r.err = fmt.Errorf("resource already set to %q, cannot change to %q", r.resource, resource)
		return r
	}
	r.resource = resource
	return r
}

// Name sets the name of a resource to access (<resource>/[ns/<namespace>/]<name>)
func (r *Request) Name(resourceName string) *Request {
	if r.err != nil {
		return r
	}
	if len(resourceName) == 0 {
		r.err = fmt.Errorf("resource name may not be empty")
		return r
	}
	if len(r.resourceName) != 0 {
		r.err = fmt.Errorf("resource name already set to %q, cannot change to %q", r.resourceName, resourceName)
		return r
	}
	r.resourceName = resourceName
	return r
}

// Body makes the request use obj as the body. Optional.
func (r *Request) Body(obj any) *Request {
	if r.err != nil {
		return r
	}
	switch t := obj.(type) {
	case io.Reader:
		r.body = t
		r.bodyBytes = nil
	case []byte:
		logger.NewLogger().Debugf("Request Body: %s", string(t))
		r.body = nil
		r.bodyBytes = t
	default:
		b, err := json.Marshal(obj)
		if err != nil {
			r.err = err
			return r
		}
		logger.NewLogger().Debugf("Request Body: %s", string(b))
		r.bodyBytes = b
	}
	return r
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	p := r.pathPrefix
	if len(r.resource) != 0 {
		p = path.Join(p, strings.ToLower(r.resource))
	}
	// Join trims trailing slashes, so preserve r.pathPrefix's trailing slash for backwards compatibility if nothing was changed
	if len(r.resourceName) != 0 {
		p = path.Join(p, r.resourceName)
	}

	finalURL := &url.URL{}
	if r.c.base != nil {
		*finalURL = *r.c.base
	}
	finalURL.Path = p

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	// timeout is handled specially here.
	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}
	finalURL.RawQuery = query.Encode()
	return finalURL
}

func (r *Request) Do(ctx context.Context) Result {
	if r.err != nil {
		return Result{err: r.err}
	}

	var body io.Reader
	switch {
	case r.body != nil && r.bodyBytes != nil:
		return Result{err: fmt.Errorf("cannot set both body and bodyBytes")}
	case r.body != nil:
		body = r.body
	case r.bodyBytes != nil:
		// Create a new reader specifically for this request.
		// Giving each request a dedicated reader allows retries to avoid races resetting the request body.
		body = bytes.NewReader(r.bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, r.verb, r.URL().String(), body)
	if err != nil {
		return Result{err: err}
	}

	c := r.c.Client
	if c == nil {
		c = http.DefaultClient
	}

	resp, err := c.Do(req)
	if err != nil {
		return Result{err: err}
	}
	defer resp.Body.Close()

	result := Result{
		contentType: resp.Header.Get("Content-Type"),
		statusCode:  resp.StatusCode,
	}
	if result.body, err = io.ReadAll(resp.Body); err != nil {
		result.err = err
		return result
	}

	logger.NewLogger().Debugf("Response Body: %s", string(result.body))

	// get structured error from response body
	if ee := new(Error); json.Unmarshal(result.body, ee) == nil && ee.Code != "" {
		result.err = ee
	}

	return result
}

// Result contains the result of calling Request.Do().
type Result struct {
	body        []byte
	contentType string
	err         error
	statusCode  int
}

func (r Result) Into(obj interface{}) error {
	if r.err != nil {
		return r.Error()
	}
	if len(r.body) == 0 {
		return fmt.Errorf("0-length response with status code: %d and content-type: %s", r.statusCode, r.contentType)
	}

	return json.Unmarshal(r.body, obj)
}

func (r Result) Error() error {
	return r.err
}

func (r Result) StatusCode() int {
	return r.statusCode
}

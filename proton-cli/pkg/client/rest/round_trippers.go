package rest

import (
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

// requestInfo keeps track of information about a request/response combination
type requestInfo struct {
	RequestHeaders http.Header `datapolicy:"token"`
	RequestVerb    string
	RequestURL     string

	ResponseStatus  string
	ResponseHeaders http.Header
	ResponseErr     error

	Duration time.Duration
}

// newRequestInfo creates a new RequestInfo based on an http request
func newRequestInfo(req *http.Request) *requestInfo {
	return &requestInfo{
		RequestURL:     req.URL.String(),
		RequestVerb:    req.Method,
		RequestHeaders: req.Header,
	}
}

// complete adds information about the response to the requestInfo
func (r *requestInfo) complete(response *http.Response, err error) {
	if err != nil {
		r.ResponseErr = err
		return
	}
	r.ResponseStatus = response.Status
	r.ResponseHeaders = response.Header
}

type debugRoundTripper struct {
	log logrus.FieldLogger
	rt  http.RoundTripper
}

// NewDebugRoundTripper wraps a round tripper and logs based on the current log level.
func NewDebugRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &debugRoundTripper{
		log: logger.NewLogger(),
		rt:  rt,
	}
}

var knownAuthTypes = map[string]bool{
	"bearer":    true,
	"basic":     true,
	"negotiate": true,
}

// maskValue masks credential content from authorization headers
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
func maskValue(key string, value string) string {
	if !strings.EqualFold(key, "Authorization") {
		return value
	}
	if len(value) == 0 {
		return ""
	}
	var authType string
	if i := strings.Index(value, " "); i > 0 {
		authType = value[0:i]
	} else {
		authType = value
	}
	if !knownAuthTypes[strings.ToLower(authType)] {
		return "<masked>"
	}
	if len(value) > len(authType)+1 {
		value = authType + " <masked>"
	} else {
		value = authType
	}
	return value
}

// RoundTrip implements http.RoundTripper.
func (rt *debugRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqInfo := newRequestInfo(req)

	rt.log.Debugf("%s %s", reqInfo.RequestVerb, reqInfo.RequestURL)
	rt.log.Debugf("Request Headers:")
	for key, values := range reqInfo.RequestHeaders {
		for _, value := range values {
			value = maskValue(key, value)
			rt.log.Debugf("    %s: %s", key, value)
		}
	}

	startTime := time.Now()

	resp, err := rt.rt.RoundTrip(req)
	reqInfo.Duration = time.Since(startTime)

	reqInfo.complete(resp, err)

	rt.log.Debugf("%s %s %s in %d milliseconds", reqInfo.RequestVerb, reqInfo.RequestURL, reqInfo.ResponseStatus, reqInfo.Duration.Milliseconds())
	rt.log.Debugf("Response Headers:")
	for key, values := range reqInfo.ResponseHeaders {
		for _, value := range values {
			value = maskValue(key, value)
			rt.log.Debugf("    %s: %s", key, value)
		}
	}

	return resp, err
}

var _ http.RoundTripper = &debugRoundTripper{}

type basicAuthRoundTripper struct {
	username string
	password string `datapolicy:"password"`
	rt       http.RoundTripper
}

var _ http.RoundTripper = &basicAuthRoundTripper{}

// NewBasicAuthRoundTripper will apply a BASIC auth authorization header to a
// request unless it has already been set.
func NewBasicAuthRoundTripper(username, password string, rt http.RoundTripper) http.RoundTripper {
	return &basicAuthRoundTripper{username, password, rt}
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.rt.RoundTrip(req)
	}
	req.SetBasicAuth(rt.username, rt.password)
	return rt.rt.RoundTrip(req)
}

type userAgentRoundTripper struct {
	agent string
	rt    http.RoundTripper
}

var _ http.RoundTripper = &userAgentRoundTripper{}

// NewUserAgentRoundTripper will add User-Agent header to a request unless it has already been set.
func NewUserAgentRoundTripper(agent string, rt http.RoundTripper) http.RoundTripper {
	return &userAgentRoundTripper{agent, rt}
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("User-Agent")) != 0 {
		return rt.rt.RoundTrip(req)
	}
	req.Header.Set("User-Agent", rt.agent)
	return rt.rt.RoundTrip(req)
}

package v1alpha1

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/utils/clock"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

// HTTPClientFor returns an http.Client that will provide the authentication or
// transport level security defined by the provided Config.
func HTTPClientFor(config *Config) (*http.Client, error) {
	transport, err := TransportFor(config)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

// TransportFor returns an http.RoundTripper that will provide the
// authentication or transport level security defined by the provided Config.
func TransportFor(config *Config) (http.RoundTripper, error) {
	var rt http.RoundTripper = http.DefaultTransport
	rt = NewDebuggingRoundTripper(rt)
	rt = NewUserAgentRoundTripper(DefaultUserAgent(), rt)
	rt = NewAuthRoundTripper(config.Username, config.Password, rt)
	rt = NewRetryOnNetworkErrorRoundTripper(rt)
	rt = NewRetryOnServerSideErrorRoundTripper(rt)
	return rt, nil
}

type authRoundTripper struct {
	username string
	password string
	rt       http.RoundTripper
}

// NewAuthRoundTripper will apply a BASIC auth authorization header to a request
// unless it has already been set.
func NewAuthRoundTripper(username, password string, rt http.RoundTripper) http.RoundTripper {
	return &authRoundTripper{username, password, rt}
}

const AuthHeaderKey = "admin-key"

func (rt *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get(AuthHeaderKey) != "" {
		return rt.rt.RoundTrip(req)
	}

	req.Header.Set(AuthHeaderKey, base64.StdEncoding.EncodeToString([]byte(rt.username+":"+rt.password)))
	return rt.rt.RoundTrip(req)
}

type debuggingRoundTripper struct {
	log logrus.FieldLogger
	rt  http.RoundTripper
}

// NewDebuggingRoundTripper allows to display in the logs output debug
// information on the API requests performed by the client.
func NewDebuggingRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &debuggingRoundTripper{logger.NewLogger(), rt}
}

func (rt *debuggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.log == nil {
		return rt.rt.RoundTrip(req)
	}

	rt.log.WithFields(logrus.Fields{"method": req.Method, "url": req.URL}).Debug("http request")
	for k, values := range req.Header {
		for _, v := range values {
			rt.log.WithFields(logrus.Fields{"key": k, "value": v}).Debug("http request header")
		}
	}

	startTime := time.Now()
	response, err := rt.rt.RoundTrip(req)
	duration := time.Since(startTime)

	if err != nil {
		return response, err
	}

	rt.log.WithFields(logrus.Fields{"method": req.Method, "url": req.URL, "duration": duration, "status": response.StatusCode}).Debug("http response")
	for k, values := range response.Header {
		for _, v := range values {
			rt.log.WithFields(logrus.Fields{"key": k, "value": v}).Debug("http response header")
		}
	}

	return response, err
}

type retryOnNetworkErrorRoundTripper struct {
	// This is useful for testing.
	clock clock.Clock

	rt     http.RoundTripper
	logger logrus.FieldLogger
}

// NewRetryOnNetworkErrorRoundTripper allows to resend request if there is a
// network error.
func NewRetryOnNetworkErrorRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &retryOnNetworkErrorRoundTripper{new(clock.RealClock), rt, logger.NewLogger()}
}

func (rt *retryOnNetworkErrorRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// attempt to deepcopy request body
	flagReqBodyIsNil := true
	var reqBody []byte
	if req.Body != nil {
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		flagReqBodyIsNil = false
	}
	req.Body = io.NopCloser(bytes.NewReader(bytesClone(reqBody)))
	resp, err = rt.rt.RoundTrip(req)
	for i := 0; i < 8 && errors.As(err, new(net.Error)); i++ {
		rt.logger.WithError(err).Debug("re-send http request on network error")
		rt.clock.Sleep(time.Second << i)
		if !flagReqBodyIsNil {
			req.Body = io.NopCloser(bytes.NewReader(bytesClone(reqBody)))
		}
		resp, err = rt.rt.RoundTrip(req)
	}
	return
}

type retryOnServerSideErrorRoundTripper struct {
	// This is useful for testing.
	clock clock.Clock

	rt     http.RoundTripper
	logger logrus.FieldLogger
}

func NewRetryOnServerSideErrorRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &retryOnServerSideErrorRoundTripper{new(clock.RealClock), rt, logger.NewLogger()}
}

func (rt *retryOnServerSideErrorRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// attempt to deepcopy request body
	flagReqBodyIsNil := true
	var reqBody []byte
	if req.Body != nil {
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		flagReqBodyIsNil = false
	}
	req.Body = io.NopCloser(bytes.NewReader(bytesClone(reqBody)))
	resp, err = rt.rt.RoundTrip(req)
	for i := 0; err == nil && i < 8 && 500 <= resp.StatusCode && resp.StatusCode < 600; i++ {
		rt.logger.WithField("status", resp.Status).Debug("re-send http request on server side error")
		rt.clock.Sleep(time.Second << i)
		if !flagReqBodyIsNil {
			req.Body = io.NopCloser(bytes.NewReader(bytesClone(reqBody)))
		}
		resp, err = rt.rt.RoundTrip(req)
	}
	return
}

type userAgentRoundTripper struct {
	agent string
	rt    http.RoundTripper
}

func NewUserAgentRoundTripper(agent string, rt http.RoundTripper) http.RoundTripper {
	return &userAgentRoundTripper{agent, rt}
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if len(req.Header.Get("user-agent")) != 0 {
		return rt.rt.RoundTrip(req)
	}
	req.Header.Set("user-agent", rt.agent)
	return rt.rt.RoundTrip(req)
}

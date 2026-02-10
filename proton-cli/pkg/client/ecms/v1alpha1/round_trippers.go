package v1alpha1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

type debuggingRoundTripper struct {
	logger *logrus.Logger
	rt     http.RoundTripper
}

func newDebuggingRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &debuggingRoundTripper{
		logger: logger.NewLogger(),
		rt:     rt,
	}
}

// RoundTrip implements http.RoundTripper.
func (rt *debuggingRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	rt.logger.Debugf("%s %s", req.Method, req.URL)
	rt.logger.Debug("Request Headers:")
	for k, values := range req.Header {
		for _, v := range values {
			rt.logger.Debugf("  %s: %s", k, v)
		}
	}
	s := time.Now()
	resp, err = rt.rt.RoundTrip(req)
	if err != nil {
		return
	}
	d := time.Since(s)

	rt.logger.Debugf("Response Status %s in %d milliseconds", resp.Status, d.Milliseconds())
	rt.logger.Debug("Response Headers:")
	for k, values := range resp.Header {
		for _, v := range values {
			rt.logger.Debugf("  %s: %s", k, v)
		}
	}

	return
}

var _ http.RoundTripper = &debuggingRoundTripper{}

type simpleAuthRoundTripper struct {
	username string

	// underlying round tripper
	rt http.RoundTripper
}

func newSimpleAuthRoundTripper(rt http.RoundTripper, username string) http.RoundTripper {
	return &simpleAuthRoundTripper{username: username, rt: rt}
}

// RoundTrip implements http.RoundTripper.
func (rt *simpleAuthRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	h := req.Header.Get("authorization")
	if h != "" {
		return rt.rt.RoundTrip(req)
	}

	epoch := time.Now().Unix()
	epoch_min := epoch - epoch%60
	raw := fmt.Appendf(nil, "%s:%d", rt.username, epoch_min)
	sum := sha256.Sum256(raw)
	password := hex.EncodeToString(sum[:])
	req.SetBasicAuth(rt.username, password)
	return rt.rt.RoundTrip(req)
}

var _ http.RoundTripper = &simpleAuthRoundTripper{}

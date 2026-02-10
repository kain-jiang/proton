package v1alpha1

import (
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clock "k8s.io/utils/clock/testing"
)

func Test_authRoundTripper_RoundTrip(t *testing.T) {
	header := make(http.Header)
	header.Set(AuthHeaderKey, "12450")

	fakeError := errors.New("something wrong")

	type args struct {
		username string
		password string
		rt       *FakeRoundTripper
	}
	tests := []struct {
		name      string
		args      args
		req       *http.Request
		wantValue string
		want      *http.Response
		wantErr   error
	}{
		{
			name:      "set",
			args:      args{username: "hello", password: "world", rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}}},
			req:       &http.Request{Header: make(http.Header)},
			wantValue: "aGVsbG86d29ybGQ=",
			want:      &http.Response{StatusCode: http.StatusOK},
		},
		{
			name:      "already exist",
			args:      args{username: "hello", password: "world", rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}}},
			req:       &http.Request{Header: header},
			wantValue: header.Get(AuthHeaderKey),
			want:      &http.Response{StatusCode: http.StatusOK},
		},
		{
			name:      "underlying error",
			args:      args{username: "hello", password: "world", rt: &FakeRoundTripper{Err: fakeError}},
			req:       &http.Request{Header: make(http.Header)},
			wantValue: "aGVsbG86d29ybGQ=",
			wantErr:   fakeError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tt.args.rt.Assertions = a
			rt := NewAuthRoundTripper(tt.args.username, tt.args.password, tt.args.rt)

			got, err := rt.RoundTrip(tt.req)
			if tt.wantErr != nil {
				a.ErrorIs(err, tt.wantErr)
			} else {
				a.NoError(err)
			}

			a.Equal(tt.want, got)

			tt.args.rt.AssertRequestHead(AuthHeaderKey, tt.wantValue)
		})
	}
}

func Test_debuggingRoundTripperRoundTripper_RoundTrip(t *testing.T) {
	var fakeError = errors.New("something wrong")
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name      string
		transport *FakeRoundTripper
		args      args
		want      *http.Response
		wantErr   error
	}{
		{
			name:      "example",
			transport: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}},
			args:      args{req: &http.Request{}},
			want:      &http.Response{StatusCode: http.StatusOK},
		},
		{
			name:      "underlying error",
			transport: &FakeRoundTripper{Err: fakeError},
			args:      args{req: &http.Request{}},
			wantErr:   fakeError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			tt.transport.Assertions = a

			rt := NewDebuggingRoundTripper(tt.transport)

			got, err := rt.RoundTrip(tt.args.req)
			if tt.wantErr != nil {
				a.ErrorIs(err, tt.wantErr)
			} else {
				a.NoError(err)
			}

			a.Equal(tt.want, got)
		})
	}
}

func Test_retryOnNetworkErrorRoundTripper_RoundTrip(t *testing.T) {
	var fakeError = errors.New("something wrong")
	var fakeNetworkError = &net.OpError{
		Op:   "dial",
		Net:  "tcp",
		Addr: &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 12450},
		Err:  fakeError,
	}

	type args struct {
		rt *FakeRoundTripper
	}
	tests := []struct {
		name      string
		args      args
		want      *http.Response
		wantCalls int
		wantErr   error
	}{
		{
			name:      "success",
			args:      args{rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}}},
			want:      &http.Response{StatusCode: http.StatusOK},
			wantCalls: 1,
		},
		{
			name:      "network error",
			args:      args{rt: &FakeRoundTripper{Err: fakeNetworkError}},
			wantCalls: 9,
			wantErr:   fakeNetworkError,
		},
		{
			name:      "other error",
			args:      args{rt: &FakeRoundTripper{Err: fakeError}},
			wantCalls: 1,
			wantErr:   fakeError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tt.args.rt.Assertions = a
			rt := NewRetryOnNetworkErrorRoundTripper(tt.args.rt)
			if rt, ok := rt.(*retryOnNetworkErrorRoundTripper); ok {
				rt.clock = clock.NewFakeClock(time.Now())
			}
			got, err := rt.RoundTrip(&http.Request{})
			if tt.wantErr != nil {
				a.ErrorIs(err, tt.wantErr)
			} else {
				a.NoError(err)
			}
			a.Equal(tt.want, got)
			tt.args.rt.AssertCalls(tt.wantCalls)
		})
	}
}

func Test_retryOnServerSideErrorRoundTripper_RoundTripper(t *testing.T) {
	var fakeError = errors.New("something wrong")
	// var fakeServerSideError = &Error{Code: 500000000, Message: "内部错误"}
	// var fakeClientSideError = &Error{Code: 401000000, Message: "没有权限"}
	type args struct {
		rt *FakeRoundTripper
	}
	tests := []struct {
		name      string
		args      args
		want      *http.Response
		wantCalls int
		wantErr   error
	}{
		{
			name:      "success",
			args:      args{rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}}},
			want:      &http.Response{StatusCode: http.StatusOK},
			wantCalls: 1,
		},
		{
			name:      "server side error",
			args:      args{rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusInternalServerError}}},
			want:      &http.Response{StatusCode: http.StatusInternalServerError},
			wantCalls: 9,
		},
		{
			name:      "client side error",
			args:      args{rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusUnauthorized}}},
			want:      &http.Response{StatusCode: http.StatusUnauthorized},
			wantCalls: 1,
		},
		{
			name:      "underlying error",
			args:      args{rt: &FakeRoundTripper{Err: fakeError}},
			wantCalls: 1,
			wantErr:   fakeError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tt.args.rt.Assertions = a
			rt := NewRetryOnServerSideErrorRoundTripper(tt.args.rt)
			if rt, ok := rt.(*retryOnServerSideErrorRoundTripper); ok {
				rt.clock = clock.NewFakeClock(time.Now())
			}
			got, err := rt.RoundTrip(&http.Request{})
			if tt.wantErr != nil {
				a.ErrorIs(err, tt.wantErr)
			} else {
				a.NoError(err)
			}
			a.Equal(tt.want, got)
			tt.args.rt.AssertCalls(tt.wantCalls)
		})
	}
}

func Test_userAgentRoundTripper(t *testing.T) {
	var header = make(http.Header)
	header.Set("user-agent", "hello/0.1")

	var fakeError = errors.New("something wrong")

	type args struct {
		agent string
		rt    *FakeRoundTripper
	}
	tests := []struct {
		name      string
		args      args
		req       *http.Request
		wantValue string
		want      *http.Response
		wantErr   error
	}{
		{
			name:      "set",
			args:      args{agent: "hello/1.0", rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}}},
			req:       &http.Request{Header: make(http.Header)},
			wantValue: "hello/1.0",
			want:      &http.Response{StatusCode: http.StatusOK},
		},
		{
			name:      "already exists",
			args:      args{rt: &FakeRoundTripper{Response: &http.Response{StatusCode: http.StatusOK}}},
			req:       &http.Request{Header: header},
			wantValue: "hello/0.1",
			want:      &http.Response{StatusCode: http.StatusOK},
		},
		{
			name:      "underlying error",
			args:      args{agent: "hello/1.0", rt: &FakeRoundTripper{Err: fakeError}},
			req:       &http.Request{Header: make(http.Header)},
			wantValue: "hello/1.0",
			wantErr:   fakeError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tt.args.rt.Assertions = a
			rt := NewUserAgentRoundTripper(tt.args.agent, tt.args.rt)
			got, err := rt.RoundTrip(tt.req)
			if tt.wantErr != nil {
				a.ErrorIs(err, tt.wantErr)
			} else {
				a.NoError(err)
			}
			a.Equal(tt.want, got)
			tt.args.rt.AssertRequestHead("user-agent", tt.wantValue)
		})
	}
}

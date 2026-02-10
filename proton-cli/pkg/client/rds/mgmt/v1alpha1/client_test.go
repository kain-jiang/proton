package v1alpha1

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

func TestNewClient(t *testing.T) {
	base := &url.URL{Scheme: "https", Host: "rds-mgmt.example.org:8888"}
	httpClient := &http.Client{Transport: http.DefaultTransport}

	client, err := NewClient(base, httpClient)
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range deep.Equal(client, &Client{Base: base, HTTP: httpClient, Logger: logger.NewLogger()}) {
		t.Errorf("NewClient() got != want: %v", d)
	}
}

type FakeRoundTripper struct {
	Assertions *assert.Assertions

	Request  *http.Request
	Response *http.Response
	Err      error

	// The number of times RoundTrip was called
	calls int
}

// RoundTrip implements http.RoundTripper.
func (rt *FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.calls++
	rt.Request = req
	return rt.Response, rt.Err
}

var _ http.RoundTripper = (*FakeRoundTripper)(nil)

func (rt *FakeRoundTripper) AssertRequestMethod(method string) bool {
	if !rt.Assertions.NotNil(rt.Request, "Request is missing") {
		return false
	}
	return rt.Assertions.Equal(method, rt.Request.Method)
}
func (rt *FakeRoundTripper) AssertRequestURL(u *url.URL) bool {
	if !rt.Assertions.NotNil(rt.Request, "Request is missing") {
		return false
	}
	return rt.Assertions.Equal(u, rt.Request.URL)
}
func (rt *FakeRoundTripper) AssertRequestHead(key, value string) bool {
	if !rt.Assertions.NotNil(rt.Request, "Request is missing") {
		return false
	}
	return rt.Assertions.Equal(value, rt.Request.Header.Get(key))
}
func (rt *FakeRoundTripper) AssertRequestBody(body io.Reader) bool {
	if !rt.Assertions.NotNil(rt.Request, "Request is missing") {
		return false
	}
	var expect, actual []byte
	var err error
	if body != nil {
		if expect, err = io.ReadAll(body); !rt.Assertions.NoError(err) {
			return false
		}
	}
	if rt.Request.Body != nil {
		if actual, err = io.ReadAll(rt.Request.Body); !rt.Assertions.NoError(err) {
			return false
		}
	}
	return rt.Assertions.Equal(expect, actual)
}
func (rt *FakeRoundTripper) AssertCalls(n int) bool {
	return rt.Assertions.Equal(n, rt.calls)
}

func TestClient_request(t *testing.T) {
	var log = logger.NewLogger()
	var header = make(http.Header)
	header.Add("content-type", "application/json")
	type strikebreaker struct {
		Name string `json:"name,omitempty"`
		Age  int    `json:"age,omitempty"`
	}
	type args struct {
		method           string
		apiPath          string
		input            interface{}
		expectStatusCode int
		output           interface{}
	}
	tests := []struct {
		name              string
		base              *url.URL
		args              args
		response          *http.Response
		wantRequestMethod string
		wantRequestURL    *url.URL
		wantRequestBody   io.Reader
		wantOutput        interface{}
	}{
		{
			name:              "get",
			base:              &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888"},
			args:              args{method: http.MethodGet, apiPath: "/api/proton-rds-mgmt/v2/strikebreakers/example", expectStatusCode: http.StatusOK, output: new(strikebreaker)},
			response:          &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(`{"name":"hello","age":12450}`)))},
			wantRequestMethod: http.MethodGet,
			wantRequestURL:    &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/api/proton-rds-mgmt/v2/strikebreakers/example"},
			wantOutput:        &strikebreaker{Name: "hello", Age: 12450},
		},
		{
			name:              "post",
			base:              &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888"},
			args:              args{method: http.MethodPost, apiPath: "/api/proton-rds-mgmt/v2/strikebreakers", input: &strikebreaker{Name: "hello", Age: 12450}, expectStatusCode: http.StatusCreated, output: new(strikebreaker)},
			response:          &http.Response{StatusCode: http.StatusCreated, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(`{"name":"hello","age":12450}`)))},
			wantRequestMethod: http.MethodPost,
			wantRequestURL:    &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/api/proton-rds-mgmt/v2/strikebreakers"},
			wantRequestBody:   bytes.NewReader([]byte(`{"name":"hello","age":12450}`)),
			wantOutput:        &strikebreaker{Name: "hello", Age: 12450},
		},
		{
			name:              "delete",
			base:              &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888"},
			args:              args{method: http.MethodDelete, apiPath: "/api/proton-rds-mgmt/v2/strikebreakers/example", expectStatusCode: http.StatusOK},
			response:          &http.Response{StatusCode: http.StatusOK, Header: header},
			wantRequestMethod: http.MethodDelete,
			wantRequestURL:    &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/api/proton-rds-mgmt/v2/strikebreakers/example"},
		},
		{
			name:              "base include path",
			base:              &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/endpoint/test"},
			args:              args{method: http.MethodPost, apiPath: "/api/proton-rds-mgmt/v2/strikebreakers", input: &strikebreaker{Name: "hello", Age: 12450}, expectStatusCode: http.StatusCreated, output: new(strikebreaker)},
			response:          &http.Response{StatusCode: http.StatusCreated, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(`{"name":"hello","age":12450}`)))},
			wantRequestMethod: http.MethodPost,
			wantRequestURL:    &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/endpoint/test/api/proton-rds-mgmt/v2/strikebreakers"},
			wantRequestBody:   bytes.NewReader([]byte(`{"name":"hello","age":12450}`)),
			wantOutput:        &strikebreaker{Name: "hello", Age: 12450},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}

			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}

			if err := c.request(context.TODO(), tt.args.method, tt.args.apiPath, tt.args.input, tt.args.expectStatusCode, tt.args.output); a.NoError(err) {
				a.Equal(tt.wantOutput, tt.args.output)
			}
			tr.AssertRequestMethod(tt.wantRequestMethod)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)
		})
	}
}

func TestClient_request_failure(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	var headerWithContentTypeEmpty = make(http.Header)
	headerWithContentTypeEmpty.Add("content-type", "")

	var headerWithContentTypePlainText = make(http.Header)
	headerWithContentTypePlainText.Add("content-type", "plain/text")

	type strikebreaker struct {
		Name string `json:"name,omitempty"`
		Age  int    `json:"age,omitempty"`
	}
	type args struct {
		method           string
		input            interface{}
		expectStatusCode int
		output           interface{}
	}
	tests := []struct {
		name         string
		args         args
		response     *http.Response
		responseErr  error
		errSubString string
	}{
		{
			name:         "json encoding failure",
			args:         args{input: func() { panic("unimplemented") }},
			errSubString: "json: unsupported type: func()",
		},
		{
			name:         "invalid method",
			args:         args{method: " "},
			errSubString: `net/http: invalid method " "`,
		},
		{
			name:         "connection refused",
			responseErr:  &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", errors.New("connection refused"))},
			errSubString: "connect: connection refused",
		},
		{
			name:         "invalid content type",
			response:     &http.Response{Header: headerWithContentTypeEmpty},
			errSubString: "mime: no media type",
		},
		{
			name:         "content type plain text",
			response:     &http.Response{Header: headerWithContentTypePlainText},
			errSubString: `invalid media type "plain/text"`,
		},
		{
			name:         "error response",
			args:         args{expectStatusCode: http.StatusOK},
			response:     &http.Response{StatusCode: http.StatusNotFound, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(`{"code":404012009,"message":"用户不存在"}`)))},
			errSubString: "404012009: 用户不存在",
		},
		{
			name:         "json decoding error failure",
			args:         args{expectStatusCode: http.StatusOK},
			response:     &http.Response{StatusCode: http.StatusNotFound, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(`{"code":404012009,"message":"用户不存在"`)))},
			errSubString: `decode response body to error fail: unexpected end of JSON input, raw: {"code":404012009,"message":"用户不存在"`,
		},
		{
			name:         "json decoding output failure",
			args:         args{expectStatusCode: http.StatusOK, output: new(strikebreaker)},
			response:     &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(`{"name":"12dora","age":"12450"}`)))},
			errSubString: "json: cannot unmarshal string into Go struct field strikebreaker.age of type int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response, Err: tt.responseErr}
			c := &Client{Base: new(url.URL), HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			err := c.request(context.TODO(), tt.args.method, "/a/b/c/", tt.args.input, tt.args.expectStatusCode, tt.args.output)
			a.ErrorContains(err, tt.errSubString)
		})
	}
}

func TestClient_CreateDatabase(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	type args struct {
		db *Database
	}
	tests := []struct {
		name             string
		base             *url.URL
		args             args
		response         *http.Response
		wantRequestURL   *url.URL
		wantRequestBody  io.Reader
		wantErr          bool
		wantErrSubString string
	}{
		{
			name:            "success",
			base:            &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:            args{db: &Database{DBName: "test", Charset: CharsetUTF8MB4, Collation: CollationUTF8MB4GeneralCI}},
			response:        &http.Response{StatusCode: http.StatusCreated, Header: header},
			wantRequestURL:  &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/dbs/test"},
			wantRequestBody: strings.NewReader(`{"db_name":"test","charset":"utf8mb4","collate":"utf8mb4_general_ci"}`),
		},
		{
			name:             "failure",
			base:             &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:             args{db: &Database{DBName: "test", Charset: CharsetUTF8MB4, Collation: CollationUTF8MB4GeneralCI}},
			response:         &http.Response{StatusCode: http.StatusInternalServerError, Header: header, Body: io.NopCloser(strings.NewReader(`{"code":500000000,"message":"内部错误"}`))},
			wantRequestURL:   &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/dbs/test"},
			wantRequestBody:  strings.NewReader(`{"db_name":"test","charset":"utf8mb4","collate":"utf8mb4_general_ci"}`),
			wantErr:          true,
			wantErrSubString: "500000000: 内部错误",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}
			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			if err := c.CreateDatabase(context.TODO(), tt.args.db); tt.wantErr {
				a.ErrorContains(err, tt.wantErrSubString)
			} else {
				a.NoError(err)
			}

			tr.AssertRequestMethod(http.MethodPut)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)
		})
	}
}

func TestClient_DeleteDatabase(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	type args struct {
		name string
	}
	tests := []struct {
		name             string
		base             *url.URL
		args             args
		response         *http.Response
		wantRequestURL   *url.URL
		wantRequestBody  io.Reader
		wantErr          bool
		wantErrSubString string
	}{
		{
			name:           "success",
			base:           &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:           args{name: "test"},
			response:       &http.Response{StatusCode: http.StatusNoContent, Header: header},
			wantRequestURL: &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/dbs/test"},
		},
		{
			name:             "failure",
			base:             &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:             args{name: "test"},
			response:         &http.Response{StatusCode: http.StatusInternalServerError, Header: header, Body: io.NopCloser(strings.NewReader(`{"code":500000000,"message":"内部错误"}`))},
			wantRequestURL:   &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/dbs/test"},
			wantErr:          true,
			wantErrSubString: "500000000: 内部错误",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}
			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			if err := c.DeleteDatabase(context.TODO(), tt.args.name); tt.wantErr {
				a.ErrorContains(err, tt.wantErrSubString)
			} else {
				a.NoError(err)
			}

			tr.AssertRequestMethod(http.MethodDelete)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)
		})
	}
}

func TestClient_ListDatabases(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	tests := []struct {
		name             string
		base             *url.URL
		response         *http.Response
		wantRequestURL   *url.URL
		wantRequestBody  io.Reader
		wantDatabases    []Database
		wantErr          bool
		wantErrSubString string
	}{
		{
			name:           "success",
			base:           &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			response:       &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(`[{"db_name":"test_0","charset":"utf8mb4","collate":"utf8mb4_general_ci"},{"db_name":"test_1","charset":"utf8mb4","collate":"utf8mb4_general_ci"}]`))},
			wantRequestURL: &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/dbs"},
			wantDatabases:  []Database{{DBName: "test_0", Charset: "utf8mb4", Collation: "utf8mb4_general_ci"}, {DBName: "test_1", Charset: "utf8mb4", Collation: "utf8mb4_general_ci"}},
		},
		{
			name:             "failure",
			base:             &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			response:         &http.Response{StatusCode: http.StatusInternalServerError, Header: header, Body: io.NopCloser(strings.NewReader(`{"code":500000000,"message":"内部错误"}`))},
			wantRequestURL:   &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/dbs"},
			wantErr:          true,
			wantErrSubString: "500000000: 内部错误",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}
			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			databases, err := c.ListDatabases(context.TODO())
			if tt.wantErr {
				a.ErrorContains(err, tt.wantErrSubString)
			} else {
				a.NoError(err)
			}
			a.Equal(tt.wantDatabases, databases)

			tr.AssertRequestMethod(http.MethodGet)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)
		})
	}
}

func TestClient_CreateUser(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	type args struct {
		username string
		password string
	}
	tests := []struct {
		name             string
		base             *url.URL
		args             args
		response         *http.Response
		wantRequestURL   *url.URL
		wantRequestBody  io.Reader
		wantErr          bool
		wantErrSubString string
	}{
		{
			name:            "success",
			base:            &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:            args{username: "hello", password: "world"},
			response:        &http.Response{StatusCode: http.StatusOK, Header: header},
			wantRequestURL:  &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/users/hello"},
			wantRequestBody: strings.NewReader(`{"password":"d29ybGQ="}`),
		},
		{
			name:             "failure",
			base:             &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:             args{username: "hello", password: "world"},
			response:         &http.Response{StatusCode: http.StatusInternalServerError, Header: header, Body: io.NopCloser(strings.NewReader(`{"code":500000000,"message":"内部错误"}`))},
			wantRequestURL:   &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/users/hello"},
			wantRequestBody:  strings.NewReader(`{"password":"d29ybGQ="}`),
			wantErr:          true,
			wantErrSubString: "500000000: 内部错误",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}
			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			if err := c.CreateUser(context.TODO(), tt.args.username, tt.args.password); tt.wantErr {
				a.ErrorContains(err, tt.wantErrSubString)
			} else {
				a.NoError(err)
			}
			tr.AssertRequestMethod(http.MethodPut)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)

		})
	}
}

func TestClient_ListUsers(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	tests := []struct {
		name             string
		base             *url.URL
		response         *http.Response
		wantRequestURL   *url.URL
		wantRequestBody  io.Reader
		wantUsers        []User
		wantErr          bool
		wantErrSubString string
	}{
		{
			name:           "success",
			base:           &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			response:       &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(`[{"username":"user_0","privileges":[{"db_name":"*","privilege_type":"None"}],"ssl_type":"None"},{"username":"user_1","privileges":[{"db_name":"*","privilege_type":"None"},{"db_name":"db_0","privilege_type":"ReadWrite"}],"ssl_type":"Any"}]`))},
			wantRequestURL: &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/users"},
			wantUsers: []User{
				{Username: "user_0", Privileges: []Privilege{{DBName: "*", PrivilegeType: PrivilegeNone}}, SSLType: SSLNone},
				{Username: "user_1", Privileges: []Privilege{{DBName: "*", PrivilegeType: PrivilegeNone}, {DBName: "db_0", PrivilegeType: PrivilegeReadWrite}}, SSLType: SSLAny},
			},
		},
		{
			name:             "failure",
			base:             &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			response:         &http.Response{StatusCode: http.StatusInternalServerError, Header: header, Body: io.NopCloser(strings.NewReader(`{"code":500000000,"message":"内部错误"}`))},
			wantRequestURL:   &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/users"},
			wantErr:          true,
			wantErrSubString: "500000000: 内部错误",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}
			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			users, err := c.ListUsers(context.TODO())
			if tt.wantErr {
				a.ErrorContains(err, tt.wantErrSubString)
			} else {
				a.NoError(err)
			}
			a.Equal(tt.wantUsers, users)

			tr.AssertRequestMethod(http.MethodGet)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)
		})
	}
}

func TestClient_PatchUserPrivileges(t *testing.T) {
	var log = logger.NewLogger()

	var header = make(http.Header)
	header.Add("content-type", "application/json")

	type args struct {
		username   string
		privileges []Privilege
	}
	tests := []struct {
		name             string
		base             *url.URL
		args             args
		response         *http.Response
		wantRequestURL   *url.URL
		wantRequestBody  io.Reader
		wantErr          bool
		wantErrSubString string
	}{
		{
			name:            "success",
			base:            &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:            args{username: "hello", privileges: []Privilege{{DBName: "test", PrivilegeType: PrivilegeReadOnly}}},
			response:        &http.Response{StatusCode: http.StatusNoContent, Header: header},
			wantRequestURL:  &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/users/hello/privileges"},
			wantRequestBody: strings.NewReader(`[{"db_name":"test","privilege_type":"ReadOnly"}]`),
		},
		{
			name:             "failure",
			base:             &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix"},
			args:             args{username: "hello", privileges: []Privilege{{DBName: "test", PrivilegeType: PrivilegeReadOnly}}},
			response:         &http.Response{StatusCode: http.StatusInternalServerError, Header: header, Body: io.NopCloser(strings.NewReader(`{"code":500000000,"message":"内部错误"}`))},
			wantRequestURL:   &url.URL{Scheme: "http", Host: "rds-mgmt.example.org:8888", Path: "/prefix/api/proton-rds-mgmt/v2/users/hello/privileges"},
			wantRequestBody:  strings.NewReader(`[{"db_name":"test","privilege_type":"ReadOnly"}]`),
			wantErr:          true,
			wantErrSubString: "500000000: 内部错误",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			tr := &FakeRoundTripper{Assertions: a, Response: tt.response}
			c := &Client{Base: tt.base, HTTP: &http.Client{Transport: tr}, Logger: log.WithField("test", t.Name())}
			if err := c.PatchUserPrivileges(context.TODO(), tt.args.username, tt.args.privileges); tt.wantErr {
				a.ErrorContains(err, tt.wantErrSubString)
			} else {
				a.NoError(err)
			}
			tr.AssertRequestMethod(http.MethodPatch)
			tr.AssertRequestURL(tt.wantRequestURL)
			tr.AssertRequestBody(tt.wantRequestBody)

		})
	}
}

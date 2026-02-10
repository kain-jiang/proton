package client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type HttpClient struct {
	client *http.Client
}

// NewHttpClient 创建HTTP客户端对象
func NewHttpClient(timeout time.Duration) *HttpClient {

	return &HttpClient{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: timeout * time.Second,
		},
	}
}

// Get http client get
func (c *HttpClient) Get(url string, headers map[string]string) (respCode int, respParam interface{}, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Post http client post
func (c *HttpClient) Post(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	sourceReq := c.prepareBody(headers, reqParam)
	req, err := http.NewRequest("POST", url, sourceReq)
	if err != nil {
		return
	}
	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Put http client put
func (c *HttpClient) Put(url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	reqBody, err := jsoniter.Marshal(reqParam)
	if err != nil {
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Delete http client delete
func (c *HttpClient) Delete(url string, headers map[string]string) (respParam interface{}, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	_, respParam, err = c.httpDo(req, headers)
	return
}

// httpDo
func (c *HttpClient) httpDo(req *http.Request, headers map[string]string) (respCode int, respParam interface{}, err error) {
	if c.client == nil {
		return 0, nil, errors.New("http client is unavailable")
	}

	c.addHeaders(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	respCode = resp.StatusCode

	if len(body) != 0 {
		err = jsoniter.Unmarshal(body, &respParam)
	}
	return
}
func (c *HttpClient) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
}

func (c *HttpClient) prepareBody(headers map[string]string, reqParam interface{}) (body io.Reader) {
	var contentType string
	if nil != headers {
		if v, ok := headers["Content-Type"]; ok {
			contentType = v
		}
	}
	switch contentType {
	case "application/x-www-form-urlencoded":
		req := reqParam.(map[string]interface{})
		if nil != req {
			reader := make([]string, 0)
			for k, v := range req {
				reader = append(reader, fmt.Sprintf("%v=%v", k, v))
			}
			return strings.NewReader(strings.Join(reader, "&"))
		}
		return
	case "application/octet-stream":
		return bytes.NewReader(reqParam.([]byte))
	case "application/xml":
		req := reqParam.(string)
		if "" != req {
			return bytes.NewReader([]byte(req))
		}
	default:
		reqBody, err := jsoniter.Marshal(reqParam)
		if nil != err {
			return
		}
		return bytes.NewReader(reqBody)
	}
	return
}

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"taskrunner/trait"
)

func DoJsonHTTP(cli *http.Client, method string, url string, obj any, headers map[string]string, receiver any, wantCode int) *trait.Error {
	var body io.Reader
	if obj != nil {
		bs, rerr := json.Marshal(obj)
		if rerr != nil {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      rerr,
				Detail:   fmt.Sprintf("encode obj for [%s] [%s] error", method, url),
			}
		}
		body = bytes.NewReader(bs)
	}

	req, rerr := http.NewRequest(method, url, body)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      rerr,
			Detail:   fmt.Sprintf("new http request for [%s] [%s] error", method, url),
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, rerr := cli.Do(req)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      rerr,
		}
	}
	defer resp.Body.Close()
	msg, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ECNetUnknow,
			Detail:   fmt.Sprintf("read response from [%s] [%s] error", method, url),
			Err:      rerr,
		}
	}

	switch resp.StatusCode {
	case 404:
		return &trait.Error{
			Internal: trait.ErrNotFound,
			Detail:   fmt.Sprintf("resource for [%s] [%s] not found", method, url),
			Err:      errors.New(string(msg)),
		}
	case 403:
		return &trait.Error{
			Internal: trait.ECNoAuthorized,
			Detail:   fmt.Sprintf("resource: [%s] [%s], status_code: [%d]", method, url, resp.StatusCode),
			Err:      errors.New(string(msg)),
		}
	case 401:
		return &trait.Error{
			Internal: trait.ECInvalidAuthorized,
			Detail:   fmt.Sprintf("resource: [%s] [%s], status_code: [%d]", method, url, resp.StatusCode),
			Err:      errors.New(string(msg)),
		}
	case wantCode:
		if receiver != nil {
			if rerr := json.Unmarshal(msg, receiver); rerr != nil {
				return &trait.Error{
					Internal: trait.ErrComponentDecodeError,
					Detail:   fmt.Sprintf("decode resource [%s] [%s] error", method, url),
					Err:      rerr,
				}
			}
		}
		return nil
	default:
		return &trait.Error{
			Internal: trait.ECNULL,
			Detail:   resp.StatusCode,
			Err:      errors.New(string(msg)),
		}
	}
}

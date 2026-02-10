package exec

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
)

type Client struct {
	HTTPClient *http.Client

	Base *url.URL
}

// Execute implements Interface.
func (c *Client) Execute(ctx context.Context, command []string, input io.Reader, opts ExecuteOptions) (out []byte, err error) {
	// generate url
	var u url.URL = *c.Base
	var q = make(url.Values)
	q.Set("stdin", strconv.FormatBool(opts.Stdin))
	q.Set("stdout", strconv.FormatBool(opts.Stdout))
	q.Set("stderr", strconv.FormatBool(opts.Stderr))
	for _, cc := range command {
		q.Add("command", cc)
	}
	u.RawQuery = q.Encode()
	// body
	if input == nil {
		input = http.NoBody
	}
	// create http request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), input)
	if err != nil {
		return
	}
	// send http request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	// ok
	case http.StatusOK:
		out, err = io.ReadAll(resp.Body)
		if rc, err := strconv.Atoi(resp.Header.Get("x-exit-code")); err != nil {
			return nil, fmt.Errorf("invalid exit code: %v", resp.Header.Get("x-exit-code"))
		} else if rc != 0 {
			return nil, &ExitError{Output: out, ExitCode: rc}
		}
		return
	// not found
	case http.StatusNotFound:
		return nil, &exec.Error{Name: command[0], Err: exec.ErrNotFound}
	default:
		err = fmt.Errorf("invalid status: %s", resp.Status)
		return
	}
}

var _ Interface = &Client{}

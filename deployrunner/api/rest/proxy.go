package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// // Client rest http client for engine
// type Client struct {
// 	client
// }

// // StartJob start the job
// func (c *Client) StartJob(ctx context.Context, jid int, owner int) error {
// 	return c.StartJob(ctx, jid, owner)
// }

// // CancelJob stop job execute
// func (c *Client) CancelJob(ctx context.Context, jid int, owner int) error {
// 	return c.CancelJob(ctx, jid, owner)
// }

// client rest http client for engine
type client struct {
	host     string
	interval time.Duration
	maxRetry int
}

// StartJob start the job
func (c *client) StartJob(ctx context.Context, jid int, owner int) *HTTPError {
	url := fmt.Sprintf(c.host+"/job/executor/%d", owner, jid)
	res, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return UnknownError.From(err.Error())
	}
	return c.doRequest(ctx, res)
}

// CancelJob stop job execute
func (c *client) CancelJob(ctx context.Context, jid int, owner int) *HTTPError {
	url := fmt.Sprintf(c.host+"/job/executor/%d", owner, jid)
	res, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return UnknownError.From(err.Error())
	}
	return c.doRequest(ctx, res)
}

func (c *client) doRequest(ctx context.Context, res *http.Request) *HTTPError {
	res = res.WithContext(ctx)
	cli := http.DefaultClient
	var resp *http.Response
	var err0 error
	if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
		resp, err0 = cli.Do(res)
		if err0 != nil {
			switch err0 {
			case context.Canceled:
				return false, &trait.Error{
					Internal: trait.ECContextEnd,
					Err:      engineProxyTimeoutError.From(err0.Error()),
				}
			case context.DeadlineExceeded:
				return false, &trait.Error{
					Internal: trait.ECContextEnd,
					Err:      ClientTimeoutError.From(err0.Error()),
				}
			}
			return true, &trait.Error{
				Internal: trait.ECNetUnknow,
				Err:      engineProxyConnectError.From(err0.Error()),
			}

		}
		return false, nil
	}, c.maxRetry, c.interval); err != nil {
		return err.Err.(*HTTPError)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err0 := &HTTPError{
			StatusCode: resp.StatusCode,
		}
		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			return UnknownError.From(err.Error())
		}
		err = json.Unmarshal(bs, err0)
		if err != nil {
			return UnknownError.From(err.Error())
		}
		return err0
	}
	return nil
}

// engineProxy proxy the request into the engine
// TODO sink into cluster manager
type engineProxy struct {
	client
}

func (e *engineProxy) StartJob(ctx *gin.Context, jid, owner int) {
	// TODO sink
	err := e.client.StartJob(ctx, jid, owner)
	if err != nil {
		err.AbortGin(ctx)
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}

func (e *engineProxy) CancelJob(ctx *gin.Context, jid, owner int) {
	// TODO sink
	err := e.client.CancelJob(ctx, jid, owner)
	if err != nil {
		err.AbortGin(ctx)
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}

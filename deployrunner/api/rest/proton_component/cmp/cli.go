package cmp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"component-manage/pkg/models/types"
	"taskrunner/pkg/utils"
	"taskrunner/trait"
)

// TODO move this into module

// Client operate component by proton component mamangement
type Client struct {
	BaseUrl string
	cli     *http.Client
}

func NewClient(ns string) *Client {
	if ns == "" {
		ns = "resource"
	}
	return &Client{
		cli: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns: 3,
			},
		},
		BaseUrl: fmt.Sprintf("http://component-manage.%s/api/component-manage/v1/components/release", ns),
	}
}

type ComponentGeneric interface {
	types.ComponentOpensearch |
		types.ComponentKafka |
		types.ComponentZookeeper |
		types.ComponentMongoDB |
		types.ComponentETCD |
		types.ComponentNebula |
		types.ComponentRedis |
		types.ComponentPolicyEngine |
		types.ComponentMariaDB
}

type ComponentInstance[T ComponentGeneric] struct {
	ComponentInstanceMeta
	Instance T
}
type ComponentInstanceMeta struct {
	trait.System `json:",inline"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

// upgradeKafa upgrade kafka release
func Upgrade[T ComponentGeneric](ctx context.Context, c *Client, obj *ComponentInstance[T]) *trait.Error {
	bs, rerr := json.Marshal(obj.Instance)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("encode %s obj fail", obj.Name),
			Internal: trait.ErrParam,
		}
	}
	req, rerr := http.NewRequest(http.MethodPut, fmt.Sprintf(c.BaseUrl+"/%s/%s", obj.Type, obj.Name), bytes.NewReader(bs))
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("create http request for upgrade %s fail", obj.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	req = req.WithContext(ctx)
	resp, rerr := http.DefaultClient.Do(req)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("http request for upgrade %s fail", obj.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	defer resp.Body.Close()
	bs, rerr = io.ReadAll(resp.Body)
	if rerr != nil {
		return &trait.Error{
			Err:      fmt.Errorf("'compnent-management' upgrade %s request return status code [%d], msg: [%s]", obj.Name, resp.StatusCode, string(bs)),
			Internal: trait.ECNetUnknow,
			Detail:   "read response from compnent-management http request",
		}
	}
	if resp.StatusCode == 400 {
		return &trait.Error{
			Internal: trait.ErrParam,
			Detail:   string(bs),
		}
	}
	if resp.StatusCode != 200 {
		return &trait.Error{
			Err:      fmt.Errorf("upgrade %s request return status code [%d], msg: [%s]", obj.Name, resp.StatusCode, string(bs)),
			Internal: trait.ECHTTPAPIRawError,
			Detail:   resp.StatusCode,
		}
	}
	if rerr := json.Unmarshal(bs, &obj.Instance); rerr != nil {
		return &trait.Error{
			Err:      fmt.Errorf("decode instance %s error %s, resp: %s", obj.Name, rerr.Error(), string(bs)),
			Internal: trait.ErrComponentDecodeError,
			Detail:   "read response from compnent-management http request",
		}
	}
	return nil
}

// upgradeKafa upgrade kafka release
func New[T ComponentGeneric](ctx context.Context, c *Client, obj *ComponentInstance[T]) *trait.Error {
	bs, rerr := json.Marshal(obj.Instance)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("encode %s obj fail", obj.Name),
			Internal: trait.ErrParam,
		}
	}
	req, rerr := http.NewRequest(http.MethodPost, fmt.Sprintf(c.BaseUrl+"/%s/%s", obj.Type, obj.Name), bytes.NewReader(bs))
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("create http request for upgrade %s fail", obj.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	req = req.WithContext(ctx)
	resp, rerr := http.DefaultClient.Do(req)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("http request for upgrade %s fail", obj.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	defer resp.Body.Close()
	bs, rerr = io.ReadAll(resp.Body)
	if rerr != nil {
		return &trait.Error{
			Err:      fmt.Errorf("'compnent-management' upgrade %s request return status code [%d], msg: [%s]", obj.Name, resp.StatusCode, string(bs)),
			Internal: trait.ECNetUnknow,
			Detail:   "decode response from compnent-management http request",
		}
	}
	if resp.StatusCode == 400 {
		return &trait.Error{
			Internal: trait.ErrParam,
			Detail:   string(bs),
		}
	}
	if resp.StatusCode != 201 {
		return &trait.Error{
			Err:      fmt.Errorf("upgrade %s request return status code [%d], msg: [%s]", obj.Name, resp.StatusCode, string(bs)),
			Internal: trait.ECHTTPAPIRawError,
			Detail:   resp.StatusCode,
		}
	}

	if rerr := json.Unmarshal(bs, &obj.Instance); rerr != nil {
		return &trait.Error{
			Err:      fmt.Errorf("decode instance %s error %s, resp: %s", obj.Name, rerr.Error(), string(bs)),
			Internal: trait.ErrComponentDecodeError,
			Detail:   "decode response from compnent-management http request",
		}
	}
	return nil
}

// Update upgrade instance release
func Update[T ComponentGeneric](ctx context.Context, c *Client, obj *ComponentInstance[T]) *trait.Error {
	err := Get(ctx, c, &ComponentInstance[T]{
		ComponentInstanceMeta: obj.ComponentInstanceMeta,
	})
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return New(ctx, c, obj)
	}
	if err != nil {
		return err
	}
	return Upgrade(ctx, c, obj)
}

// Get get instance and set into input obj
func Get[T ComponentGeneric](ctx context.Context, c *Client, ins *ComponentInstance[T]) *trait.Error {
	req, rerr := http.NewRequest(http.MethodGet, fmt.Sprintf(c.BaseUrl+"/%s/%s", ins.Type, ins.Name), nil)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("create http request for get %s fail", ins.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	resp, rerr := http.DefaultClient.Do(req)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("http request for get %s fail", ins.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	defer resp.Body.Close()
	bs, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("read http request for get %s fail", ins.Name),
			Internal: trait.ECNetUnknow,
		}
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return &trait.Error{
				Internal: trait.ErrNotFound,
				Detail:   ins.ComponentInstanceMeta,
			}
		}
		return &trait.Error{
			Err:      fmt.Errorf("'compnent-management' get %s request return status code [%d], msg: [%s]", ins.Name, resp.StatusCode, string(bs)),
			Internal: trait.ECHTTPAPIRawError,
			Detail:   resp.StatusCode,
		}
	}
	if rerr := json.Unmarshal(bs, &ins.Instance); rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Errorf("decode %s from component-management response: %s", ins.Name, string(bs)),
			Internal: trait.ECNULL,
		}
	}
	return nil
}

func All(ctx context.Context, c *Client) (ls []ComponentInstanceMeta, err *trait.Error) {
	ls = make([]ComponentInstanceMeta, 0)
	err = utils.DoJsonHTTP(http.DefaultClient, http.MethodGet, c.BaseUrl+"/all", nil, nil, &ls, 200)
	return
}

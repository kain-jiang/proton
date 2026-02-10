package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"taskrunner/trait"

	"github.com/sirupsen/logrus"
)

const (
	hdyraIntrospectURL = "http://hydra-admin:4445/admin/oauth2/introspect"
	userURL            = "http://user-management-private:30980/api/user-management/v1/users/%s/roles"
)

// Client deployweb and as oauth process
type Client struct {
	*logrus.Logger
	HydraOauthClientID     string `json:"oauthClientID"`
	HydraOauthClientSecret string `json:"oauthClientSecret"`
	hydraAuthHeader        string
}

// NewClient create a Client
func NewClient(log *logrus.Logger, id, secret string) Client {
	authHeader := base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("Basic %s:%s", id, secret)))
	return Client{
		Logger:                 log,
		HydraOauthClientID:     id,
		HydraOauthClientSecret: secret,
		hydraAuthHeader:        authHeader,
	}
}

type introspectBody struct {
	UserID string `json:"sub"`
	Active bool   `json:"active"`
}

// UserInfo user info store ing as
type UserInfo struct {
	Roles []string `json:"roles"`
}

// GetUserID get userID with token
func (c *Client) GetUserID(ctx context.Context, token string) (string, *trait.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hdyraIntrospectURL, strings.NewReader(fmt.Sprintf("token=%s", token)))
	if err != nil {
		return "", &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      err,
			Detail:   "new request for get user id from hydra",
		}
	}
	c.Logger.Tracef("get userid with token %s ", token)

	req.Header.Set("cache-control", "no-cahce")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("authorization", c.hydraAuthHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      err,
			Detail:   "do request for get user id from hydra",
		}
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      err,
			Detail:   " read request response for get user id from hydra",
		}
	}
	if resp.StatusCode != 200 {
		return "", &trait.Error{
			Err:      errors.New(string(respBody)),
			Detail:   resp.StatusCode,
			Internal: trait.ECHTTPAPIRawError,
		}
	}
	res := &introspectBody{}
	err = json.Unmarshal(respBody, res)
	if err != nil {
		err = fmt.Errorf("decode hydra introspect response error: %s", err.Error())
		return "", &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "decode use id response for get user id from hydra",
		}
	}
	if !res.Active {
		return "", &trait.Error{
			Err: fmt.Errorf("token %s has expired, get user id fail", token),
			// token 失效，需要返回403，以供退出登录
			Detail:   403,
			Internal: trait.ECHTTPAPIRawError,
		}
	}
	return res.UserID, nil
}

// GetUserRole get user role with id
func (c *Client) GetUserRole(ctx context.Context, id string) ([]string, *trait.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(userURL, id), nil)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      err,
			Detail:   "new request for get user role from user-management",
		}
	}
	c.Logger.Debugf("get user roles with userid %s", id)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      err,
			Detail:   "do request for get user role from user-management",
		}
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNetUnknow,
			Err:      err,
			Detail:   "read response for get user role from user-management",
		}
	}
	if resp.StatusCode != 200 {
		return nil, &trait.Error{
			Err:      errors.New(string(respBody)),
			Detail:   resp.StatusCode,
			Internal: trait.ECHTTPAPIRawError,
		}
	}
	res := []UserInfo{}
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		err = fmt.Errorf("decode user-manager %s response error: %s", userURL, err.Error())
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "decode use role response for get user role from user-management",
		}
	}
	if len(res) < 1 {
		// warn no role mean return no assign role
		return nil, nil
	}
	roles := res[0].Roles
	return roles, nil
}

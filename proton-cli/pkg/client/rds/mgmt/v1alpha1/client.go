package v1alpha1

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

type Client struct {
	// Base is the root URL for all invocations of the client
	Base *url.URL

	// Underlying http client. If not set http.DefaultClient will be used.
	HTTP *http.Client

	Logger logrus.FieldLogger
}

// NewClient creates a new Client.
func NewClient(baseURL *url.URL, client *http.Client) (*Client, error) {
	return &Client{
		Base:   baseURL,
		HTTP:   client,
		Logger: logger.NewLogger(),
	}, nil
}

// API request connects to the server and decode result into specific output
// when a server response is received. It handles retry behavior.
func (c *Client) request(ctx context.Context, method string, apiPath string, input interface{}, expectStatusCode int, output interface{}) error {
	var u url.URL = *c.Base
	u.Path = path.Join(u.Path, apiPath)

	var body io.Reader
	if input != nil {
		b, err := json.Marshal(input)
		if err != nil {
			return err
		}
		c.Logger.WithField("content", string(b)).Debug("http request body")
		body = bytes.NewBuffer(b)
	} else {
		c.Logger.WithField("content", input).Debug("http request body is nil")
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("content-type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	c.Logger.WithField("content", string(content)).Debug("http response body")

	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if err != nil {
		return fmt.Errorf("unable to parse media type %q: %w", resp.Header.Get("content-type"), err)
	}
	if mediaType != "application/json" {
		return fmt.Errorf("invalid media type %q", mediaType)
	}

	if resp.StatusCode != expectStatusCode {
		var e Error
		if err := json.Unmarshal(content, &e); err != nil {
			return fmt.Errorf("decode response body to error fail: %w, raw: %s", err, string(content))
		}
		return &e
	}

	if output != nil {
		return json.Unmarshal(content, output)
	}

	return nil
}

// CreateDatabase implements Interface.
func (c *Client) CreateDatabase(ctx context.Context, db *Database) error {
	return c.request(ctx, http.MethodPut, path.Join("/api/proton-rds-mgmt/v2/dbs", db.DBName), db, http.StatusCreated, nil)
}

func (c *Client) DeleteDatabase(ctx context.Context, name string) error {
	return c.request(ctx, http.MethodDelete, path.Join("/api/proton-rds-mgmt/v2/dbs", name), nil, http.StatusNoContent, nil)
}

// ListDatabases implements Interface.
func (c *Client) ListDatabases(ctx context.Context) (databases []Database, err error) {
	err = c.request(ctx, http.MethodGet, "/api/proton-rds-mgmt/v2/dbs", nil, http.StatusOK, &databases)
	return
}

func (c *Client) CreateUser(ctx context.Context, username, password string) error {
	type request struct {
		Password string `json:"password,omitempty"`
	}
	return c.request(ctx, http.MethodPut, path.Join("/api/proton-rds-mgmt/v2/users", username), &request{base64.StdEncoding.EncodeToString([]byte(password))}, http.StatusOK, nil)
}

// ListUsers implements Interface.
func (c *Client) ListUsers(ctx context.Context) (users []User, err error) {
	err = c.request(ctx, http.MethodGet, "/api/proton-rds-mgmt/v2/users", nil, http.StatusOK, &users)
	return
}

// PatchUserPrivileges implements Interface.
func (c *Client) PatchUserPrivileges(ctx context.Context, username string, privileges []Privilege) error {
	return c.request(ctx, http.MethodPatch, path.Join("/api/proton-rds-mgmt/v2/users", username, "privileges"), privileges, http.StatusNoContent, nil)
}

var _ Interface = (*Client)(nil)

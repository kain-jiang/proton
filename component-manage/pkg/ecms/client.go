package ecms

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type cli struct {
	cli *resty.Client
}

type Client interface {
	DirectoryExist(dir string) (bool, error)
	DirectoryCreate(dir string) error
	DirectoryDelete(dir string) error
}

func New(hostPort string) Client {
	address := fmt.Sprintf("http://%s", hostPort)
	client := resty.New().
		SetTimeout(1 * time.Minute).
		SetBaseURL(address).
		SetRetryCount(3).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			switch err0 := err.(type) {
			case *net.DNSError:
				return true
			case *net.OpError:
				return err0.Op == "dial"
			default:
				return false
			}
		})
	return &cli{
		cli: client,
	}
}

// DirectoryCreate implements Client.
func (c *cli) DirectoryCreate(dir string) error {
	resp, err := c.cli.R().SetBody(struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}{
		Path: dir,
		Type: "directory",
	}).Post("/api/ecms/v1/file")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

// DirectoryDelete implements Client.
func (c *cli) DirectoryDelete(dir string) error {
	resp, err := c.cli.R().SetBody(struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}{
		Path: dir,
		Type: "directory",
	}).Delete("/api/ecms/v1/file")
	if err != nil {
		return err
	}
	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

// DirectoryExist implements Client.
func (c *cli) DirectoryExist(dir string) (bool, error) {
	var result bool

	resp, err := c.cli.R().
		SetQueryParam("path", dir).
		SetQueryParam("type", "file").
		SetResult(&result).
		Get("/api/ecms/v1/file")
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	if result {
		return false, fmt.Errorf("the directory is a file: %s", dir)
	}

	resp, err = c.cli.R().
		SetQueryParam("path", dir).
		SetQueryParam("type", "directory").
		SetResult(&result).
		Get("/api/ecms/v1/file")
	if err != nil {
		return false, err
	}
	if resp.IsError() {
		return false, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}

	return result, nil
}

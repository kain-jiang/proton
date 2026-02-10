package cms

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type cli struct {
	cli *resty.Client
	ns  string
}

type Client interface {
	Get(n string) (map[string]interface{}, error)
	Set(n string, data map[string]interface{}) error
	Del(n string) error
}

type CMSObject struct {
	Name         string                            `json:"name"`
	Use          string                            `json:"use"`
	Data         map[string]map[string]interface{} `json:"data"`
	EncryptField []string                          `json:"encrypt_field"`
}

func New(host, ns string) Client {
	address := fmt.Sprintf("http://%s", host)
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
		ns:  ns,
	}
}

func (c *cli) Get(n string) (map[string]interface{}, error) {
	rel := new(CMSObject)
	resp, err := c.cli.R().
		SetResult(rel).
		SetQueryParam("namespace", c.ns).
		SetPathParam("service", n).
		Get("/api/cms/v1/configuration/service/{service}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		if resp.StatusCode() == 404 {
			return nil, &NotFoundError{
				msg: fmt.Sprintf("code: %d, resp: %s", resp.StatusCode(), resp.Body()),
			}
		}
		return nil, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}

	if data, ok := rel.Data[rel.Use]; ok {
		return data, nil
	}

	if data, ok := rel.Data[strings.ReplaceAll(n, "-", "_")]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("data not found in cms %s", n)
}

func (c *cli) Set(n string, data map[string]interface{}) error {
	// 尝试更新再尝试创建,因为升级场景一般次数大于安装场景
	resp, err := c.cli.R().
		SetBody(CMSObject{
			Name: n,
			Use:  "default",
			Data: map[string]map[string]interface{}{
				"default": data,
			},
			EncryptField: []string{},
		}).
		SetQueryParam("namespace", c.ns).
		SetPathParam("service", n).
		Patch("/api/cms/v1/configuration/service/{service}")
	if err != nil {
		return err
	}

	if resp.IsError() {
		if resp.StatusCode() == 404 {
			return c.create(n, data)
		}
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

func (c *cli) create(n string, data map[string]interface{}) error {
	resp, err := c.cli.R().
		SetBody(CMSObject{
			Name: n,
			Use:  "default",
			Data: map[string]map[string]interface{}{
				"default": data,
			},
			EncryptField: []string{},
		}).
		SetQueryParam("namespace", c.ns).
		SetPathParam("service", n).
		Post("/api/cms/v1/configuration/service/{service}")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

func (c *cli) Del(n string) error {
	resp, err := c.cli.R().
		SetQueryParam("namespace", c.ns).
		SetPathParam("service", n).
		Delete("/api/cms/v1/configuration/service/{service}")
	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

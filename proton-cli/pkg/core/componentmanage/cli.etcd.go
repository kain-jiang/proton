package componentmanage

import (
	"fmt"
	"net/http"
)

func (c *cli) EnableETCD(chart, version string) error {
	resp, err := c.restyCli.R().SetBody(
		map[string]any{
			"chart_name":    chart,
			"chart_version": version,
		},
	).Post("/api/component-manage/v1/components/plugin/etcd")

	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("enable etcd failed: %w", err)
	}
	return nil
}

func (c *cli) CreateETCD(name string, reqData map[string]any) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{"params": reqData}).
		SetResult(&result).SetPathParam("name", name).
		Post("/api/component-manage/v1/components/release/etcd/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("create etcd failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) GetETCD(name string) (map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().SetResult(&result).SetPathParam("name", name).
		Get("/api/component-manage/v1/components/release/etcd/{name}")

	// 不存在
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("create etcd failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) UpgradeETCD(name string, reqData map[string]any) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{"params": reqData}).
		SetResult(&result).SetPathParam("name", name).
		Put("/api/component-manage/v1/components/release/etcd/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("create etcd failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) DeleteETCD(name string) error {
	resp, err := c.restyCli.R().
		SetPathParam("name", name).
		SetQueryParam("clean", "true").
		Delete("/api/component-manage/v1/components/release/etcd/{name}")
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// 未找到报错
		return nil
	}
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("delete etcd failed: %w", err)
	}
	return nil
}

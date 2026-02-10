package componentmanage

import (
	"fmt"
	"net/http"
)

func (c *cli) EnableKafka(chart, version string) error {
	resp, err := c.restyCli.R().SetBody(
		map[string]any{
			"chart_name":    chart,
			"chart_version": version,
		},
	).Post("/api/component-manage/v1/components/plugin/kafka")

	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("enable kafka failed: %w", err)
	}
	return nil
}

func (c *cli) CreateKafka(name string, reqData map[string]any, zkName string) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
			"dependencies": {
				"zookeeper": zkName,
			}}).
		SetResult(&result).SetPathParam("name", name).
		Post("/api/component-manage/v1/components/release/kafka/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("create kafka failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) GetKafka(name string) (map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().SetResult(&result).SetPathParam("name", name).
		Get("/api/component-manage/v1/components/release/kafka/{name}")

	// 不存在
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("get kafka failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) UpgradeKafka(name string, reqData map[string]any, zkName string) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
			"dependencies": {
				"zookeeper": zkName,
			}}).
		SetResult(&result).SetPathParam("name", name).
		Put("/api/component-manage/v1/components/release/kafka/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("upgrade kafka failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) DeleteKafka(name string) error {
	resp, err := c.restyCli.R().
		SetPathParam("name", name).
		SetQueryParam("clean", "true").
		Delete("/api/component-manage/v1/components/release/kafka/{name}")
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// 未找到报错
		return nil
	}
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("delete kafka failed: %w", err)
	}
	return nil
}

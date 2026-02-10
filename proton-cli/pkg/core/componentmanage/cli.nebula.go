package componentmanage

import (
	"fmt"
	"net/http"
)

type NebulaPluginImagesInfo struct {
	GraphD   string `json:"graphd"`
	MetaD    string `json:"metad"`
	StorageD string `json:"storaged"`
	Exporter string `json:"exporter"`
}
type NebulaPluginInfo struct {
	ChartName    string                 `json:"chart_name"`
	ChartVersion string                 `json:"chart_version"`
	Namespace    string                 `json:"namespace"`
	Images       NebulaPluginImagesInfo `json:"images"`
}

func (c *cli) EnableNebula(info NebulaPluginInfo) error {
	resp, err := c.restyCli.R().SetBody(info).Post("/api/component-manage/v1/components/plugin/nebula")
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("enable nebula failed: %w", err)
	}
	return nil
}

func (c *cli) CreateNebula(name string, reqData map[string]any) (map[string]any, map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
		}).
		SetResult(&result).SetPathParam("name", name).
		Post("/api/component-manage/v1/components/release/nebula/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, nil, fmt.Errorf("create nebula failed: %w", err)
	}
	return result.Params, result.Info, nil
}

func (c *cli) UpgradeNebula(name string, reqData map[string]any) (map[string]any, map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
		}).
		SetResult(&result).SetPathParam("name", name).
		Put("/api/component-manage/v1/components/release/nebula/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, nil, fmt.Errorf("upgrade nebula failed: %w", err)
	}
	return result.Params, result.Info, nil
}

func (c *cli) GetNebula(name string) (map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().SetResult(&result).SetPathParam("name", name).
		Get("/api/component-manage/v1/components/release/nebula/{name}")

	// 不存在
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("get nebula failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) DeleteNebula(name string) error {
	resp, err := c.restyCli.R().
		SetPathParam("name", name).
		SetQueryParam("clean", "true").
		Delete("/api/component-manage/v1/components/release/nebula/{name}")
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// 未找到报错
		return nil
	}
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("delete nebula failed: %w", err)
	}
	return nil

}

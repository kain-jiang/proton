package componentmanage

import (
	"fmt"
	"net/http"
)

type MongoDBPluginImagesInfo struct {
	MongoDB   string `json:"mongodb"`
	Logrotate string `json:"logrotate"`
	Exporter  string `json:"exporter"`
	Mgmt      string `json:"mgmt"`
}
type MongoDBPluginInfo struct {
	ChartName    string                  `json:"chart_name"`
	ChartVersion string                  `json:"chart_version"`
	Namespace    string                  `json:"namespace"`
	Images       MongoDBPluginImagesInfo `json:"images"`
}

func (c *cli) EnableMongoDB(info MongoDBPluginInfo) error {
	resp, err := c.restyCli.R().SetBody(info).Post("/api/component-manage/v1/components/plugin/mongodb")
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("enable mongodb failed: %w", err)
	}
	return nil
}

func (c *cli) CreateMongoDB(name string, reqData map[string]any) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
		}).
		SetResult(&result).SetPathParam("name", name).
		Post("/api/component-manage/v1/components/release/mongodb/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("create mongodb failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) UpgradeMongoDB(name string, reqData map[string]any) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
		}).
		SetResult(&result).SetPathParam("name", name).
		Put("/api/component-manage/v1/components/release/mongodb/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("upgrade mongodb failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) GetMongoDB(name string) (map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().SetResult(&result).SetPathParam("name", name).
		Get("/api/component-manage/v1/components/release/mongodb/{name}")

	// 不存在
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("get mongodb failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) DeleteMongoDB(name string) error {
	resp, err := c.restyCli.R().
		SetPathParam("name", name).
		SetQueryParam("clean", "true").
		Delete("/api/component-manage/v1/components/release/mongodb/{name}")
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// 未找到报错
		return nil
	}
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("delete mongodb failed: %w", err)
	}
	return nil

}

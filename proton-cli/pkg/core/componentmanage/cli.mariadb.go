package componentmanage

import (
	"fmt"
	"net/http"
)

type MariaDBPluginImagesInfo struct {
	MariaDB  string `json:"mariadb"`
	ETCD     string `json:"etcd"`
	Exporter string `json:"exporter"`
	Mgmt     string `json:"mgmt"`
}
type MariaDBPluginInfo struct {
	ChartName    string                  `json:"chart_name"`
	ChartVersion string                  `json:"chart_version"`
	Namespace    string                  `json:"namespace"`
	Images       MariaDBPluginImagesInfo `json:"images"`
}

func (c *cli) EnableMariaDB(info MariaDBPluginInfo) error {
	resp, err := c.restyCli.R().SetBody(info).Post("/api/component-manage/v1/components/plugin/mariadb")
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("enable mariadb failed: %w", err)
	}
	return nil
}
func (c *cli) CreateMariaDB(name string, reqData map[string]any) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
		}).
		SetResult(&result).SetPathParam("name", name).
		Post("/api/component-manage/v1/components/release/mariadb/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("create mariadb failed: %w", err)
	}
	return result.Info, nil
}
func (c *cli) UpgradeMariaDB(name string, reqData map[string]any) (map[string]any, error) {
	var result struct {
		Info map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().
		SetBody(map[string]map[string]any{
			"params": reqData,
		}).
		SetResult(&result).SetPathParam("name", name).
		Put("/api/component-manage/v1/components/release/mariadb/{name}")
	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("upgrade mariadb failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) GetMariaDB(name string) (map[string]any, error) {
	var result struct {
		Params map[string]any `json:"params"`
		Info   map[string]any `json:"info"`
	}
	resp, err := c.restyCli.R().SetResult(&result).SetPathParam("name", name).
		Get("/api/component-manage/v1/components/release/mariadb/{name}")

	// 不存在
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	err = errorOf(resp, err)
	if err != nil {
		return nil, fmt.Errorf("get mariadb failed: %w", err)
	}
	return result.Info, nil
}

func (c *cli) DeleteMariaDB(name string) error {
	resp, err := c.restyCli.R().
		SetPathParam("name", name).
		SetQueryParam("clean", "true").
		Delete("/api/component-manage/v1/components/release/mariadb/{name}")
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// 未找到报错
		return nil
	}
	err = errorOf(resp, err)
	if err != nil {
		return fmt.Errorf("delete mariadb failed: %w", err)
	}
	return nil

}

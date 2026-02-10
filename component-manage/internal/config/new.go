package config

import (
	"fmt"
	"os"

	"component-manage/pkg/k8s"

	"gopkg.in/yaml.v2"
)

func NewConfig(configPath string) (*Config, error) {
	// 解析yaml 配置
	var result Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	// 内置配置
	result.Internal.ClusterDomain = k8s.ClusterDomain()
	return &result, nil
}

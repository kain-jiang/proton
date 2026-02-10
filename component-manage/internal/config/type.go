package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"component-manage/pkg/helm3"

	"helm.sh/helm/v3/pkg/repo"
)

/*
log:
  level: debug
  format: json

server:
  host: "0.0.0.0"
  port: 8888

config:
  enableDualStack: false
  chartmuseum:
    enable: false
    url: http://chartmuseum.aishu.cn:15001
    username: ""
    password: ""
  oci:
    enable: false
    registry: "registry.aishu.cn:15000"
    username: ""
    password: ""
  registry: registry.aishu.cn:15000

persist:
  secret_components_name: persist-component-manage-components
  secret_plugins_name: persist-component-manage-plugins
  secret_namespace: resource


*/

type Config struct {
	Internal struct {
		ClusterDomain string `yaml:"-" json:"-"`
	}
	Log struct {
		Level         string `yaml:"level" json:"level"`
		Format        string `yaml:"format" json:"format"`
		DisableHealth bool   `yaml:"disable_health" json:"disable_health"`
	} `yaml:"log" json:"log"`
	Server struct {
		Host string `yaml:"host" json:"host"`
		Port int    `yaml:"port" json:"port"`
	} `yaml:"server" json:"server"`
	Config struct {
		EnableDualStack bool `yaml:"enableDualStack" json:"enableDualStack"`
		Chartmuseum     struct {
			Enable   bool   `yaml:"enable" json:"enable"`
			URL      string `yaml:"url" json:"url"`
			Username string `yaml:"username" json:"username"`
			Password string `yaml:"password" json:"password"`
		} `yaml:"chartmuseum" json:"chartmuseum"`
		OCI struct {
			Enable    bool   `yaml:"enable" json:"enable"`
			Registry  string `yaml:"registry" json:"registry"`
			PlainHTTP bool   `yaml:"plain_http" json:"plain_http"`
			Username  string `yaml:"username" json:"username"`
			Password  string `yaml:"password" json:"password"`
		} `yaml:"oci" json:"oci"`
		Registry string `yaml:"registry" json:"registry"`
	} `yaml:"config" json:"config"`
	Persist struct {
		SecretComponentsName string `yaml:"secret_components_name" json:"secret_components_name"`
		SecretPluginsName    string `yaml:"secret_plugins_name" json:"secret_plugins_name"`
		SecretNamespace      string `yaml:"secret_namespace" json:"secret_namespace"`
	} `yaml:"persist" json:"persist"`
}

func (this *Config) ServerHost() string {
	return net.JoinHostPort(this.Server.Host, strconv.Itoa(this.Server.Port))
}

func (this *Config) configChartmuseumURL() string {
	// 判断是否有 HOST_IP 环境变量
	hostIP := os.Getenv("HOST_IP")
	// 判断url 的域名是否为 chartmuseum.aishu.cn
	if hostIP != "" && strings.Contains(this.Config.Chartmuseum.URL, "chartmuseum.aishu.cn") {
		// 如果有，则替换为 HOST_IP
		return strings.ReplaceAll(this.Config.Chartmuseum.URL, "chartmuseum.aishu.cn", hostIP)
	}
	return this.Config.Chartmuseum.URL
}

func (this *Config) configOCIRegistry() string {
	// 判断是否有 HOST_IP 环境变量
	hostIP := os.Getenv("HOST_IP")
	// 判断url 的域名是否为 chartmuseum.aishu.cn
	if hostIP != "" && strings.Contains(this.Config.OCI.Registry, "registry.aishu.cn") {
		// 如果有，则替换为 HOST_IP
		return strings.ReplaceAll(this.Config.OCI.Registry, "registry.aishu.cn", hostIP)
	}
	return this.Config.OCI.Registry
}

func (this *Config) ConfigChartmuseumToRepoEntry() *repo.Entry {
	if !this.Config.Chartmuseum.Enable {
		return nil
	}
	return &repo.Entry{
		Name:     "helm_repos",
		URL:      this.configChartmuseumURL(),
		Username: this.Config.Chartmuseum.Username,
		Password: this.Config.Chartmuseum.Password,
	}
}

func (this *Config) ConfigOCIRegistryInfo() *helm3.OCIRegistryConfig {
	if !this.Config.OCI.Enable {
		return nil
	}
	return &helm3.OCIRegistryConfig{
		PlainHTTP: this.Config.OCI.PlainHTTP,
		Registry:  this.configOCIRegistry(),
		Username:  this.Config.OCI.Username,
		Password:  this.Config.OCI.Password,
	}
}

func (this *Config) ServiceSuffix() string {
	return fmt.Sprintf("svc.%s.", this.Internal.ClusterDomain)
}

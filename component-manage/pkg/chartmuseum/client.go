package chartmuseum

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	urlParse "net/url"
	"sort"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/go-resty/resty/v2"
)

type cli struct {
	cli     *resty.Client
	baseUri string
}

type Client interface {
	GetNewest(name string) (*chart.Chart, error)
	Get(name, version string) (*chart.Chart, error)
	Push(chartPath string) error
	PushAfterDel(chartPath string) error
	Del(name, version string) error
	SearchRepoUrl(name, version string, repos []string) (string, error)
}

type ChartInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	ApiVersion  string    `json:"apiVersion"`
	AppVersion  string    `json:"appVersion"`
	Urls        []string  `json:"urls"`
	Created     time.Time `json:"created"`
	Digest      string    `json:"digest"`
}

func New(url, username, password string) Client {
	u, _ := urlParse.Parse(url)
	baseUrl := &urlParse.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}
	uriSplit := append([]string{"/api"}, strings.Trim(u.Path, "/"), "charts")
	if strings.Trim(u.Path, "/") == "" {
		uriSplit = append([]string{"/api"}, "charts")
	}
	uri := strings.Join(uriSplit, "/")
	client := resty.New().SetDisableWarn(true).
		SetBasicAuth(username, password).SetCookieJar(nil).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetBaseURL(baseUrl.String()).
		SetTimeout(1 * time.Minute)
	return &cli{
		cli:     client,
		baseUri: uri,
	}
}

// GetNewest 获取名为 name 的最新上传的 chart
func (c cli) GetNewest(name string) (*chart.Chart, error) {
	charts := make([]ChartInfo, 0)
	resp, err := c.cli.R().SetResult(&charts).Get(fmt.Sprintf("/api/charts/%s", name))
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() || len(charts) == 0 {
		return nil, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	sort.SliceStable(charts, func(i, j int) bool {
		return charts[i].Created.After(charts[j].Created)
	})

	resp, err = c.cli.R().Get(charts[0].Urls[0])
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return loader.LoadArchive(bytes.NewReader(resp.Body()))
}

// Get 获取名为 name, 版本为 version 的 chart
func (c cli) Get(name, version string) (*chart.Chart, error) {
	var ci ChartInfo
	resp, err := c.cli.R().SetResult(&ci).Get(fmt.Sprintf("/api/charts/%s/%s", name, version))
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}

	resp, err = c.cli.R().Get(ci.Urls[0])
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}

	return loader.LoadArchive(bytes.NewReader(resp.Body()))
}

// Push 上传 chartPath 的 chart 到仓库
func (c cli) Push(chartPath string) error {
	resp, err := c.cli.R().SetFile("chart", chartPath).Post(c.baseUri)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

// PushAfterDel 删除 chartPath 在仓库中的 chart 后上传
func (c cli) PushAfterDel(chartPath string) error {
	ci, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	err = c.Del(ci.Metadata.Name, ci.Metadata.Version)
	if err != nil {
		return err
	}

	return c.Push(chartPath)
}

// Del 删除名为 name, 版本为 version 的 chart
func (c cli) Del(name, version string) error {
	resp, err := c.cli.R().Delete(fmt.Sprintf("%s/%s/%s", c.baseUri, name, version))
	if err != nil {
		return err
	}
	if !resp.IsSuccess() && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
	}
	return nil
}

// search 查找名为 name, 版本为 version 的 chart 所在的 repo 地址
func (c cli) SearchRepoUrl(name, version string, repos []string) (string, error) {
	uriSplit := strings.Split(strings.Trim(c.baseUri, "/"), "/")
	if c.baseUri == "/api/charts" {
		return c.cli.BaseURL, nil
	} else if len(repos) == 1 {
		return fmt.Sprintf("%s/%s", c.cli.BaseURL, strings.Join(uriSplit[1:len(uriSplit)-1], "/")), nil
	}
	var repoUrl string
	for _, repo := range repos {
		uriSplit = append(uriSplit[:len(uriSplit)-2], repo, uriSplit[len(uriSplit)-1])
		uri := strings.Join(uriSplit, "/")
		resp, err := c.cli.R().Get(fmt.Sprintf("/%s/%s/%s", uri, name, version))
		if err != nil {
			return "", err
		}
		if resp.IsSuccess() {
			repoUrl = fmt.Sprintf("%s/%s", c.cli.BaseURL, strings.Join(uriSplit[1:len(uriSplit)-1], "/"))
			break
		} else if !resp.IsSuccess() && resp.StatusCode() == http.StatusNotFound {
			continue
		} else {
			return "", fmt.Errorf("code: %d, resp: %s", resp.StatusCode(), resp.Body())
		}
	}
	if repoUrl == "" {
		return repoUrl, fmt.Errorf("cannot found %s[%s] in repos(%v)", name, version, repos)
	}
	return repoUrl, nil
}

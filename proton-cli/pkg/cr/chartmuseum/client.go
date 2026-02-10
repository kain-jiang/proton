package chartmuseum

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/chartmuseum/helm-push/pkg/chartmuseum"
	"github.com/go-resty/resty/v2"
	"helm.sh/helm/v3/pkg/repo"
)

// Client 是访问 chartmuseum 所需的客户端
type Client struct {
	// base is the root URL for all invocations of the client
	base *url.URL

	client *resty.Client

	cm *chartmuseum.Client
}

// NewClient 创建访问 chartmuseum 所需的客户端
func NewClient(host, username, password string) (*Client, error) {
	base, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("unable to pasrse chartmuseum host: %w", err)
	}

	client := resty.New()
	if username != "" {
		client.SetBasicAuth(username, password)
	}

	cm, err := chartmuseum.NewClient(
		chartmuseum.URL(host),
		chartmuseum.Username(username),
		chartmuseum.Password(password),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create chartmuseum client: %w", err)
	}

	return &Client{
		base:   base,
		client: client,
		cm:     cm,
	}, err
}

const indexFileName = "index-proton-cli-*.yaml"

// IndexFile 获取 chartmuseum 的 Index File
func (c *Client) IndexFile() (*repo.IndexFile, error) {
	api := &url.URL{}
	*api = *c.base
	api.Path = path.Join(api.Path, "index.yaml")

	resp, err := c.client.SetOutputDirectory(os.TempDir()).R().Get(api.String())
	if err != nil {
		return nil, fmt.Errorf("unable to invoke rest api: %w", err)
	}
	if resp.IsError() {
		return nil, ErrorFromStatusCodeAndBody(resp.StatusCode(), resp.Body())
	}

	f, err := os.CreateTemp("", indexFileName)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporary index: %w, path: %v", err, f.Name())
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(resp.Body()); err != nil {
		return nil, fmt.Errorf("unable to create temporary index: %w", err)
	}

	index, err := repo.LoadIndexFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to load index file: %w", err)
	}

	return index, nil
}

// Get 返回指定 chart 的元数据。
func (c *Client) Get(name, version string) (*repo.ChartVersion, error) {
	return nil, nil
}

// PushOptions 是向 chartmuseum 推送 chart 的参数
type PushOptions struct {
	// true: 覆盖 chartmusem 可能已经存在的 chart。
	Force bool
}

// PushChartFile 推送 chart 文件到 chartmuseum
func (c *Client) PushChartFile(name string, opts PushOptions) error {
	resp, err := c.cm.UploadChartPackage(name, opts.Force)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return ErrorFromStatusCodeAndBody(resp.StatusCode, b)
	}

	return nil
}

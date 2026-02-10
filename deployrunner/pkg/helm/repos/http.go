package helm

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	"taskrunner/trait"
)

type basicAuth struct {
	AuthPasswd string `json:"authPasswd"`
	AuthUser   string `json:"authUSer"`
}

// HarBorRepo harbor uri is different from chartsum, need special operate with /api/charts uri
type HarBorRepo struct {
	HTTPHelmRepo
	StoreURI string
}

// NewHarborRepo create a harbor repo  from normal charsum http repo
func NewHarborRepo(r HTTPHelmRepo) (*HarBorRepo, *trait.Error) {
	parsed, err := neturl.Parse(r.URL)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   fmt.Sprintf("parse helm repo url: %s", r.URL),
		}
	}

	return &HarBorRepo{
		HTTPHelmRepo: r,
		StoreURI:     fmt.Sprintf("%s://%s/api%s/charts", parsed.Scheme, parsed.Host, parsed.Path),
	}, nil
}

// Store impl helm repo interface
func (r *HarBorRepo) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	return r.store(ctx, r.StoreURI, chart, data)
}

type RepoConf struct {
	*HTTPHelmRepo `json:"http,omitempty"`
	*OCi          `json:"oci,omitempty"`
}

func (r *RepoConf) CreateRealHelmRepo(ctx context.Context) (helm.Repo, *trait.Error) {
	if r.HTTPHelmRepo != nil {
		return r.HTTPHelmRepo.CreateRealHelmRepo(ctx)
	}
	return r.OCi.Init(ctx)
}

// HTTPHelmRepo is a repo config
type HTTPHelmRepo struct {
	SourceType string    `json:"source_type"`
	RepoName   string    `json:"name"`
	URL        string    `json:"url"`
	ShouldPush bool      `json:"shouldPush"`
	AuthType   string    `json:"authType"`
	BasicAuth  basicAuth `json:"auth"`
	RetryCount int       `json:"retryCount"`
	// RetryDelay ms
	RetryDelay int `json:"retryDelay"`
}

// CreateRealHelmRepo create real helm repo with current config
func (r *HTTPHelmRepo) CreateRealHelmRepo(ctx context.Context) (helm.Repo, *trait.Error) {
	if r.SourceType == "external" {
		return NewHarborRepo(*r)
	}
	return r, nil
}

func retryN(ctx context.Context, f func() (bool, *trait.Error), n int, delay time.Duration) *trait.Error {
	ok, err := f()
	for i := 1; i < n && ok; i++ {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return context.Cause(ctx).(*trait.Error)
		}
		time.Sleep(delay)
		ok, err = f()
	}
	return err
}

func retryBytesN(ctx context.Context, f func() (bool, []byte, *trait.Error), n int, delay time.Duration) ([]byte, *trait.Error) {
	ok, bs, err := f()
	for i := 1; i < n && ok; i++ {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return bs, err
		}
		time.Sleep(delay)
		ok, bs, err = f()
	}
	return bs, err
}

func (r *HTTPHelmRepo) store(ctx context.Context, url string, chart *component.HelmComponent, data []byte) *trait.Error {
	if !r.ShouldPush {
		return nil
	}

	// buf := bytes.NewBuffer(nil)
	// w := multipart.NewWriter(buf)
	// if err := w.WriteField("chart", string(data)); err != nil {
	// 	w.Close()
	// 	return err
	// }
	// if err := w.Close(); err != nil {
	// 	return err
	// }

	// bs := buf.Bytes()

	f := func() (bool, *trait.Error) {
	upload:
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return false, &trait.Error{
				Err:      err,
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("new POST request error"),
			}
		}
		// req.Header.Set("Content-Type", w.FormDataContentType())
		r.setAuth(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return true, &trait.Error{
				Internal: trait.ErrHelmRepoUnknow,
				Err:      err,
				Detail:   fmt.Errorf("do POST request error"),
			}
		}
		defer resp.Body.Close()
		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, &trait.Error{
				Err:      err,
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("read http request body error"),
			}
		}

		if resp.StatusCode == 409 {
			oURL := fmt.Sprintf("%s/%s/%s", url, chart.Name, chart.Version)
			fmt.Println(oURL)
			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, oURL, nil)
			if err != nil {
				return true, &trait.Error{
					Internal: trait.ErrHelmRepoUnknow,
					Err:      err,
					Detail:   fmt.Errorf("new DELETE request error"),
				}
			}
			r.setAuth(req)
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return true, &trait.Error{
					Internal: trait.ErrHelmRepoUnknow,
					Err:      err,
					Detail:   fmt.Errorf("do POST request error"),
				}
			}
			defer resp.Body.Close()
			bs, err := io.ReadAll(resp.Body)
			if err != nil {
				return true, &trait.Error{
					Err:      err,
					Internal: trait.ErrHelmRepoUnknow,
					Detail:   fmt.Errorf("read http request body error"),
				}
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				err = fmt.Errorf("delete chart error, status code: [%d], msg: [%s]", resp.StatusCode, bs)
				return false, &trait.Error{
					Err:      err,
					Internal: trait.ErrHelmRepoUnknow,
					Detail:   fmt.Errorf("helm repo api error"),
				}
			}
			goto upload
		} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err = fmt.Errorf("upload chart error, status code: [%d], msg: [%s]", resp.StatusCode, bs)
			return false, &trait.Error{
				Err:      err,
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("helm repo api error"),
			}
		}
		return false, nil
	}
	return retryN(ctx, f, r.RetryCount, time.Microsecond*time.Duration(r.RetryDelay))
}

// Store store chart
func (r *HTTPHelmRepo) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	if !r.ShouldPush {
		return nil
	}
	url := fmt.Sprintf("%s/api/charts", r.URL)
	return r.store(ctx, url, chart, data)
}

func (r *HTTPHelmRepo) setAuth(req *http.Request) {
	authType := strings.ToLower(r.AuthType)
	switch r.AuthType {
	case "":
		// no need auth
	case "basic":
		req.SetBasicAuth(r.BasicAuth.AuthUser, r.BasicAuth.AuthPasswd)
	default:
		panic(fmt.Sprintf("don't support [%s] authType", authType))
	}
}

// Name return repo name
func (r *HTTPHelmRepo) Name() string {
	return r.RepoName
}

// Fetch fetch chart from repo
func (r *HTTPHelmRepo) Fetch(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	url := fmt.Sprintf("%s/charts/%s-%s.tgz", r.URL, c.Name, c.Version)
	f := func() (bool, []byte, *trait.Error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return false, nil, &trait.Error{
				Err:      err,
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("fetch chart %s:%s error", c.Name, c.Version),
			}
		}
		r.setAuth(req)
		cli := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		res, err := cli.Do(req)
		if err != nil {
			return true, nil, &trait.Error{
				Err:      err,
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("fetch chart %s:%s rest client error", c.Name, c.Version),
			}
		}
		defer res.Body.Close()

		bs, err := io.ReadAll(res.Body)
		if err != nil {
			return true, nil, &trait.Error{
				Err:      fmt.Errorf("download chart error, read resp with error: %s", err.Error()),
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("fetch chart %s:%s read response error", c.Name, c.Version),
			}
		}

		if res.StatusCode != 200 {
			if res.StatusCode == 404 {
				err := fmt.Errorf("download chart %s:%s error, recevie statusCode: %d, resp: %s", c.Name, c.Version, res.StatusCode, bs)
				return false, bs, &trait.Error{
					Internal: trait.ErrHelmChartNoFound,
					Err:      err,
					Detail:   err,
				}
			}
			return false, bs, &trait.Error{
				Err:      err,
				Internal: trait.ErrHelmRepoUnknow,
				Detail:   fmt.Errorf("download chart error, recevie statusCode: %d, resp: %s", res.StatusCode, bs),
			}
		}
		return false, bs, nil
	}
	return retryBytesN(ctx, f, r.RetryCount, time.Microsecond*time.Duration(r.RetryDelay))
}

type haMutilUploadRepo struct {
	repos []helm.Repo
}

// NewHaMutilUploadRepo return the mutil upload for ha repo
func NewHaMutilUploadRepo(repos []RepoConf) (helm.Repo, *trait.Error) {
	repo := &haMutilUploadRepo{
		repos: make([]helm.Repo, 0, len(repos)),
	}
	for _, hr := range repos {
		r := hr
		rel, err := r.CreateRealHelmRepo(context.Background())
		if err != nil {
			return nil, err
		}
		repo.repos = append(repo.repos, rel)
	}
	return repo, nil
}

func (r *haMutilUploadRepo) Name() string {
	return "haUpload"
}

func (r *haMutilUploadRepo) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	for _, repo := range r.repos {
		if err := repo.Store(ctx, chart, data); err != nil {
			return err
		}
	}
	return nil
}

func (r *haMutilUploadRepo) Fetch(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	err := &trait.Error{
		Internal: trait.ErrHelmChartNoFound,
		Err:      fmt.Errorf("download chart %s:%s not found", c.Name, c.Version),
	}
	err.Detail = err.Err
	for _, repo := range r.repos {
		bs, err0 := repo.Fetch(ctx, c)
		if trait.IsInternalError(err0, trait.ErrHelmChartNoFound) {
			continue
		} else if err0 != nil {
			return nil, err0
		}
		return bs, err0
	}
	return nil, err
}

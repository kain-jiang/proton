package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"taskrunner/pkg/component"
	"taskrunner/test"
	testcharts "taskrunner/test/charts"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HTTPRepoMock struct {
	testcharts.MemoryHelmRepoMock
	prefix string
	path   string
	addr   string
	route  http.Server
	cache  map[string]bool
}

func (m *HTTPRepoMock) Serve() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Handle(http.MethodGet, m.prefix+"/charts/:path", m.GetChart)
	router.Handle(http.MethodPost, m.prefix+"/api/charts", m.UploadChart)
	router.Handle(http.MethodDelete, m.prefix+"/api/charts/:name/:version", m.delete)

	m.route = http.Server{
		Addr:    "127.0.0.1:30002",
		Handler: router,
	}
	m.cache = map[string]bool{}
	_ = m.route.ListenAndServe()
	return nil
}

func (m *HTTPRepoMock) GetChart(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.Status(400)
		//nolint: juse test fake
		_ = c.AbortWithError(400, fmt.Errorf("the path must not nil"))
		return
	}

	bs, err := m.MemoryHelmRepoMock.FS.ReadFile(filepath.Join(m.path, path))
	if err != nil {
		logrus.Warn(err)
		//nolint: juse test fake
		_ = c.AbortWithError(500, err)
		return
	}
	if _, err := c.Writer.Write(bs); err != nil {
		//nolint: juse test fake
		_ = c.AbortWithError(500, err)
	}
}

func (m *HTTPRepoMock) delete(c *gin.Context) {
	delete(m.cache, "python3-0.1.0.tgz")
	c.Status(200)
}

func (m *HTTPRepoMock) UploadChart(c *gin.Context) {
	path := "python3-0.1.0.tgz"
	if _, ok := m.cache[path]; ok {
		c.AbortWithStatus(409)
		return
	}
	bs, err := m.MemoryHelmRepoMock.FS.ReadFile(filepath.Join(m.path, path))
	if err != nil {
		logrus.Warn(err)
		_ = c.AbortWithError(500, err)
		return
	}
	bs0, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Warn(err)
		_ = c.AbortWithError(500, err)
	}
	if string(bs) != string(bs0) {
		logrus.Warn("input is not want")
		_ = c.AbortWithError(400, fmt.Errorf("input erro"))
	}
	m.cache[path] = true
}

func getTestHTTPMock(_ *testing.T) *HTTPRepoMock {
	m := HTTPRepoMock{
		MemoryHelmRepoMock: testcharts.MemoryHelmRepoMock{
			RepoName: "test",
			FS:       test.TestCharts,
		},
		prefix: "/test",
		path:   "testdata/charts",
		addr:   "127.0.0.1:30002",
	}
	return &m
}

func TestHTTPHelmRepo(t *testing.T) {
	m := getTestHTTPMock(t)
	//nolint: test no need check
	go func() {
		_ = m.Serve()
	}()
	time.Sleep(100 * time.Millisecond)
	defer func() {
		_ = m.route.Shutdown(context.Background())
	}()

	r := HTTPHelmRepo{
		RepoName:   "test",
		URL:        "http://127.0.0.1:30002/test",
		ShouldPush: true,
	}

	if r.RepoName != r.Name() {
		t.Fatal(r.Name())
	}

	if _, err := r.Fetch(context.Background(), &component.HelmComponent{
		HelmComponentSpec: component.HelmComponentSpec{
			Repository: r.RepoName,
		},
		ComponentMeta: trait.ComponentMeta{
			ComponentNode: trait.ComponentNode{
				Name:    "python3",
				Version: "9999.0.0",
			},
		},
	}); err == nil {
		t.Fatal("want error")
	}

	hc := &component.HelmComponent{
		HelmComponentSpec: component.HelmComponentSpec{
			Repository: r.RepoName,
		},
		ComponentMeta: trait.ComponentMeta{
			ComponentNode: trait.ComponentNode{
				Name:    "python3",
				Version: "0.1.0",
			},
		},
	}
	bs, err := r.Fetch(context.Background(), hc)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Store(context.Background(), hc, bs); err != nil {
		t.Fatal(err.Error())
	}

	if err := r.Store(context.Background(), hc, bs); err != nil {
		t.Fatal(err.Error())
	}
}

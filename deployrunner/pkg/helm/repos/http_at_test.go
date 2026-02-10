package helm

import (
	"context"
	"io"
	"os"
	"testing"

	"taskrunner/pkg/component"
	"taskrunner/test"
	"taskrunner/trait"
)

func TestHTTPHelmRepoAt(t *testing.T) {
	url := os.Getenv("HELM_CHART_URL")
	if url == "" {
		t.SkipNow()
	}
	baseUser := os.Getenv("HELM_CHART_USER")
	basePass := os.Getenv("HELM_CHART_PASS")
	authtype := ""
	tt := test.TestingT{T: t}

	if baseUser != "" {
		authtype = "basic"
	}
	r := &HTTPHelmRepo{
		URL:        url,
		ShouldPush: true,
		BasicAuth: basicAuth{
			AuthUser:   baseUser,
			AuthPasswd: basePass,
		},
		AuthType:   authtype,
		RetryCount: 3,
		RetryDelay: 1,
	}

	tcharts := test.TestCharts
	fpath := "testdata/charts/python3-0.1.0.tgz"
	fin, err := tcharts.Open(fpath)
	tt.AssertNil(err)
	defer fin.Close()
	bs, err := io.ReadAll(fin)
	tt.AssertNil(err)

	ch := &component.HelmComponent{
		ComponentMeta: trait.ComponentMeta{
			ComponentNode: trait.ComponentNode{
				Name:                "python3",
				Version:             "0.1.0",
				ComponentDefineType: component.ComponentHelmServiceType,
			},
		},
	}
	r0, err := NewHarborRepo(*r)
	tt.AssertNil(err)

	err = r0.Store(context.Background(), ch, bs)
	tt.AssertNil(err)
	// ch.Name = "deploy-service"
	// ch.Version = "2.12.0-mission"
	bs0, err := r0.Fetch(context.Background(), ch)
	tt.AssertNil(err)
	tt.Assert(bs, bs0)
}

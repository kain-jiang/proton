package helm

import (
	"context"
	"testing"

	"taskrunner/pkg/component"
	"taskrunner/test"

	"github.com/stretchr/testify/assert"
)

func TestOci(t *testing.T) {
	// 无harbor可用，暂时掠过该harbor测试用例
	t.SkipNow()
	tt := test.TestingT{T: t}
	ctx := context.Background()
	c := OCi{
		Registry: "harbor.aishu.eu.org/ict",
		Password: "Harbor12345",
		Username: "admin",
	}
	charts := test.TestCharts
	bs, err := charts.ReadFile("testdata/charts/python3-0.1.0.tgz")
	tt.AssertNil(err)
	hr, err := c.Init(ctx)
	tt.AssertNil(err)
	ch := &component.HelmComponent{}
	ch.Version = "0.1.0"
	ch.Name = "python3"

	err = hr.Store(ctx, ch, bs)
	tt.AssertNil(err)
	bs0, err := hr.Fetch(ctx, ch)
	tt.AssertNil(err)
	assert.Equal(t, bs, bs0, "chart bytes error")
}

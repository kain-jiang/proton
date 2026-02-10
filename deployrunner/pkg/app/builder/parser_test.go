package builder

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"taskrunner/pkg/component"
	"taskrunner/pkg/utils"
	"taskrunner/test"
	testdata "taskrunner/test"
	"taskrunner/trait"
)

func TestParseHelmChart(t *testing.T) {
	tt := test.TestingT{T: t}
	bs, err := testdata.TestCharts.ReadFile("testdata/charts/python3-0.1.0.tgz")
	tt.AssertNil(err)

	name := trait.HelmChartDir + "test/" + "python3-0.1.0.tgz"
	c := ParseHelmChartMeta(name)
	if c == nil {
		t.Fatal(name)
	}
	buf := bytes.NewBuffer(nil)
	w := utils.NewTGzWriter(buf)
	defer w.Close()
	graph, ver, err := ParseHelmChart(c, bs, w, nil)
	tt.AssertNil(err)
	if ver == "" {
		t.Fatal(ver)
	}
	w.Close()
	bs = buf.Bytes()
	_, err = utils.NewTGZReader(bytes.NewReader(bs))
	tt.AssertNil(err)
	if len(graph) != 1 {
		t.Fatal(graph)
	}

	name = trait.HelmChartDir + "test/wopi-proxy-7.5.1-story-490348-chart.tgz"
	c = ParseHelmChartMeta(name)
	if c == nil {
		t.Fatal(name)
	}
	if c.Name != "wopi-proxy" || c.Version != "7.5.1-story-490348-chart" {
		t.Fatal(c.Name, c.Version)
	}
}

func TestParseHelmChartMeta(t *testing.T) {
	name := "helm_charts/test/qwe-qwe-qwe-1.2.3:qwe-.tgz"
	c := ParseHelmChartMeta(name)
	if c == nil {
		t.Fatal(name)
	}
	if c.Name != "qwe-qwe-qwe" || c.Version != "1.2.3:qwe-" || c.Repository != "test" {
		t.Logf("%s %s %s", c.Name, c.Version, c.Repository)
		t.FailNow()
	}

	testcase := []string{
		"asdjkha/qwe",
		"test/qwe1.2.3:qwe.tgz",
	}
	for _, name := range testcase {
		if c := ParseHelmChartMeta(name); c != nil {
			t.Fatal(c)
		}
	}
}

func TestParseApplicationAT(t *testing.T) {
	pck := os.Getenv("TASKRUNNER_UPLOAD_APP")
	if pck == "" {
		t.SkipNow()
	}
	tt := test.TestingT{T: t}

	bs, err := os.ReadFile(pck)
	tt.AssertNil(err)
	a, fs, err := ParseApplication(bytes.NewReader(bs))
	tt.AssertNil(err)

	for _, meta := range a.Components() {
		if meta.ComponentNode.ComponentDefineType == component.ComponentHelmServiceType {
			hc := &component.HelmComponent{
				ComponentMeta: *meta,
			}
			if err := hc.Decode(meta.Spec); err != nil {
				err := fmt.Errorf("the component instance %s:%s decode error [%s]: %s", meta.Name, meta.Version, err.Error(), string(meta.Spec))
				t.Fatal(err)
			}
		}
	}

	count := 0
	for h, bs := range fs {
		c := ParseHelmChartMeta(h.Name)
		if c == nil {
			continue
		}
		count++
		_, ver, err := ParseHelmChart(c, bs, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if ver == "" {
			t.Fatal(ver)
		}
	}
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"taskrunner/pkg/app/builder"
	"taskrunner/pkg/component"
	"taskrunner/test"
)

func TestBuilder(t *testing.T) {
	tt := test.TestingT{T: t}
	appPath := os.Getenv("TASKRUNNER_APP")
	if appPath == "" {
		t.SkipNow()
	}
	cmd := newBuildCmd()
	f, err := os.CreateTemp("", "testapp")
	tt.AssertNil(err)
	defer os.Remove(f.Name())
	f.Close()
	cmd.SetArgs([]string{"-i", "/dev/null", "-a", appPath, "-d", f.Name()})
	t.Log(f.Name())
	err = cmd.Execute()
	tt.AssertNil(err)

	bs, err := os.ReadFile(f.Name())
	tt.AssertNil(err)
	a, fs, err := builder.ParseApplication(bytes.NewReader(bs))
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
		c := builder.ParseHelmChartMeta(h.Name)
		if c == nil {
			continue
		}
		count++
		_, ver, err := builder.ParseHelmChart(c, bs, nil, nil)
		if err != nil {
			t.Fatalf("解析打包后的helmchart失败:%s", err)
		}
		if ver == "" {
			t.Fatal(ver)
		}
	}
}

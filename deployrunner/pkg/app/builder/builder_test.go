package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"taskrunner/pkg/component"
	testdata "taskrunner/test"
	testchart "taskrunner/test/charts"
	"taskrunner/trait"
)

// func testAppConfiguration() Configuration {
// 	return Configuration{
// 		Application: trait.Application{
// 			Type: trait.AppBetav1Type,
// 			ApplicationMeta: trait.ApplicationMeta{
// 				Name:    "test",
// 				Version: "0.0.1",
// 			},
// 			Component: trait.Conponents{
// 				HelmComponents: map[string][]trait.HelmComponent{
// 					"test": {
// 						{
// 							ComponentMeta: trait.ComponentMeta{
// 								ConponentDefineType: "helm/task",
// 								Component: trait.Component{
// 									Name:    "python3",
// 									Version: "0.1.0",
// 								},
// 								Images: []string{
// 									"python6",
// 									"python4",
// 								},
// 								// Deps: []trait.Component{
// 								// 	{
// 								// 		Name:    "python4",
// 								// 		Version: "0.1.0",
// 								// 	},
// 								// },
// 							},
// 						},
// 						{
// 							ComponentMeta: trait.ComponentMeta{
// 								ConponentDefineType: "helm/service",
// 								Component: trait.Component{
// 									Name:    "python4",
// 									Version: "0.0.1",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		HelmRepos: []*helmimpy.HTTPHelmRepo{
// 			{
// 				RepoName: "test",
// 				URL:      "http://127.0.0.1:30002/test",
// 			}},
// 	}
// }

func TestBuildError(t *testing.T) {
	tt := testdata.TestingT{T: t}
	w := bytes.NewBuffer(nil)
	// w, err := os.OpenFile("tmp.tgz", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	// tt.AssertError(nil, err)
	// defer w.Close()
	imgOut := bytes.NewBuffer(nil)

	apf := bytes.NewReader(testdata.TestAPP)

	cfg, err := LoadConfiguration(apf)
	tt.AssertNil(err)

	repo := &testchart.MemoryHelmRepoMock{
		RepoName: "test",
		FS:       testdata.TestCharts,
	}
	cfg.Components.ResourceComponents = []*component.HoleComponent{
		{
			ComponentMeta: trait.ComponentMeta{
				ComponentNode: trait.ComponentNode{
					Name:                "dup",
					ComponentDefineType: "base",
				},
			},
		},
		{
			ComponentMeta: trait.ComponentMeta{
				ComponentNode: trait.ComponentNode{
					Name:                "dup",
					ComponentDefineType: "base",
				},
			},
		},
	}

	b, err := NewApplicationBuilder(&cfg, w, imgOut, repo)
	tt.AssertNil(err)
	err = b.Build(context.Background())
	tt.AssertError(trait.ErrComponentDup, err)

	cfg.Components.ResourceComponents = cfg.Components.ResourceComponents[:1]
	cfg.Application.Component = nil
	cfg.Graph = append(cfg.Graph, trait.Edge{
		From: trait.ComponentNode{
			Name:                "python4",
			ComponentDefineType: "base",
		},
		To: trait.ComponentNode{
			Name:                "python3",
			ComponentDefineType: "base",
		},
	})

	b, err = NewApplicationBuilder(&cfg, w, imgOut, repo)
	tt.AssertNil(err)
	err = b.Build(context.Background())
	if !trait.IsInternalError(err, trait.ErrAPPlicationComponentTortuous) {
		t.Fatal(err)
	}
}

func TestBuildAndParse(t *testing.T) {
	// cfg := testAppConfiguration()
	// bs, _ := json.MarshalIndent(cfg, "", "  ")
	// fmt.Println(string(bs))
	// cfg0 := &Configuration{}
	// if err := json.Unmarshal(bs, cfg0); err != nil {
	// 	t.Fatal(err)
	// }

	tt := testdata.TestingT{T: t}
	w := bytes.NewBuffer(nil)
	// w, err := os.OpenFile("tmp.tgz", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	// tt.AssertError(nil, err)
	// defer w.Close()
	imgOut := bytes.NewBuffer(nil)

	apf := bytes.NewReader(testdata.TestAPP)

	cfg, err := LoadConfiguration(apf)
	tt.AssertNil(err)

	repo := &testchart.MemoryHelmRepoMock{
		RepoName: "test",
		FS:       testdata.TestCharts,
	}
	f, err0 := os.CreateTemp(os.TempDir(), "test-config-tempalte-*.json")
	tt.AssertNil(err0)
	defer f.Close()
	defer os.Remove(f.Name())
	cfgt := trait.AppliacationConfigTemplate{
		AppliacationConfigTemplateMeta: trait.AppliacationConfigTemplateMeta{
			Aname:    "test",
			Aversion: "~v2.12.0-123",
			Tname:    "test0",
			Tversion: "qwe",
		},
		Config: trait.ApplicationConfigSet{
			AppConfig: map[string]interface{}{
				"qwe": 123,
			},
		},
	}
	bs0, err0 := json.Marshal(cfgt)
	tt.AssertNil(err0)
	_, err0 = f.Write(bs0)
	tt.AssertNil(err0)
	tt.AssertNil(f.Sync())

	b, err := NewApplicationBuilder(&cfg, w, imgOut, repo)
	b.ConfigTemplatePath = f.Name()
	tt.AssertNil(err)

	err = b.Build(context.Background())
	tt.AssertNil(err)

	wantImg := map[string]int{
		"python4":      0,
		"python3":      0,
		"nginx:1.16.0": 0,
		"python6":      0,
		"python5":      0,
	}
	imgs := strings.Split(strings.Trim(imgOut.String(), "\n"), "\n")

	for _, img := range imgs {
		if _, ok := wantImg[img]; !ok {
			t.Fatal(imgs, img)
		}
		wantImg[img]++
	}

	for img, c := range wantImg {
		if c == 0 {
			t.Fatal(imgs, img)
		}
	}

	bs := w.Bytes()
	a, fs, err0 := ParseApplication(bytes.NewReader(bs))
	tt.AssertNil(err0)
	if a.AName != "test" {
		t.Fatal(a)
	}

	for _, meta := range a.Components() {
		if meta.ComponentDefineType != component.ComponentHelmServiceType && meta.ComponentDefineType != component.ComponentHelmTaskType {
			continue
		}
		hc := &component.HelmComponent{
			ComponentMeta: *meta,
		}
		if err := hc.Decode(meta.Spec); err != nil {
			err := fmt.Errorf("the component instance %s:%s decode error [%s]: %s", meta.Name, meta.Version, err.Error(), string(meta.Spec))
			t.Fatal(err)
		}
		if hc.HelmChartAPIVersion != "v2" {
			t.Fatal(hc.HelmChartAPIVersion, hc.Name)
		}
		if hc.Repository != "test" {
			t.Fatal(hc.Repository, hc.Name)
		}
		if hc.Version != "0.1.0" {
			t.Fatal(hc.Version, hc.Name)
		}
	}

	err0 = SetAPPUISchema(&a)
	tt.AssertNil(err0)

	fmt.Println(string(a.UISchema))

	count := 0
	for h, bs := range fs {
		c := ParseHelmChartMeta(h.Name)
		if c != nil {
			count++
			_, ver, err := ParseHelmChart(c, bs, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if ver == "" {
				t.Fatal(ver)
			}
		} else if strings.HasPrefix(h.Name, trait.ConfigTemplateDir) {
			cfgt := &trait.AppliacationConfigTemplate{}
			tt.AssertNil(json.Unmarshal(bs, cfgt))
			tt.AssertNil(cfgt.Validate())
		}

	}
}

func TestDumpHelmCompoentImagesReal(t *testing.T) {
	t.SkipNow()
	fp := "/root/src/mycode/taskrunner/test/testdata/charts/python3.1.tgz"

	tt := testdata.TestingT{T: t}
	ch := &component.HelmComponent{}

	buf := bytes.NewBuffer(nil)
	b := &Builder{
		imagesOut: buf,
	}
	fin, err := os.Open(fp)
	tt.AssertNil(err)
	err = b.dumpHelmCompoentImages(fin, ch, "v1")
	tt.AssertNil(err)
	if buf.String() != "nginx:1.16.0\n" {
		t.Error(buf.String())
	}
}

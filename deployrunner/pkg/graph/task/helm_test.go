package task

import (
	"context"
	"reflect"
	"testing"

	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	testdata "taskrunner/test"
	test "taskrunner/test/charts"
	"taskrunner/trait"

	"github.com/ghodss/yaml"
	"helm.sh/helm/v3/pkg/action"
)

type helmClientMock struct {
	err *trait.Error
}

func (hc *helmClientMock) Install(ctx context.Context, name, ns string, chart *helm.Chart, cfg map[string]interface{}, timeout int, log action.DebugLog) *trait.Error {
	return hc.err
}

func (hc *helmClientMock) Uninstall(ctx context.Context, name, ns string, timeout int, log action.DebugLog) *trait.Error {
	return hc.err
}

func (hc *helmClientMock) Values(ctx context.Context, name, ns string) (map[string]interface{}, *trait.Error) {
	panic("no imply")
}

func TestHelmConfig(t *testing.T) {
	yamlText := []byte(`
test: 123
qwe:
  asd: 123
`)
	tt := testdata.TestingT{T: t}
	cfg := map[string]interface{}{}
	tt.AssertNil(yaml.Unmarshal(yamlText, &cfg))
	hc := &helmClientMock{}
	hr := &test.MemoryHelmRepoMock{
		RepoName: "test",
		FS:       testdata.TestCharts,
	}

	cins := &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			AIID: 2,
			Component: trait.ComponentNode{
				Name:    "useleescompoentnode",
				Version: "0.1.1",
			},
		},
		Config:    cfg,
		AppConfig: cfg,
		Attribute: cfg,
	}

	ht := &HelmTask{
		System: &cluster.SystemContext{
			System: trait.System{
				NameSpace: "test",
				SID:       0,
			},
			HelmManagerInterface: cluster.HelmManagerInterface{
				HelmRepo:   hr,
				HelmClient: hc,
			},
		},
		Base: Base{
			ComponentInsData: &trait.ComponentInstance{
				ComponentInstanceMeta: trait.ComponentInstanceMeta{
					Component: trait.ComponentNode{
						Name:    "python3",
						Version: "0.1.0",
					},
					AIID: 1,
				},
				Config:    cfg,
				AppConfig: cfg,
			},
			Topology: []*trait.ComponentInstance{
				{
					ComponentInstanceMeta: trait.ComponentInstanceMeta{
						System: trait.System{
							NameSpace: "test",
						},
						Component: trait.ComponentNode{
							Name:    "test0",
							Version: "1.0.1",
						},
					},
					Attribute: cfg,
				},
			},
		},
		HelmComponent: &component.HelmComponent{
			ComponentMeta: trait.ComponentMeta{
				ComponentNode: trait.ComponentNode{
					Name:    "python3",
					Version: "0.1.0",
				},
			},
			HelmComponentSpec: component.HelmComponentSpec{
				HelmChartAPIVersion: "v2",
			},
		},
	}
	//nolint:it won't return error
	_ = ht.SetComponentIns(cins)

	bs, err := ht.config()
	tt.AssertNil(err)
	// v := &chartValues{}
	// err = yaml.Unmarshal(bs, v)
	// testWantError(t, nil, err)

	// if _, ok := v.AppConfig["qwe"]; !ok {
	// 	fmt.Println(string(bs))
	// 	fmt.Printf("%#v", v)
	// 	t.Fatal()
	// }

	if _, ok := bs["qwe"]; !ok {
		t.Fatalf("%#v\n", bs)
	}

	dTrait := bs["deployTrait"].(map[string]interface{})
	if dTrait["aiid"] != ht.ComponentInsData.AIID || dTrait["version"] != "0.1.0" {
		t.Fatalf("%v\n", bs)
	}

	deps := bs["depServices"].(map[string]interface{})
	child := deps["test0"].(map[string]interface{})

	cdTrait := child["deployTrait"].(map[string]interface{})
	if cdTrait["version"] != "1.0.1" {
		t.Fatalf("%#v\n", child)
	}
	delete(child, "deployTrait")

	if !reflect.DeepEqual(child, cfg) {
		t.Fatalf("%#v\n", child)
	}

	tt.AssertNil(ht.Install(context.Background()))

	tt.AssertNil(ht.Uninstall(context.Background()))

	ht.HelmComponent.HelmChartAPIVersion = "qwe"
	err = ht.Install(context.Background())
	if err.Err.Error() != "chart apiVersion 'qwe' 不支持" {
		t.Fatal(err)
	}
	ht.HelmComponent.HelmChartAPIVersion = "v2"

	ht.HelmComponent.ComponentNode.Name = "test"
	err = ht.Install(context.Background())
	tt.AssertError(trait.ErrHelmChartNoFound, err)
}

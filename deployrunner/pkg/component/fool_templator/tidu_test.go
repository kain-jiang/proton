package templator

import (
	"embed"
	"encoding/json"
	"path"
	"testing"

	"taskrunner/pkg/component"
	"taskrunner/test"
	"taskrunner/trait"
)

//go:embed testcase/fool
var testcase embed.FS

// TODO add a testcase dir for complex situation
func getTestDeploy(tt test.TestingT) *component.FoolDeployment {
	root := "testcase/fool"

	entrys, err := testcase.ReadDir(root)
	tt.AssertNil(err)
	for _, e := range entrys {
		if e.IsDir() {
			continue
		}
		bs, err := testcase.ReadFile(path.Join(root, e.Name()))
		tt.AssertNil(err)
		c := &component.FoolComponent{}
		err = json.Unmarshal(bs, c)
		tt.AssertNil(err)
		err = c.Check()
		tt.AssertNil(err)
		return &c.Deploys[0]
	}
	return nil
}

func TestTiduFoolTemplate(t *testing.T) {
	tt := test.TestingT{T: t}
	c := &component.FoolComponent{
		ComponentMeta: trait.ComponentMeta{
			ComponentNode: trait.ComponentNode{
				Name:                "testc",
				Version:             "0.0.1",
				ComponentDefineType: "fool",
			},
		},
		FoolComponentSpec: component.FoolComponentSpec{
			Deploys: []component.FoolDeployment{
				*getTestDeploy(tt),
			},
		},
	}

	err := TiduFoolTemplate(c, "/tmp/tidu/charts")
	tt.AssertNil(err)
}

func TestTiduDeployTemplate(t *testing.T) {
	tt := test.TestingT{T: t}
	cname := "test"
	deploy := getTestDeploy(tt)
	_, err := tiduDeployTemplate(cname, deploy)
	// fmt.Println(err.Error())
	tt.AssertNil(err)
	// println(string(bs))
}

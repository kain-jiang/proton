package store

import (
	"context"
	"os"
	"reflect"
	"testing"

	"taskrunner/pkg/component"
	"taskrunner/test"
	"taskrunner/trait"

	"github.com/ghodss/yaml"
)

func TestAutoTest(t *testing.T) {
	testcasePath, ok := os.LookupEnv("PROTON_CONf_TESTCASE")
	if !ok {
		t.SkipNow()
	}
	tt := test.TestingT{T: t}
	ctx := context.Background()

	namesapce, ok := os.LookupEnv("PROTON_CONF_NAMESPACE")
	if !ok {
		t.SkipNow()
	}
	confName, ok := os.LookupEnv("PROTON_CONF_Name")
	if !ok {
		confName = _DefaultConfName
	}
	pcli, err := NewProtonCli(namesapce, confName, "")
	tt.AssertNil(err)

	tc := []struct {
		Obj struct {
			Config map[string]interface{} `json:"config"`
			Type   string                 `json:"resourceType"`
		} `json:"obj"`
		Want map[string]interface{} `json:"want"`
	}{}
	bs, rerr := os.ReadFile(testcasePath)
	tt.AssertNil(rerr)
	//nolint:no need check error
	_ = yaml.Unmarshal(bs, &tc)

	conf, err := pcli.GetConf(ctx)
	tt.AssertNil(err)

	for _, ts := range tc {
		cins := &trait.ComponentInstance{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{
					Name:                "",
					Type:                ts.Obj.Type,
					ComponentDefineType: component.ComponentProtonResourceType,
				},
			},

			Config: ts.Obj.Config,
		}
		obj, err := replaceAttribute(ctx, nil, cins, conf)
		tt.AssertNil(err)
		if !reflect.DeepEqual(obj.Attribute, ts.Want) {
			t.Fatalf("\nwant: %#v\n get: %#v", ts.Want, obj.Attribute)
		}
	}
}

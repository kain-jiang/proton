package store

import (
	"testing"
)

func TestToCRComponent(t *testing.T) {
	conf := &ProtonConf{
		Cr: &CR{
			Local: &localCR{
				Hosts: []string{
					"node-71-59",
					"node-71-58",
				},
			},
		},
		Nodes: []Node{
			{
				Name: "node-71-60",
				IP4:  "123.123.123.123",
			},
			{
				Name: "node-71-59",
				IP4:  "123.123.123.124",
			},
			{
				Name: "node-71-58",
				IP4:  "123.123.123.122",
			},
		},
	}
	cr := conf.ToCRComponent()
	if len(cr.HelmRepo) != 2 {
		t.Error(cr.HelmRepo)
	}
	// t.Log(cr.HelmRepo)
}

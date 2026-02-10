package component

import (
	"encoding/json"
	"testing"

	"taskrunner/test"
)

func TestFoolComponentDump(t *testing.T) {
	tt := test.TestingT{T: t}
	c := &FoolComponent{
		FoolComponentSpec: FoolComponentSpec{
			Deploys: []FoolDeployment{
				{
					Name: "test",
					Replica: FoolReplica{
						Custom:         true,
						DefaultReplica: 1,
					},
					Containers: []FoolContainer{
						{
							FoolInitContainer: FoolInitContainer{
								Name:    "test",
								Command: []string{"sleep"},
								Image: RepoImage{
									Registry: "test",
									Image:    "test",
									Tag:      "0.1.2",
								},
							},
							Ports: []int{1},
							ReadinessProbe: FoolProbe{
								Exec:             &FoolExecAction{Command: "echo 0"},
								FailureThreshold: 1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								TimeoutSeconds:   10,
							},
						},
					},
				},
			},
		},
	}
	bs, err := json.MarshalIndent(c, "", "  ")
	tt.AssertNil(err)
	// fmt.Println(string(bs))

	c0 := &FoolComponent{}
	err = json.Unmarshal(bs, c0)
	tt.AssertNil(err)

	if err := c0.check(); err != nil {
		t.Fatal(err)
	}
}

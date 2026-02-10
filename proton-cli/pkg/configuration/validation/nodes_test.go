package validation

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateNodes(t *testing.T) {
	type args struct {
		nodes   []configuration.Node
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid-ipv4",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP4:  "192.168.0.1",
					},
					{
						Name: "node-1",
						IP4:  "192.168.0.2",
					},
					{
						Name: "node-2",
						IP4:  "192.168.0.3",
					},
				},
			},
		},
		{
			name: "valid-ipv6",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP6:  "fe80::250:56ff:fe82:c102",
					},
					{
						Name: "node-1",
						IP6:  "fe80::250:56ff:fe82:c103",
					},
					{
						Name: "node-2",
						IP6:  "fe80::250:56ff:fe82:c104",
					},
				},
			},
		},
		{
			name: "valid-dual-stack",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP4:  "192.168.0.1",
						IP6:  "fe80::250:56ff:fe82:c102",
					},
					{
						Name: "node-1",
						IP4:  "192.168.0.2",
						IP6:  "fe80::250:56ff:fe82:c103",
					},
					{
						Name: "node-2",
						IP4:  "192.168.0.3",
						IP6:  "fe80::250:56ff:fe82:c104",
					},
				},
			},
		},
		{
			name: "invalid-empty-name",
			args: args{
				nodes: []configuration.Node{
					{
						IP4: "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-name-with-heading-dash",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "-node",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-name-with-heading-numeric-character",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "0-node",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-name-with-heading-space",
			args: args{
				nodes: []configuration.Node{
					{
						Name: " node-0",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-name-with-trailing-space",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0 ",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-name-with-trailing-dash",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0-",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-name-with-dot",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node.0",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-duplicate-name",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP4:  "192.168.0.1",
					},
					{
						Name: "node-0",
						IP4:  "192.168.0.2",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-non-ipv4-or-ipv6",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-ipv4",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP4:  "192.168.0.1111",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-duplicate-ipv4",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP4:  "192.168.0.1",
					},
					{
						Name: "node-1",
						IP4:  "192.168.0.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-ipv6",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP6:  "fe80::250:56ff:fe82:c10222",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-duplicate-ipv6",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
						IP6:  "fe80::250:56ff:fe82:c102",
					},
					{
						Name: "node-1",
						IP6:  "fe80::250:56ff:fe82:c102",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateNodes(tt.args.nodes, tt.args.fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				t.Errorf("ValidateNodes() len(errList) = %v, wantErr %v", len(errList), tt.wantErr)
				for i, err := range errList {
					t.Errorf("ValidateNodes() errList[%d] = %v", i, err)
				}
			}
		})
	}
}

func TestNewNodeNameSet(t *testing.T) {
	type args struct {
		nodes []configuration.Node
	}
	tests := []struct {
		name string
		args args
		want sets.Set[string]
	}{
		{
			name: "empty",
			args: args{},
			want: sets.Set[string]{},
		},
		{
			name: "single",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
					},
				},
			},
			want: sets.New[string]("node-0"),
		},
		{
			name: "multi",
			args: args{
				nodes: []configuration.Node{
					{
						Name: "node-0",
					},
					{
						Name: "node-1",
					},
					{
						Name: "node-2",
					},
				},
			},
			want: sets.New[string](
				"node-0",
				"node-1",
				"node-2",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeNameSet(tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeNameSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

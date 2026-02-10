package migrate

import (
	"testing"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidatePDConfig4ECeph(t *testing.T) {
	commonProtonCLIConfig := &configuration.ClusterConfig{
		Nodes: []configuration.Node{
			{
				Name:        "node-71-59",
				IP4:         "10.4.71.59",
				IP6:         "",
				Internal_ip: "10.4.71.59",
			},
		},
	}
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "empty",
			input:   map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "has all top level elems",
			input: map[string]interface{}{
				"apiVersion": nil,
				"hosts":      nil,
				"slb":        nil,
				"eceph":      nil,
				"as_vip":     nil,
			},
			wantErr: true,
		},
		{
			name: "no hosts",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts":      map[string]interface{}{},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
		{
			name: "non existent hostname",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
		{
			name: "non existent ssh_ip",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71": map[string]interface{}{
						"ssh_ip":      "127.0.0.1",
						"internal_ip": "127.0.0.1",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
		{
			name: "missing slb>ha",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "normal",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: false,
		},
		{
			name: "normal but no internal vip",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: false,
		},
		{
			name: "normal but multiple nodes without ext vip",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
					"node-0-1": map[string]interface{}{
						"ssh_ip":      "127.0.0.1/24",
						"internal_ip": "127.0.0.1/24",
					},
				},
				"slb": map[string]interface{}{},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
		{
			name: "missing eceph>lb>vip",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePDConfig4ECeph(tt.input, commonProtonCLIConfig)
			if err != nil && !tt.wantErr {
				t.Errorf("validatePDConfig4ECeph() unexpected error in unit test %v : %v", tt.name, err)
			}
		})
	}
}

func TestEnforceDefaultPDConfig4ECeph(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "normal",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: false,
		},
		{
			name: "custom ha label",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "iv",
						},
					},
				},
				"eceph": map[string]interface{}{
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "iv",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
		{
			name: "custom namespace",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"namespace": "anyshare",
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
		{
			name: "custom slb_listen",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"hosts": map[string]interface{}{
					"node-71-59": map[string]interface{}{
						"ssh_ip":      "10.4.71.59/24",
						"internal_ip": "10.4.71.59/24",
					},
				},
				"slb": map[string]interface{}{
					"ha": []interface{}{
						map[string]interface{}{
							"vip":   "10.20.30.40/31",
							"label": "ivip",
						},
					},
				},
				"eceph": map[string]interface{}{
					"slb_listen": 1234,
					"hosts": []interface{}{
						"node-71-59",
					},
					"lb": map[string]interface{}{
						"vip": "ivip",
					},
				},
				"as_vip": "10.4.71.59",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnforceDefaultPDConfig4ECeph(tt.input)
			if err != nil && !tt.wantErr {
				t.Errorf("EnforceDefaultPDConfig4ECeph() unexpected error in unit test %v : %v", tt.name, err)
			}
		})
	}
}

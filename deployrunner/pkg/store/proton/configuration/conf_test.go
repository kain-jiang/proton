package configuration

import (
	"testing"
)

func TestNode_IP(t *testing.T) {
	type fields struct {
		Name string
		IP4  string
		IP6  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ipv4-only",
			fields: fields{
				IP4: "1.1.1.1",
			},
			want: "1.1.1.1",
		},
		{
			name: "ipv6-only",
			fields: fields{
				IP6: "fc99:3088::a03:5b65",
			},
			want: "fc99:3088::a03:5b65",
		},
		{
			name: "dual-stack",
			fields: fields{
				IP4: "2.2.2.2",
				IP6: "fc99:3088::a03:5b66",
			},
			want: "2.2.2.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Name: tt.fields.Name,
				IP4:  tt.fields.IP4,
				IP6:  tt.fields.IP6,
			}
			if got := n.IP(); got != tt.want {
				t.Errorf("Node.IP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNodeNameByIP(t *testing.T) {
	tests := []struct {
		name  string
		ip    string
		nodes []Node
		want  string
	}{
		{
			name: "not-exist",
			ip:   "1.1.1.1",
			want: "",
		},
		{
			name: "IPv4",
			nodes: []Node{
				{
					Name: "ipv4-1-1-1-1",
					IP4:  "1.1.1.1",
				},
			},
			ip:   "1.1.1.1",
			want: "ipv4-1-1-1-1",
		},
		{
			name: "IPv6",
			nodes: []Node{
				{
					Name: "ipv6-c102",
					IP6:  "fe80::250:56ff:fe82:c102",
				},
			},
			ip:   "fe80::250:56ff:fe82:c102",
			want: "ipv6-c102",
		},
		{
			name: "dual-stack/ipv4",
			nodes: []Node{
				{
					Name: "node-dual-stack",
					IP4:  "1.1.1.1",
					IP6:  "fe80::250:56ff:fe82:c102",
				},
			},
			ip:   "1.1.1.1",
			want: "node-dual-stack",
		},
		{
			name: "dual-stack/ipv6",
			nodes: []Node{
				{
					Name: "node-dual-stack",
					IP4:  "1.1.1.1",
					IP6:  "fe80::250:56ff:fe82:c102",
				},
			},
			ip:   "fe80::250:56ff:fe82:c102",
			want: "node-dual-stack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNodeNameByIP(tt.ip, tt.nodes); got != tt.want {
				t.Errorf("ClusterConfig.NodeNameFromIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIPByNodeName(t *testing.T) {
	tests := []struct {
		name     string
		nodeName string
		nodeList []Node
		want     string
	}{
		{
			name:     "non-exist",
			nodeName: "node-0",
		},
		{
			name:     "ipv4-only",
			nodeName: "node-1",
			nodeList: []Node{
				{
					Name: "node-1",
					IP4:  "1.1.1.1",
				},
			},
			want: "1.1.1.1",
		},
		{
			name:     "ipv6-only",
			nodeName: "node-2",
			nodeList: []Node{
				{
					Name: "node-2",
					IP6:  "fe80::250:56ff:fe82:c355",
				},
			},
			want: "fe80::250:56ff:fe82:c355",
		},
		{
			name:     "dual-stack",
			nodeName: "node-3",
			nodeList: []Node{
				{
					Name: "node-3",
					IP4:  "3.3.3.3",
					IP6:  "fe80::250:56ff:fe82:c355",
				},
			},
			want: "3.3.3.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetIPByNodeName(tt.nodeName, tt.nodeList); got != tt.want {
				t.Errorf("GetIPByNodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

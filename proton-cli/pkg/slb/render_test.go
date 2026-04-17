package slb

import (
	"strings"
	"testing"

	slbv1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
	slbv2 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestRenderHAProxyConfig(t *testing.T) {
	conf := configuration.HaProxyConf{
		Conf: configuration.Conf{
			Global: configuration.Global{
				Log:     "/dev/log local0",
				Maxconn: "50000",
			},
			Defaults: configuration.Defaults{
				Maxconn: "5000",
				Mode:    "tcp",
				Timeout: "tunnel 86400s",
				Option:  "dontlognull",
			},
			Frontend: configuration.Frontend{
				CsFrontend: configuration.CsFrontend{
					Bind:           ":::8443 v4v6",
					Mode:           "tcp",
					DefaultBackend: "cs",
				},
			},
			Backend: configuration.Backend{
				CsBackend: configuration.CsBackend{
					Option:        []string{"httpchk GET /readyz HTTP/1.0"},
					DefaultServer: "verify none check-ssl",
					HTTPCheck:     "expect status 200",
					Balance:       "first",
					Server:        []string{"node1 192.168.1.10:6443 check"},
					Mode:          "tcp",
				},
			},
		},
	}

	rendered := string(renderHAProxyConfig(conf))
	for _, want := range []string{
		"frontend cs",
		"bind :::8443 v4v6",
		"backend cs",
		"server node1 192.168.1.10:6443 check",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered haproxy config does not contain %q:\n%s", want, rendered)
		}
	}
}

func TestRenderNginxHTTPServer(t *testing.T) {
	server := &slbv1.NginxHTTP{
		Name: "eceph_https",
		Conf: slbv1.NginxHTTPConf{
			Server: slbv1.NginxServer{
				Listen:            "10443 ssl",
				ClientMaxBodySize: "0",
				Locations: []map[string]map[string]interface{}{
					{
						"/": {
							"proxy_pass": "http://eceph_https",
							"if": []map[string]map[string]string{
								{
									`($request_method = POST)`: {
										"return": "204",
									},
								},
							},
						},
					},
				},
				AddHeaders: []string{"Access-Control-Allow-Origin * always"},
			},
			Upstream: map[string]slbv1.NginxUpstream{
				"eceph_https": {
					Check:     "interval=10000",
					Keepalive: "300",
					Servers:   []string{"10.0.0.1:7480", "10.0.0.2:7480"},
				},
			},
		},
	}

	rendered, err := renderNginxHTTPServer(server)
	if err != nil {
		t.Fatalf("renderNginxHTTPServer() error = %v", err)
	}

	text := string(rendered)
	for _, want := range []string{
		"upstream eceph_https",
		"server 10.0.0.1:7480;",
		"location / {",
		`if ($request_method = POST) {`,
		"add_header Access-Control-Allow-Origin * always;",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered nginx config does not contain %q:\n%s", want, text)
		}
	}
}

func TestParseKeepalivedConfigAndUpdateInstance(t *testing.T) {
	conf, err := parseKeepalivedConfig([]byte(`global_defs {
    router_id test
}

vrrp_instance eceph_vip {
    state BACKUP
    virtual_router_id 51
    priority 100
    notify_master /etc/slb/scripts/eceph_vip/entering_master.py
}
`))
	if err != nil {
		t.Fatalf("parseKeepalivedConfig() error = %v", err)
	}

	block := conf.block("vrrp_instance", "eceph_vip")
	if block == nil {
		t.Fatal("expected vrrp_instance eceph_vip")
	}

	updateKeepalivedHAInstanceBlock(block, &slbv2.KeepalivedHA{
		Interface:        "ens160",
		UnicastSRC_IP:    "192.168.1.10",
		UnicastPeer:      []string{"192.168.1.11", "192.168.1.12"},
		VirtualRouterID:  "61",
		Priority:         "110",
		VirtualIPAddress: map[string]string{"192.168.1.100/24": "label ens160:vip dev ens160"},
	})

	rendered := string(conf.render())
	for _, want := range []string{
		"interface ens160",
		"virtual_router_id 61",
		"priority 110",
		"unicast_peer {",
		"192.168.1.11",
		"notify_master /etc/slb/scripts/eceph_vip/entering_master.py",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered keepalived config does not contain %q:\n%s", want, rendered)
		}
	}
}

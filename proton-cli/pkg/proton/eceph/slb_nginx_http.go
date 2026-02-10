package eceph

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"path/filepath"
	"strconv"

	"github.com/go-test/deep"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func (m *Manager) reconcileNGINXServers() error {
	m.Logger.Info("reconcile nginx servers")

	var upstreams []string
	for _, n := range m.Nodes {
		upstreams = append(upstreams, net.JoinHostPort(n.InternalIP().String(), strconv.Itoa(CephRADOSGatewayPort)))
	}

	var nginxServers = []slb.NginxHTTP{
		generateNGINXServerECephHTTP(upstreams, m.Nodes[0].IPVersion()),
		generateNGINXServerECephHTTPS(upstreams, m.Nodes[0].IPVersion()),
	}

	for _, n := range m.Nodes {
		log := m.Logger.WithField("node", n.Name())

		for _, f := range []struct {
			path string
			data []byte
		}{
			{
				path: NGINXServerCertificatePath,
				data: m.Spec.TLS.CertificateData,
			},
			{
				path: NGINXServerKeyPath,
				data: m.Spec.TLS.KeyData,
			},
		} {
			if err := reconcileNodeFile(n, f.path, f.data, log); err != nil {
				return err
			}
		}

		for _, s := range nginxServers {
			if err := reconcileNodeNGINXServer(n.SLB_V1().NginxHTTPs(), &s, log); err != nil {
				return err
			}
		}
	}
	return nil
}

func reconcileNodeFile(n node.Interface, path string, data []byte, log logrus.FieldLogger) error {
	var ctx = context.TODO()
	equal, err := nodeFileEqual(n, path, data)
	if errors.Is(err, fs.ErrNotExist) {
		if d, err := n.ECMS().Files().Stat(ctx, filepath.Dir(path)); errors.Is(err, fs.ErrNotExist) {
			log.WithField("path", filepath.Dir(path)).Info("create directory")
			if err := n.ECMS().Files().Create(ctx, filepath.Dir(path), true, nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !d.IsDir() {
			return fmt.Errorf("%s is not a directory", d.Name())
		}

		log.WithField("path", path).Info("create file")
		return n.ECMS().Files().Create(ctx, path, false, data)
	} else if err != nil {
		return err
	}

	if equal {
		log.WithField("path", path).Debug("file is already satisfied")
		return nil
	}

	log.WithField("path", path).Debug("update file content")
	return n.ECMS().Files().Create(ctx, path, false, data)
}

func nodeFileEqual(n node.Interface, path string, data []byte) (bool, error) {
	var ctx = context.TODO()
	got, err := n.ECMS().Files().ReadFile(ctx, path)
	if err != nil {
		return false, err
	}

	return bytes.Equal(got, data), nil
}

func generateNGINXServerECephHTTP(upstreams []string, ipVersion string) slb.NginxHTTP {
	slbListen := ""
	if ipVersion == configuration.IPVersionIPV4 {
		slbListen = fmt.Sprintf("%d default_server", NGINXServerPortECephHTTP)
	} else if ipVersion == configuration.IPVersionIPV6 {
		slbListen = fmt.Sprintf("[::]:%d", NGINXServerPortECephHTTP)
	} else {
		panic("unsupported IP Version at generateNGINXServerECephHTTP")
	}
	return slb.NginxHTTP{
		Name: NGINXServerNameECephHTTP,
		Conf: slb.NginxHTTPConf{
			Server: slb.NginxServer{
				Listen:            slbListen,
				ClientMaxBodySize: "0",
				Locations: []map[string]map[string]interface{}{
					{
						"= /crossdomain.xml": {
							"root": "html",
						},
					},
					{
						"/": {
							"proxy_pass":               "http://" + NGINXServerNameECephHTTP,
							"proxy_set_header":         "Host $http_host",
							"proxy_max_temp_file_size": "0",
						},
					},
				},
				ProxyRequestBuffering: "off",
				AccessLog:             "off",
			},
			Upstream: map[string]slb.NginxUpstream{
				NGINXServerNameECephHTTP: {
					CheckHTTPSend: `"HEAD / HTTP/1.1\r\nConnection: keep-alive\r\n\r\n"`,
					Check:         "interval=10000 rise=2 fall=3 timeout=1000 type=http  default_down=true",
					Servers:       upstreams,
					Keepalive:     "300",
				},
			},
		},
	}
}

func generateNGINXServerECephHTTPS(upstreams []string, ipVersion string) slb.NginxHTTP {
	slbListen := ""
	if ipVersion == configuration.IPVersionIPV4 {
		slbListen = fmt.Sprintf("%d ssl", NGINXServerPortECephHTTPS)
	} else if ipVersion == configuration.IPVersionIPV6 {
		slbListen = fmt.Sprintf("[::]:%d ssl", NGINXServerPortECephHTTPS)
	} else {
		panic("unsupported IP Version at generateNGINXServerECephHTTP")
	}
	return slb.NginxHTTP{
		Name: NGINXServerNameECephHTTPS,
		Conf: slb.NginxHTTPConf{
			Server: slb.NginxServer{
				Listen:            slbListen,
				ClientMaxBodySize: "0",
				Locations: []map[string]map[string]interface{}{
					{
						"/": {
							"set":              "$real_request_method $request_method",
							"proxy_method":     "$real_request_method",
							"proxy_set_header": "Host $http_host",
							"proxy_pass":       "http://" + NGINXServerNameECephHTTPS,
							"if": []map[string]map[string]string{
								{
									"($http_x_aishu_real_method = PUT )": {
										"set": "$real_request_method $http_x_aishu_real_method",
									},
								},
								{
									"($http_x_aishu_real_method = DELETE )": {
										"set": "$real_request_method $http_x_aishu_real_method",
									},
								},
								{
									"($request_method != POST )": {
										"set": "$real_request_method $request_method",
									},
								},
							},
						},
					},
					{
						"= /eceph/cert/eceph-server.crt": {
							"alias": "/usr/local/slb-nginx/ssl/eceph-server.crt",
							"if": map[string]interface{}{
								`($request_filename ~* ^.*?\.(txt|pdf|doc|xls|crt)$)`: map[string]interface{}{
									"add_header": `Content-Disposition "attachment;"`,
								},
							},
						},
					},
				},
				ProxyRequestBuffering: "off",
				AccessLog:             "off",
				SSLCertificate:        "/usr/local/slb-nginx/ssl/eceph-server.crt",
				SSLCertificateKey:     "/usr/local/slb-nginx/ssl/eceph-server.key",
				AddHeaders: []string{
					"Access-Control-Allow-Headers * always",
					"Access-Control-Expose-Headers Location,Content-Range,Content-Length,Accept-Ranges,Etag always",
					"Access-Control-Allow-Origin * always",
					"Access-Control-Allow-Methods GET,PUT,POST always",
				},
				IF: map[string]interface{}{
					`($request_method = "OPTIONS")`: map[string]interface{}{
						"return": "204",
					},
				},
			},
			Upstream: map[string]slb.NginxUpstream{
				NGINXServerNameECephHTTPS: {
					CheckHTTPSend: `"HEAD / HTTP/1.1\r\nConnection: keep-alive\r\n\r\n"`,
					Check:         "interval=10000 rise=2 fall=3 timeout=1000 type=http  default_down=true",
					Servers:       upstreams,
					Keepalive:     "300",
				},
			},
		},
	}
}

func reconcileNodeNGINXServer(c slb.NginxHTTPInterface, s *slb.NginxHTTP, log logrus.FieldLogger) error {
	names, err := c.List(context.TODO())
	if err != nil {
		return err
	}

	if !slices.Contains(names, s.Name) {
		log.WithField("server", s).Info("create proton slb nginx http server")
		return c.Create(context.TODO(), s)
	}

	actual, err := c.Get(context.TODO(), s.Name)
	if err != nil {
		return err
	}

	differences := deep.Equal(actual.Conf, s.Conf)
	if differences == nil {
		log.WithField("server", actual).Debug("skip updating proton slb nginx http server")
		return nil
	}

	for _, d := range differences {
		log.WithField("diff", d).Debug("unexpected proton slb nginx http server")
	}

	log.WithField("server", actual).Info("update proton slb nginx http server")
	return c.Update(context.TODO(), s)
}

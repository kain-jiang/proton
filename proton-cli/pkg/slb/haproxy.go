package slb

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

const (
	haProxyConfigPath = "/usr/local/haproxy/haproxy.cfg"
	haProxyService    = "haproxy"
)

func (r *Remote) EnsureHAProxy(ctx context.Context, conf configuration.HaProxyConf) error {
	rendered := renderHAProxyConfig(conf)
	current, exists, err := r.readFileIfExists(ctx, haProxyConfigPath)
	if err != nil {
		return err
	}

	if !exists || !bytes.Equal(current, rendered) {
		if err := r.writeFile(ctx, haProxyConfigPath, rendered); err != nil {
			return err
		}

		if msg, err := r.testHAProxyConfig(); err != nil {
			if exists {
				_ = r.writeFile(ctx, haProxyConfigPath, current)
			} else {
				_ = r.deleteFileIfExists(ctx, haProxyConfigPath)
			}
			if msg != "" {
				return fmt.Errorf("invalid haproxy config: %s", msg)
			}
			return fmt.Errorf("invalid haproxy config: %w", err)
		}
	}

	r.resetFailedService(haProxyService)
	if err := r.exec.Command("systemctl", "reload-or-restart", haProxyService).Run(); err != nil {
		return err
	}
	return r.exec.Command("systemctl", "enable", haProxyService).Run()
}

func (r *Remote) testHAProxyConfig() (string, error) {
	msg, err := r.testCommand("/usr/local/haproxy/sbin/haproxy -c -f " + haProxyConfigPath + " 2>&1")
	return strings.TrimSpace(msg), err
}

func renderHAProxyConfig(conf configuration.HaProxyConf) []byte {
	var b strings.Builder

	writeSimpleBlock(&b, "global", []directive{
		{name: "log", value: conf.Conf.Global.Log},
		{name: "maxconn", value: conf.Conf.Global.Maxconn},
	})
	writeSimpleBlock(&b, "defaults", []directive{
		{name: "maxconn", value: conf.Conf.Defaults.Maxconn},
		{name: "mode", value: conf.Conf.Defaults.Mode},
		{name: "timeout", value: conf.Conf.Defaults.Timeout},
		{name: "option", value: conf.Conf.Defaults.Option},
	})

	for _, fe := range []struct {
		name string
		conf frontendConfig
	}{
		{name: "cs", conf: frontendConfig(conf.Conf.Frontend.CsFrontend)},
		{name: "cr-chartmuseum", conf: frontendConfig(conf.Conf.Frontend.CrChartmuseumFrontend)},
		{name: "cr-registry", conf: frontendConfig(conf.Conf.Frontend.CrRegistryFrontend)},
		{name: "cr-rpm", conf: frontendConfig(conf.Conf.Frontend.CrRpmFrontend)},
	} {
		writeSimpleBlock(&b, "frontend "+fe.name, []directive{
			{name: "bind", value: fe.conf.Bind},
			{name: "mode", value: fe.conf.Mode},
			{name: "default_backend", value: fe.conf.DefaultBackend},
		})
	}

	for _, be := range []struct {
		name string
		conf backendConfig
	}{
		{name: "cs", conf: backendConfig(conf.Conf.Backend.CsBackend)},
		{name: "cr-chartmuseum", conf: backendConfig(conf.Conf.Backend.CrChartmuseumBackend)},
		{name: "cr-registry", conf: backendConfig(conf.Conf.Backend.CrRegistryBackend)},
		{name: "cr-rpm", conf: backendConfig(conf.Conf.Backend.CrRpmBackend)},
	} {
		writeHAProxyBackend(&b, be.name, be.conf)
	}

	return []byte(b.String())
}

type directive struct {
	name  string
	value string
}

type frontendConfig struct {
	Bind           string
	Mode           string
	DefaultBackend string
}

type backendConfig struct {
	Option        []string
	DefaultServer string
	HTTPCheck     string
	Balance       string
	Server        []string
	Mode          string
}

func writeSimpleBlock(b *strings.Builder, header string, directives []directive) {
	b.WriteString(header)
	b.WriteByte('\n')
	for _, item := range directives {
		if item.value == "" {
			continue
		}
		b.WriteString("    ")
		b.WriteString(item.name)
		b.WriteByte(' ')
		b.WriteString(item.value)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
}

func writeHAProxyBackend(b *strings.Builder, name string, conf backendConfig) {
	b.WriteString("backend ")
	b.WriteString(name)
	b.WriteByte('\n')
	for _, option := range conf.Option {
		if option == "" {
			continue
		}
		b.WriteString("    option ")
		b.WriteString(option)
		b.WriteByte('\n')
	}
	for _, item := range []directive{
		{name: "default-server", value: conf.DefaultServer},
		{name: "http-check", value: conf.HTTPCheck},
		{name: "balance", value: conf.Balance},
	} {
		if item.value == "" {
			continue
		}
		b.WriteString("    ")
		b.WriteString(item.name)
		b.WriteByte(' ')
		b.WriteString(item.value)
		b.WriteByte('\n')
	}
	for _, server := range conf.Server {
		if server == "" {
			continue
		}
		b.WriteString("    server ")
		b.WriteString(server)
		b.WriteByte('\n')
	}
	if conf.Mode != "" {
		b.WriteString("    mode ")
		b.WriteString(conf.Mode)
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
}

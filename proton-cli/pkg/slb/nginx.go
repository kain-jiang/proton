package slb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	slbv1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
)

const (
	nginxConfigPath        = "/usr/local/slb-nginx/conf/nginx.conf"
	nginxDefaultConfigPath = "/usr/local/slb-nginx/conf/nginx.conf.default"
	nginxHTTPIncludeDir    = "/usr/local/slb-nginx/conf.d/http"
	nginxLogDir            = "/var/log/slb-nginx"
	nginxService           = "slb-nginx"
	nginxBinary            = "/usr/local/slb-nginx/sbin/slb-nginx"
)

func (r *Remote) EnsureNginxHTTPServer(ctx context.Context, server *slbv1.NginxHTTP) error {
	if err := r.ensureNginxBaseConfig(ctx); err != nil {
		return err
	}

	path := filepath.Join(nginxHTTPIncludeDir, server.Name+".conf")
	rendered, err := renderNginxHTTPServer(server)
	if err != nil {
		return err
	}

	current, exists, err := r.readFileIfExists(ctx, path)
	if err != nil {
		return err
	}
	if exists && bytes.Equal(current, rendered) {
		return nil
	}

	if err := r.writeFile(ctx, path, rendered); err != nil {
		return err
	}

	if msg, err := r.testNginxConfig(); err != nil {
		if exists {
			_ = r.writeFile(ctx, path, current)
		} else {
			_ = r.deleteFileIfExists(ctx, path)
		}
		if msg != "" {
			return fmt.Errorf("invalid nginx config: %s", msg)
		}
		return fmt.Errorf("invalid nginx config: %w", err)
	}

	r.resetFailedService(nginxService)
	if err := r.exec.Command("systemctl", "reload", nginxService).Run(); err != nil {
		return r.exec.Command("systemctl", "restart", nginxService).Run()
	}
	return nil
}

func (r *Remote) ensureNginxBaseConfig(ctx context.Context) error {
	if _, err := r.files.Stat(ctx, nginxConfigPath); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := r.ensureDir(ctx, nginxHTTPIncludeDir); err != nil {
		return err
	}
	if err := r.ensureDir(ctx, nginxLogDir); err != nil {
		return err
	}

	defaultConfig, err := r.files.ReadFile(ctx, nginxDefaultConfigPath)
	if err != nil {
		return err
	}
	if err := r.writeFile(ctx, nginxConfigPath, defaultConfig); err != nil {
		return err
	}

	if msg, err := r.testNginxConfig(); err != nil {
		_ = r.deleteFileIfExists(ctx, nginxConfigPath)
		if msg != "" {
			return fmt.Errorf("invalid nginx config: %s", msg)
		}
		return fmt.Errorf("invalid nginx config: %w", err)
	}

	r.resetFailedService(nginxService)
	if err := r.exec.Command("systemctl", "enable", nginxService).Run(); err != nil {
		return err
	}
	return r.exec.Command("systemctl", "restart", nginxService).Run()
}

func (r *Remote) testNginxConfig() (string, error) {
	msg, err := r.testCommand(nginxBinary + " -t 2>&1")
	return strings.TrimSpace(msg), err
}

func renderNginxHTTPServer(server *slbv1.NginxHTTP) ([]byte, error) {
	var b strings.Builder

	upstreamNames := make([]string, 0, len(server.Conf.Upstream))
	for name := range server.Conf.Upstream {
		upstreamNames = append(upstreamNames, name)
	}
	sort.Strings(upstreamNames)

	for _, name := range upstreamNames {
		upstream := server.Conf.Upstream[name]
		b.WriteString("upstream ")
		b.WriteString(name)
		b.WriteString(" {\n")
		if upstream.CheckHTTPSend != "" {
			writeIndentedDirective(&b, 1, "check_http_send", upstream.CheckHTTPSend)
		}
		if upstream.Check != "" {
			writeIndentedDirective(&b, 1, "check", upstream.Check)
		}
		for _, item := range stringifySlice(upstream.Servers) {
			writeIndentedDirective(&b, 1, "server", item)
		}
		if upstream.Keepalive != "" {
			writeIndentedDirective(&b, 1, "keepalive", upstream.Keepalive)
		}
		b.WriteString("}\n\n")
	}

	b.WriteString("server {\n")
	s := server.Conf.Server
	for _, item := range []directive{
		{name: "listen", value: s.Listen},
		{name: "client_max_body_size", value: s.ClientMaxBodySize},
		{name: "proxy_request_buffering", value: s.ProxyRequestBuffering},
		{name: "access_log", value: s.AccessLog},
		{name: "ssl_certificate", value: s.SSLCertificate},
		{name: "ssl_certificate_key", value: s.SSLCertificateKey},
	} {
		if item.value != "" {
			writeIndentedDirective(&b, 1, item.name, item.value)
		}
	}
	for _, item := range s.AddHeaders {
		writeIndentedDirective(&b, 1, "add_header", item)
	}
	if err := renderNginxIf(&b, 1, s.IF); err != nil {
		return nil, err
	}
	for _, location := range s.Locations {
		if err := renderNginxLocation(&b, location); err != nil {
			return nil, err
		}
	}
	b.WriteString("}\n")

	return []byte(b.String()), nil
}

func renderNginxLocation(b *strings.Builder, location map[string]map[string]interface{}) error {
	names := make([]string, 0, len(location))
	for name := range location {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		b.WriteString("    location ")
		b.WriteString(name)
		b.WriteString(" {\n")
		if err := renderNginxDirectiveMap(b, 2, location[name]); err != nil {
			return err
		}
		b.WriteString("    }\n")
	}
	return nil
}

func renderNginxDirectiveMap(b *strings.Builder, indent int, directives map[string]interface{}) error {
	keys := make([]string, 0, len(directives))
	for key := range directives {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if key == "if" {
			if err := renderNginxIf(b, indent, directives[key]); err != nil {
				return err
			}
			continue
		}
		if err := renderNginxValue(b, indent, key, directives[key]); err != nil {
			return err
		}
	}
	return nil
}

func renderNginxIf(b *strings.Builder, indent int, value interface{}) error {
	if value == nil {
		return nil
	}

	for _, item := range reflectSlice(value) {
		switch typed := item.(type) {
		case map[string]interface{}:
			if err := renderNginxIfBlockMap(b, indent, typed); err != nil {
				return err
			}
		case map[string]string:
			converted := make(map[string]interface{}, len(typed))
			for key, val := range typed {
				converted[key] = val
			}
			if err := renderNginxIfBlockMap(b, indent, converted); err != nil {
				return err
			}
		case map[string]map[string]string:
			converted := make(map[string]interface{}, len(typed))
			for key, val := range typed {
				body := make(map[string]interface{}, len(val))
				for innerKey, innerVal := range val {
					body[innerKey] = innerVal
				}
				converted[key] = body
			}
			if err := renderNginxIfBlockMap(b, indent, converted); err != nil {
				return err
			}
		case map[string]map[string]interface{}:
			converted := make(map[string]interface{}, len(typed))
			for key, val := range typed {
				converted[key] = val
			}
			if err := renderNginxIfBlockMap(b, indent, converted); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported nginx if value type %T", value)
		}
	}

	return nil
}

func renderNginxIfBlockMap(b *strings.Builder, indent int, blocks map[string]interface{}) error {
	conditions := make([]string, 0, len(blocks))
	for cond := range blocks {
		conditions = append(conditions, cond)
	}
	sort.Strings(conditions)

	for _, cond := range conditions {
		writeIndent(b, indent)
		b.WriteString("if ")
		b.WriteString(cond)
		b.WriteString(" {\n")

		body, ok := blocks[cond]
		if !ok {
			continue
		}
		switch typed := body.(type) {
		case map[string]interface{}:
			if err := renderNginxDirectiveMap(b, indent+1, typed); err != nil {
				return err
			}
		case map[string]string:
			converted := make(map[string]interface{}, len(typed))
			for key, val := range typed {
				converted[key] = val
			}
			if err := renderNginxDirectiveMap(b, indent+1, converted); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported nginx if block type %T", body)
		}

		writeIndent(b, indent)
		b.WriteString("}\n")
	}
	return nil
}

func renderNginxValue(b *strings.Builder, indent int, name string, value interface{}) error {
	switch typed := value.(type) {
	case string:
		writeIndentedDirective(b, indent, name, typed)
		return nil
	case []string:
		for _, item := range typed {
			writeIndentedDirective(b, indent, name, item)
		}
		return nil
	case map[string]interface{}:
		if name == "if" {
			return renderNginxIf(b, indent, typed)
		}
	case map[string]string:
		if name == "if" {
			converted := make(map[string]interface{}, len(typed))
			for key, val := range typed {
				converted[key] = val
			}
			return renderNginxIf(b, indent, converted)
		}
	}

	for _, item := range reflectSlice(value) {
		if text, ok := item.(string); ok {
			writeIndentedDirective(b, indent, name, text)
			continue
		}
		return fmt.Errorf("unsupported nginx directive type %T for %s", value, name)
	}
	return nil
}

func writeIndentedDirective(b *strings.Builder, indent int, name, value string) {
	writeIndent(b, indent)
	b.WriteString(name)
	if value != "" {
		b.WriteByte(' ')
		b.WriteString(value)
	}
	b.WriteString(";\n")
}

func writeIndent(b *strings.Builder, indent int) {
	for range indent {
		b.WriteString("    ")
	}
}

func stringifySlice(value interface{}) []string {
	items := reflectSlice(value)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func reflectSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return []interface{}{value}
	}

	result := make([]interface{}, 0, v.Len())
	for i := range v.Len() {
		result = append(result, v.Index(i).Interface())
	}
	return result
}

package app

import (
	"reflect"
	"testing"
)

func TestAppInstallCompletedMessage(t *testing.T) {
	got := appInstallCompletedMessage("release-manifests/0.5.0/kweaver-dip.yaml", "kweaver")
	want := "App install completed successfully: manifest=release-manifests/0.5.0/kweaver-dip.yaml namespace=kweaver"
	if got != want {
		t.Fatalf("appInstallCompletedMessage() = %q, want %q", got, want)
	}
}

func TestParseAccessAddressURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want map[string]interface{}
	}{
		{
			name: "https host only uses default https port and root path",
			raw:  "https://1.1.1.1/",
			want: map[string]interface{}{
				"host":   "1.1.1.1",
				"port":   443,
				"scheme": "https",
				"path":   "/",
			},
		},
		{
			name: "http path uses default http port",
			raw:  "http://1.1.1.1/api",
			want: map[string]interface{}{
				"host":   "1.1.1.1",
				"port":   80,
				"scheme": "http",
				"path":   "/api",
			},
		},
		{
			name: "explicit port is preserved",
			raw:  "https://1.1.1.1:8443/",
			want: map[string]interface{}{
				"host":   "1.1.1.1",
				"port":   8443,
				"scheme": "https",
				"path":   "/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAccessAddressURL(tt.raw)
			if err != nil {
				t.Fatalf("parseAccessAddressURL() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseAccessAddressURL() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseAccessAddressURLRejectsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "missing scheme",
			raw:  "1.1.1.1",
		},
		{
			name: "unsupported scheme",
			raw:  "tcp://1.1.1.1:443",
		},
		{
			name: "query not allowed",
			raw:  "https://1.1.1.1/?a=1",
		},
		{
			name: "fragment not allowed",
			raw:  "https://1.1.1.1/#frag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseAccessAddressURL(tt.raw); err == nil {
				t.Fatalf("parseAccessAddressURL(%q) error = nil, want error", tt.raw)
			}
		})
	}
}

func TestResolveAccessAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		flags        appInstallFlags
		existing     map[string]interface{}
		detectedHost string
		want         map[string]interface{}
	}{
		{
			name:  "preserves existing when no override flags are set",
			flags: appInstallFlags{},
			existing: map[string]interface{}{
				"host":   "existing.example.com",
				"port":   9443,
				"scheme": "https",
				"path":   "/portal",
			},
			want: map[string]interface{}{
				"host":   "existing.example.com",
				"port":   9443,
				"scheme": "https",
				"path":   "/portal",
			},
		},
		{
			name: "full access address overrides all fields",
			flags: appInstallFlags{
				accessAddress: "https://1.1.1.1:8443/",
			},
			existing: map[string]interface{}{
				"host":   "existing.example.com",
				"port":   9443,
				"scheme": "https",
				"path":   "/portal",
			},
			want: map[string]interface{}{
				"host":   "1.1.1.1",
				"port":   8443,
				"scheme": "https",
				"path":   "/",
			},
		},
		{
			name: "full access address wins over split flags",
			flags: appInstallFlags{
				accessAddress:   "https://1.1.1.1:8443/",
				accessHost:      "override.example.com",
				accessPort:      8080,
				accessPortSet:   true,
				accessScheme:    "http",
				accessSchemeSet: true,
			},
			existing: map[string]interface{}{
				"host":   "existing.example.com",
				"port":   9443,
				"scheme": "https",
				"path":   "/portal",
			},
			want: map[string]interface{}{
				"host":   "1.1.1.1",
				"port":   8443,
				"scheme": "https",
				"path":   "/",
			},
		},
		{
			name: "split flags override selected fields and preserve the rest",
			flags: appInstallFlags{
				accessHost:      "override.example.com",
				accessScheme:    "http",
				accessSchemeSet: true,
			},
			existing: map[string]interface{}{
				"host":   "existing.example.com",
				"port":   9443,
				"scheme": "https",
				"path":   "/portal",
			},
			want: map[string]interface{}{
				"host":   "override.example.com",
				"port":   9443,
				"scheme": "http",
				"path":   "/portal",
			},
		},
		{
			name:         "falls back to detected host and defaults",
			flags:        appInstallFlags{},
			detectedHost: "10.0.0.10",
			want: map[string]interface{}{
				"host":   "10.0.0.10",
				"port":   443,
				"scheme": "https",
				"path":   "/",
			},
		},
		{
			name:  "falls back to localhost when detection is empty",
			flags: appInstallFlags{},
			want: map[string]interface{}{
				"host":   "localhost",
				"port":   443,
				"scheme": "https",
				"path":   "/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAccessAddress(tt.existing, &tt.flags, func() string {
				return tt.detectedHost
			})
			if err != nil {
				t.Fatalf("resolveAccessAddress() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("resolveAccessAddress() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseAppInstallSetFlags(t *testing.T) {
	t.Parallel()
	entries := []string{
		"auth.enabled=false",
		"businessDomain.enabled=false",
		"services.replicas=3",
		"name=prod",
		"auth.tls.enabled=true",
	}
	got, err := parseAppInstallSet(entries)
	if err != nil {
		t.Fatalf("parseAppInstallSet() error = %v", err)
	}
	want := map[string]interface{}{
		"auth": map[string]interface{}{
			"enabled": false,
			"tls": map[string]interface{}{
				"enabled": true,
			},
		},
		"businessDomain": map[string]interface{}{"enabled": false},
		"services":        map[string]interface{}{"replicas": 3},
		"name":            "prod",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAppInstallSet() = %#v, want %#v", got, want)
	}
}

func TestParseAppInstallSetFlagsRejectsInvalid(t *testing.T) {
	t.Parallel()
	invalid := []string{
		"noequals",
		"auth..enabled=true",
		".leading=1",
		"trailing.=",
		"=missingkey",
	}
	for _, input := range invalid {
		if _, err := parseAppInstallSet([]string{input}); err == nil {
			t.Fatalf("parseAppInstallSet(%q) error = nil, want error", input)
		}
	}
}

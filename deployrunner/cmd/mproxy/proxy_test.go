package main

import (
	"strings"
	"testing"
)

func TestRe(t *testing.T) {
	raw := "/api/deploy-mproxy/api/eacp/v1/auth1/login-configs"
	prefix := ""
	removePrefix := "/api/deploy-mproxy"
	backend := prefix + strings.TrimPrefix(raw, removePrefix)
	if backend != "/api/eacp/v1/auth1/login-configs" {
		t.Error(backend)
		t.FailNow()
	}
}

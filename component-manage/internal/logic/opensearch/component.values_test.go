package opensearch

import (
	"testing"

	"component-manage/internal/config"
	"component-manage/internal/global"
	"component-manage/pkg/models/types"
)

func TestGenerateInfoUsesExpectedDefaultCredentials(t *testing.T) {
	previousConfig := global.Config
	t.Cleanup(func() {
		global.Config = previousConfig
	})

	global.Config = &config.Config{}
	global.Config.Internal.ClusterDomain = "cluster.local"

	info := generateInfo("opensearch", &types.OpensearchComponentParams{
		Namespace: "resource",
	})

	if info.Username != "admin" {
		t.Fatalf("unexpected username: %q", info.Username)
	}
	if info.Password != "eisoo.com123" {
		t.Fatalf("unexpected password: %q", info.Password)
	}
	if info.Hosts != "opensearch-master.resource.svc.cluster.local." {
		t.Fatalf("unexpected hosts: %q", info.Hosts)
	}
	if info.Port != 9200 {
		t.Fatalf("unexpected port: %d", info.Port)
	}
	if info.Protocol != "http" {
		t.Fatalf("unexpected protocol: %q", info.Protocol)
	}
}

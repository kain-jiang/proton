package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestUpgradeConfigForDisplayKeepsDeployOnlyForNamespaceAndServiceAccount(t *testing.T) {
	conf := &configuration.ClusterConfig{
		Deploy: &configuration.Deploy{
			Namespace:      "resource",
			ServiceAccount: "svc-account",
		},
		ResourceConnectInfo: &configuration.ResourceConnectInfo{},
	}

	upgradeConfigForDisplay(conf)

	require.NotNil(t, conf.Deploy)
	require.Equal(t, "resource", conf.Deploy.Namespace)
	require.Equal(t, "svc-account", conf.Deploy.ServiceAccount)
	require.NotNil(t, conf.ComponentManage)
}

func TestUpgradeConfigForDisplayDoesNotCreateDeployWhenMissing(t *testing.T) {
	conf := &configuration.ClusterConfig{
		ResourceConnectInfo: &configuration.ResourceConnectInfo{},
	}

	upgradeConfigForDisplay(conf)

	require.Nil(t, conf.Deploy)
	require.NotNil(t, conf.ComponentManage)
}

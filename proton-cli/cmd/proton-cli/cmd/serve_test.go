package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/require"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
)

func TestDecodeInitRequestSupportsLegacyClusterConfig(t *testing.T) {
	body := []byte(`{"apiVersion":"v1","nodes":[],"chrony":{"mode":"usermanaged"},"firewall":{"mode":"usermanaged"},"cs":{"provisioner":"external","namespace":"ns","serviceaccount":"sa","addons":[]},"cr":{"external":{"chart_repository":"chartmuseum","image_repository":"registry","registry":{"host":"","username":"","password":""},"chartmuseum":{"host":"","username":"","password":""},"oci":{"registry":"","username":"","password":"","plain_http":false}}},"component_management":{},"resource_connect_info":{"rds":{"source_type":"external","rds_type":"MySQL","auto_create_database":false,"admin_user":"","admin_passwd":"","hosts":"","port":3306,"username":"","password":""},"redis":{"source_type":"external","connect_type":"standalone","username":"","password":"","sentinel_hosts":"","sentinel_port":26379,"master_group_name":"","master_hosts":"","master_port":6379,"slave_hosts":"","slave_port":6379,"hosts":"","port":6379},"opensearch":{"source_type":"external","hosts":"","port":9200,"username":"","password":"","distribution":"elasticsearch","version":"2.11.0"},"mq":{"source_type":"external","mq_type":"kafka","mq_hosts":"","mq_port":9092,"mq_lookupd_hosts":"","mq_lookupd_port":4161,"auth":{"username":"","password":"","mechanism":"PLAIN"}}}}`)

	req, err := decodeInitRequest(body)
	require.NoError(t, err)
	require.Empty(t, req.ServicePackageDir)
	require.NotNil(t, req.ClusterConfig)
	require.Equal(t, "v1", req.ClusterConfig.ApiVersion)
}

func TestDecodeInitRequestSupportsWrappedPayload(t *testing.T) {
	body := []byte(`{"service_package_dir":"/srv/packages/service-package","cluster_config":{"apiVersion":"v1","nodes":[],"chrony":{"mode":"usermanaged"},"firewall":{"mode":"usermanaged"},"cs":{"provisioner":"external","namespace":"ns","serviceaccount":"sa","addons":[]},"cr":{"external":{"chart_repository":"chartmuseum","image_repository":"registry","registry":{"host":"","username":"","password":""},"chartmuseum":{"host":"","username":"","password":""},"oci":{"registry":"","username":"","password":"","plain_http":false}}},"component_management":{},"resource_connect_info":{"rds":{"source_type":"external","rds_type":"MySQL","auto_create_database":false,"admin_user":"","admin_passwd":"","hosts":"","port":3306,"username":"","password":""},"redis":{"source_type":"external","connect_type":"standalone","username":"","password":"","sentinel_hosts":"","sentinel_port":26379,"master_group_name":"","master_hosts":"","master_port":6379,"slave_hosts":"","slave_port":6379,"hosts":"","port":6379},"opensearch":{"source_type":"external","hosts":"","port":9200,"username":"","password":"","distribution":"elasticsearch","version":"2.11.0"},"mq":{"source_type":"external","mq_type":"kafka","mq_hosts":"","mq_port":9092,"mq_lookupd_hosts":"","mq_lookupd_port":4161,"auth":{"username":"","password":"","mechanism":"PLAIN"}}}}}`)

	req, err := decodeInitRequest(body)
	require.NoError(t, err)
	require.Equal(t, "/srv/packages/service-package", req.ServicePackageDir)
	require.NotNil(t, req.ClusterConfig)
	require.Equal(t, "v1", req.ClusterConfig.ApiVersion)
}

func TestInitialHandlerUsesRequestScopedServicePackage(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	originalServicePackage := global.ServicePackage
	defer func() {
		global.ServicePackage = originalServicePackage
	}()

	global.ServicePackage = "./service-package"

	patches.ApplyFunc(configuration.UpdateProtonCliEnvConfig, func(string) error {
		return nil
	})

	appliedServicePackage := ""
	patches.ApplyFunc(applyClusterConfigForTestHook, func(conf *configuration.ClusterConfig) error {
		appliedServicePackage = global.ServicePackage
		require.Equal(t, "v1", conf.ApiVersion)
		return nil
	})

	handler := &initialHandler{cancel: func() {}}
	requestBody := `{"service_package_dir":"/srv/packages/service-package","cluster_config":{"apiVersion":"v1","nodes":[],"chrony":{"mode":"usermanaged"},"firewall":{"mode":"usermanaged"},"cs":{"provisioner":"external","namespace":"ns","serviceaccount":"sa","addons":[]},"cr":{"external":{"chart_repository":"chartmuseum","image_repository":"registry","registry":{"host":"","username":"","password":""},"chartmuseum":{"host":"","username":"","password":""},"oci":{"registry":"","username":"","password":"","plain_http":false}}},"component_management":{},"resource_connect_info":{"rds":{"source_type":"external","rds_type":"MySQL","auto_create_database":false,"admin_user":"","admin_passwd":"","hosts":"","port":3306,"username":"","password":""},"redis":{"source_type":"external","connect_type":"standalone","username":"","password":"","sentinel_hosts":"","sentinel_port":26379,"master_group_name":"","master_hosts":"","master_port":6379,"slave_hosts":"","slave_port":6379,"hosts":"","port":6379},"opensearch":{"source_type":"external","hosts":"","port":9200,"username":"","password":"","distribution":"elasticsearch","version":"2.11.0"},"mq":{"source_type":"external","mq_type":"kafka","mq_hosts":"","mq_port":9092,"mq_lookupd_hosts":"","mq_lookupd_port":4161,"auth":{"username":"","password":"","mechanism":"PLAIN"}}}}}`

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/init", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	response := recorder.Result()
	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode, string(respBody))
	require.Equal(t, "/srv/packages/service-package", appliedServicePackage)
	require.Equal(t, "./service-package", global.ServicePackage)
}


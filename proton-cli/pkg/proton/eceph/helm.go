package eceph

import (
	"context"
	"net"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/eceph/helm"
)

const IPV4AnyAddress string = "0.0.0.0"
const IPV6AnyAddress string = "::"
const IPV4Localhost string = "127.0.0.1"

func (m *Manager) chartPath(name string) string {
	return path.Join(m.PkgECeph.BaseDir(), m.PkgECeph.Charts().Get(name, "").Path)
}

func (m *Manager) reconcileHelmReleases() error {
	m.Logger.Info("reconcile helm releases")

	rgwHost, rgwPort, rgwProtocol := generateRGWConnectionInfo(m.Spec.Keepalived, m.Nodes[0].IP())
	bindAddr := IPV4AnyAddress
	tenantMgrHost := IPV4Localhost
	if m.Nodes[0].IPVersion() == configuration.IPVersionIPV6 {
		tenantMgrHost = IPV6AnyAddress
		bindAddr = IPV6AnyAddress
	}

	for _, r := range []struct {
		release string
		chart   string
		values  *helm.Values4ECeph
	}{
		{
			release: helm.ReleaseNameConfigManager,
			chart:   m.chartPath(helm.ChartNameConfigManager),
			values:  helm.ValuesForConfigManager(m.Registry, m.RDS, rgwHost, rgwPort, rgwProtocol, bindAddr),
		},
		{
			release: helm.ReleaseNameConfigWeb,
			chart:   m.chartPath(helm.ChartNameConfigWeb),
			values:  helm.ValuesForConfigWeb(m.Registry),
		},
		{
			release: helm.ReleaseNameNodeAgent,
			chart:   m.chartPath(helm.ChartNameNodeAgent),
			values:  helm.ValuesForNodeAgent(m.Registry, bindAddr),
		},
		{
			release: helm.ReleaseNameTenantWeb,
			chart:   m.chartPath(helm.ChartNameTenantWeb),
			values:  helm.ValuesForTenantWeb(m.Registry, rgwHost, rgwPort, rgwProtocol, bindAddr, tenantMgrHost),
		},
	} {
		m.Logger.WithFields(logrus.Fields{"release": r.release, "chart": r.chart}).Debug("reconcile helm release")
		if err := m.Helm.Reconcile(context.TODO(), r.release, r.chart, r.values.ToMap()); err != nil {
			return err
		}
	}

	return nil
}

func generateRGWConnectionInfo(k *configuration.ECephKeepalived, firstNodeIP net.IP) (host string, port int, protocol string) {
	if k != nil && len(k.External) > 0 {
		host = strings.Split(k.External, "/")[0]
	} else {
		host = strings.Split(firstNodeIP.String(), "/")[0]
	}

	// ECeph services in Kubernetes access RGW through NGINX HTTP
	port = NGINXServerPortECephHTTP

	protocol = "http"

	return
}

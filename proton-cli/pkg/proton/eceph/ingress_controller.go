package eceph

import (
	"context"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/eceph/helm"
)

func (m *Manager) reconcileIngressControllers() error {
	m.Logger.Info("reconcile kubernetes ingress controllers")
	for _, r := range []struct {
		release string
		chart   string
		port    int
		class   string
	}{
		{
			release: helm.ReleaseNameNGINX_ECeph,
			chart:   m.chartPath(helm.ChartNameNGINXIngressController),
			port:    NGINXIngressControllerPortProtonECeph,
			class:   IngressClassNGINX_ECeph,
		},
		{
			release: helm.ReleaseNameNGINX_ECephTenantWeb,
			chart:   m.chartPath(helm.ChartNameNGINXIngressController),
			port:    NGINXIngressControllerPortProtonECephTenantWeb,
			class:   IngressClassNGINX_ECephTenantWeb,
		},
	} {
		m.Logger.WithFields(logrus.Fields{"release": r.release, "chart": r.chart}).Debug("reconcile kubernetes ingress controller")
		if err := m.Helm.Reconcile(context.TODO(), r.release, r.chart, helm.ValuesForNGINXIngressController(m.Registry, r.port, r.class).ToMap()); err != nil {
			return err
		}
	}
	return nil
}

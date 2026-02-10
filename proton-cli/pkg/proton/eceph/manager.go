package eceph

import (
	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"

	"sigs.k8s.io/controller-runtime/pkg/client"

	helm "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2"
	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	rds_mgmt "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// Manager implements Apply and Reset methods.
type Manager struct {
	Spec    *configuration.ECeph
	OldSpec *configuration.ECeph

	// container registry address
	Registry string

	Nodes []node.Interface

	Kube client.Client

	// RDS connection info
	RDS *configuration.RdsInfo
	// Cache of created clients to avoid repeated creation.
	rdsMGMTClient rds_mgmt.Interface
	// Whether to initialize the database
	InitDatabase bool
	// Function to create a rds mgmt client
	RDS_MGMTClientCreateFunc func() (rds_mgmt.Interface, error)

	Helm helm.Interface

	Logger logrus.FieldLogger

	PkgECeph *servicepackage.ServicePackage
}

func (m *Manager) Apply() error {
	steps := []func() error{
		m.reconcileCertificate,
		m.reconcileNGINXServers,
		m.reconcileDatabase,
		m.reconcileDatabaseUser,
		m.reconcileHelmReleases,
		m.reconcileIngresses,
		m.reconcileIngressControllers,
		m.executeECephAgentConfig,
		m.reconcileSystemdUnits,
		m.reconcileAllowedLocalPorts,
		m.reconcileKeepalivedHAInstances,
	}
	if m.Spec.SkipECephUpdate {
		m.Spec.SkipECephUpdate = false
		m.Logger.Info("skip_eceph_update is true, not performing ECeph update operations this time")
	} else {
		for _, s := range steps {
			if err := s(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Reset() error {
	steps := []func() error{
		m.resetKeepalivedHAInstances,
	}
	for _, s := range steps {
		if err := s(); err != nil {
			return err
		}
	}
	return nil
}

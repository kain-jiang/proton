package eceph

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	firewalld_v1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/firewalld/v1alpha1"
)

func (m *Manager) reconcileAllowedLocalPorts() error {
	m.Logger.Info("reconcile allowed local ports")
	for _, n := range m.Nodes {
		log := m.Logger.WithField("node", n.Name())

		// permanent config
		if err := reconcileNodeAllowedLocalPorts(n.Firewalld(), true, log); err != nil {
			return err
		}

		// skip changing runtime config if firewalld is not running
		if r, err := n.Firewalld().IsRunning(); err != nil {
			return err
		} else if !r {
			log.Debug("skip changing runtime because firewalld is not running")
			continue
		}

		// runtime config
		if err := reconcileNodeAllowedLocalPorts(n.Firewalld(), false, log); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) reconcileFirewalldVirtualIPs() error {
	m.Logger.Info("reconcile that firewalld allows for keepalived virtual ip")
	oldInternalVIP := ""
	oldExternalVIP := ""
	if m.OldSpec != nil && m.OldSpec.Keepalived != nil {
		oldInternalVIP = m.OldSpec.Keepalived.Internal
		oldExternalVIP = m.OldSpec.Keepalived.External
	}

	for _, n := range m.Nodes {
		log := m.Logger.WithField("node", n.Name())

		// permanent config
		if len(m.Spec.Keepalived.Internal) > 0 {
			if err := n.Firewalld().RemoveAndAddFirewallSource(m.Spec.Keepalived.Internal, oldInternalVIP, true); err != nil {
				return err
			}
		}
		if len(m.Spec.Keepalived.External) > 0 {
			if err := n.Firewalld().RemoveAndAddFirewallSource(m.Spec.Keepalived.External, oldExternalVIP, true); err != nil {
				return err
			}
		}

		// skip changing runtime config if firewalld is not running
		if r, err := n.Firewalld().IsRunning(); err != nil {
			return err
		} else if !r {
			log.Debug("skip changing runtime because firewalld is not running")
			continue
		}

		// runtime config
		if len(m.Spec.Keepalived.Internal) > 0 {
			if err := n.Firewalld().RemoveAndAddFirewallSource(m.Spec.Keepalived.Internal, oldInternalVIP, false); err != nil {
				return err
			}
		}
		if len(m.Spec.Keepalived.External) > 0 {
			if err := n.Firewalld().RemoveAndAddFirewallSource(m.Spec.Keepalived.External, oldExternalVIP, false); err != nil {
				return err
			}
		}
	}

	return nil
}

func reconcileNodeAllowedLocalPorts(c firewalld_v1alpha1.Interface, permanent bool, log logrus.FieldLogger) error {
	const (
		// allowed local ports in default zone. If option zone is empty, use
		// default zone.
		z = ""
		// Empty option timeout means forever
		t = ""
	)

	log = log.WithFields(logrus.Fields{"zone": z, "timeout": t, "permanent": permanent})

	// list ports of default zone
	ports, err := c.ListPorts(permanent, z)
	if err != nil {
		return err
	}

	for _, p := range AllowedAccessLocalPorts {
		if slices.Contains(ports, p) {
			log.WithField("port", p).Debug("skip already existing port")
			continue
		}

		log.WithField("port", p).Info("add allowed local port")
		if err := c.AddPort(p, permanent, z, t); err != nil {
			return err
		}
	}

	return nil
}

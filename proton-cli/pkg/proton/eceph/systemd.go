package eceph

import "github.com/sirupsen/logrus"

func (m *Manager) reconcileSystemdUnits() error {
	m.Logger.Info("reconcile systemd units")
	for _, n := range m.Nodes {
		for _, u := range SystemdUnits {
			log := m.Logger.WithFields(logrus.Fields{"node": n.Name(), "unit": u})

			enabled, err := n.Systemd().IsEnabled(u)
			if err != nil {
				return err
			}
			active, err := n.Systemd().IsActive(u)
			if err != nil {
				return err
			}

			switch {
			case !active && !enabled:
				log.Info("enable and start systemd unit")
				err = n.Systemd().Enabled(u, true)
			case !active && enabled:
				log.Info("start systemd unit")
				err = n.Systemd().Start(u)
			case active && !enabled:
				log.Info("enable systemd unit")
				err = n.Systemd().Enabled(u, false)
			default:
				log.Debug("systemd unit is already active and enabled")
			}

			if err != nil {
				return err
			}
		}

	}
	return nil
}

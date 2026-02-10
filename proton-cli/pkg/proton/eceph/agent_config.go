package eceph

func (m *Manager) executeECephAgentConfig() error {
	for _, n := range m.Nodes {
		m.Logger.WithField("node", n.Name()).Info("execute ECeph agent_config")
		if err := n.ECephAgentConfig().ShouldExecuteAgentConfig(); err != nil {
			m.Logger.WithField("node", n.Name()).Info("ECeph Config Agent is not running on this node, should try running agent_config binary")
			if err := n.ECephAgentConfig().ExecuteWithTimeout(ExecuteECephAgentConfigTimeout, n.IP().String(), n.InternalIP().String(), n.IP().String(), n.Name()); err != nil {
				return err
			}
		} else {
			m.Logger.WithField("node", n.Name()).Info("ECeph Agent Config is healthy on this node, skip running agent_config binary")
		}
	}
	return nil
}

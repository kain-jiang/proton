package helm3

import (
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type Client interface {
	Install(release string, chart *chart.Chart, opts ...InstallOption) (*release.Release, error)
	Upgrade(release string, chart *chart.Chart, opts ...UpgradeOption) (*release.Release, error)
	Uninstall(release string, opts ...UninstallOption) (*release.Release, error)

	GetRelease(name string, opts ...GetOption) (*release.Release, error)
	HistoryRelease(name string, opts ...HistoryOption) ([]*release.Release, error)

	// NameSpace 切换命名空间
	NameSpace(ns string) Client

	PullChart(name, version string, reg *OCIRegistryConfig) (string, func(), error)
	Push(f string, reg *OCIRegistryConfig) error
}

type (
	L = []any
	M = map[string]any
)

type helmv3 struct {
	actionConfig *action.Configuration
	settings     *helmcli.EnvSettings
	namespace    string
	log          *logrus.Entry
}

func NewCli(namespace string, log *logrus.Entry) (Client, error) {
	if log == nil {
		_log := logrus.New()
		_log.SetLevel(logrus.DebugLevel)
		log = _log.WithField("helm", "v3")
	}

	settings := helmcli.New()
	actionConfig := new(action.Configuration)

	err := actionConfig.Init(settings.RESTClientGetter(), namespace, "", log.Debugf)
	if err != nil {
		log.WithError(err).Errorln("action config init failed")
		return nil, err
	}

	return &helmv3{
		actionConfig: actionConfig,
		settings:     settings,
		namespace:    namespace,
		log:          log,
	}, nil
}

func (c *helmv3) NameSpace(ns string) Client {
	if c.namespace == ns {
		return c
	}
	nsActionConfig := new(action.Configuration)
	_ = nsActionConfig.Init(c.settings.RESTClientGetter(), ns, "", c.log.Debugf)
	// helmDriver 为 sql 时才有可能出现错误
	return &helmv3{
		actionConfig: nsActionConfig,
		settings:     c.settings,
		namespace:    ns,
		log:          c.log,
	}
}

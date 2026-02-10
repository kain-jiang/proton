package helm3

import (
	"encoding/json"
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type upgradeParams struct {
	values          map[string]interface{}
	force           bool
	wait            bool
	timeout         time.Duration
	dryRun          bool
	install         bool
	atomic          bool
	createNamespace bool
	skipCRDs        bool
	reCreatePods    bool
}

type UpgradeOption func(upgradeParams *upgradeParams)

func WithUpgradeValues(valuesMap map[string]interface{}) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.values = valuesMap
	}
}

func WithUpgradeValuesAny(object interface{}) UpgradeOption {
	bytes, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	values := make(map[string]interface{})
	err = json.Unmarshal(bytes, &values)
	if err != nil {
		panic(err)
	}
	return func(upgradeParams *upgradeParams) {
		upgradeParams.values = values
	}
}

func WithUpgradeForce(force bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.force = force
	}
}

func WithUpgradeWait(wait bool, timeout time.Duration) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.wait = wait
		upgradeParams.timeout = timeout
	}
}

func WithUpgradeDryRun(dryRun bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.dryRun = dryRun
	}
}

func WithUpgradeInstall(install bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.install = install
	}
}

func WithUpgradeRecreatePods(recreatepods bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.reCreatePods = recreatepods
	}
}

func WithUpgradeAtoMic(atomic bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.atomic = atomic
		// 使用sdk的实现， wait 将自动打开, 需要设置 timeout
		upgradeParams.timeout = 5 * time.Minute
	}
}

func WithUpgradeCreateNamespace(createNamespace bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.createNamespace = createNamespace
	}
}

func WithUpgradeSkipCRDs(skipCRDs bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.skipCRDs = skipCRDs
	}
}

func (c *helmv3) Upgrade(release string, chartRef *ChartRef, opts ...UpgradeOption) error {

	log := c.log.WithField("release", release).WithField("operation", "upgrade")
	param := &upgradeParams{}
	for _, opt := range opts {
		opt(param)
	}
	log.Debugf("upgrade with values: %+v", param.values)

	if param.install {
		// 获取当前旧版本
		historier := action.NewHistory(c.actionConfig)
		historier.Max = 1
		historyReleases, err := historier.Run(release)
		if err == nil && len(historyReleases) > 0 {
			param.install = false
		} else {
			log.WithError(err).Debugln("get release history failed")
		}
	}

	upgrader := action.NewUpgrade(c.actionConfig)
	upgrader.Namespace = c.namespace
	upgrader.Force = param.force
	upgrader.Wait = param.wait
	upgrader.Timeout = param.timeout
	upgrader.DryRun = param.dryRun
	upgrader.SkipCRDs = param.skipCRDs
	upgrader.Atomic = param.atomic
	upgrader.Recreate = param.reCreatePods

	if param.install {
		installer := action.NewInstall(c.actionConfig)
		// copy write from: helm.sh/helm/v3/cmd/helm/upgrade.go
		installer.ReleaseName = release
		installer.CreateNamespace = param.createNamespace
		installer.ChartPathOptions = upgrader.ChartPathOptions
		installer.Force = upgrader.Force
		installer.DryRun = upgrader.DryRun
		installer.DisableHooks = upgrader.DisableHooks
		installer.SkipCRDs = upgrader.SkipCRDs
		installer.Timeout = upgrader.Timeout
		installer.Wait = upgrader.Wait
		installer.WaitForJobs = upgrader.WaitForJobs
		installer.Devel = upgrader.Devel
		installer.Namespace = upgrader.Namespace
		installer.Atomic = upgrader.Atomic
		installer.PostRenderer = upgrader.PostRenderer
		installer.DisableOpenAPIValidation = upgrader.DisableOpenAPIValidation
		installer.SubNotes = upgrader.SubNotes
		installer.Description = upgrader.Description
		installer.DependencyUpdate = upgrader.DependencyUpdate
		installer.EnableDNS = upgrader.EnableDNS
		installer.Atomic = upgrader.Atomic

		if chartRef.File == "" {
			log.Debugln("chart file not provided, need locate")
			installer.ChartPathOptions = chartRef.ChartPathOptions
			cp, err := installer.ChartPathOptions.LocateChart(chartRef.Name, c.settings)
			if err != nil {
				log.WithError(err).WithField("chart", chartRef.Name).Errorln("locate chart failed")
				return fmt.Errorf("locate chart %s file failed: %w", chartRef.Name, err)
			}
			chartRef.File = cp
		}

		chart, err := loader.Load(chartRef.File)
		if err != nil {
			log.WithError(err).WithField("file", chartRef.File).Errorln("load chart file failed")
			return fmt.Errorf("load chart file %s failed: %w", chartRef.File, err)
		}
		_, err = installer.Run(chart, param.values)
		if err != nil {
			log.WithError(err).Errorln("install release failed")
			return fmt.Errorf("install release %s failed: %w", release, err)
		}
		return nil // 走install逻辑
	}

	if chartRef.File == "" {
		log.Debugln("chart file not provided, need locate")
		upgrader.ChartPathOptions = chartRef.ChartPathOptions
		cp, err := upgrader.ChartPathOptions.LocateChart(chartRef.Name, c.settings)
		if err != nil {
			log.WithError(err).WithField("chart", chartRef.Name).Errorln("locate chart failed")
			return fmt.Errorf("locate chart %s file failed: %w", chartRef.Name, err)
		}
		chartRef.File = cp
	}

	chart, err := loader.Load(chartRef.File)
	if err != nil {
		log.WithError(err).WithField("file", chartRef.File).Errorln("load chart file failed")
		return fmt.Errorf("load chart file %s failed: %w", chartRef.File, err)
	}

	_, err = upgrader.Run(release, chart, param.values)
	if err != nil {
		return fmt.Errorf("upgrade release %s failed: %w", release, err)
	}
	return nil
}

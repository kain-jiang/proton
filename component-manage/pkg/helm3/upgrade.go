package helm3

import (
	"encoding/json"
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
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
	recreatePods    bool
}

type UpgradeOption func(upgradeParams *upgradeParams)

func WithUpgradeValues(valuesMap map[string]interface{}) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.values = valuesMap
	}
}

func WithUpgradeRecreatePods(recreatePods bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.recreatePods = recreatePods
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

func WithUpgradeAtoMic(atomic bool) UpgradeOption {
	return func(upgradeParams *upgradeParams) {
		upgradeParams.atomic = atomic
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

func (c *helmv3) Upgrade(release string, ch *chart.Chart, opts ...UpgradeOption) (*release.Release, error) {
	log := c.log.WithField("release", release).WithField("operate", "upgrade")
	param := &upgradeParams{}
	for _, opt := range opts {
		opt(param)
	}
	log.Debugf("upgrade with values: %+v", param.values)

	var historyVersion int = 0

	if param.atomic || param.install {
		// 获取当前旧版本
		historier := action.NewHistory(c.actionConfig)
		historier.Max = 1
		historyReleases, err := historier.Run(release)
		if err != nil || len(historyReleases) == 0 {
			log.WithError(err).Debugln("get release history failed")
			param.atomic = false
		} else {
			historyVersion = historyReleases[0].Version
			param.install = false
		}
	}

	upgrader := action.NewUpgrade(c.actionConfig)
	upgrader.Namespace = c.namespace
	upgrader.Force = param.force
	upgrader.Wait = param.wait
	upgrader.Timeout = param.timeout
	upgrader.DryRun = param.dryRun
	upgrader.SkipCRDs = param.skipCRDs
	upgrader.Recreate = param.recreatePods

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

		rls, err := installer.Run(ch, param.values)
		if err != nil {
			log.WithError(err).Errorln("install release failed")
			return nil, fmt.Errorf("install release %s failed: %w", release, err)
		}
		return rls, nil // 走install逻辑
	}

	rls, err := upgrader.Run(release, ch, param.values)
	if err != nil {
		log.WithError(err).Errorln("upgrade release failed")
		if param.atomic {
			log.Debugf("atomic enable, need rollback to %d", historyVersion)
			rollbacker := action.NewRollback(c.actionConfig)
			rollbacker.Force = param.force
			rollbacker.Wait = param.wait
			rollbacker.DryRun = param.dryRun
			rollbacker.Version = historyVersion
			rErr := rollbacker.Run(release)
			if rErr != nil {
				log.WithError(rErr).Errorln("rollback release failed")
				return nil, fmt.Errorf("rollback release %s failed: %w", release, rErr)
			}
		}
		return nil, fmt.Errorf("upgrade release %s failed: %w", release, err)
	}
	return rls, nil
}

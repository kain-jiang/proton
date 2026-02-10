package helm3

import (
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type installParams struct {
	values   map[string]interface{}
	force    bool
	wait     bool
	timeout  time.Duration
	dryRun   bool
	skipCRDs bool
}

type InstallOption func(installParam *installParams)

func WithInstallValues(valuesMap map[string]interface{}) InstallOption {
	return func(installParam *installParams) {
		installParam.values = valuesMap
	}
}

func WithInstallWait(wait bool, timeout time.Duration) InstallOption {
	return func(installParam *installParams) {
		installParam.wait = wait
		installParam.timeout = timeout
	}
}

func WithInstallForce(force bool) InstallOption {
	return func(installParam *installParams) {
		installParam.force = force
	}
}

func WithInstallDryRun(dryRun bool) InstallOption {
	return func(installParam *installParams) {
		installParam.dryRun = dryRun
	}
}

func WithInstallSkipCRDs(skipCRDs bool) InstallOption {
	return func(installParam *installParams) {
		installParam.skipCRDs = skipCRDs
	}
}

func (c *helmv3) Install(release string, ch *chart.Chart, opts ...InstallOption) (*release.Release, error) {
	log := c.log.WithField("release", release).WithField("operate", "install")

	param := &installParams{}
	for _, opt := range opts {
		opt(param)
	}

	installer := action.NewInstall(c.actionConfig)
	installer.ReleaseName = release
	installer.Namespace = c.namespace
	installer.Wait = param.wait
	installer.Timeout = param.timeout
	installer.DryRun = param.dryRun
	installer.SkipCRDs = param.skipCRDs

	rls, err := installer.Run(ch, param.values)
	if err != nil {
		log.WithError(err).Errorln("install release failed")
		return nil, fmt.Errorf("install release %s failed: %w", release, err)
	}
	return rls, nil
}

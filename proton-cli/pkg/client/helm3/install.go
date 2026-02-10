package helm3

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type installParams struct {
	values   map[string]interface{}
	force    bool
	wait     bool
	dryRun   bool
	skipCRDs bool
}

type InstallOption func(installParam *installParams)

func WithInstallValues(valuesMap map[string]interface{}) InstallOption {
	return func(installParam *installParams) {
		installParam.values = valuesMap
	}
}

func WithInstallWait(wait bool) InstallOption {
	return func(installParam *installParams) {
		installParam.wait = wait
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

func WithInstallForce(force bool) InstallOption {
	return func(installParam *installParams) {
		installParam.force = force
	}
}

func (c *helmv3) Install(release string, chartRef *ChartRef, opts ...InstallOption) error {
	log := c.log.WithField("release", release).WithField("operation", "install")

	param := &installParams{}
	for _, opt := range opts {
		opt(param)
	}

	installer := action.NewInstall(c.actionConfig)
	installer.ReleaseName = release
	installer.Namespace = c.namespace
	installer.Wait = param.wait
	installer.DryRun = param.dryRun
	installer.SkipCRDs = param.skipCRDs
	installer.Force = param.force

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
	return nil
}

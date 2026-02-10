package helm3

import (
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type uninstallParams struct {
	dryRun         bool
	keepHistory    bool
	ignoreNotFound bool
}

type UninstallOption func(uninstallParams *uninstallParams)

func WithUninstallDryRun(dryRun bool) UninstallOption {
	return func(uninstallParams *uninstallParams) {
		uninstallParams.dryRun = dryRun
	}
}

func WithUninstallKeepHistory(keepHistory bool) UninstallOption {
	return func(uninstallParams *uninstallParams) {
		uninstallParams.keepHistory = keepHistory
	}
}

func WithUninstallIgnoreNotFound(ignoreNotFound bool) UninstallOption {
	return func(uninstallParams *uninstallParams) {
		uninstallParams.ignoreNotFound = ignoreNotFound
	}
}

func (c *helmv3) Uninstall(release string, opts ...UninstallOption) error {
	log := c.log.WithField("release", release).WithField("operation", "uninstall")

	param := &uninstallParams{}
	for _, opt := range opts {
		opt(param)
	}

	uninstaller := action.NewUninstall(c.actionConfig)
	uninstaller.DryRun = param.dryRun
	uninstaller.KeepHistory = param.keepHistory

	_, err := uninstaller.Run(release)
	if errors.Is(err, driver.ErrReleaseNotFound) && param.ignoreNotFound {
		err = nil
	}
	if err != nil {
		log.WithError(err).Errorln("uninstall release failed")
		return fmt.Errorf("uninstall release %s failed: %w", release, err)
	}
	return nil
}

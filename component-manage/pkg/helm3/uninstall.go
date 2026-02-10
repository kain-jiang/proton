package helm3

import (
	"errors"
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type uninstallParams struct {
	dryRun         bool
	keepHistory    bool
	ignoreNotFound bool
	wait           bool
	timeout        time.Duration
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

func WithUninstallWait(wait bool, timeout time.Duration) UninstallOption {
	return func(uninstallParams *uninstallParams) {
		uninstallParams.wait = wait
		uninstallParams.timeout = timeout
	}
}

func (c *helmv3) Uninstall(release string, opts ...UninstallOption) (*release.Release, error) {
	log := c.log.WithField("release", release).WithField("operate", "uninstall")

	param := &uninstallParams{}
	for _, opt := range opts {
		opt(param)
	}

	uninstaller := action.NewUninstall(c.actionConfig)
	uninstaller.DryRun = param.dryRun
	uninstaller.KeepHistory = param.keepHistory
	uninstaller.Wait = param.wait
	uninstaller.Timeout = param.timeout

	rls, err := uninstaller.Run(release)
	if errors.Is(err, driver.ErrReleaseNotFound) && param.ignoreNotFound {
		err = nil
	}
	if err != nil {
		log.WithError(err).Errorln("uninstall release failed")
		return nil, fmt.Errorf("uninstall release %s failed: %w", release, err)
	}
	return rls.Release, nil
}

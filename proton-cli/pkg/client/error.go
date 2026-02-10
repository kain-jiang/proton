package client

import (
	"errors"
	"regexp"
	"strings"

	utilsexec "k8s.io/utils/exec"
)

// Deprecated: use ErrHelmReleaseNotFound instead
var ErrNotFound = ErrHelmReleaseNotFound

const (
	//
	ChartRepositoryNotFoundStderrFmt = "Error: no repo named \".*\" found\n"
	// HelmRepositoryUnauthorizedStderrSuffix 是未通过 repository 鉴权时的 stderr 的 suffix
	HelmRepositoryUnauthorizedStderrSuffix = "401 Unauthorized"
	// HelmNotFindTillerStderrSuffix 是未找到 tiller pod 时的 stderr 的 suffix
	HelmNotFindTillerStderrSuffix = "could not find tiller"
	// HelmNotFindReadyTillerPodStderrSuffix 是未找到 ready tiller pod 时的 stderr 的 suffix
	HelmNotFindReadyTillerPodStderrSuffix = "could not find a ready tiller pod"
)

var (
	ChartRepositoryNotFoundStderrRegexp = regexp.MustCompile("^" + ChartRepositoryNotFoundStderrFmt + "$")
)

var (
	// ErrChartRepositoryNotFound 表示未在 helm 本地配置中找到指定的 chart repository
	ErrChartRepositoryNotFound = errors.New("chart repository not found")
	// ErrHelmRepositoryUnauthorized 表示调用 helm 命令访问 repository 时返回 `401 Unauthorized`
	ErrHelmRepositoryUnauthorized = errors.New("unauthorized")
	// ErrHelmReleaseNotFound 表示调用 helm 命令时未找到指定的 release
	ErrHelmReleaseNotFound = errors.New("not found")
	// ErrHelmNotFindTiller 表示调用 helm 命令时未找到 tiller
	ErrHelmNotFindTiller = errors.New("could not find tiller")
	// ErrHelmNotFindReadyTillerPod 表示调用 helm 命令时未找到 ready 的 tiller pod
	ErrHelmNotFindReadyTillerPod = errors.New("could not find a ready tiller pod")
)

func handleHelmStderr(err error) error {
	ee := new(utilsexec.ExitErrorWrapper)
	if !errors.As(err, &ee) {
		return err
	}

	stderr := strings.TrimSpace(string(ee.Stderr))
	switch {
	case ChartRepositoryNotFoundStderrRegexp.Match(ee.Stderr):
		return ErrChartRepositoryNotFound
	case strings.HasSuffix(stderr, HelmRepositoryUnauthorizedStderrSuffix):
		return ErrHelmRepositoryUnauthorized
	case strings.HasSuffix(stderr, HelmNotFindTillerStderrSuffix):
		return ErrHelmNotFindTiller
	case strings.HasSuffix(stderr, HelmNotFindReadyTillerPodStderrSuffix):
		return ErrHelmNotFindReadyTillerPod
	case strings.HasPrefix(stderr, "Error: release:") && strings.HasSuffix(stderr, "not found"):
		return ErrHelmReleaseNotFound
	default:
		return err
	}
}

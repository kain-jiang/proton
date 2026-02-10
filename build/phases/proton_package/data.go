package proton_package

import (
	"k8s.io/utils/exec"
)

type ProtonPackageData interface {
	Executor() exec.Interface

	Version() string
	Architecture() string // amd64, arm64
	DistroArch() string   // x86_64, aarch64 ...

	ReleaseName() string

	RPMRepositoryURL() string
	RPMRepositoryVersion() string

	ToolRepositoryURL() string
	Harbor() string
	ChartSourceDir() string

	ProtonCLIPath() string

	WorkspaceDir() string
	ReleaseDir() string
}

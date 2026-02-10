package main

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/utils/exec"

	phases "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/phases/proton_package"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/workflow"
)

const (
	defaultOutput                string = "_output"
	defaultRPMRepoArchiveURL     string = "http://repo.proton.aishu.cn/yum"
	defaultRPMRepoArchiveVersion string = "3.8.0_beta.20260127.1"
	defaultToolRepoURL           string = "https://ftp-ict.aishu.cn/proton/tools"
	defaultHarbor                string = "https://acr.aishu.cn"
	defaultChartSourceDir        string = "charts"
	defaultProtonCLIPath         string = "proton-cli"
)

type options struct {
	version string
	// azure devops predefined variable Build.BuildNumber
	// buildNumber string
	// CPU 架构
	architecture string
	// 输出目录的路径，构建中间、最终产物将被写入这个目录。
	output string
	// RPM 仓库，归档的 URL
	rpmRepoArchiveURL string
	// RPM 仓库版本
	rpmRepoArchiveVersion string
	// static binaries of tools url
	toolRepoURL string
	// Harbor 地址，例如 https://acr.aishu.cn
	harbor string
	// Helm chart 源码目录，目录中包含多个 chart
	chartSourceDir string
	// proton-cli path
	protonCLIPath string
}

func newDefaultOptions() *options {
	return &options{
		version:               "0.0.0",
		architecture:          runtime.GOARCH,
		output:                defaultOutput,
		rpmRepoArchiveURL:     defaultRPMRepoArchiveURL,
		rpmRepoArchiveVersion: defaultRPMRepoArchiveVersion,
		toolRepoURL:           defaultToolRepoURL,
		harbor:                defaultHarbor,
		chartSourceDir:        defaultChartSourceDir,
		protonCLIPath:         defaultProtonCLIPath,
	}
}

func (opts *options) AddFlags(s *pflag.FlagSet) {
	s.StringVar(&opts.version, "version", opts.version, "build version")
	s.StringVar(&opts.architecture, "architecture", opts.architecture, "Architecture target, one of amd64 or arm64")
	s.StringVar(&opts.output, "output", opts.output, "Path to the output directory where build intermediates and final artifacts are written")
	s.StringVar(&opts.rpmRepoArchiveURL, "rpm", opts.rpmRepoArchiveURL, "RPM repository archive url")
	s.StringVar(&opts.rpmRepoArchiveVersion, "rpm-version", opts.rpmRepoArchiveVersion, "RPM repository archive version")
	s.StringVar(&opts.toolRepoURL, "tool", opts.toolRepoURL, "Static binary tool repository url")
	s.StringVar(&opts.harbor, "harbor", opts.harbor, "Harbor url")
	s.StringVar(&opts.chartSourceDir, "chart", opts.chartSourceDir, "Helm chart source directory")
	s.StringVar(&opts.protonCLIPath, "proton-cli-path", opts.protonCLIPath, "Proton CLI Path")
}

type protonPackageData struct {
	executor     exec.Interface
	output       string
	version      string
	architecture string
	distroArch   string

	releaseName string

	rpmRepositoryURL     string
	rpmRepositoryVersion string

	toolRepositoryURL string
	harbor            string
	chartSourceDir    string

	protonCLIPath string
}

var _ phases.ProtonPackageData = &protonPackageData{}

func newData(opts *options) (*protonPackageData, error) {
	distroArch, err := distroArchFromGOArch(opts.architecture)
	if err != nil {
		return nil, err
	}

	d := &protonPackageData{
		executor:             exec.New(),
		output:               opts.output,
		version:              opts.version,
		architecture:         opts.architecture,
		distroArch:           distroArch,
		releaseName:          fmt.Sprintf("ProtonDeps-%s.%s", opts.version, distroArch),
		rpmRepositoryURL:     opts.rpmRepoArchiveURL,
		rpmRepositoryVersion: opts.rpmRepoArchiveVersion,
		toolRepositoryURL:    opts.toolRepoURL,
		harbor:               opts.harbor,
		chartSourceDir:       opts.chartSourceDir,
		protonCLIPath:        opts.protonCLIPath,
	}

	return d, nil
}

// Executor implements phases.Data.
func (d *protonPackageData) Executor() exec.Interface { return d.executor }

// Version implements proton_package.ProtonPackageData.
func (d *protonPackageData) Version() string { return d.version }

// Architecture implements phases.Data.
func (d *protonPackageData) Architecture() string { return d.architecture }

// DistroArch implements phases.Data.
func (d *protonPackageData) DistroArch() string { return d.distroArch }

// ReleaseDir implements phases.Data.
func (d *protonPackageData) ReleaseDir() string {
	return filepath.Join(d.output, "releases", "proton-package", d.version)
}

// WorkspaceDir implements phases.Data.
func (d *protonPackageData) WorkspaceDir() string {
	return filepath.Join(d.output, "workspaces", "proton-package", d.version)
}

// ReleaseName implements phases.Data.
func (d *protonPackageData) ReleaseName() string { return d.releaseName }

// RPMRepositoryURL implements phases.Data.
func (d *protonPackageData) RPMRepositoryURL() string { return d.rpmRepositoryURL }

// RPMRepositoryVersion implements phases.Data.
func (d *protonPackageData) RPMRepositoryVersion() string { return d.rpmRepositoryVersion }

// ToolRepositoryURL implements phases.Data.
func (d *protonPackageData) ToolRepositoryURL() string { return d.toolRepositoryURL }

// Harbor implements phases.Data.
func (d *protonPackageData) Harbor() string { return d.harbor }

// ChartSourceDir implements proton_package.ProtonPackageData.
func (d *protonPackageData) ChartSourceDir() string { return d.chartSourceDir }

// ProtonCLIPath implements proton_package.ProtonPackageData.
func (d *protonPackageData) ProtonCLIPath() string { return d.protonCLIPath }

func distroArchFromGOArch(in string) (out string, err error) {
	m := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}

	out, ok := m[in]
	if !ok {
		err = fmt.Errorf("unsupported go arch %q", in)
	}
	return
}

func newCommandProtonPackage() *cobra.Command {
	opts := newDefaultOptions()

	runner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "proton-package",
		Short: "ProtonPackage (ProtonDeps)",
		Aliases: []string{
			"ProtonPackage",
			"ProtonDeps",
		},
		GroupID: groupTargets.ID,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := runner.InitData(args); err != nil {
				return err
			}

			return runner.Run(args)
		},
		SilenceUsage: true,
	}

	// Add flags to command
	opts.AddFlags(cmd.Flags())

	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newData(opts)
	})

	// Add phases to runner
	runner.AppendPhase(phases.NewPhaseWorkspace())
	runner.AppendPhase(phases.NewPhaseRPMs())
	runner.AppendPhase(phases.NewPhaseImages())
	runner.AppendPhase(phases.NewPhaseCharts())
	runner.AppendPhase(phases.NewPhaseTools())
	runner.AppendPhase(phases.NewPhaseTarball())

	runner.BindToCommand(cmd)

	return cmd
}

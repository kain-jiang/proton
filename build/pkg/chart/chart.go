package chart

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"k8s.io/utils/exec"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/utils"
)

type Reference struct {
	Project string `json:"project,omitzero"`
	Name    string `json:"name,omitzero"`
	Version string `json:"version,omitzero"`
}

func (r Reference) String() string {
	repoAndName := path.Join(r.Project, r.Name)
	if r.Version == "" {
		return repoAndName
	}
	return repoAndName + ":" + r.Version
}

func (r *Reference) AppendText(b []byte) ([]byte, error) { return []byte(r.String()), nil }
func (r *Reference) MarshalText() ([]byte, error)        { return r.AppendText(nil) }
func (r *Reference) UnmarshalText(data []byte) error {
	// project
	if i := bytes.IndexRune(data, '/'); i != -1 {
		r.Project = string(data[:i])
		data = data[i+1:]
	}
	// version
	if i := bytes.IndexRune(data, ':'); i != -1 {
		r.Version = string(data[i+1:])
		data = data[:i]
	}
	// name
	r.Name = string(data)

	return nil
}

//go:embed proton-package.yaml
var protonPackageBytes []byte

var protonPackageReferences []Reference = utils.Must(utils.UnmarshalYAML[[]Reference](protonPackageBytes))

func GenerateChartReferences() []Reference { return protonPackageReferences }

func generateProtonPackageChartsDirectoryPath(workspace string) string {
	return filepath.Join(workspace, "proton-packages", "service-package", "charts")
}

func generateProtonPackageChartPath(workspace string, ref *Reference) string {
	return filepath.Join(generateProtonPackageChartsDirectoryPath(workspace), ref.Name+"-"+ref.Version+".tgz")
}

func CreateProtonPackageChartsDirectoryInWorkspace(workspace string) (string, error) {
	charts := generateProtonPackageChartsDirectoryPath(workspace)
	slog.Info("Create helm charts directory", "path", charts)
	if err := os.MkdirAll(charts, 0755); err != nil {
		return "", err
	}
	return charts, nil
}

func PullFromHarbor(harbor, workspace string, ref *Reference) error {
	path := generateProtonPackageChartPath(workspace, ref)
	url := generateChartURLForHarbor(harbor, ref)
	slog.Info("Pull helm chart from harbor", "chart", ref, "url", url)
	if err := utils.Download(url, path); err != nil {
		return err
	}
	return nil
}

// Package all helm charts under specified directory
func PackageDirectoryAll(executor exec.Interface, version, destination, directory string) error {
	var charts []string

	if err := filepath.Walk(directory, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "Chart.yaml" {
			charts = append(charts, filepath.Dir(p))
			return filepath.SkipDir
		}

		return nil
	}); err != nil {
		return err
	}

	if charts == nil {
		slog.Warn("No charts found", "directory", directory)
		return nil
	}

	return Package(executor, version, destination, charts...)
}

// Package packages multi helm charts with specific version and output destination
func Package(executor exec.Interface, version, destination string, charts ...string) error {
	var args []string
	args = append(args, "package")
	args = append(args, "--destination", destination)
	args = append(args, "--version", version)
	args = append(args, charts...)

	slog.Info("Package helm charts", "version", version, "destination", destination, "charts", charts)
	if _, err := executor.Command("helm", args...).Output(); err != nil {
		if ee := new(exec.ExitErrorWrapper); errors.As(err, &ee) {
			slog.Error("package helm chart fail", "return", ee.ExitCode(), "stderr", string(ee.Stderr))
		}
		return err
	}

	return nil
}

func generateChartURLForHarbor(harbor string, ref *Reference) string {
	return fmt.Sprintf("%s/chartrepo/%s/charts/%s-%s.tgz", harbor, ref.Project, ref.Name, ref.Version)
}

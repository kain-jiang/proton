package image

import (
	"bytes"
	_ "embed"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"k8s.io/utils/exec"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/build/pkg/utils"
)

type Reference struct {
	Repository string `json:"repository,omitzero"`
	Tag        string `json:"tag,omitzero"`
}

func (r Reference) String() string {
	if r.Tag == "" || r.Tag == "latest" {
		return r.Repository
	}
	return r.Repository + ":" + r.Tag
}

func (r *Reference) AppendText(b []byte) ([]byte, error) { return []byte(r.String()), nil }
func (r *Reference) MarshalText() ([]byte, error)        { return r.AppendText(nil) }
func (r *Reference) UnmarshalText(data []byte) error {
	index := bytes.IndexRune(data, ':')
	if index == -1 {
		r.Repository = string(data)
		return nil
	}

	r.Repository = string(data[:index])
	r.Tag = string(data[index+1:])
	return nil
}

//go:embed proton-package.yaml
var protonPackageBytes []byte

var protonPackageReferences []Reference = utils.Must(utils.UnmarshalYAML[[]Reference](protonPackageBytes))

func GenerateImageReferences() []Reference { return protonPackageReferences }

func generateProtonPackageImagesDirectoryPath(workspace string) string {
	return filepath.Join(workspace, "proton-packages", "service-package", "images")
}

func CreateProtonPackageImagesDirectoryInWorkspace(workspace string) error {
	images := generateProtonPackageImagesDirectoryPath(workspace)
	slog.Info("Create container images directory", "path", images)
	return os.MkdirAll(images, 0755)
}

func PullFromHarbor(executor exec.Interface, harbor, workspace string, ref *Reference, arch string) error {
	images := generateProtonPackageImagesDirectoryPath(workspace)
	slog.Info("Pull image via skopeo", "image", ref)
	return skopeoCopyFromDockerToOCI(executor, hostFromURL(harbor), images, ref, arch)
}

func hostFromURL(in string) string {
	u, err := url.Parse(in)
	if err != nil {
		return ""
	}
	return u.Host
}

package offline_package

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manifest 定义离线包构建所需的完整目标清单。
type Manifest struct {
	meta.TypeMeta
	meta.ObjectMeta `json:"metadata,omitzero"`

	Spec ManifestSpec `json:"spec,omitzero"`
}

// ManifestSpec 描述离线包内应收集的架构信息与各类制品列表。
type ManifestSpec struct {
	Architecture string `json:"architecture,omitzero"`

	Components Component `json:"components,omitzero"`

	Binaries []Artifact `json:"binaries,omitzero"`

	Charts []Artifact `json:"charts,omitzero"`

	Images []Artifact `json:"images,omitzero"`

	RPMs []Artifact `json:"rpms,omitzero"`
}

// Component 描述离线包关联的组件标识与版本信息。
type Component struct {
	Name string `json:"name,omitzero"`

	Version string `json:"version,omitzero"`
}

// Artifact 描述一个需要被拉取并写入离线包的制品。
type Artifact struct {
	Name string `json:"name,omitzero"`

	Source
}

// Source 表示制品来源，目前支持 HTTP 下载和 OCI 拉取两种方式。
type Source struct {
	HTTP *HTTPSource `json:"http,omitzero"`

	OCI *OCISource `json:"oci,omitzero"`

	File *FileSource `json:"file,omitzero"`
}

// HTTPSource 定义通过 HTTP 获取制品时所需的地址和解包信息。
type HTTPSource struct {
	URL string `json:"url,omitzero"`

	Format string `json:"format,omitzero"`

	Path string `json:"path,omitzero"`
}

// OCISource 定义通过 OCI 仓库拉取制品时使用的镜像引用。
type OCISource struct {
	Reference string `json:"reference,omitzero"`
}

// FileSource 定义通过本地文件获取制品
type FileSource struct {
	Path string `json:"path,omitzero"`
}

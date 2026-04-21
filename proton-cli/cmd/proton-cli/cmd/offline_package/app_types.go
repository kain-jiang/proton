package offline_package

import "time"

type appManifest struct {
	APIVersion     string                        `json:"apiVersion,omitempty"`
	Kind           string                        `json:"kind,omitempty"`
	Product        string                        `json:"product,omitempty"`
	Version        string                        `json:"version,omitempty"`
	Source         appManifestSource             `json:"source"`
	Dependencies   []appManifestDependency       `json:"dependencies,omitempty"`
	Releases       map[string]appManifestRelease `json:"releases,omitempty"`
	Platform       string                        `json:"platform,omitempty"`
	OfflinePackage *appPackageMetadata           `json:"offlinePackage,omitempty"`
}

type appManifestSource struct {
	HelmRepoName string `json:"helmRepoName,omitempty"`
	HelmRepoURL  string `json:"helmRepoUrl,omitempty"`
}

type appManifestDependency struct {
	Product        string `json:"product,omitempty"`
	Version        string `json:"version,omitempty"`
	Manifest       string `json:"manifest,omitempty"`
	Optional       bool   `json:"optional,omitempty"`
	DefaultEnabled bool   `json:"defaultEnabled,omitempty"`
}

type appManifestRelease struct {
	Chart   string `json:"chart,omitempty"`
	Version string `json:"version,omitempty"`
}

type appPackageMetadata struct {
	Platform         string            `json:"platform,omitempty"`
	Platforms        []string          `json:"platforms,omitempty"`
	OverrideRegistry string            `json:"overrideRegistry,omitempty"`
	ExportedAt       time.Time         `json:"exportedAt"`
	Images           []appPackageImage `json:"images,omitempty"`
	ImageErrors      []string          `json:"imageErrors,omitempty"`
}

type appPackageImage struct {
	Source             string   `json:"source,omitempty"`
	PullSource         string   `json:"pullSource,omitempty"`
	Repository         string   `json:"repository,omitempty"`
	Tag                string   `json:"tag,omitempty"`
	LocalRef           string   `json:"localRef,omitempty"`
	RequestedPlatforms []string `json:"requestedPlatforms,omitempty"`
	ExportedPlatforms  []string `json:"exportedPlatforms,omitempty"`
	Exported           bool     `json:"exported,omitempty"`
	Error              string   `json:"error,omitempty"`
}

type appChartArtifact struct {
	Name    string
	Version string
	RepoURL string
	URL     string
	Path    string
}

type appManifestDocument struct {
	Location string
	Bytes    []byte
	Manifest *appManifest
}

type appImageRef struct {
	Registry   string
	Source     string
	Repository string
	Tag        string
}

func (r appImageRef) LocalRef() string {
	return r.Repository + ":" + r.Tag
}

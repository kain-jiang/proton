package app

// VersionSet 是 deploy.kweaver.ai/v1alpha1 VersionSet manifest 的 Go 结构体，
// 对应 release-manifests/*.yaml 文件格式。
type VersionSet struct {
	APIVersion   string                  `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind         string                  `json:"kind,omitempty" yaml:"kind,omitempty"`
	Product      string                  `json:"product,omitempty" yaml:"product,omitempty"`
	Version      string                  `json:"version,omitempty" yaml:"version,omitempty"`
	Source       VersionSetSource        `json:"source" yaml:"source,omitempty"`
	Dependencies []VersionSetDependency  `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Releases     map[string]ReleaseEntry `json:"releases,omitempty" yaml:"releases,omitempty"`
}

// VersionSetSource 描述 Helm chart 来源
type VersionSetSource struct {
	HelmRepoName string `json:"helmRepoName,omitempty" yaml:"helmRepoName,omitempty"`
	HelmRepoURL  string `json:"helmRepoUrl,omitempty" yaml:"helmRepoUrl,omitempty"`
}

// VersionSetDependency 描述一个依赖的产品
type VersionSetDependency struct {
	Product string `json:"product,omitempty" yaml:"product,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Manifest 是相对于当前 YAML 文件的路径，或绝对路径
	Manifest       string `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Optional       bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
	DefaultEnabled bool   `json:"defaultEnabled,omitempty" yaml:"defaultEnabled,omitempty"`
	EnabledIf      string `json:"enabledIf,omitempty" yaml:"enabledIf,omitempty"`
}

// ReleaseEntry 描述单个 Helm release 的安装参数
type ReleaseEntry struct {
	Chart   string `json:"chart,omitempty" yaml:"chart,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Stage 控制安装顺序。"pre" 表示需先于同组其他 release 安装完成。
	// 空字符串表示普通 release，与其他普通 release 顺序安装。
	Stage string `json:"stage,omitempty" yaml:"stage,omitempty"`
	// DependsOn 仅用于 app install 的显式顺序约束，表示当前 release 依赖同一 VersionSet 中的其他 release。
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	// Values 是该 release 特定的额外 Helm values，会与基础 values 合并。
	// 用于应用特定配置（如 dip-studio 的 studio.openclaw），避免与基础服务配置重复。
	Values map[string]any `json:"values,omitempty" yaml:"values,omitempty"`
}

// InstallPlan 是经过依赖展开和 stage 排序后的安装计划，每个 Step 为可并发执行的一批 release。
type InstallPlan struct {
	// Steps 中每个元素是一个可以并发执行的 release 集合。
	// Steps[0] 先执行完毕后，再执行 Steps[1]，以此类推。
	Steps [][]InstallItem
}

// InstallItem 是安装计划中的一个 release 条目
type InstallItem struct {
	// ReleaseName 是 helm release 名称（也是 VersionSet.Releases 的 key）
	ReleaseName string
	// ChartName 是 Helm chart 名称
	ChartName string
	// ChartVersion 是 Helm chart 版本
	ChartVersion string
	// HelmRepoName 是 Helm repo 名称，用于 helm pull/upgrade
	HelmRepoName string
	// HelmRepoURL 是 Helm repo 地址
	HelmRepoURL string
	// Product 来自哪个 VersionSet 产品
	Product string
	// Stage 来自 manifest 的 release.stage，用于控制等待策略。
	Stage string
	// DependsOn 来自 manifest 的 release.dependsOn，用于安装计划拓扑排序。
	DependsOn []string
	// WaitForReady 表示当前 release 必须等待 ready，通常因为它是 stage=pre 或被其他 release 依赖。
	WaitForReady bool
	// Values 是该 release 特定的额外 Helm values（从 manifest 中的 ReleaseEntry.Values 传递）
	Values map[string]any
}

// InstallOptions 控制 app install 行为
type InstallOptions struct {
	// Namespace 是安装目标的 K8s namespace
	Namespace string
	// Timeout 是每个 helm release 的安装超时
	Timeout string
	// DryRun 只打印计划，不执行安装
	DryRun bool
	// CreateNamespace 是否自动创建 namespace
	CreateNamespace bool
	// HelmRepoName 覆盖 manifest 中的 helmRepoName（用于离线部署时指向内置仓库）
	HelmRepoName string
	// HelmRepoURL 覆盖 manifest 中的 helmRepoURL（用于离线部署时指向内置仓库）
	HelmRepoURL string
	// SetValues 保存来自 CLI --set 的 install-time values overrides。
	SetValues map[string]any
}

// UninstallOptions 控制 app uninstall 行为
type UninstallOptions struct {
	// Namespace 是卸载目标的 K8s namespace
	Namespace string
	// DryRun 只打印卸载计划，不执行卸载
	DryRun bool
	// Timeout 是每个 helm release 的卸载超时
	Timeout string
}

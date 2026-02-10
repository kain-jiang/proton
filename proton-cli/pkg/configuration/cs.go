package configuration

// AddonName 定义 Proton CS 的插件的名称
type CSAddonName string

const (
	CSAddonNameNodeExporter CSAddonName = "node-exporter"
	CSAddonNameStateMetrics CSAddonName = "kube-state-metrics"
)

// AllCSAddons 是 Proton CS 所有的插件的名称列表
var AllCSAddons = []CSAddonName{
	CSAddonNameNodeExporter,
	CSAddonNameStateMetrics,
}

// DefaultCSAddons 是 Proton CS 默认启用的插件的名称列表
var DefaultCSAddons = AllCSAddons

package plugins

type PluginObject struct {
	Name    string         `json:"name" yaml:"name"`
	Type    string         `json:"type" yaml:"type"`
	Version string         `json:"version" yaml:"version"`
	Config  map[string]any `json:"config" yaml:"config"`
}

type chartPluginConfig struct {
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
}

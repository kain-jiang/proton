package plugins

type NebulaPluginConfig struct {
	ChartName    string `json:"chart_name" yaml:"chart_name" binding:"required"`
	ChartVersion string `json:"chart_version" yaml:"chart_version" binding:"required"`
	Namespace    string `json:"namespace" yaml:"namespace" binding:"required"`
	Images       struct {
		GraphD   string `json:"graphd" yaml:"graphd" binding:"required"`
		MetaD    string `json:"metad" yaml:"metad" binding:"required"`
		StorageD string `json:"storaged" yaml:"storaged" binding:"required"`
		Exporter string `json:"exporter" yaml:"exporter" binding:"required"`
	} `json:"images" yaml:"images" binding:"required"`
}

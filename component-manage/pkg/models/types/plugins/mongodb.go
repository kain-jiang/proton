package plugins

type MongoDBPluginConfig struct {
	ChartName    string `json:"chart_name" binding:"required"`
	ChartVersion string `json:"chart_version" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
	Images       struct {
		MongoDB   string `json:"mongodb" binding:"required"`
		Logrotate string `json:"logrotate" binding:"required"`
		Exporter  string `json:"exporter" binding:"required"`
		Mgmt      string `json:"mgmt" binding:"required"`
	} `json:"images" binding:"required"`
}

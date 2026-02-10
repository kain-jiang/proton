package plugins

type MariaDBPluginConfig struct {
	ChartName    string `json:"chart_name" binding:"required"`
	ChartVersion string `json:"chart_version" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
	Images       struct {
		MariaDB  string `json:"mariadb" binding:"required"`
		ETCD     string `json:"etcd" binding:"required"`
		Exporter string `json:"exporter" binding:"required"`
		Mgmt     string `json:"mgmt" binding:"required"`
	} `json:"images" binding:"required"`
}

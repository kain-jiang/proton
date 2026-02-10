package grafana

const (
	// 默认数据目录的绝对路径
	DefaultDataPath = "/sysvol/grafana"

	// Helm chart's name
	ChartName = "proton-grafana"

	// Helm release's name
	HelmReleaseName = "grafana"

	// Grafana's replicas
	Replicas = 1

	// Grafana is exposed on this node port
	NodePort = 30002
)

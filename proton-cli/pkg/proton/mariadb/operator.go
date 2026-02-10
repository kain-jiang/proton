package mariadb

const (
	OperatorHelmChartName   = "rds-mariadb-operator"
	OperatorHelmReleaseName = "rds-mariadb-operator"
	// Chart proton-rds-mariadb-operator 不支持安装在指定的 namespace，这里与 chart 内硬编码的 namespace 保持一致
	OperatorHelmReleaseNamespace = "proton-rds-mariadb-operator-system"
)

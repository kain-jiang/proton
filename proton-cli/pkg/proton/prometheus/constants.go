package prometheus

const (
	// 默认数据目录的绝对路径
	DefaultDataPath = "/sysvol/prometheus"
	// 最多的节点数量
	MaxNodeNumber = 2

	// helm chart's name
	ChartName = "proton-prometheus"

	// helm release's name
	HelmReleaseName = "prometheus"

	// replicas of prometheus on the hosted kubernetes
	ReplicasForHostedKubernetes = 2
)

// 以下是给Proton ETCD和K8S ETCD生成证书过程中用到的名称常量
const (
	PrometheusETCDCommonName = "proton-prometheus-observability"

	ProtonETCDCACertSecret = "etcdssl-secret"
	ProtonETCDCACertName   = "ca.crt"
	ProtonETCDCACertKey    = "etcdssl-secret-key"
	ProtonETCDCAKeyName    = "ca.key"
	K8SETCDCACertPath      = "/etc/kubernetes/pki/etcd/ca.crt"
	K8SETCDCACertKey       = "/etc/kubernetes/pki/etcd/ca.key"

	ProtonETCDResultSecretName = "etcdssl-secret-for-prometheus"
	ProtonETCDResultCAName     = "ca-protonetcd.crt"
	ProtonETCDResultCertName   = "prometheus-metrics-protonetcd.crt"
	ProtonETCDResultKeyName    = "prometheus-metrics-protonetcd.key"
	K8SETCDResultSecretName    = "k8s-etcdssl-secret-for-prometheus"
	K8SETCDResultCAName        = "ca-k8setcd.crt"
	K8SETCDResultCertName      = "prometheus-metrics-k8setcd.crt"
	K8SETCDResultKeyName       = "prometheus-metrics-k8setcd.key"
)

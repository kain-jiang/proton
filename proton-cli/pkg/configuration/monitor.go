package configuration

import (
	v1 "k8s.io/api/core/v1"
)

// ProtonMonitor defines the configuration for the proton-monitor component
type ProtonMonitor struct {
	// Config contains the configuration for proton-monitor components
	Config *MonitorConfig `json:"config,omitempty"`
	// 运行 proton-monitor 的节点的名称
	Hosts []string `json:"hosts,omitempty"`
	// 数据目录的路径，需要是绝对路径
	DataPath string `json:"data_path,omitempty"`
	// Resources contains the resource requirements for proton-monitor components
	Resources *MonitorResources `json:"resources,omitempty"`
	// Storage contains the storage configuration for proton-monitor
	Storage *MonitorStorage `json:"storage,omitempty"`
}

// MonitorConfig defines the configuration for proton-monitor components
type MonitorConfig struct {
	// DcgmExporter contains the configuration for the DCGM exporter
	DcgmExporter *DcgmExporterConfig `json:"dcgmExporter,omitempty"`
	// NodeExporter contains the configuration for the node exporter
	NodeExporter *NodeExporterConfig `json:"nodeExporter,omitempty"`
	// Fluentbit contains the configuration for fluentbit
	Fluentbit *FluentbitConfig `json:"fluentbit,omitempty"`
	// Vmagent contains the configuration for vmagent
	Vmagent *VmagentConfig `json:"vmagent,omitempty"`
	// Vmetrics contains the configuration for vmetrics
	Vmetrics *VmetricsConfig `json:"vmetrics,omitempty"`
	// Vlogs contains the configuration for vlogs
	Vlogs *VlogsConfig `json:"vlogs,omitempty"`
	// Grafana contains the configuration for Grafana
	Grafana *GrafanaConfig `json:"grafana,omitempty"`
}

// DcgmExporterConfig defines the configuration for the DCGM exporter
type DcgmExporterConfig struct {
	// Port specifies the port for the DCGM exporter
	Port int32 `json:"port,omitempty"`
}

// NodeExporterConfig defines the configuration for the node exporter
type NodeExporterConfig struct {
	// Port specifies the port for the node exporter
	Port int32 `json:"port,omitempty"`
}

// FluentbitConfig defines the configuration for fluentbit
type FluentbitConfig struct {
	// Port specifies the port for fluentbit
	Port int32 `json:"port,omitempty"`
	// Namespaces specifies the namespaces to collect logs from
	Namespaces []string `json:"namespaces,omitempty"`
	// RemoteLogServers contains the configuration for remote log servers
	RemoteLogServers []RemoteLogServerConfig `json:"remoteLogServers,omitempty"`
}

// RemoteLogServerConfig defines the configuration for the remote log server
type RemoteLogServerConfig struct {
	// Host specifies the host for the remote log server
	Host string `json:"host,omitempty"`
	// Port specifies the port for the remote log server
	Port int32 `json:"port,omitempty"`
	// URI specifies the URI for the remote log server
	URI string `json:"uri,omitempty"`
}

// VmagentConfig defines the configuration for vmagent
type VmagentConfig struct {
	// ScrapeInterval specifies the scrape interval for vmagent
	ScrapeInterval string `json:"scrape_interval,omitempty"`
	// ScrapeTimeout specifies the scrape timeout for vmagent
	ScrapeTimeout string `json:"scrape_timeout,omitempty"`
	// K8sEtcdCerts specifies the name of the secret containing etcd certificates
	K8sEtcdCerts string `json:"k8sEtcdCerts,omitempty"`
	// RemoteWrite contains the configuration for remote write
	RemoteWrite *RemoteWriteConfig `json:"remoteWrite,omitempty"`
	// Port specifies the port for vmagent
	Port int32 `json:"port,omitempty"`
}

// RemoteWriteConfig defines the configuration for remote write
type RemoteWriteConfig struct {
	// Host specifies the host for remote write
	Host string `json:"host,omitempty"`
	// Port specifies the port for remote write
	Port int32 `json:"port,omitempty"`
	// Path specifies the path for remote write
	Path string `json:"path,omitempty"`
	// ExtraServers specifies additional remote write servers
	ExtraServers []RemoteWriteServer `json:"extraServers,omitempty"`
}

// RemoteWriteServer defines a remote write server
type RemoteWriteServer struct {
	// URL specifies the URL for the remote write server
	URL string `json:"url,omitempty"`
}

// VmetricsConfig defines the configuration for vmetrics
type VmetricsConfig struct {
	// Retention specifies the retention period for vmetrics
	Retention string `json:"retention,omitempty"`
	// Port specifies the port for vmetrics
	Port int32 `json:"port,omitempty"`
}

// VlogsConfig defines the configuration for vlogs
type VlogsConfig struct {
	// Retention specifies the retention period for vlogs
	Retention string `json:"retention,omitempty"`
	// Port specifies the port for vlogs
	Port int32 `json:"port,omitempty"`
}

// GrafanaConfig defines the configuration for Grafana
type GrafanaConfig struct {
	// SMTP contains the SMTP configuration for Grafana
	SMTP *SMTPConfig `json:"smtp,omitempty"`
	// Port specifies the port for Grafana
	Port int32 `json:"port,omitempty"`
	// NodePort specifies the node port for Grafana
	NodePort int32 `json:"nodePort,omitempty"`
}

// SMTPConfig defines the SMTP configuration for Grafana
type SMTPConfig struct {
	// Enabled specifies whether SMTP is enabled
	Enabled bool `json:"enabled,omitempty"`
	// Host specifies the SMTP host
	Host string `json:"host,omitempty"`
	// User specifies the SMTP user
	User string `json:"user,omitempty"`
	// Password specifies the SMTP password
	Password string `json:"password,omitempty"`
	// SkipVerify specifies whether to skip SSL verification
	SkipVerify bool `json:"skip_verify,omitempty"`
	// From specifies the sender email address
	From string `json:"from,omitempty"`
	// FromName specifies the sender name
	FromName string `json:"from_name,omitempty"`
	// StartTLSPolicy specifies the StartTLS policy
	StartTLSPolicy string `json:"startTLS_policy,omitempty"`
	// EnableTracing specifies whether to enable tracing
	EnableTracing bool `json:"enable_tracing,omitempty"`
}

// MonitorStorage defines the storage configuration for proton-monitor
type MonitorStorage struct {
	// Capacity specifies the storage capacity
	Capacity string `json:"capacity,omitempty"`
	// StorageClassName specifies the storage class name
	StorageClassName string `json:"storageClassName,omitempty"`
	// Local contains the local storage configuration
	Local map[string]LocalStorage `json:"local,omitempty"`
}

// LocalStorage defines the local storage configuration
type LocalStorage struct {
	// Host specifies the host for local storage
	Host string `json:"host,omitempty"`
	// Path specifies the path for local storage
	Path string `json:"path,omitempty"`
}

// MonitorResources defines the resource requirements for proton-monitor components
type MonitorResources struct {
	// Fluentbit contains the resource requirements for fluentbit
	Fluentbit *v1.ResourceRequirements `json:"fluentbit,omitempty"`
	// DcgmExporter contains the resource requirements for the DCGM exporter
	DcgmExporter *v1.ResourceRequirements `json:"dcgmExporter,omitempty"`
	// NodeExporter contains the resource requirements for the node exporter
	NodeExporter *v1.ResourceRequirements `json:"nodeExporter,omitempty"`
	// Grafana contains the resource requirements for Grafana
	Grafana *v1.ResourceRequirements `json:"grafana,omitempty"`
	// Vmetrics contains the resource requirements for vmetrics
	Vmetrics *v1.ResourceRequirements `json:"vmetrics,omitempty"`
	// Vlogs contains the resource requirements for vlogs
	Vlogs *v1.ResourceRequirements `json:"vlogs,omitempty"`
	// Vmagent contains the resource requirements for vmagent
	Vmagent *v1.ResourceRequirements `json:"vmagent,omitempty"`
}

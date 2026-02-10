package configuration

import (
	"component-manage/pkg/models/types"

	v1 "k8s.io/api/core/v1"
)

// cluster conf
type ClusterConfig struct {
	ApiVersion           string                             `json:"apiVersion"`
	Deploy               map[string]any                     `json:"deploy,omitempty"`
	Nodes                []Node                             `json:"nodes"`
	Chrony               *Chrony                            `json:"chrony,omitempty"`
	Cs                   *Cs                                `json:"cs,omitempty"`
	Cr                   *Cr                                `json:"cr,omitempty"`
	Proton_mariadb       *types.MariaDBComponentParams      `json:"proton_mariadb,omitempty"`
	Proton_mongodb       *types.MongoDBComponentParams      `json:"proton_mongodb,omitempty"`
	Proton_redis         *types.RedisComponentParams        `json:"proton_redis,omitempty"`
	Proton_mq_nsq        *ProtonDataConf                    `json:"proton_mq_nsq,omitempty"`
	Proton_policy_engine *types.PolicyEngineComponentParams `json:"proton_policy_engine,omitempty"`
	Proton_etcd          *types.ETCDComponentParams         `json:"proton_etcd,omitempty"`
	OpenSearch           *types.OpensearchComponentParams   `json:"opensearch,omitempty"`
	OrientDB             *OrientDB                          `json:"orientdb,omitempty"`

	// AnyShare 的 configuration management service
	CMS *CMS `json:"cms,omitempty"`
	// AnyShare 的 installer service
	Installer_Service *InstallerService `json:"installer_service,omitempty"`
	ComponentManage   map[string]any    `json:"component_management,omitempty"`
	// nvidia-device-plugin for GPU computation on Kubernetes
	NvidiaDevicePlugin *NvidiaDevicePlugin `json:"nvidia_device_plugin,omitempty"`

	// AnyRobot 使用的 Kafka
	Kafka *types.KafkaComponentParams `json:"kafka,omitempty"`
	// AnyRobot 使用的 ZooKeeper
	ZooKeeper *types.ZookeeperComponentParams `json:"zookeeper,omitempty"`

	// Proton 坯观测性朝务使用的 Prometheus
	Prometheus *Prometheus `json:"prometheus,omitempty"`
	// Proton 坯观测性朝务使用的 Grafana
	Grafana *Grafana `json:"grafana,omitempty"`

	// Nebula Graph 的部署酝置，如果为 nil 则丝会安装 Nebula Graph
	Nebula *types.NebulaComponentParams `json:"nebula,omitempty"`

	//  基础组件连接信息都在此保存，替代cms的保存
	ResourceConnectInfo *ResourceConnectInfo `json:"resource_connect_info,omitempty" mapstructure:"resourceConnectInfo"`

	// Proton 包管睆朝务
	PackageStore *PackageStore `json:"package-store,omitempty"`

	// Proton ECeph
	ECeph *ECeph `json:"eceph,omitempty"`
}

type Node struct {
	Name        string `json:"name"`
	IP4         string `json:"ip4,omitempty"`
	IP6         string `json:"ip6,omitempty"`
	Internal_ip string `json:"internal_ip,omitempty"`
}

// IP 返回节点的 IPv4 或 IPv6 地址，优先返回 IPv4 地址。
func (n *Node) IP() string {
	if n.IP4 != "" {
		return n.IP4
	}
	return n.IP6
}

// GetNodeNameByIP 返回指定节点列表包坫指定 IP 地址的节点坝称，如果节点丝存在则返回空字符串
func GetNodeNameByIP(ip string, nodes []Node) string {
	for _, node := range nodes {
		if node.IP4 == ip || node.IP6 == ip {
			return node.Name
		}
	}
	return ""
}

// GetIPByNodeName 返回指定节点列表中指定节点坝称的节点 IP 地址，如果节点丝存在则返回空字符串
func GetIPByNodeName(name string, nodes []Node) string {
	for _, node := range nodes {
		if node.Name == name {
			return node.IP()
		}
	}
	return ""
}

type ByGetName []Node

func (a ByGetName) Len() int           { return len(a) }
func (a ByGetName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByGetName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type Cs struct {
	// Kubernetes 的杝供者
	Provisioner KubernetesProvisioner `json:"provisioner,omitempty"`
	// IPFamilies is a list of IP families (e.g. IPv4, IPv6) used by kubernetes.
	IPFamilies []v1.IPFamily `json:"ipFamilies,omitempty"`

	Master            []string     `json:"master"`
	Host_network      *HostNetWork `json:"host_network"`
	Ha_port           int          `json:"ha_port"`
	Etcd_data_dir     string       `json:"etcd_data_dir"`
	Docker_data_dir   string       `json:"docker_data_dir"`
	Cs_controller_dir string       `json:"cs_controller_dir"`
	// Proton CS 坯用的杒件列表, 因为需覝区分 `nil` 和 `[]` 所以丝能使用
	// omitempty
	Addons []CSAddonName `json:"addons"`
}

type Chrony struct {
	// 时间服务器相关的配置
	Mode   string   `json:"mode,omitempty"`
	Server []string `json:"server,omitempty"`
}

type KubernetesProvisioner string

// Valid kubernetes provisioner
const (
	// Proton CLI 创建的 kubernetes
	KubernetesProvisionerLocal KubernetesProvisioner = "local"
	// 外部的 kubernetes，Proton CLI 仅作使用
	KubernetesProvisionerExternal KubernetesProvisioner = "external"
)

type HostNetWork struct {
	Bip              string `json:"bip"`
	Pod_network_cidr string `json:"pod_network_cidr"`
	Service_cidr     string `json:"service_cidr"`
	Ipv4_interface   string `json:"ipv4_interface,omitempty"`
	Ipv6_interface   string `json:"ipv6_interface,omitempty"`
}

// Cr contains elements describing Cr configuration.
type Cr struct {
	// Local provides configuration knobs for configuring the local cr instance
	// Local and External are mutually exclusive
	Local *LocalCR `json:"local,omitempty"`

	// External describes how to connect to an external cr cluster
	// Local and External are mutually exclusive
	External *ExternalCR `json:"external,omitempty"`
}

// LocalCR describes that proton-cli should run an cr cluster locally
type LocalCR struct {
	Hosts    []string `json:"hosts"`
	Ports    Ports    `json:"ports"`
	Ha_ports Ports    `json:"ha_ports"`
	Storage  string   `json:"storage"`
}

// ExternalCR describes an external cr cluster.
type ExternalCR struct {
	// Registry holds configuration for registry.
	Registry Registry `json:"registry"`
	// Chartmuseum holds configuration for chartmuseum.
	Chartmuseum Chartmuseum `json:"chartmuseum"`

	// OCI holds configuration for oci, it can be used for image/chart repository
	OCI *OCI `json:"oci,omitempty"`

	ChartRepo string `json:"chart_repository"`
	ImageRepo string `json:"image_repository"`
}

type OCI struct {
	Registry  string `json:"registry,omitempty"`
	PlainHTTP bool   `json:"plain_http,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

// Registry describes an external container registry.
type Registry struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Chartmuseum describes an external chartmuseum.
type Chartmuseum struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Ports struct {
	Chartmuseum int `json:"chartmuseum"`
	Registry    int `json:"registry"`
	Rpm         int `json:"rpm"`
	Cr_manager  int `json:"cr_manager"`
}

// HelmComponent 包坫 helm 组件的通用酝置
type HelmComponent struct {
	// ExtraValues 是传递给 Proton OpenSearch 的额外的 Helm Values
	ExtraValues map[string]interface{} `json:"extraValues,omitempty"`
}

type ProtonMariaDB struct {
	ReplicaCount int                   `json:"replica_count,omitempty"`
	Hosts        []string              `json:"hosts"`
	Config       *ProtonMariaDBConfigs `json:"config"`
	Admin_user   string                `json:"admin_user"`
	Admin_passwd string                `json:"admin_passwd"`
	Data_path    string                `json:"data_path"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}

const (
	// redis 的 Chart 坝称
	ChartNameRedis = "proton-redis"
	// redis 的 Helm release 坝称
	ReleaseNameRedis = "proton-redis"
)

type ProtonDB struct {
	ReplicaCount int      `json:"replica_count,omitempty"`
	Hosts        []string `json:"hosts"`
	Admin_user   string   `json:"admin_user"`
	Admin_passwd string   `json:"admin_passwd"`
	Data_path    string   `json:"data_path"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`

	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

const (
	// policyEngine 的 Chart 坝称
	ChartNamePolicyEngine = "proton-policy-engine"
	// policyEngine 的 Helm release 坝称
	ReleaseNamePolicyEngine = "proton-policy-engine"
)

type ProtonDataConf struct {
	ReplicaCount int      `json:"replica_count,omitempty"`
	Hosts        []string `json:"hosts"`
	Data_path    string   `json:"data_path"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`

	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ProtonMariaDBConfigs struct {
	LowerCaseTableNames int `json:"lower_case_table_names,omitempty"`

	Thread_handling          string `json:"thread_handling,omitempty"`
	Innodb_buffer_pool_size  string `json:"innodb_buffer_pool_size"`
	Resource_requests_memory string `json:"resource_requests_memory"`
	Resource_limits_memory   string `json:"resource_limits_memory"`
}

type OpenSearch struct {
	HelmComponent `json:",inline"`

	ReplicaCount int               `json:"replica_count,omitempty"`
	Hosts        []string          `json:"hosts"`
	Data_path    string            `json:"data_path"`
	Mode         OpenSearchMode    `json:"mode"`
	Config       OpenSearchConfigs `json:"config"`

	// OpenSearch 的原生坂数
	// 	Detail: "https://opensearch.org/docs/latest/security/configuration/yaml/#opensearchyml"
	Settings map[string]interface{} `json:"settings,omitempty"`
	// OpenSearch 需求的资溝
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	ExporterResources *v1.ResourceRequirements `json:"exporter_resources,omitempty"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}

type OpenSearchMode string

type Version string

type OpenSearchConfigs struct {
	JvmOptions              string `json:"jvmOptions"`
	HanlpRemoteextDict      string `json:"hanlpRemoteextDict"`
	HanlpRemoteextStopwords string `json:"hanlpRemoteextStopwords"`
}

const (
	// orientdb 的 Chart 坝称
	ChartNameOrientDB = "kg-orientdb"

	// orientdb 的 Helm release 坝称，与 chart 坝称一致
	ReleaseNameOrientDB = "kg-orientdb"
)

// OrientDB 是杝述 OrientDB 酝置的结构
type OrientDB struct {
	// 部署 OrientDB 的节点坝称列表
	// OrientDB 当剝仅支挝坕节点部署，为保挝兼容性（已绝实现了Migrate）
	// 这里沿用列表的设计，在validate部分坚检查确保Hosts中坪有1个
	Hosts []string `json:"hosts,omitempty"`

	// OrientDB 使用主机的此目录作为数杮目录
	DataPath string `json:"data_path,omitempty"`

	// 使用的 storage class 的坝称，空字符串代表丝使用 storage class
	StorageClassName string `json:"storageClassName,omitempty"`
	StorageCapacity  string `json:"storage_capacity,omitempty"`
}

// CMS 代表 AnyShare 的 configuration management service
type CMS struct{}

const (
	// installer-service 的 Chart 名称
	ChartNameInstallerService = "installer-service"

	// installer-service 的 Helm release 名称，与 chart 名称一致
	ReleaseNameInstallerService = "installer-service"
)

// InstallerService 代表 AnyShare 的 installer service
type InstallerService struct{}

const (
	// nvidia-device-plugin 的 Chart 名称
	ChartNameNvDevPlugin = "nvidia-device-plugin"

	// nvidia-device-plugin 的 Helm release 名称，与 chart 名称一致
	ReleaseNameNvDevPlugin = "nvidia-device-plugin"
)

// NvidiaDevicePlugin 代表 nvidia-device-plugin，在K8S集群中使用英伟达GPU时需要使用
type NvidiaDevicePlugin struct{}

// slb
type HaProxyConf struct {
	Conf Conf `json:"conf"`
}

type Conf struct {
	Global   `json:"global"`
	Defaults `json:"defaults"`
	Frontend `json:"frontend"`
	Backend  `json:"backend"`
}
type Global struct {
	Log     string `json:"log"`
	Maxconn string `json:"maxconn"`
}
type Defaults struct {
	Maxconn string `json:"maxconn"`
	Mode    string `json:"mode"`
	Timeout string `json:"timeout"`
	Option  string `json:"option"`
}

type Frontend struct {
	CsFrontend            `json:"cs"`
	CrChartmuseumFrontend `json:"cr-chartmuseum"`
	CrRpmFrontend         `json:"cr-rpm"`
	CrRegistryFrontend    `json:"cr-registry"`
}
type CsFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}
type CrChartmuseumFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}
type CrRpmFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}
type CrRegistryFrontend struct {
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"default_backend"`
}

type Backend struct {
	CsBackend            `json:"cs"`
	CrRpmBackend         `json:"cr-rpm"`
	CrRegistryBackend    `json:"cr-registry"`
	CrChartmuseumBackend `json:"cr-chartmuseum"`
}

type CsBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}
type CrRpmBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}
type CrRegistryBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}
type CrChartmuseumBackend struct {
	Option        []string `json:"option"`
	DefaultServer string   `json:"default-server"`
	HTTPCheck     string   `json:"http-check"`
	Balance       string   `json:"balance"`
	Server        []string `json:"server"`
	Mode          string   `json:"mode"`
}

type CrConf struct {
	ConfigFile ConfigFile `json:"configfile,omitempty"`
	Port       Port       `json:"port,omitempty"`
	Storage    string     `json:"storage,omitempty"`
}

type ConfigFile struct {
	Chartmuseum string `json:"chartmuseum,omitempty"`
	Registry    string `json:"registry,omitempty"`
	Rpm         string `json:"rpm,omitempty"`
}

type Port struct {
	Chartmuseum int `json:"chartmuseum,omitempty"`
	Crmanager   int `json:"crmanager,omitempty"`
	Registry    int `json:"registry,omitempty"`
	Rpm         int `json:"rpm,omitempty"`
}

type ChartInfo struct {
	ApiVersion  string `json:"apiVersion"`
	AppVersion  string `json:"appVersion"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Version     string `json:"version"`
}

// kubeconfig configmap struct

type ClusterConfiguration struct {
	// metav1.TypeMeta
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	// Etcd holds configuration for etcd.
	Etcd Etcd `json:"etcd"`

	// Networking holds configuration for the networking topology of the cluster.
	Networking Networking `json:"networking"`
	// KubernetesVersion is the target version of the control plane.
	KubernetesVersion string `json:"kubernetesVersion"`

	// ControlPlaneEndpoint sets a stable IP address or DNS name for the control plane; it
	// can be a valid IP address or a RFC-1123 DNS subdomain, both with optional TCP port.
	// In case the ControlPlaneEndpoint is not specified, the AdvertiseAddress + BindPort
	// are used; in case the ControlPlaneEndpoint is specified but without a TCP port,
	// the BindPort is used.
	// Possible usages are:
	// e.g. In a cluster with more than one control plane instances, this field should be
	// assigned the address of the external load balancer in front of the
	// control plane instances.
	// e.g.  in environments with enforced node recycling, the ControlPlaneEndpoint
	// could be used for assigning a stable DNS to the control plane.
	ControlPlaneEndpoint string `json:"controlPlaneEndpoint"`

	// APIServer contains extra settings for the API server control plane component
	APIServer APIServer `json:"apiServer"`

	// ControllerManager contains extra settings for the controller manager control plane component
	ControllerManager ControlPlaneComponent `json:"controllerManager"`

	// Scheduler contains extra settings for the scheduler control plane component
	Scheduler ControlPlaneComponent `json:"scheduler"`

	// DNS defines the options for the DNS add-on installed in the cluster.
	DNS DNS `json:"dns"`

	// CertificatesDir specifies where to store or look for all required certificates.
	CertificatesDir string `json:"certificatesDir"`

	// ImageRepository sets the container registry to pull images from.
	// If empty, `k8s.gcr.io` will be used by default; in case of kubernetes version is a CI build (kubernetes version starts with `ci/`)
	// `gcr.io/k8s-staging-ci-images` will be used as a default for control plane components and for kube-proxy, while `k8s.gcr.io`
	// will be used for all the other images.
	ImageRepository string `json:"imageRepository"`

	// CIImageRepository is the container registry for core images generated by CI.
	// Useful for running kubeadm with images from CI builds.
	// +k8s:conversion-gen=false
	// CIImageRepository string

	// FeatureGates enabled by the user.
	// FeatureGates map[string]bool

	// The cluster name
	ClusterName string `json:"clusterName"`
}

type Etcd struct {
	// Local provides configuration knobs for configuring the local etcd instance
	// Local and External are mutually exclusive
	Local *LocalEtcd `json:"local"`

	// External describes how to connect to an external etcd cluster
	// Local and External are mutually exclusive
	// External *ExternalEtcd
}

type LocalEtcd struct {
	// ImageMeta allows to customize the container used for etcd
	// ImageMeta `json:",inline"`

	// DataDir is the directory etcd will place its data.
	// Defaults to "/var/lib/etcd".
	DataDir string `json:"dataDir"`

	// ExtraArgs are extra arguments provided to the etcd binary
	// when run inside a static pod.
	// A key in this map is the flag name as it appears on the
	// command line except without leading dash(es).
	ExtraArgs map[string]string `json:"extraArgs"`

	// ServerCertSANs sets extra Subject Alternative Names for the etcd server signing cert.
	// ServerCertSANs []string
	// PeerCertSANs sets extra Subject Alternative Names for the etcd peer signing cert.
	// PeerCertSANs []string
}

type ExternalEtcd struct {
	// Endpoints of etcd members. Useful for using external etcd.
	// If not provided, kubeadm will run etcd in a static pod.
	Endpoints []string
	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	CAFile string
	// CertFile is an SSL certification file used to secure etcd communication.
	CertFile string
	// KeyFile is an SSL key file used to secure etcd communication.
	KeyFile string
}

type Networking struct {
	// ServiceSubnet is the subnet used by k8s services. Defaults to "10.96.0.0/12".
	ServiceSubnet string `json:"serviceSubnet"`
	// PodSubnet is the subnet used by pods.
	PodSubnet string `json:"podSubnet"`
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain"`
}

type APIServer struct {
	ControlPlaneComponentExtra `json:",inline"`

	// CertSANs sets extra Subject Alternative Names for the API Server signing cert.
	// CertSANs []string

	// TimeoutForControlPlane controls the timeout that we use for API server to appear
	TimeoutForControlPlane string `json:"timeoutForControlPlane"`
}

type ControlPlaneComponent struct {
	// ExtraArgs is an extra set of flags to pass to the control plane component.
	// A key in this map is the flag name as it appears on the
	// command line except without leading dash(es).
	// TODO: This is temporary and ideally we would like to switch all components to
	// use ComponentConfig + ConfigMaps.
	ExtraArgs map[string]string `json:"extraArgs,inline"`

	// ExtraVolumes is an extra set of host volumes, mounted to the control plane component.
	// ExtraVolumes []HostPathMount
}

type ControlPlaneComponentExtra struct {
	// ExtraArgs is an extra set of flags to pass to the control plane component.
	// A key in this map is the flag name as it appears on the
	// command line except without leading dash(es).
	// TODO: This is temporary and ideally we would like to switch all components to
	// use ComponentConfig + ConfigMaps.
	ExtraArgs map[string]string `json:"extraArgs"`

	// ExtraVolumes is an extra set of host volumes, mounted to the control plane component.
	// ExtraVolumes []HostPathMount
}

type HostPathMount struct {
	// Name of the volume inside the pod template.
	Name string
	// HostPath is the path in the host that will be mounted inside
	// the pod.
	HostPath string
	// MountPath is the path inside the pod where hostPath will be mounted.
	MountPath string
	// ReadOnly controls write access to the volume
	ReadOnly bool
	// PathType is the type of the HostPath.
	PathType v1.HostPathType
}

type DNS struct {
	// Type defines the DNS add-on to be used
	// TODO: Used only in validation over the internal type. Remove with v1beta2 https://github.com/kubernetes/kubeadm/issues/2459
	// Type DNSAddOnType

	// ImageMeta allows to customize the image used for the DNS component
	// ImageMeta `json:",inline"`
}
type DNSAddOnType string

type ImageMeta struct {
	// ImageRepository sets the container registry to pull images from.
	// if not set, the ImageRepository defined in ClusterConfiguration will be used instead.
	ImageRepository string

	// ImageTag allows to specify a tag for the image.
	// In case this value is set, kubeadm does not change automatically the version of the above components during upgrades.
	ImageTag string
}

const KubeadmJoinDefault = `
apiVersion: kubeadm.k8s.io/v1beta2
controlPlane:
  localAPIEndpoint:
    advertiseAddress: ''
  certificateKey: e6a2eb8581237ab72a4f494f30285ec12a9694d750b9785706a83bfcbbbd2204
discovery:
  bootstrapToken:
    apiServerEndpoint: proton-cs.lb.aishu.cn:9443
    token: 783bde.3f89s0fje9f38fhf
    unsafeSkipCAVerification: true
kind: JoinConfiguration
nodeRegistration:
  taints: []
`

const KubeadmInitDefault = `
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
bootstrapTokens:
  - description: another bootstrap token
    token: 783bde.3f89s0fje9f38fhf
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  taints: []
  kubeletExtraArgs:
    allowed-unsafe-sysctls: net.core.somaxconn
    node-ip: 10.4.71.158

localAPIEndpoint:
  advertiseAddress: ''
certificateKey: e6a2eb8581237ab72a4f494f30285ec12a9694d750b9785706a83bfcbbbd2204
`

const KubeadmKubeletDefault = `
apiVersion: kubelet.config.k8s.io/v1beta1
evictionHard:
  nodefs.available: 0%
  imagefs.available: 0%
imageGCHighThresholdPercent: 100
imageGCLowThresholdPercent: 99
kind: KubeletConfiguration
`

// JoinConfiguration
type KubeadmInitDefaultStruct struct {
	APIVersion       string               `json:"apiVersion"`
	Kind             string               `json:"kind"`
	BootstrapTokens  []BootstrapTokens    `json:"bootstrapTokens"`
	NodeRegistration NodeRegistrationInit `json:"nodeRegistration"`
	LocalAPIEndpoint LocalAPIEndpoint     `json:"localAPIEndpoint"`
	CertificateKey   string               `json:"certificateKey"`
}
type BootstrapTokens struct {
	Description string `json:"description"`
	Token       string `json:"token"`
}
type NodeRegistrationInit struct {
	CriSocket        string           `json:"criSocket"`
	Taints           []interface{}    `json:"taints"`
	KubeletExtraArgs KubeletExtraArgs `json:"kubeletExtraArgs"`
}
type KubeletExtraArgs struct {
	AllowedUnsafeSysctls string `json:"allowed-unsafe-sysctls"`
	NodeIP               string `json:"node-ip"`
}
type LocalAPIEndpoint struct {
	AdvertiseAddress string `json:"advertiseAddress"`
}

// JoinConfiguration

type KubeadmJoinDefaultStruct struct {
	APIVersion       string               `json:"apiVersion"`
	ControlPlane     ControlPlane         `json:"controlPlane"`
	Discovery        Discovery            `json:"discovery"`
	Kind             string               `json:"kind"`
	NodeRegistration NodeRegistrationJoin `json:"nodeRegistration"`
}

type ControlPlane struct {
	LocalAPIEndpoint LocalAPIEndpoint `json:"localAPIEndpoint"`
	CertificateKey   string           `json:"certificateKey"`
}
type BootstrapToken struct {
	Token                    string `json:"token"`
	UnsafeSkipCAVerification bool   `json:"unsafeSkipCAVerification"`
	ApiServerEndpoint        string `json:"apiServerEndpoint"`
}
type Discovery struct {
	BootstrapToken BootstrapToken `json:"bootstrapToken"`
}

type NodeRegistrationJoin struct {
	Taints []interface{} `json:"taints"`
}

// KubeletConfiguration

type KubeadmKubeletDefaultStruct struct {
	APIVersion                  string       `json:"apiVersion"`
	EvictionHard                EvictionHard `json:"evictionHard"`
	ImageGCHighThresholdPercent int          `json:"imageGCHighThresholdPercent"`
	ImageGCLowThresholdPercent  int          `json:"imageGCLowThresholdPercent"`
	Kind                        string       `json:"kind"`
}
type EvictionHard struct {
	NodefsAvailable  string `json:"nodefs.available"`
	ImagefsAvailable string `json:"imagefs.available"`
}

// kube config struct /etc/kubernetes/kubelet.conf
type KubeConfig struct {
	ApiVersion      string        `json:"apiVersion"`
	Clusters        []ClusterInfo `json:"clusters"`
	Contexts        []ContextInfo `json:"contexts"`
	Current_context string        `json:"current-context"`
	Kind            string        `json:"kind"`
	Preferences     interface{}   `json:"preferences"`
	Users           []UserInfo    `json:"users"`
}

type ClusterInfo struct {
	Cluster Cluster `json:"cluster"`
	Name    string  `json:"name"`
}
type Cluster struct {
	Certificate_authority_data string `json:"certificate-authority-data"`
	Server                     string `json:"server"`
}

type ContextInfo struct {
	Context Context `json:"context"`
	Name    string  `json:"name"`
}
type Context struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type UserInfo struct {
	Name string `json:"name"`
	User User   `json:"user"`
}
type User struct {
	Client_certificate      string `json:"client-certificate,omitempty"`
	Client_key              string `json:"client-key,omitempty"`
	Client_certificate_data string `json:"client-certificate-data,omitempty"`
	Client_key_data         string `json:"client-key-data,omitempty"`
}

// 基础组件连接信杯
type ResourceConnectInfo struct {
	Rds          *types.MariaDBComponentInfo      `json:"rds,omitempty"`
	Mongodb      *types.MongoDBComponentInfo      `json:"mongodb,omitempty"`
	Redis        *types.RedisComponentInfo        `json:"redis,omitempty"`
	Mq           *MqInfo                          `json:"mq,omitempty"`
	OpenSearch   *types.OpensearchComponentInfo   `json:"opensearch,omitempty"`
	PolicyEngine *types.PolicyEngineComponentInfo `json:"policy_engine,omitempty"`
	Etcd         *types.ETCDComponentInfo         `json:"etcd,omitempty"`
}

type SourceType string

const (
	Internal SourceType = "internal"
	External SourceType = "external"
)

// Valid RDS type
type RDSType string

const (
	// 支挝的关系型数杮库
	DM8      RDSType = "DM8"
	MySQL    RDSType = "MySQL"
	MariaDB  RDSType = "MariaDB"
	GoldenDB RDSType = "GoldenDB"
	TiDB     RDSType = "TiDB"
)

// 当剝支挝的关系型数杮类型集
var RdsTypeList = []RDSType{DM8, MariaDB, MySQL, GoldenDB, TiDB}

type RdsInfo struct {
	SourceType SourceType `json:"source_type,omitempty"`

	RdsType RDSType `json:"rds_type,omitempty" mapstructure:"dbType" `

	Hosts    string `json:"hosts,omitempty" mapstructure:"host"`
	Port     int    `json:"port,omitempty" mapstructure:"port"`
	Username string `json:"username,omitempty" mapstructure:"user"`
	Password string `json:"password,omitempty" mapstructure:"password"`

	// 内置时扝有的字段
	HostsRead string `json:"hosts_read,omitempty" mapstructure:"hostRead"`
	PortRead  int    `json:"port_read,omitempty" mapstructure:"portRead"`
}

type MongodbInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`

	Hosts      string      `json:"hosts,omitempty" mapstructure:"host"`
	Port       int         `json:"port,omitempty" mapstructure:"port"`
	ReplicaSet string      `json:"replica_set,omitempty" mapstructure:"replicaSet"`
	Username   string      `json:"username,omitempty" mapstructure:"user"`
	Password   string      `json:"password,omitempty" mapstructure:"password"`
	SSL        bool        `json:"ssl,omitempty" mapstructure:"ssl"`
	AuthSource string      `json:"auth_source,omitempty" mapstructure:"authSource"`
	Options    interface{} `json:"options,omitempty" mapstructure:"options"`
}

type ConnectType string

const (
	SentinelMode     ConnectType = "sentinel"
	MasterSlaverMode ConnectType = "master-slave"
	StandAlonelMode  ConnectType = "standalone"
)

type RedisInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"connectType"`

	ConnectType ConnectType `json:"connect_type,omitempty"`

	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`

	// masterSlave 字段
	MasterHosts string `json:"master_hosts,omitempty" mapstructure:"masterHost"`
	MasterPort  int    `json:"master_port,omitempty" mapstructure:"masterPort"`
	SlaveHosts  string `json:"slave_hosts,omitempty" mapstructure:"slaveHost"`
	SlavePort   int    `json:"slave_port,omitempty" mapstructure:"slavePort"`
	// sentinel字段
	SentinelHosts    string `json:"sentinel_hosts,omitempty" mapstructure:"sentinelHost"`
	SentinelPort     int    `json:"sentinel_port,omitempty" mapstructure:"sentinelPort"`
	SentinelUsername string `json:"sentinel_username,omitempty" mapstructure:"sentinelUsername"`
	SentinelPassword string `json:"sentinel_password,omitempty" mapstructure:"sentinelPassword"`
	MasterGroupName  string `json:"master_group_name,omitempty" mapstructure:"masterGroupName"`
	// standAlone字段
	Hosts string `json:"hosts,omitempty" mapstructure:"host"`
	Port  int    `json:"port,omitempty" mapstructure:"port"`
}

type MqType string

type MechanType string

const (
	// 支挝的mq类型
	Nsq       MqType = "nsq"
	Tonglink  MqType = "tonglink"
	Htp2      MqType = "htp20"
	KafkaType MqType = "kafka"
	Besmq     MqType = "bmq"

	// 支挝的算法
	Sha512 MechanType = "SCRAM-SHA-512"
	Sha256 MechanType = "SCRAM-SHA-256"
	Plain  MechanType = "PLAIN"
)

// 当剝支挝的消杯中间件类型集
var MqTypeList = [5]MqType{Nsq, Tonglink, Htp2, KafkaType, Besmq}

// 当剝支挝的tls坝议加密类型
var MechanTypeList = [3]MechanType{Plain, Sha256, Sha512}

type MqInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`

	// 消息队列类型
	MQType MqType `json:"mq_type,omitempty" mapstructure:"mqType"`

	MQHosts        string `json:"mq_hosts,omitempty" mapstructure:"mqHost"`
	MQPort         int    `json:"mq_port,omitempty" mapstructure:"mqPort"`
	MQLookupdHosts string `json:"mq_lookupd_hosts,omitempty" mapstructure:"mqLookupdHost"`
	// lookupd仅有nsq使用，其它中间件为空
	MQLookupdPort int `json:"mq_lookupd_port,omitempty" mapstructure:"mqLookupdPort"`
	// 当仅有kafka支持配置auth，其它中间件为空
	Auth *types.KafkaAuth `json:"auth,omitempty" mapstructure:"auth"`
}

func (i *MqInfo) FromKafkaInfo(obj types.KafkaComponentInfo) {
	i.SourceType = "internal"
	i.MQType = "kafka"
	i.MQHosts = obj.MQHosts
	i.MQPort = obj.MQPort
	i.MQLookupdHosts = obj.MQLookupdHosts
	i.MQLookupdPort = obj.MQLookupdPort
	i.Auth = &obj.Auth
}

type Auth struct {
	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`

	Mechanism MechanType `json:"mechanism,omitempty" mapstructure:"mechanism"`
}

type OpensearchInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`
	Hosts      string     `json:"hosts,omitempty" mapstructure:"host"`
	Port       int        `json:"port,omitempty" mapstructure:"port"`
	Username   string     `json:"username,omitempty" mapstructure:"user"`
	Password   string     `json:"password,omitempty" mapstructure:"password"`
	Protocol   string     `json:"protocol,omitempty" mapstructure:"protocol"`
	Version    Version    `json:"version,omitempty" mapstructure:"version"`
}

type PolicyEngineInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`
	Hosts      string     `json:"hosts,omitempty" mapstructure:"host"`
	Port       int        `json:"port,omitempty" mapstructure:"port"`
}

type EtcdInfo struct {
	SourceType SourceType `json:"source_type,omitempty" mapstructure:"sourceType"`

	Hosts  string `json:"hosts,omitempty" mapstructure:"host"`
	Port   int    `json:"port,omitempty" mapstructure:"port"`
	Secret string `json:"secret,omitempty" mapstructure:"secret"`
}

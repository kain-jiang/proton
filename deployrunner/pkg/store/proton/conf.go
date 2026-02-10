package store

import (
	"encoding/base64"
	"fmt"

	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component/resources"
	helm "taskrunner/pkg/helm/repos"
	"taskrunner/trait"
)

// ResourceInfo proton conf resources connect info
type ResourceInfo struct {
	Rds        *resources.RDS         `json:"rds,omitempty"`
	Mongodb    *resources.MongoDB     `json:"mongodb,omitempty"`
	Redis      *resources.Redis       `json:"redis,omitempty"`
	MQ         *resources.MQ          `json:"mq,omitempty"`
	Opensearch *resources.Opensearch  `json:"opensearch,omitempty"`
	POA        map[string]interface{} `json:"policy_engine,omitempty"`
	Etcd       *resources.Etcd        `json:"etcd,omitempty"`
}

// ProtonConf proton-cli conf
type ProtonConf struct {
	Apiversion string          `json:"apiVersion"`
	Nodes      []Node          `json:"nodes"`
	Cs         CS              `json:"cs"`
	Cr         *CR             `json:"cr,omitempty"`
	Mariadb    *mariadb        `json:"proton_mariadb,omitempty"`
	Redis      *protonDB       `json:"proton_redis,omitempty"`
	MongoDB    *protonDB       `json:"proton_mongodb,omitempty"`
	NSQ        *protonDataConf `json:"proton_mq_nsq,omitempty"`
	POA        *protonDataConf `json:"proton_policy_engine,omitempty"`
	Etcd       *protonDB       `json:"proton_etcd,omitempty"`
	Opensearch *protonDB       `json:"opensearch,omitempty"`
	Kafka      *protonDB       `json:"kafka,omitempty"`
	Nebula     *Nebula         `json:"nebula,omitempty"`
	Resources  ResourceInfo    `json:"resource_connect_info"`
	DeployConf DeployConf      `json:"deploy"`
}

// DeployConf 部署配置
type DeployConf struct {
	// 统一的外置非自建k8s账户
	Serviceaccount string `json:"serviceaccount,omitempty"`
	// proton的统一命名空间,非空时proton基础服务和配置将设置于该空间
	Namespace string `json:"namespace,omitempty"`
}

func (c *DeployConf) init() {
	if c.Namespace == "" {
		c.Namespace = "resource"
	}
}

// CS container engine info
type CS struct {
	IPFamilies []string `json:"ipFamilies"`
}

func (c *CS) Have(protocol string) bool {
	for _, i := range c.IPFamilies {
		if i == protocol {
			return true
		}
	}
	return false
}

// Node proton node ingo
type Node struct {
	Name string `json:"name"`
	IP4  string `json:"ip4"`
	IP6  string `json:"ip6"`
}

// Nebula nebula install info
type Nebula struct {
	Hosts    []string `json:"hosts,omitempty"`
	Password string   `json:"password,omitempty"`
}

// Mongodb mongodb conf
type Mongodb struct {
	Hosts []string `json:"hosts"`
}

// CR cr conf
type CR struct {
	Local    *localCR    `json:"local,omitempty"`
	External *externalCR `json:"external,omitempty"`
}

// ToCRComponent return cr info
func (conf *ProtonConf) ToCRComponent() *resources.CR {
	cr := &resources.CR{}
	hrs := []helm.RepoConf{}
	ir := cluster.ImageRepo{}

	craw := conf.Cr
	if craw.External != nil {
		hr := helm.RepoConf{}
		if craw.External.ChartRepo == "oci" {
			hr.OCi = &helm.OCi{
				Registry:  craw.External.OCI.Registry,
				PlainHTTP: craw.External.OCI.PlainHTTP,
				Username:  craw.External.OCI.Username,
				Password:  craw.External.OCI.Password,
				RepoName:  "oci",
			}
		} else {
			hr.HTTPHelmRepo = &helm.HTTPHelmRepo{}
			hr.SourceType = "external"
			hr.HTTPHelmRepo.RepoName = "external"
			hr.URL = craw.External.Chartmuseum.Host
			if craw.External.Chartmuseum.UserName != "" {
				hr.AuthType = "basic"
				hr.BasicAuth.AuthPasswd = craw.External.Chartmuseum.Password
				hr.BasicAuth.AuthUser = craw.External.Chartmuseum.UserName
			}
			hr.ShouldPush = true
			hr.RetryCount = 3
			hr.RetryDelay = 100
		}

		hrs = append(hrs, hr)

		if craw.External.ImageRepo == "oci" {
			ir.Repo = craw.External.OCI.Registry
		} else {
			ir.Repo = craw.External.Registry.Host
		}

	} else {
		for _, hostname := range conf.Cr.Local.Hosts {
			for _, node := range conf.Nodes {
				addr := node.IP4
				if node.Name != hostname {
					continue
				}
				if addr == "" || !conf.Cs.Have("IPv4") {
					if node.IP6 != "" && conf.Cs.Have("IPv6") {
						addr = node.IP6
					}
				}

				hr := helm.HTTPHelmRepo{}
				hr.SourceType = "internal"
				hr.RepoName = "internal"
				hr.URL = fmt.Sprintf("http://%s:%d", addr, craw.Local.Ports.Chartmuseum)

				hr.ShouldPush = true
				hr.RetryCount = 3
				hr.RetryDelay = 100
				ir.Repo = fmt.Sprintf("registry.aishu.cn:%d", craw.Local.HaPorts.Registry)
				hrs = append(hrs, helm.RepoConf{
					HTTPHelmRepo: &hr,
				})
				break
			}
		}
	}
	cr.HelmRepo = hrs
	cr.ImageRepo = ir

	return cr
}

func (conf *ProtonConf) GetRegistries() []*OCI {
	craw := conf.Cr
	result := make([]*OCI, 0)
	if craw.External != nil {
		switch craw.External.ImageRepo {
		case "registry", "":
			result = append(result, &OCI{
				Registry:  craw.External.Registry.Host,
				Username:  craw.External.Registry.Username,
				Password:  craw.External.Registry.Password,
				PlainHTTP: false,
			})
		case "oci":
			result = append(result, &craw.External.OCI)
		default:
		}
	} else if craw.Local != nil {
		for _, hostname := range conf.Cr.Local.Hosts {
			for _, node := range conf.Nodes {
				addr := node.IP4
				if node.Name != hostname {
					continue
				}
				if addr == "" || !conf.Cs.Have("IPv4") {
					if node.IP6 != "" && conf.Cs.Have("IPv6") {
						addr = node.IP6
					}
				}

				h := fmt.Sprintf("%s:%d", addr, craw.Local.Ports.Registry)
				result = append(result,
					&OCI{
						Registry:  h,
						PlainHTTP: true,
					},
				)
			}
		}
	}
	return result
}

// // ToCRComponent to cr object
// func (c *CR) ToCRComponent() *resources.CR {
// 	cr := &resources.CR{}
// 	hr := helm.HTTPHelmRepo{}
// 	ir := cluster.ImageRepo{}

// 	if c.External != nil {
// 		hr.SourceType = "external"
// 		hr.RepoName = "external"
// 		hr.URL = fmt.Sprintf(c.External.Chartmuseum.Host)
// 		if c.External.Chartmuseum.UserName != "" {
// 			hr.AuthType = "basic"
// 			hr.BasicAuth.AuthPasswd = c.External.Chartmuseum.Password
// 			hr.BasicAuth.AuthUser = c.External.Chartmuseum.UserName
// 		}
// 		hr.ShouldPush = true
// 		hr.RetryCount = 3
// 		hr.RetryDelay = 100

// 		ir.Repo = c.External.Registry.Host

// 	} else {
// 		hr.SourceType = "internal"
// 		hr.RepoName = "internal"
// 		podNodeIP := os.Getenv("K8S_NODE_IP")
// 		if podNodeIP != "" {
// 			hr.URL = fmt.Sprintf("http://%s:%d", podNodeIP, c.Local.HaPorts.Chartmuseum)
// 		} else {
// 			hr.URL = fmt.Sprintf("http://chartmuseum.aishu.cn:%d", c.Local.HaPorts.Chartmuseum)
// 		}
// 		hr.ShouldPush = true
// 		hr.RetryCount = 3
// 		hr.RetryDelay = 100
// 		ir.Repo = fmt.Sprintf("registry.aishu.cn:%d", c.Local.HaPorts.Registry)
// 	}
// 	cr.HelmRepo = hr
// 	cr.ImageRepo = ir

// 	return cr
// }

type localCR struct {
	Hosts   []string `json:"hosts"`
	Ports   ports    `json:"ports"`
	HaPorts ports    `json:"ha_ports"`
	Storage string   `json:"storage"`
}

type externalCR struct {
	Chartmuseum chartmuseum `json:"chartmuseum"`
	Registry    registry    `json:"registry"`
	ChartRepo   string      `json:"chart_repository"`
	ImageRepo   string      `json:"image_repository"`
	OCI         OCI         `json:"oci"`
}

type OCI struct {
	Registry  string `json:"registry,omitempty"`
	PlainHTTP bool   `json:"plain_http,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

type registry struct {
	Host     string `json:"host,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type chartmuseum struct {
	Host     string `json:"host"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

type ports struct {
	Registry    int `json:"registry"`
	Chartmuseum int `json:"chartmuseum"`
}

type mariadb struct {
	Version string `json:"version,omitempty"`

	Hosts []string `json:"hosts"`
	// Config
	Passwd string `json:"admin_passwd"`
	User   string `json:"admin_user"`
}

// GetRDSComponent get rds component
func (conf *ProtonConf) GetRDSComponent() (*resources.RDS, *trait.Error) {
	rds := conf.Resources.Rds
	if rds == nil {
		return rds, &trait.Error{
			Err:      fmt.Errorf("rds not found, the instance not installed"),
			Internal: trait.ErrComponentNotFound,
			Detail:   "rds",
		}
	}

	if rds.SourceType == "internal" {
		if conf.Mariadb == nil {
			return rds, &trait.Error{
				Err:      fmt.Errorf("rds not found, the instance not installed"),
				Internal: trait.ErrComponentNotFound,
				Detail:   "rds",
			}
		}

		user := conf.Mariadb.User
		pass := conf.Mariadb.Passwd

		rds.AdminKey = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass)))
	}
	return rds, nil
}

type protonDB struct {
	AdminUser   string   `json:"admin_user"`
	AdminPasswd string   `json:"admin_passwd"`
	Hosts       []string `json:"hosts"`
}

type protonDataConf struct {
	Hosts []string `json:"hosts"`
}

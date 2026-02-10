package componentmanage

import (
	"encoding/json"
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

const (
	// ComponentManage 的 Chart 名称
	ChartName = "component-manage"
	// ComponentManage 的 Helm release 名称，与 chart 名称一致
	ReleaseName = "component-manage"
)

type ComponentManageParams struct {
	Release   string
	ChartFile string
	Namespace string
	Helm3     helm3.Client
	Values    map[string]interface{}
}

type Applier struct {
	ComponentManageParams

	OldCfg, NewCfg *configuration.ClusterConfig
	charts         servicepackage.Charts
	registry       string

	onlyInitComponent bool
	extraImages       []string
}

type Resetter struct {
	Namespace string
}

type canMust interface {
	configuration.MqInfo | configuration.RdsInfo |
		configuration.Kafka | *configuration.Kafka |
		configuration.ZooKeeper | *configuration.ZooKeeper |
		configuration.OpenSearch | *configuration.OpenSearch |
		configuration.ProtonDB | *configuration.ProtonDB |
		configuration.ProtonDataConf | *configuration.ProtonDataConf |
		configuration.PolicyEngineInfo | *configuration.PolicyEngineInfo |
		configuration.RedisInfo | *configuration.RedisInfo |
		configuration.OpensearchInfo | *configuration.OpensearchInfo |
		configuration.EtcdInfo | *configuration.EtcdInfo |
		configuration.ProtonMariaDB | *configuration.ProtonMariaDB |
		configuration.MongodbInfo | *configuration.MongodbInfo |
		configuration.Nebula | *configuration.Nebula
}

func mustToMap[T canMust](val T) map[string]interface{} {
	rel := make(map[string]interface{})
	d, err := json.Marshal(val)
	if err != nil {
		// 不可达
		panic(err)
	}
	err = json.Unmarshal(d, &rel)
	if err != nil {
		// 不可达
		panic(err)
	}
	return rel
}

func mustFromMap[T canMust](mq map[string]interface{}) *T {
	var rel T
	d, err := json.Marshal(mq)
	if err != nil {
		// 不可达
		panic(err)
	}
	err = json.Unmarshal(d, &rel)
	if err != nil {
		// 不可达
		panic(err)
	}
	return &rel
}

func NewManager(helm3 helm3.Client, oldCfg, newCfg *configuration.ClusterConfig, registry string, servicePackage string, charts servicepackage.Charts, images []string, namespace string) *Applier {
	myChart := charts.Get(ChartName, "")
	return &Applier{
		ComponentManageParams: ComponentManageParams{
			Release:   ReleaseName,
			ChartFile: filepath.Join(servicePackage, myChart.Path),
			Namespace: namespace,
			Helm3:     helm3,
			Values: map[string]interface{}{
				"image": map[string]interface{}{
					"registry": registry,
				},
				"serviceAccount": map[string]interface{}{
					"create": newCfg.Deploy.ServiceAccount == "",
					"name":   newCfg.Deploy.ServiceAccount,
				},
				"namespace": namespace,
				"service": map[string]interface{}{
					"enableDualStack": global.EnableDualStack,
					"config":          RepoForCr(newCfg.Cr),
				},
				"nodeSelector": getNodeSelector(newCfg),
			},
		},

		charts:      charts,
		OldCfg:      oldCfg,
		NewCfg:      newCfg,
		registry:    registry,
		extraImages: images,
	}
}

// getNodeSelector safely extracts the NodeSelector from config, handling nil cases
func getNodeSelector(cfg *configuration.ClusterConfig) map[string]string {
	if cfg != nil && cfg.ComponentManage != nil {
		return cfg.ComponentManage.NodeSelector
	}
	return nil
}

func RepoForCr(cr *configuration.Cr) map[string]interface{} {
	if cr.Local != nil {
		url, username, password := global.Chartmuseum(cr)
		return map[string]interface{}{
			"chartmuseum": map[string]interface{}{
				"url":      url,
				"username": username,
				"password": password,
				"enable":   true,
			},
		}
	}

	if cr.External != nil {
		switch cr.External.ChartRepo {
		case configuration.RepoChartmuseum:
			return map[string]interface{}{
				"chartmuseum": map[string]interface{}{
					"url":      cr.External.Chartmuseum.Host,
					"username": cr.External.Chartmuseum.Username,
					"password": cr.External.Chartmuseum.Password,
					"enable":   true,
				},
			}
		case configuration.RepoOCI:
			return map[string]interface{}{
				"oci": map[string]interface{}{
					"enable":     true,
					"plain_http": cr.External.OCI.PlainHTTP,
					"registry":   cr.External.OCI.Registry,
					"username":   cr.External.OCI.Username,
					"password":   cr.External.OCI.Password,
				},
			}
		case configuration.RepoDefault:
			return map[string]interface{}{
				"chartmuseum": map[string]interface{}{
					"url":      cr.External.Chartmuseum.Host,
					"username": cr.External.Chartmuseum.Username,
					"password": cr.External.Chartmuseum.Password,
					"enable":   true,
				},
			}
		}
	}
	return nil
}

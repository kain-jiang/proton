package reset

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cs"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/node"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/componentmanage"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/eceph"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/mq"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/store"
)

// var log = logger.NewLogger()

// 模块通过此接坣实现針置环境
type Restter interface {
	Reset() error
}

type module struct {
	name     string
	resetter Restter
}

func Reset(cfg *configuration.ClusterConfig) error {
	log := logger.NewLogger()
	log.Info("reset cluster")

	// 需覝針置的模块
	var modules []module

	_, kube := client.NewK8sClient()
	if kube == nil {
		return client.ErrKubernetesClientSetNil
	}

	// create node/v1alpha1
	var nodes []v1alpha1.Interface
	for _, node := range cfg.Nodes {
		n, err := v1alpha1.New(&node)
		if err != nil {
			return fmt.Errorf("%v: create node/v1alpha1 fail: %w", node.Name, err)
		}
		nodes = append(nodes, n)
	}

	// The helm client is nil, if the kubernetes is the kubernetes's provisioner is local.
	var helm helm3.Client
	// Create helm client with chart repository client.
	if cfg.Cs.Provisioner == configuration.KubernetesProvisionerExternal {
		var err error
		if helm, err = helm3.NewCli(configuration.GetProtonResourceNSFromFile(), log.WithField("helm", "v3")); err != nil {
			return fmt.Errorf("create helm3 client failed: %w", err)
		}
	}

	// 追加坯选模块
	modules = appendOptionalModules(modules, cfg, nodes, helm, log, kube)
	// 追加必覝模块
	modules = appendRequiredModules(modules, cfg, nodes)

	for _, m := range modules {
		log.Infof("reset module %s", m.name)
		if err := m.resetter.Reset(); err != nil {
			log.Warningf("reset module %s fail: %v", m.name, err)
		}
	}

	return nil
}

// 追加必覝的模块
//
//  1. cs
//  2. node
//  3. cr
func appendRequiredModules(modules []module, cfg *configuration.ClusterConfig, nodes []v1alpha1.Interface) []module {
	modules = append(modules, module{
		name: "cs",
		resetter: &cs.Cs{
			Logger:      logger.NewLogger(),
			ClusterConf: cfg,
			AllNodes:    nodes,
		},
	})
	modules = append(modules, module{
		name: "node",
		resetter: &node.Node{
			Logger:      logger.NewLogger(),
			ClusterConf: cfg,
			HttpClient:  client.NewHttpClient(30),
		},
	})
	modules = append(modules, module{
		name: "cr",
		resetter: &cr.Cr{
			Logger:      logger.NewLogger(),
			ClusterConf: cfg,
		},
	})
	return modules
}

// 追加坯选模块
//
//  1. opensearch
//  2. proton_etcd
//  3. proton_mariadb <delete>
//  4. proton_mongodb <delete>
//  5. proton_mq_nsq
//  6. proton_policy_engine
//  7. proton_redis
//  8. zookeeper <delete>
//  9. kafka <delete>
//  10. orientdb
//  11. prometheus
//  12. package store
func appendOptionalModules(modules []module, clusterConf *configuration.ClusterConfig, nodes []v1alpha1.Interface, helm helm3.Client, log logrus.FieldLogger, kube kubernetes.Interface) []module {
	if clusterConf.ComponentManage != nil {
		modules = append(modules, module{
			name: "component_manage",
			resetter: &componentmanage.Resetter{
				Namespace: configuration.GetProtonResourceNSFromFile(),
			},
		})
	}

	if clusterConf.Proton_mq_nsq != nil {
		modules = append(modules, module{
			name: "proton_mq_nsq",
			resetter: mq.
				New(clusterConf.Proton_mq_nsq).
				Hosts(clusterConf.Nodes),
		})
	}
	if clusterConf.Prometheus != nil {
		// The helm client is nil if the kubernetes's provisioner is external.
		if helm != nil {
			helm = helm.NameSpace(configuration.GetProtonResourceNSFromFile())
		}
		modules = append(modules, module{
			name:     "prometheus",
			resetter: prometheus.New(clusterConf.Prometheus, "", nodes, helm, log.WithField("module", "prometheus"), clusterConf.Cs.Provisioner, &configuration.Node{}, kube, (clusterConf.Proton_etcd != nil), nodes, nil, configuration.GetProtonResourceNSFromFile()),
		})
	}
	if spec := clusterConf.PackageStore; spec != nil {
		var selected []v1alpha1.Interface
		for _, n := range nodes {
			if slices.Contains(spec.Hosts, n.Name()) {
				selected = append(selected, n)
			}
		}
		modules = append(modules, module{
			name: "package-store",
			resetter: &store.Manager{
				Spec:   spec,
				Nodes:  selected,
				Logger: log.WithField("module", "package-store"),
			},
		})
	}
	if spec := clusterConf.ECeph; spec != nil {
		var selected []v1alpha1.Interface
		for _, n := range nodes {
			if slices.Contains(spec.Hosts, n.Name()) {
				selected = append(selected, n)
			}
		}
		modules = append(modules, module{
			name: "eceph",
			resetter: &eceph.Manager{
				Spec:   spec,
				Nodes:  selected,
				Logger: log.WithField("module", "eceph"),
			},
		})
	}
	return modules
}

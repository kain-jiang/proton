package apply

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	controller_runtime_client "sigs.k8s.io/controller-runtime/pkg/client"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
	v2 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	rds_mgmt "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/registry"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration/completion"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration/validation"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cs"
	csutil "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cs/util"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/firewall"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/node"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/cms"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/componentmanage"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/eceph"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/grafana"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/monitor"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/mq"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/nvidiadeviceplugin"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/store"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

// 模块通过此接坣实现安装和更新
type Applier interface {
	Apply() error
}

type module struct {
	name    string
	applier Applier
}

func Apply(clusterConf *configuration.ClusterConfig) error {
	// logger 的级别在创建时指定。修改日志级别丝影哝已绝创建的 logger 的 level，
	// 所以 logger 丝能作为 module 级别的坘針
	var log = logger.NewLogger()

	// 針置 global 坘針
	// TODO: 使用其他方法替杢这秝使用 module 坘針的行为，比如函数中的坘針
	global.NodeAuthSecret = make([]*corev1.Secret, 0)
	global.ChartInfoList = make([]configuration.ChartInfo, 0)
	global.EnableDualStack = clusterConf.Cs.EnableDualStack

	var olClusterConf *configuration.ClusterConfig

	if _, k := client.NewK8sClient(); k != nil {
		// 此时丝能保话 kubernetes 存在或坯用，所以从获坖旧酝置坯能会因为 kubernetes 无法访问或旧酝置丝存在而失败
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		c, err := configuration.LoadFromKubernetes(ctx, k)
		if err != nil {
			log.Debugf("unable load old cluster conf: %v", err)
		} else {
			olClusterConf = c
		}
	}

	pkg := new(servicepackage.ServicePackage)
	if err := pkg.Load(global.ServicePackage); err != nil {
		log.Errorf("unable load service package: %v", err)
		return err
	}
	pkgECeph := new(servicepackage.ServicePackage)
	if clusterConf.ECeph != nil && clusterConf.ECeph.SkipECephUpdate {
		if olClusterConf != nil && olClusterConf.ECeph != nil && len(olClusterConf.ECeph.Hosts) > 0 {
			clusterConf.ECeph = olClusterConf.ECeph
			clusterConf.ECeph.SkipECephUpdate = true
			log.Infoln("skip_eceph_update is true, skipping ECeph update discarding all changes in currently applying config for this time")
		} else {
			return fmt.Errorf("skip_eceph_update is true but ECeph seems not installed")
		}
	} else {
		if err := pkgECeph.Load(global.ServicePackageECeph); err != nil {
			log.Infof("unable to load ECeph service package: %v", err)
			if clusterConf.ECeph != nil && len(clusterConf.ECeph.Hosts) > 0 {
				log.Errorln("Cannot continue with ECeph installation if proton-cli is unable to load ECeph service package")
				return err
			}
		}
	}

	global.ChartInfoList = []configuration.ChartInfo{}
	fmt.Printf("\033[1;37;42m%s\033[0m\n", "start apply cluster conf")

	// 补全 cluster config 的酝置
	completion.CompleteClusterConfig(clusterConf, olClusterConf, pkg)

	if err := validation.ValidateClusterConfig(clusterConf); err != nil {
		log.Errorf("invalid cluster config: %v", err)
		return &validation.InvalidError{ErrorList: err}
	}

	if olClusterConf != nil {
		if err := validation.ValidateClusterConfigUpdate(olClusterConf, clusterConf); err != nil {
			log.Errorf("invalid cluster config update: %v", err)
			return &validation.InvalidError{ErrorList: err}
		}
	}

	// create helm client with chart repository client
	helm3client, err := helm3.NewCli(configuration.GetProtonResourceNSFromFile(), log.WithField("helm", "v3"))
	if err != nil {
		return fmt.Errorf("create helm3 client failed: %w", err)
	}

	var nodes []v1alpha1.Interface
	for _, node := range clusterConf.Nodes {
		n, err := v1alpha1.New(&node)
		if err != nil {
			return fmt.Errorf("create node/v1alpha1 %v fail: %w", node.Name, err)
		}
		nodes = append(nodes, n)
	}

	var helm2client *v2.Client
	// Helm2客户端基于命令执行，且可以在集群的任意节点上执行
	if clusterConf.Cs.Provisioner == configuration.KubernetesProvisionerLocal {
		helm2client = v2.New(exec.NewLocalShellExecutor())
	}

	// do checks and completion that requires clients created after initial checks here
	completion.CompleteClusterConfigPost(clusterConf, pkg, nodes)

	if err := validation.ValidateClusterConfigPost(clusterConf); err != nil {
		log.Errorf("invalid cluster config: %v", err)
		return &validation.InvalidError{ErrorList: err}
	}

	// 必选模块
	for _, m := range appendRequiredModules(nil, clusterConf, olClusterConf, nodes) {
		if err := m.applier.Apply(); err != nil {
			return fmt.Errorf("apply module %s fail: %w", m.name, err)
		}
	}

	// Kubernetes 是必选模块之一，所以 Kubernetes 客户端需覝在必选模块安装完戝坎创建
	_, kube := client.NewK8sClient()
	if kube == nil {
		return client.ErrKubernetesClientSetNil
	}

	registry, err := registry.New(registry.ConfigForCR(clusterConf.Cr))
	if err != nil {
		return fmt.Errorf("create registry client fail: %w", err)
	}
	restConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return fmt.Errorf("create rest config fail: %w", err)
	}

	controllerClient, err := controller_runtime_client.New(restConfig, controller_runtime_client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return fmt.Errorf("create controller runtime client fail: %w", err)
	}

	// 可选模块
	var mECeph *module
	for _, m := range appendOptionalModules(nil, kube, helm3client, helm2client, controllerClient, clusterConf, olClusterConf, registry, pkg, pkgECeph, nodes) {
		if m.name == "eceph" {
			mECeph = &m
			log.Infof("ECeph apply process will be executed last after all other operations are done")
		} else {
			if err := m.applier.Apply(); err != nil {
				return fmt.Errorf("apply module %s fail: %w", m.name, err)
			}
		}
	}

	nodeManager := &node.Node{
		Logger:         logger.NewLogger(),
		ClusterConf:    clusterConf,
		OldClusterConf: olClusterConf,
		HttpClient:     client.NewHttpClient(30),
	}
	// 移除 proton-cli 1.2 坊更旧版本生戝的酝置文件
	// TODO: 1.4 丝冝处睆 1.2坊更旧版本生戝的酝置文件
	if err := nodeManager.RemoveConf(); err != nil {
		return fmt.Errorf("remove conf fail: %w", err)
	}
	if mECeph == nil {
		if err := configuration.UploadToKubernetes(context.Background(), clusterConf, kube); err != nil {
			return fmt.Errorf("unable to upload cluster config to kubernetes: %w", err)
		}
	} else {
		clusterConfWithoutECeph := *clusterConf
		if olClusterConf != nil && olClusterConf.ECeph != nil {
			clusterConfWithoutECeph.ECeph = olClusterConf.ECeph
		} else {
			clusterConfWithoutECeph.ECeph = nil
		}
		var ns string
		if clusterConf.Deploy != nil {
			ns = clusterConf.Deploy.Namespace
		}
		if err := configuration.UploadToKubernetes(context.Background(), &clusterConfWithoutECeph, kube, ns); err != nil {
			return fmt.Errorf("unable to upload cluster config without ECeph to kubernetes: %w", err)
		}
		log.Info("Apply conf except ECeph success, now applying ECeph")
		if err := mECeph.applier.Apply(); err != nil {
			return fmt.Errorf("apply module %s fail: %w", mECeph.name, err)
		}
		if err := configuration.UploadToKubernetes(context.Background(), clusterConf, kube); err != nil {
			return fmt.Errorf("unable to upload cluster config to kubernetes: %w", err)
		}
	}

	if err := csutil.SyncEtcdDataDir(log, clusterConf, nodes, kube); err != nil {
		return fmt.Errorf("failed to sync etcd data directory: %w", err)
	}

	log.Info("Apply conf success")

	fmt.Printf("\033[1;37;42m%s\033[0m\n", "Apply conf success")
	return nil
}

// 追加必覝的模块
// 坄模块执行顺庝根杮以下依赖关系得到
//
//	cr -> firewall
//	cr -> nodes
//	cs -> cr
//	cs -> firewall
//	cs -> nodes
//	nodes -> firewall
//
//	1. firewall
//	2. nodes
//	3. cr
//	4. cs
func appendRequiredModules(modules []module, clusterConf, olClusterConf *configuration.ClusterConfig, nodes []v1alpha1.Interface) []module {
	modules = append(modules, module{
		name: "firewall",
		applier: firewall.New(
			&clusterConf.Firewall,
			nodes,
			clusterConf.Cs.Host_network.Pod_network_cidr,
			logger.NewLogger(),
		),
	})
	modules = append(modules, module{
		name: "nodes",
		applier: &node.Node{
			Logger:         logger.NewLogger(),
			ClusterConf:    clusterConf,
			OldClusterConf: olClusterConf,
			HttpClient:     client.NewHttpClient(30),
		},
	})
	modules = append(modules, module{
		name: "cr",
		applier: &cr.Cr{
			Logger:        logger.NewLogger(),
			ClusterConf:   clusterConf,
			PrePullImages: false,
		},
	})
	modules = append(modules, module{
		name: "cs",
		applier: &cs.Cs{
			Logger:         logger.NewLogger(),
			ClusterConf:    clusterConf,
			OldClusterConf: olClusterConf,
			AllNodes:       nodes,
		},
	})
	return modules
}

// 追加坯选模块坄模块执行顺庝根杮以下依赖关系得到
//
//   - proton_policy_engine -> proton_etcd
//   - kafka -> zookeeper
//   - proton_policy_engine -> proton_redis
//   - proton_etcd -> prometheus
//
// 执行顺庝
//
//  1. cms(deprecated)
//  2. installer_service(deprecated)
//  3. nvidia_device_plugin
//  4. opensearch(moved to component-manage)
//  5. proton_etcd(moved to component-manage)
//  6. proton_mariadb(moved to component-manage)
//  7. proton_mongodb(moved to component-manage)
//  8. proton_mq_nsq
//  9. proton_redis(moved to component-manage)
//  10. proton_policy_engine(moved to component-manage)
//  11. zookeeper(moved to component-manage)
//  12. kafka(moved to component-manage)
//  13. orientdb(deprecated)
//  14. prometheus
//  15. grafana
//  16. nebula(moved to component-manage)
//  17. package store
//  18. ECeph(执行ECeph之前先提交一次除ECeph的proton-cli-config配置)
func appendOptionalModules(
	modules []module,
	kube kubernetes.Interface,
	helm3 helm3.Client,
	helm2 *v2.Client,
	controllerClient controller_runtime_client.Client,
	clusterConf, olClusterConf *configuration.ClusterConfig,
	registry registry.Interface,
	pkg, pkgECeph *servicepackage.ServicePackage,
	nodes []v1alpha1.Interface,
) []module {
	var resourceNamespace = configuration.GetProtonResourceNSFromFile()
	charts := pkg.Charts()
	images := pkg.Images()
	isExistProtonETCD := false
	if clusterConf.Proton_monitor != nil {
		modules = append(modules, module{
			name:    "proton_monitor",
			applier: monitor.NewManager(helm3, clusterConf.Proton_monitor, registry.Address(), global.ServicePackage, charts, clusterConf.Nodes),
		})
	}
	if clusterConf.CMS != nil {
		modules = append(modules, module{
			name:    "cms",
			applier: cms.NewManager(helm3, clusterConf.CMS, registry.Address(), global.ServicePackage, charts, clusterConf.Deploy.ServiceAccount),
		})
	}
	// TODO component-manager applier is required
	componentManageApplier := componentmanage.NewManager(
		helm3, olClusterConf, clusterConf,
		registry.Address(), global.ServicePackage,
		charts, images, resourceNamespace,
	)
	if clusterConf.ComponentManage != nil {
		modules = append(modules, module{
			name:    "component_manage",
			applier: componentManageApplier,
		})
	}
	if clusterConf.NvidiaDevicePlugin != nil {
		modules = append(modules, module{
			name:    "nvidia_device_plugin",
			applier: nvidiadeviceplugin.NewManager(helm3, clusterConf.NvidiaDevicePlugin, registry.Address(), global.ServicePackage, charts, resourceNamespace),
		})
	}

	if clusterConf.Proton_mq_nsq != nil {
		var old *configuration.ProtonDataConf
		if olClusterConf != nil {
			old = olClusterConf.Proton_mq_nsq
		}
		modules = append(modules, module{
			name: "proton_mq_nsq",
			applier: mq.
				New(clusterConf.Proton_mq_nsq).
				OldConfig(old).
				Helm3(helm3).
				Registry(registry.Address()).
				ServicePackage(global.ServicePackage).
				Charts(charts).
				Hosts(clusterConf.Nodes).
				ReleaseNamespace(resourceNamespace),
		})
	}
	// prometheus service that will be used by other modules.
	var servicePrometheus *corev1.Service
	if clusterConf.Prometheus != nil {
		// prometheus will be deployed on these nodes
		var selected []v1alpha1.Interface
		for _, n := range nodes {
			for _, h := range clusterConf.Prometheus.Hosts {
				if h == n.Name() {
					selected = append(selected, n)
				}
			}
		}
		csProvisioner := clusterConf.Cs.Provisioner
		var masterNode *configuration.Node
		for _, node := range clusterConf.Nodes {
			if node.Name == clusterConf.Cs.Master[0] {
				masterNode = &node
				break
			}
		}

		manager := prometheus.New(clusterConf.Prometheus, registry.Address(), selected, helm3, logger.NewLogger().WithField("module", "prometheus"), csProvisioner, masterNode, kube, isExistProtonETCD, nodes, pkg, resourceNamespace)
		modules = append(modules, module{
			name:    "prometheus",
			applier: manager,
		})
		servicePrometheus = manager.KubernetesService()
	}
	if clusterConf.Grafana != nil {
		// grafana will be deployed on this node
		var selected v1alpha1.Interface
		for _, n := range nodes {
			if n.Name() == clusterConf.Grafana.Hosts[0] {
				selected = n
				break
			}
		}
		modules = append(modules, module{
			name: "grafana",
			applier: &grafana.Manager{
				Registry:       registry.Address(),
				Spec:           clusterConf.Grafana,
				Node:           selected,
				Helm:           helm3,
				Prometheus:     servicePrometheus,
				ServicePackage: pkg,
				Logger:         logger.NewLogger().WithField("module", "grafana"),
				Namespace:      resourceNamespace,
			},
		})
	}

	if spec := clusterConf.PackageStore; spec != nil {
		// package store will be running on these nodes
		var selected []v1alpha1.Interface
		for _, n := range nodes {
			for _, h := range spec.Hosts {
				if h == n.Name() {
					selected = append(selected, n)
				}
			}
		}
		manager := &store.Manager{
			Registry:       registry.Address(),
			Spec:           spec,
			RDS:            clusterConf.ResourceConnectInfo.Rds,
			Nodes:          selected,
			Helm:           helm3,
			Kube:           controllerClient,
			ServicePackage: pkg,
			Logger:         logger.NewLogger().WithField("module", "package-store"),
			Namespace:      resourceNamespace,
		}
		modules = append(modules, module{name: "package-store", applier: manager})
	}
	if spec := clusterConf.ECeph; spec != nil {
		var selected []v1alpha1.Interface
		for _, n := range nodes {
			for _, h := range spec.Hosts {
				if h == n.Name() {
					selected = append(selected, n)
				}
			}
		}
		var oldSpec *configuration.ECeph
		if olClusterConf != nil {
			oldSpec = olClusterConf.ECeph
		}
		manager := &eceph.Manager{
			Registry: registry.Address(),
			Spec:     spec,
			OldSpec:  oldSpec,
			RDS:      clusterConf.ResourceConnectInfo.Rds,
			RDS_MGMTClientCreateFunc: func() (rds_mgmt.Interface, error) {
				return componentManageApplier.RDS_MGMTClient(controllerClient)
			},
			InitDatabase: clusterConf.Proton_mariadb != nil,
			Nodes:        selected,
			Helm:         helm2,
			Kube:         controllerClient,
			Logger:       logger.NewLogger().WithField("module", "eceph"),
			PkgECeph:     pkgECeph,
		}
		modules = append(modules, module{name: "eceph", applier: manager})
	}
	return modules
}

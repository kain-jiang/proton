package validation

import (
	"reflect"

	"golang.org/x/exp/slices"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	eceph "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/eceph/validation"
)

const FieldNameStorageCapacity string = "StorageCapacity"
const FieldNameReplicaCount string = "ReplicaCount"

// ValidateClusterConfig tests whether required fields in th ClusterConfig are set.
func ValidateClusterConfig(c *configuration.ClusterConfig) (allErrs field.ErrorList) {
	allErrs = append(allErrs, ValidateNodes(c.Nodes, field.NewPath("nodes"))...)
	nodeNameSet := NewNodeNameSet(c.Nodes)

	// validate that replica_count should be set in external k8s cluster mode and should not be set in internal k8s mode
	allErrs = append(allErrs, ValidateReplicaCount(c)...)

	if c.Deploy == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("deploy"), "must set deploy mode and devicespec"))
	} else {
		validModes := []string{"standard", "cloud"}
		if !slices.Contains(validModes, c.Deploy.Mode) {
			allErrs = append(allErrs, field.NotSupported(field.NewPath("deploy", "mode"), c.Deploy.Mode, validModes))
		}
	}

	if c.Cs == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("cs"), ""))
	} else {
		allErrs = append(allErrs, ValidateCS(c.Cs, c.Nodes, field.NewPath("cs"))...)
	}
	if c.Chrony != nil {
		allErrs = append(allErrs, ValidateChrony(c.Chrony, c.Cs, field.NewPath("chrony"))...)
	}
	allErrs = append(allErrs, ValidateFirewall(&c.Firewall, field.NewPath("firewall"))...)
	if c.Cr == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("cr"), ""))
	} else {
		allErrs = append(allErrs, ValidateCR(c.Cr, field.NewPath("cr"))...)
	}
	if c.Proton_mariadb != nil {
		allErrs = append(allErrs, ValidateMariaDB(c.Proton_mariadb, nodeNameSet, field.NewPath("proton_mariadb"))...)
	}
	if c.OpenSearch != nil {
		allErrs = append(allErrs, ValidateOpenSearch(c.OpenSearch, nodeNameSet, field.NewPath("opensearch"))...)
	}
	if c.Kafka != nil {
		allErrs = append(allErrs, ValidateKafka(c.Kafka, nodeNameSet, field.NewPath("kafka"))...)
		if c.ZooKeeper == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("zookeeper"), "kafka requires zookeeper"))
		}
	}
	if c.ZooKeeper != nil {
		allErrs = append(allErrs, ValidateZooKeeper(c.ZooKeeper, nodeNameSet, field.NewPath("zookeeper"))...)
	}
	if c.Proton_mq_nsq != nil {
		allErrs = append(allErrs, ValidateMQNSQ(c.Proton_mq_nsq, nodeNameSet, field.NewPath("proton_mq_nsq"))...)
	}
	if c.Proton_mongodb != nil {
		allErrs = append(allErrs, ValidateMongodb(c.Proton_mongodb, nodeNameSet, field.NewPath("mongodb"))...)
	}
	if c.Prometheus != nil {
		allErrs = append(allErrs, ValidatePrometheus(c.Prometheus, nodeNameSet, field.NewPath("prometheus"))...)
	}
	if c.Grafana != nil {
		allErrs = append(allErrs, ValidateGrafana(c.Grafana, nodeNameSet, field.NewPath("grafana"))...)
		if c.Prometheus == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("prometheus"), "grafana requires prometheus"))
		}
	}

	if c.ResourceConnectInfo != nil {
		allErrs = append(allErrs, ValidateResourceConnectInfo(c, field.NewPath("resource_connect_info"))...)
	}
	if c.PackageStore != nil {
		if c.ResourceConnectInfo == nil || c.ResourceConnectInfo.Rds == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("resource_connect_info", "rds"), ""))
		}
		allErrs = append(allErrs, ValidatePackageStore(c.PackageStore, nodeNameSet, field.NewPath("package-store"))...)
	}
	if c.Cs.Provisioner == configuration.KubernetesProvisionerLocal && c.ECeph != nil {
		allErrs = append(allErrs, eceph.Validate(c.ECeph, c.Nodes, c.ResourceConnectInfo, field.NewPath("eceph"))...)
	} else if c.ECeph != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("eceph"), "", "Cannot deploy ECeph on a non-local cluster."))
	}

	return
}

func ValidateReplicaCount(c *configuration.ClusterConfig) (allErrs field.ErrorList) {
	if c.Cs.Provisioner == configuration.KubernetesProvisionerExternal {
		invalidReplicaCountTips := "replica_count should be specified and should be >= 1 when using external k8s cluster"
		if c.Proton_mariadb != nil {
			fp := field.NewPath("proton_mariadb").Child(FieldNameReplicaCount)
			if c.Proton_mariadb.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Proton_mariadb.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.Proton_mongodb != nil {
			fp := field.NewPath("proton_mongodb").Child(FieldNameReplicaCount)
			if c.Proton_mongodb.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Proton_mongodb.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.Proton_mq_nsq != nil {
			fp := field.NewPath("proton_mq_nsq").Child(FieldNameReplicaCount)
			if c.Proton_mq_nsq.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Proton_mq_nsq.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.OpenSearch != nil {
			fp := field.NewPath("opensearch").Child(FieldNameReplicaCount)
			if c.OpenSearch.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.OpenSearch.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.Proton_policy_engine != nil {
			fp := field.NewPath("proton_policy_engine").Child(FieldNameReplicaCount)
			if c.Proton_policy_engine.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Proton_policy_engine.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.Proton_etcd != nil {
			fp := field.NewPath("etcd").Child(FieldNameReplicaCount)
			if c.Proton_etcd.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Proton_etcd.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.Proton_redis != nil {
			fp := field.NewPath("proton_redis").Child(FieldNameReplicaCount)
			if c.Proton_redis.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Proton_redis.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.Kafka != nil {
			fp := field.NewPath("kafka").Child(FieldNameReplicaCount)
			if c.Kafka.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.Kafka.ReplicaCount, invalidReplicaCountTips))
			}
		}
		if c.ZooKeeper != nil {
			fp := field.NewPath("zookeeper").Child(FieldNameReplicaCount)
			if c.ZooKeeper.ReplicaCount < 1 {
				allErrs = append(allErrs, field.Invalid(fp, c.ZooKeeper.ReplicaCount, invalidReplicaCountTips))
			}
		}
	} else if c.Cs.Provisioner == configuration.KubernetesProvisionerLocal {
		// replica_count should not be provided in local k8s deployment and they would simply be ignored
	} else {
		panic("unsupported K8S cluster provisioner")
	}
	return
}

// TODO: 补充数杮朝务的更新检查，检查数杮目录〝storage class 是坦更改
func ValidateClusterConfigUpdate(o, n *configuration.ClusterConfig) (allErrs field.ErrorList) {
	allErrs = append(allErrs, ValidateCRUpdate(o.Cr, n.Cr, field.NewPath("cr"))...)
	allErrs = append(allErrs, ValidateCSUpdate(o.Cs, n.Cs, field.NewPath("cs"))...)

	//validate that storage capacity is not changed
	immutableStorageCapacityTips := "storage_capacity is immutable and cannot be changed"
	oldConfType := reflect.TypeOf(*o)
	newConfType := reflect.TypeOf(*n)
	oldConfVal := reflect.ValueOf(*o)
	newConfVal := reflect.ValueOf(*n)
	for i := 0; i < oldConfType.NumField(); i++ {
		var oldStorageCapacity, newStorageCapacity string
		flagThisFieldIsStruct := reflect.TypeOf(oldConfVal.Field(i).Interface()) == reflect.TypeOf(reflect.Struct)
		if oldConfType.Field(i).IsExported() && flagThisFieldIsStruct {
			oldSubItemType := reflect.TypeOf(oldConfVal.Field(i).Interface())
			oldSubItemValue := reflect.ValueOf(oldConfVal.Field(i).Interface())
			_, ok := oldSubItemType.FieldByName(FieldNameStorageCapacity)
			if ok {
				v, typeOK := (oldSubItemValue.FieldByName(FieldNameStorageCapacity).Interface()).(string)
				if typeOK {
					oldStorageCapacity = v
				}
			}
		}
		_, ok := newConfType.FieldByName(oldConfType.Field(i).Name)
		if ok && flagThisFieldIsStruct {
			newSubItemType := reflect.TypeOf(newConfVal.FieldByName(oldConfType.Field(i).Name).Interface())
			newSubItemValue := reflect.ValueOf(newConfVal.FieldByName(oldConfType.Field(i).Name).Interface())
			_, ok := newSubItemType.FieldByName(FieldNameStorageCapacity)
			if ok {
				v, typeOK := (newSubItemValue.FieldByName(FieldNameStorageCapacity).Interface()).(string)
				if typeOK {
					newStorageCapacity = v
				}
			}
		}
		if oldStorageCapacity != newStorageCapacity {
			fp := field.NewPath(oldConfType.Field(i).Name).Child(FieldNameStorageCapacity)
			allErrs = append(allErrs, field.Forbidden(fp, immutableStorageCapacityTips))
		}
	}

	if o.Proton_mariadb != nil {
		fldPath := field.NewPath("proton_mariadb")
		if n.Proton_mariadb != nil {
			allErrs = append(allErrs, ValidateMariaDBUpdate(o.Proton_mariadb, n.Proton_mariadb, fldPath)...)
		} else {
			allErrs = append(allErrs, field.Required(fldPath, "mariadb doesn't support uninstall"))
		}
	}

	if o.Kafka != nil {
		if n.Kafka != nil {
			allErrs = append(allErrs, ValidateKafkaUpdate(o.Kafka, n.Kafka, field.NewPath("kafka"))...)
		} else {
			allErrs = append(allErrs, field.Invalid(field.NewPath("kafka"), n.Kafka, "kafka doesn't support uninstall"))
		}
	}

	if o.ZooKeeper != nil {
		if n.ZooKeeper != nil {
			allErrs = append(allErrs, ValidateZooKeeperUpdate(o.ZooKeeper, n.ZooKeeper, field.NewPath("zookeeper"))...)
		} else {
			allErrs = append(allErrs, field.Invalid(field.NewPath("zookeeper"), n.ZooKeeper, "zookeeper doesn't support uninstall"))
		}
	}

	if o.Proton_redis != nil {
		if n.Proton_redis == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("proton_redis"), n.Proton_redis, "proton_redis doesn't support uninstall"))
		}
	}

	if o.Proton_policy_engine != nil {
		if n.Proton_policy_engine == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("proton_policy_engine"), n.Proton_policy_engine, "proton_policy_engine doesn't support uninstall"))
		}
	}

	if o.Proton_mq_nsq != nil {
		if n.Proton_mq_nsq != nil {
			allErrs = append(allErrs, ValidateMQNSQUpdate(o.Proton_mq_nsq, n.Proton_mq_nsq, field.NewPath("proton_mq_nsq"))...)
		}
		// 支持nsq卸载（不卸载实例）
	}

	if o.OpenSearch != nil {
		if n.OpenSearch == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("openSearch"), n.OpenSearch, "openSearch doesn't support uninstall"))
		}
	}

	if o.Proton_mongodb != nil {
		if n.Proton_mongodb != nil {
			allErrs = append(allErrs, ValidateMongodbUpdate(o.Proton_mongodb, n.Proton_mongodb, field.NewPath("mongodb"))...)
		} else {
			allErrs = append(allErrs, field.Invalid(field.NewPath("mongodb"), n.Proton_mongodb, "mongodb doesn't support uninstall"))
		}
	}

	if o.Prometheus != nil {
		if n.Prometheus != nil {
			allErrs = append(allErrs, ValidatePrometheusUpdate(o.Prometheus, n.Prometheus, field.NewPath("prometheus"))...)
		} else {
			allErrs = append(allErrs, field.Invalid(field.NewPath("prometheus"), n.Prometheus, "prometheus doesn't support uninstall"))
		}
	}

	if o.Grafana != nil {
		if n.Grafana != nil {
			allErrs = append(allErrs, ValidateGrafanaUpdate(o.Grafana, n.Grafana, field.NewPath("grafana"))...)
		} else {
			allErrs = append(allErrs, field.Invalid(field.NewPath("grafana"), n.Grafana, "grafana doesn't support uninstall"))
		}
	}

	if o.PackageStore != nil && n.PackageStore == nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("package-store"), n.PackageStore, "package store doesn't support uninstall"))
	}

	allErrs = append(allErrs, eceph.ValidateUpdate(o.ECeph, n.ECeph, field.NewPath("eceph"))...)
	return
}

// Do checks that requires clients like node client here
func ValidateClusterConfigPost(c *configuration.ClusterConfig) (allErrs field.ErrorList) {
	if c.ECeph != nil && c.Cs.Provisioner == configuration.KubernetesProvisionerLocal {
		allErrs = append(allErrs, eceph.ValidatePost(c.ECeph, c.Nodes, c.ResourceConnectInfo, field.NewPath("eceph"))...)
	} else if c.ECeph != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("eceph"), "", "Cannot deploy ECeph on a non-local cluster."))
	}
	return
}

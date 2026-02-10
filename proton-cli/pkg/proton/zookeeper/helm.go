package zookeeper

import (
	"strconv"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// HelmValuesFor return helm values for given configuration and registry
func HelmValuesFor(spec *configuration.ZooKeeper, registry string) helm3.M {
	v := helm3.M{
		"namespace":    ReleaseNamespace,
		"image":        helm3.M{"registry": registry},
		"replicaCount": HelmValuesReplicaCountFor(len(spec.Hosts), spec.ReplicaCount),
		"storage":      HelmValuesStorageFor(spec.Hosts, spec.Data_path, spec.StorageClassName),
		"config": helm3.M{
			"zookeeperENV": func() helm3.M {
				rel := make(helm3.M)
				for k, v := range spec.Env {
					rel[k] = v
				}
				return rel
			}(),
		},
	}

	if spec.Resources != nil {
		v["resources"] = spec.Resources.DeepCopy()
	}
	if len(spec.StorageCapacity) > 0 {
		v["storage"].(helm3.M)["capacity"] = spec.StorageCapacity
	}
	return v
}

// HelmValuesReplicaCountFor return given count if non-zero otherwise return
// defaultReplicaCount for .Values.replicaCount.
func HelmValuesReplicaCountFor(count, defaultReplicaCount int) int {
	if count == 0 {
		return defaultReplicaCount
	}
	return count
}

// HelmValuesStorageFor return storage configuration for .Values.storage
func HelmValuesStorageFor(hosts []string, dataPath string, storageClassName string) helm3.M {
	storage := helm3.M{}

	if storageClassName != "" {
		storage["storageClassName"] = storageClassName
	}

	if dataPath != "" {
		rel := make(helm3.M)
		for i, h := range hosts {
			rel[strconv.Itoa(i)] = helm3.M{
				"host": h,
				"path": dataPath,
			}
		}
		storage["local"] = rel
	}

	return storage
}

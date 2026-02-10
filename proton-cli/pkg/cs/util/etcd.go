package util

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	k8s "k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

const etcdStaticPodPath = "/etc/kubernetes/manifests/etcd.yaml"

// SyncEtcdDataDir checks the etcd data directory defined in the etcd static pod manifest
// on a master node and updates the cluster configuration if it differs.
func SyncEtcdDataDir(lg *logrus.Logger, conf *configuration.ClusterConfig, nodes []v1alpha1.Interface, k k8s.Interface) error {
	var ctx = context.TODO()
	if len(conf.Cs.Master) == 0 {
		lg.Debug("No master nodes in config, skipping etcd data dir sync.")
		return nil
	}

	masterNodeName := conf.Cs.Master[0]
	var masterNode v1alpha1.Interface
	for _, n := range nodes {
		if n.Name() == masterNodeName {
			masterNode = n
			break
		}
	}

	if masterNode == nil {
		return fmt.Errorf("master node %s not found in the node list", masterNodeName)
	}

	lg.Infof("Reading etcd static pod manifest from %s on node %s", etcdStaticPodPath, masterNodeName)
	content, err := masterNode.ECMS().Files().ReadFile(ctx, etcdStaticPodPath)
	if err != nil {
		lg.Warnf("Could not read etcd manifest on master node %s, maybe it's a new installation. Error: %v", masterNodeName, err)
		// If the file doesn't exist, it's likely a new cluster, so we don't need to sync.
		return nil
	}

	// Define a partial struct to unmarshal the etcd.yaml content
	type EtcdManifest struct {
		Spec struct {
			Volumes []struct {
				Name     string `yaml:"name"`
				HostPath struct {
					Path string `yaml:"path"`
				} `yaml:"hostPath"`
			} `yaml:"volumes"`
		} `yaml:"spec"`
	}

	var manifest EtcdManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return fmt.Errorf("failed to unmarshal etcd manifest from node %s: %w", masterNodeName, err)
	}

	var manifestPath string
	for _, vol := range manifest.Spec.Volumes {
		if vol.Name == "etcd-data" {
			manifestPath = vol.HostPath.Path
			break
		}
	}

	if manifestPath == "" {
		lg.Warnf("Could not find 'etcd-data' volume in etcd manifest on node %s. Skipping sync.", masterNodeName)
		return nil
	}

	if conf.Cs.Etcd_data_dir != manifestPath {
		lg.Infof("Found different etcd data directory in %s on master node. Updating config: '%s' -> '%s'", etcdStaticPodPath, conf.Cs.Etcd_data_dir, manifestPath)
		conf.Cs.Etcd_data_dir = manifestPath

		// Write the updated configuration back to the secret
		if err := configuration.UploadToKubernetes(context.TODO(), conf, k); err != nil {
			return fmt.Errorf("failed to upload updated configuration to secret: %w", err)
		}

		lg.Infof("Successfully updated and saved the configuration secret.")
	}

	return nil
}

package cluster

import (
	"taskrunner/pkg/helm"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"k8s.io/client-go/kubernetes"
)

// HelmManagerInterface is the manager in the cluster
type HelmManagerInterface struct {
	HelmRepo   helm.Repo
	HelmClient helm.Client
}

// SystemContext system info
type SystemContext struct {
	trait.System
	ImageRepo
	HelmManagerInterface
	Kcli kubernetes.Interface
}

// ToMap convert systemContext into config map
func (s *SystemContext) ToMap() map[string]interface{} {
	return utils.MergeMaps(map[string]interface{}{
		"image":     s.ImageRepo.ToMap(),
		"namespace": s.NameSpace,
	}, s.Config)
}

// ImageRepo use in k8s pull image
type ImageRepo struct {
	Repo            string `json:"repo"`
	ImagePullPolicy string `json:"imagePullpolicy"`
}

// ToMap convert systemContext into config map
func (i *ImageRepo) ToMap() map[string]interface{} {
	obj := map[string]interface{}{}
	if i.ImagePullPolicy != "" {
		obj["imagePullPolicy"] = i.ImagePullPolicy
	}
	if i.Repo != "" {
		obj["registry"] = i.Repo
	}
	return obj
}

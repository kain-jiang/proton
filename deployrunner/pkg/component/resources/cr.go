package resources

import (
	"taskrunner/pkg/cluster"
	helm "taskrunner/pkg/helm/repos"
)

// CR helm repo
type CR struct {
	HelmRepo  []helm.RepoConf   `json:"HelmRepo"`
	ImageRepo cluster.ImageRepo `json:"ImageRepo"`
}

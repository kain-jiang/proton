package helm

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// Storage defines helm values's storage of proton package store.
type Storage struct {
	StorageClassName string `json:"storageClassName,omitempty"`

	Capacity resource.Quantity `json:"capacity"`

	Local map[string]Local `json:"local,omitempty"`
}

func storageFor(spec *configuration.PackageStore) Storage {
	return Storage{
		StorageClassName: spec.Storage.StorageClassName,
		Capacity:         *spec.Storage.Capacity,
		Local:            localFor(spec.Hosts, spec.Storage.Path),
	}
}

// Local defines storage.local.[index] of proton package store's helm values.
type Local struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
}

func localFor(hosts []string, path string) map[string]Local {
	if hosts == nil {
		return nil
	}

	local := make(map[string]Local)
	for i, h := range hosts {
		local[strconv.Itoa(i)] = Local{Host: h, Path: path}
	}
	return local
}

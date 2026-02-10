package configuration

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PackageStore defines the deployment configuration of proton package store.
type PackageStore struct {
	Hosts []string `json:"hosts,omitempty"`

	Replicas *int `json:"replicas,omitempty"`

	Storage PackageStoreStorage `json:"storage,omitempty"`

	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// PackageStoreStorage defines the storage configuration of proton package
// store.
type PackageStoreStorage struct {
	StorageClassName string `json:"storageClassName,omitempty"`

	Capacity *resource.Quantity `json:"capacity,omitempty"`

	Path string `json:"path,omitempty"`
}

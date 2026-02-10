package helm

import (
	corev1 "k8s.io/api/core/v1"
)

// Resources defines helm values's resources of proton package store.
type Resources struct {
	Store *corev1.ResourceRequirements `json:"store,omitempty"`
}

func resourcesFor(resources *corev1.ResourceRequirements) *Resources {
	if resources == nil {
		return nil
	}
	return &Resources{Store: resources}
}

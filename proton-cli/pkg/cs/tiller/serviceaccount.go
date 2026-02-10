package tiller

import (
	api_core_v1 "k8s.io/api/core/v1"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ServiceAccountName = "tiller"

var ServiceAccount = api_core_v1.ServiceAccount{
	ObjectMeta: api_meta_v1.ObjectMeta{
		Name:   ServiceAccountName,
		Labels: TillerLabels,
	},
}

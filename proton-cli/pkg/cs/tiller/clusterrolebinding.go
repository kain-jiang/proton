package tiller

import (
	api_rbac_v1 "k8s.io/api/rbac/v1"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClusterRoleBindingName = "tiller"

	ClusterRoleBindingRoleRefAPIGroup = api_rbac_v1.GroupName
	ClusterRoleBindingRoleRefKind     = "ClusterRole"
	ClusterRoleBindingRoleRefName     = "cluster-admin"
)

var ClusterRoleBinding = api_rbac_v1.ClusterRoleBinding{
	ObjectMeta: api_meta_v1.ObjectMeta{
		Name:   ClusterRoleBindingName,
		Labels: TillerLabels,
	},
	Subjects: []api_rbac_v1.Subject{
		{
			Kind:      api_rbac_v1.ServiceAccountKind,
			Name:      ServiceAccountName,
			Namespace: Namespace,
		},
	},
	RoleRef: api_rbac_v1.RoleRef{
		APIGroup: api_rbac_v1.GroupName,
		Kind:     ClusterRoleBindingRoleRefKind,
		Name:     ClusterRoleBindingRoleRefName,
	},
}

package baseresource

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mongodbv1 "proton-mongodb-operator/api/v1"
)

// NewStatefulSet returns a StatefulSet object configured for a name
func NewMgmtClusterRoleBinding(instance *mongodbv1.MongodbOperator) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
			Kind:     "ClusterRole",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
				Namespace: instance.GetNamespace(),
			},
		},
	}
}

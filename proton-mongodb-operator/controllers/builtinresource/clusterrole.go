package baseresource

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mongodbv1 "proton-mongodb-operator/api/v1"
)

// NewStatefulSet returns a StatefulSet object configured for a name
func NewMgmtClusterRole(instance *mongodbv1.MongodbOperator) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
					"batch",
					"apps",
					"core",
				},
				Resources: []string{
					"jobs",
					"pods",
					"statefulsets",
					"persistentvolumes",
				},
				Verbs: []string{
					"create",
					"delete",
					"get",
					"list",
					"patch",
					"update",
					"watch",
				},
			},
		},
	}
}

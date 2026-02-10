package baseresource

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mongodbv1 "proton-mongodb-operator/api/v1"
)

// NewMgmtServiceAccount returns a ServiceAccount object
func NewMgmtServiceAccount(instance *mongodbv1.MongodbOperator) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
			Namespace: instance.GetNamespace(),
		},
	}
}

package baseresource

import (
	"fmt"
	mongodbv1 "proton-mongodb-operator/api/v1"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaim returns a Persistent Volume Claims for Mongod pod
func NewMongoPersistentVolumeClaim(instance *mongodbv1.MongodbOperator) []*corev1.PersistentVolumeClaim {
	scName := ""
	storageClassname := &scName
	pvcs := []*corev1.PersistentVolumeClaim{}
	if instance.Spec.MongoDBSpec.Storage.StorageClassName != "" {
		return pvcs
	}
	var i int32
	quantity, _ := resource.ParseQuantity(instance.Spec.MongoDBSpec.Storage.Capacity)
	for i = 0; i < instance.Spec.MongoDBSpec.Replicas; i++ {
		spec := corev1.PersistentVolumeClaimSpec{
			StorageClassName: storageClassname,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
					"podindex":  strconv.Itoa(int(i)),
					"namespace": instance.Namespace,
				},
			},
		}
		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-%d", "mongodb-datadir", fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"), int(i)),
				Namespace: instance.Namespace,
			},
			Spec: spec,
		}
		pvcs = append(pvcs, pvc)
	}
	return pvcs
}

package baseresource

import (
	"fmt"
	mongodbv1 "proton-mongodb-operator/api/v1"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolume returns a Persistent Volume for Mongod pod
func NewMongoPersistentVolume(instance *mongodbv1.MongodbOperator) []*corev1.PersistentVolume {
	pvs := []*corev1.PersistentVolume{}
	if instance.Spec.MongoDBSpec.Storage.StorageClassName != "" {
		return pvs
	}
	var i int32
	quantity, _ := resource.ParseQuantity(instance.Spec.MongoDBSpec.Storage.Capacity)
	for i = 0; i < instance.Spec.MongoDBSpec.Replicas; i++ {
		spec := corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: quantity,
			},
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: instance.Spec.MongoDBSpec.Storage.VolumeSpec[i].Path,
				},
			},
			ClaimRef: &corev1.ObjectReference{
				Name:      fmt.Sprintf("%s-%s-%d", "mongodb-datadir", fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"), int(i)),
				Namespace: instance.GetNamespace(),
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{instance.Spec.MongoDBSpec.Storage.VolumeSpec[i].Host},
								},
							},
						},
					},
				},
			},
		}
		pv := &corev1.PersistentVolume{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PersistentVolume",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s-%d", instance.GetName(), fmt.Sprintf("%s-%s", "mongodb", instance.GetNamespace()), int(i)),
				Labels: map[string]string{
					"app":       fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
					"podindex":  strconv.Itoa(int(i)),
					"namespace": instance.Namespace,
				},
			},
			Spec: spec,
		}
		pvs = append(pvs, pv)
	}
	return pvs
}

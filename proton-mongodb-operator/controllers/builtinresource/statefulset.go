package baseresource

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	mongodbv1 "proton-mongodb-operator/api/v1"
)

// NewStatefulSet returns a StatefulSet object configured for a name
func NewMongoStatefulSet(instance *mongodbv1.MongodbOperator) *appsv1.StatefulSet {
	var fsgroup int64 = 1001
	quantity, _ := resource.ParseQuantity(instance.Spec.MongoDBSpec.Storage.Capacity)

	ls := map[string]string{"app": instance.GetName() + "-mongodb"}
	svcName := fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")
	podSecurityContext := &corev1.PodSecurityContext{
		FSGroup: &fsgroup,
	}
	volumes := []corev1.Volume{
		{
			Name: "mongodb-conf",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "mongodb.conf",
							Path: "mongodb.conf",
						},
					},
				},
			},
		},
	}
	if instance.Spec.MongoDBSpec.Mongodconf.TLS.Enabled {
		volumes = append(volumes, corev1.Volume{
			Name: "tls-file",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "mongo-tls-secret",
					Items: []corev1.KeyToPath{
						{Key: "ca", Path: "ca.pem"},
						{Key: "server-cert-key", Path: "server-cert-key.pem"},
					},
				}},
		})
	}
	volumemounts := []corev1.VolumeMount{
		{
			Name:      "mongodb-conf",
			MountPath: "/mongodb/mongoconfig",
		},
		{
			Name:      "mongodb-datadir",
			MountPath: "/data/mongodb_data",
		},
	}
	if instance.Spec.MongoDBSpec.Mongodconf.TLS.Enabled {
		volumemounts = append(volumemounts, corev1.VolumeMount{
			Name:      "tls-file",
			MountPath: "/mongodb/tls",
		})
	}
	containers := []corev1.Container{
		{
			Name:            "mongodb",
			Image:           instance.Spec.MongoDBSpec.Image,
			Args:            []string{"mongod", "--config", "/mongodb/mongoconfig/mongodb.conf", "--port", strconv.Itoa(28000), "--auth", "--keyFile", "/mongodb/config/mongodb.keyfile", "--replSet", instance.Spec.MongoDBSpec.Replset.Name},
			ImagePullPolicy: instance.Spec.MongoDBSpec.ImagePullPolicy,
			Ports: []corev1.ContainerPort{
				{
					Name:          "mongodb",
					ContainerPort: 28000,
				},
			},
			Env: []corev1.EnvVar{
				{
					Name: "MONGO_INITDB_ROOT_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: instance.Spec.SecretName,
							},
							Key: "username",
						},
					},
				},
				{
					Name: "MONGO_INITDB_ROOT_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: instance.Spec.SecretName,
							},
							Key: "password",
						},
					},
				},
				{
					Name:  "MONGODB_PORT",
					Value: strconv.Itoa(28000),
				},
				{
					Name:  "TRACE",
					Value: instance.Spec.MongoDBSpec.Debug,
				},
				{
					Name:  "TZ",
					Value: "Asia/Shanghai",
				},
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{IntVal: 28000},
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       5,
				TimeoutSeconds:      2,
				FailureThreshold:    10,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/healthcheck.sh",
							"--readiness",
						},
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       5,
				TimeoutSeconds:      2,
				FailureThreshold:    100,
			},
			VolumeMounts: volumemounts,
			Resources:    instance.Spec.MongoDBSpec.Resources,
		},
	}
	var selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"namespace": instance.Namespace,
		},
	}
	if instance.Spec.MongoDBSpec.Storage.StorageClassName != "" {
		selector = nil
	}
	mongosts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
			Namespace: instance.GetNamespace(),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         svcName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &instance.Spec.MongoDBSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					SecurityContext: podSecurityContext,
					RestartPolicy:   corev1.RestartPolicyAlways,
					Containers:      containers,
					Volumes:         volumes,
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mongodb-datadir",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &instance.Spec.MongoDBSpec.Storage.StorageClassName,
						AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Selector:         selector,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: quantity,
							},
						},
					},
				},
			},
		},
	}
	return mongosts
}

// NewStatefulSet returns a StatefulSet object configured for a name
func NewMgmtStatefulSet(instance *mongodbv1.MongodbOperator) *appsv1.StatefulSet {
	var fsgroup int64 = 1001
	ls := map[string]string{"app": instance.GetName() + "-mgmt"}
	svcName := fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt")
	podSecurityContext := &corev1.PodSecurityContext{
		FSGroup: &fsgroup,
	}
	volumes := []corev1.Volume{
		{
			Name: "mongodb-mgmt-conf",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "mongodb-mgmt.yaml",
							Path: "mongodb-mgmt.yaml",
						},
					},
				},
			},
		},
	}
	containers := []corev1.Container{
		{
			Name:            "mgmt",
			Image:           instance.Spec.MgmtSpec.Image,
			ImagePullPolicy: instance.Spec.MongoDBSpec.ImagePullPolicy,
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
				{
					Name:  "MONGODB_PORT",
					Value: strconv.Itoa(28000),
				},
				{
					Name:  "MONGO_NAME",
					Value: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
				},
				{
					Name:  "MONGO_SVC_NAME",
					Value: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
				},
				{
					Name:  "NAMESPACE",
					Value: instance.Namespace,
				},
				{
					Name: "MONGO_ROOT_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: instance.Spec.SecretName,
							},
							Key: "username",
						},
					},
				},
				{
					Name: "MONGO_ROOT_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: instance.Spec.SecretName,
							},
							Key: "password",
						},
					},
				},
				{
					Name:  "TZ",
					Value: "Asia/Shanghai",
				},
				{
					Name:  "CR_NAME",
					Value: instance.Name,
				},
				{
					Name:  "CR_NAMESPACE",
					Value: instance.Namespace,
				},
				{
					Name:  "IMAGE",
					Value: instance.Spec.MongoDBSpec.Image,
				},
				{
					Name: "BACKUP_HOST",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "spec.nodeName",
						},
					},
				},
				{
					Name:  "MONGO_SECRET_NAME",
					Value: instance.Spec.SecretName,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "mongodb-mgmt-conf",
					MountPath: "/etc/mongodb-mgmt-conf/",
				},
			},
		},
	}
	mgmtsts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
			Namespace: instance.GetNamespace(),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         svcName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &instance.Spec.MongoDBSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
					SecurityContext:    podSecurityContext,
					// Affinity:        PodAffinityF(instance, instance.Spec.MongoDBSpec.PodAffinity, ls),
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers:    containers,
					Volumes:       volumes,
				},
			},
		},
	}
	return mgmtsts
}

// NewStatefulSet returns a StatefulSet object configured for a name
func NewExporterStatefulSet(instance *mongodbv1.MongodbOperator) *appsv1.StatefulSet {
	ls := map[string]string{"app": instance.GetName() + "-mongodb-exporter"}
	svcName := fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-exporter")
	podSecurityContext := &corev1.PodSecurityContext{
		//TO ADD
	}
	volumes := []corev1.Volume{
		{
			Name: "mongodb-conf",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "mongodb.conf",
							Path: "mongodb.conf",
						},
					},
				},
			},
		},
		{
			Name: "mongodb-mgmt-conf",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "mongodb-mgmt.yaml",
							Path: "mongodb-mgmt.yaml",
						},
					},
				},
			},
		},
	}
	containers := []corev1.Container{
		{
			Name:            "exporter",
			Image:           instance.Spec.ExporterSpec.Image,
			ImagePullPolicy: instance.Spec.MongoDBSpec.ImagePullPolicy,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: 9216,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
				{
					Name:  "MONGO_PORT",
					Value: strconv.Itoa(28000),
				},
				{
					Name:  "MONGO_NAME",
					Value: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
				},
				{
					Name:  "MONGO_SVC_NAME",
					Value: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
				},
				{
					Name:  "NAMESPACE",
					Value: instance.Namespace,
				},
				{
					Name: "MONGO_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: instance.Spec.SecretName,
							},
							Key: "username",
						},
					},
				},
				{
					Name: "MONGO_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: instance.Spec.SecretName,
							},
							Key: "password",
						},
					},
				},
				{
					Name:  "TZ",
					Value: "Asia/Shanghai",
				},
			},
		},
	}
	exportersts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-exporter"),
			Namespace: instance.GetNamespace(),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         svcName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &instance.Spec.MongoDBSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					SecurityContext: podSecurityContext,
					// Affinity:        PodAffinityF(instance, instance.Spec.MongoDBSpec.PodAffinity, ls),
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers:    containers,
					Volumes:       volumes,
				},
			},
		},
	}
	return exportersts
}

// PodAffinity returns podAffinity options for the pod
// func PodAffinityF(instance *databaseprotonv1beta1.MykindCluster, af *databaseprotonv1beta1.PodAffinity, labels map[string]string) *corev1.Affinity {
// 	if af == nil {
// 		return nil
// 	}

// 	labelsCopy := make(map[string]string)
// 	for k, v := range labels {
// 		labelsCopy[k] = v
// 	}

// 	switch {
// 	case af.Advanced != nil:
// 		return af.Advanced
// 	case af.TopologyKey != nil:
// 		if *af.TopologyKey == databaseprotonv1beta1.AffinityOff {
// 			return nil
// 		}

// 		return &corev1.Affinity{
// 			PodAntiAffinity: &corev1.PodAntiAffinity{
// 				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
// 					{
// 						LabelSelector: &metav1.LabelSelector{
// 							MatchLabels: labelsCopy,
// 						},
// 						TopologyKey: *af.TopologyKey,
// 					},
// 				},
// 			},
// 		}
// 	}

// 	return nil
// }

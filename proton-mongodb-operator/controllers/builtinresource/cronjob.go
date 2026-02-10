package baseresource

import (
	"fmt"
	mongodbv1 "proton-mongodb-operator/api/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewMongoCronjob returns a Cronjob object configured for a mongo
func NewMongoCronjob(instance *mongodbv1.MongodbOperator) []*batchv1.CronJob {
	var i int32
	var SuccessfulJobsHistoryLimit int32 = 1
	var ActiveDeadlineSeconds int64 = 90
	var BackoffLimit int32 = 0
	cjbs := []*batchv1.CronJob{}
	for i = 0; i < instance.Spec.MongoDBSpec.Replicas; i++ {
		cjb := &batchv1.CronJob{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "CronJob",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", "logrotate-cron", int(i)),
				Namespace: instance.Namespace,
			},
			Spec: batchv1.CronJobSpec{
				Schedule:                   instance.Spec.LogrotateSpec.Schedule,
				SuccessfulJobsHistoryLimit: &SuccessfulJobsHistoryLimit,
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						ActiveDeadlineSeconds: &ActiveDeadlineSeconds,
						BackoffLimit:          &BackoffLimit,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Affinity: &corev1.Affinity{
									PodAffinity: &corev1.PodAffinity{
										RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
											{
												LabelSelector: &metav1.LabelSelector{
													MatchExpressions: []metav1.LabelSelectorRequirement{
														{
															Key:      "statefulset.kubernetes.io/pod-name",
															Operator: metav1.LabelSelectorOpIn,
															Values:   []string{fmt.Sprintf("%s-%s-%d", instance.GetName(), "mongodb", int(i))},
														},
													},
												},
												TopologyKey: "kubernetes.io/hostname",
											},
										},
									},
								},
								Containers: []corev1.Container{
									{
										Name:  "logrotate",
										Image: instance.Spec.LogrotateSpec.Image,
										Env: []corev1.EnvVar{
											{
												Name:  "MONGODB_POD_HOSTNAME",
												Value: fmt.Sprintf("%s-mongodb-%d.%s-mongodb.%s.svc.cluster.local:%d", instance.Name, i, instance.Name, instance.Namespace, 28000),
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
										},
										Args: []string{
											"1001",
											"mongodb",
											"1001",
											"mongodb",
											"mongodb-logrotate",
										},
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      fmt.Sprintf("%s-%d", "mongodblogdir", int(i)),
												MountPath: instance.Spec.MongoDBSpec.Storage.VolumeSpec[i].Path,
											}, {
												Name:      "config",
												MountPath: "/etc/logrotate.d/",
											},
										},
									},
								},
								RestartPolicy: corev1.RestartPolicyNever,
								Volumes: []corev1.Volume{
									{
										Name: "config",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
												},
												Items: []corev1.KeyToPath{
													{
														Key:  fmt.Sprintf("%s-%d", "mongodb-logrotate", int(i)),
														Path: "mongodb-logrotate",
													},
												},
											},
										},
									},
									{
										Name: fmt.Sprintf("%s-%d", "mongodblogdir", int(i)),
										VolumeSource: corev1.VolumeSource{
											PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
												ClaimName: fmt.Sprintf("%s-%s-%d", "mongodb-datadir", fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"), int(i)),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		cjbs = append(cjbs, cjb)
	}

	return cjbs
}

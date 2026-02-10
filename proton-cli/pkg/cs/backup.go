package cs

import (
	"context"
	"fmt"
	"path"

	"github.com/sirupsen/logrus"
	k8s_batch_v1 "k8s.io/api/batch/v1"
	k8s_core_v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_client_batch_v1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/utils/ptr"
)

// BackupCronJobNamePrefix 是 Kubernetes 中每个 Control Plane 节点的定时备份任务名称的前缀
const BackupCronJobNamePrefix = "proton-cs-backup-"

// BackupCronJobSchedule 是 Proton CS 备份任务的时间表。任务由
// kube-controller-manager 发起，所以任务执行时间根据 kube-controller-manager 的
// 时区计算。因为 kube-controller-manager运行在容器中，所以时区为 UTC，所以
// Asia/Shanghai 的 02:30 约等于 UTC 的 18:30。 const BackupCronJobSchedule =
// "30 18 * * *"
const BackupCronJobSchedule = "30 18 * * *"

// BackupCronJobStartingDeadlineSeconds 是备份任务可以执行的最长时间
const BackupCronJobStartingDeadlineSeconds = 3600

// BackupDirectory 是存放 Proton CS 备份的目录
const BackupDirectory = "/var/lib/proton-cs/backup"

// BackupLimit 是 Proton CS 备份保存数量的上限
const BackupLimit = 7

// BackupPrefix 是 Proton CS 备份文件名的前缀
const BackupPrefix = ""

// BackupImageRepository 是备份工具镜像在 docker 仓库中的路径
const BackupImageRepository = "proton/proton-cs-backup"

// BackupImageTag 是备份工具镜像的 tag
const BackupImageTag = "v1.1.3"

// EnsureBackupCronJobForNode 确保节点的定时备份任务存在且配置正确。
func EnsureBackupCronJobForNode(ctx context.Context, c k8s_client_batch_v1.CronJobInterface, node, registry string, log *logrus.Logger) (err error) {
	// 定时备份任务对象
	var cj = backupCronJobTemplate.DeepCopy()
	// 渲染定时备份任务模板
	cj.Name = BackupCronJobNamePrefix + node
	cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image = backupImage(registry, BackupImageRepository, BackupImageTag)
	cj.Spec.JobTemplate.Spec.Template.Spec.NodeName = node

	log.Debugf("create cronjob %s", cj.Name)
	_, err = c.Create(ctx, cj, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		log.Debugf("update cronjob %s due to already existing", cj.Name)
		_, err = c.Update(ctx, cj, metav1.UpdateOptions{})
	}
	return
}

// 定时任务模板
var backupCronJobTemplate = &k8s_batch_v1.CronJob{
	Spec: k8s_batch_v1.CronJobSpec{
		Schedule:                BackupCronJobSchedule,
		StartingDeadlineSeconds: ptr.To(int64(BackupCronJobStartingDeadlineSeconds)),
		ConcurrencyPolicy:       k8s_batch_v1.ForbidConcurrent,
		JobTemplate: k8s_batch_v1.JobTemplateSpec{
			Spec: k8s_batch_v1.JobSpec{
				Template: k8s_core_v1.PodTemplateSpec{
					Spec: k8s_core_v1.PodSpec{
						Volumes: []k8s_core_v1.Volume{
							{
								Name: "backup-directory",
								VolumeSource: k8s_core_v1.VolumeSource{
									HostPath: &k8s_core_v1.HostPathVolumeSource{
										Path: BackupDirectory,
									},
								},
							},
							{
								Name: "k8s-dir",
								VolumeSource: k8s_core_v1.VolumeSource{
									HostPath: &k8s_core_v1.HostPathVolumeSource{
										Path: "/etc/kubernetes",
									},
								},
							},
							{
								Name: "kubelet-run-directory",
								VolumeSource: k8s_core_v1.VolumeSource{
									HostPath: &k8s_core_v1.HostPathVolumeSource{
										Path: "/var/lib/kubelet",
									},
								},
							},
							// 节点的时区文件是一个符号连接，需要挂载完整的 tzdata 才可以使用。
							{
								Name: "zoneinfo",
								VolumeSource: k8s_core_v1.VolumeSource{
									HostPath: &k8s_core_v1.HostPathVolumeSource{
										Path: "/usr/share/zoneinfo",
									},
								},
							},
							// 备份文件的名称包含备份时间。为了让时间是节点的本地时间，所以需要挂载节点的时区文件。
							{
								Name: "localtime",
								VolumeSource: k8s_core_v1.VolumeSource{
									HostPath: &k8s_core_v1.HostPathVolumeSource{
										Path: "/etc/localtime",
									},
								},
							},
						},
						Containers: []k8s_core_v1.Container{
							{
								Name: "backup",
								Command: []string{
									"proton-cs-backup",
									"backup",
									fmt.Sprintf("--limit=%d", BackupLimit),
									"--directory=/var/lib/proton-cs/backup",
									fmt.Sprintf("--prefix=%s", BackupPrefix),
								},
								VolumeMounts: []k8s_core_v1.VolumeMount{
									{
										Name:      "backup-directory",
										MountPath: "/var/lib/proton-cs/backup",
									},
									{
										Name:      "k8s-dir",
										MountPath: "/etc/kubernetes",
									},
									{
										Name:      "kubelet-run-directory",
										MountPath: "/var/lib/kubelet",
									},
									{
										Name:      "zoneinfo",
										MountPath: "/usr/share/zoneinfo",
									},
									{
										Name:      "localtime",
										MountPath: "/etc/localtime",
									},
								},
							},
						},
						RestartPolicy: k8s_core_v1.RestartPolicyOnFailure,
						HostNetwork:   true,
					},
				},
			},
		},
	},
}

// NewBackupCronJobForNode 生成指定节点创建定时备份任务对象
func NewBackupCronJobForNode(node, registry string) *k8s_batch_v1.CronJob {
	return &k8s_batch_v1.CronJob{
		ObjectMeta: k8s_meta_v1.ObjectMeta{
			Name: BackupCronJobNamePrefix + node,
		},
		Spec: k8s_batch_v1.CronJobSpec{
			Schedule:                BackupCronJobSchedule,
			StartingDeadlineSeconds: ptr.To(int64(BackupCronJobStartingDeadlineSeconds)),
			ConcurrencyPolicy:       k8s_batch_v1.ForbidConcurrent,
			JobTemplate: k8s_batch_v1.JobTemplateSpec{
				Spec: k8s_batch_v1.JobSpec{
					Template: k8s_core_v1.PodTemplateSpec{
						Spec: k8s_core_v1.PodSpec{
							Volumes: []k8s_core_v1.Volume{
								{
									Name: "backup-directory",
									VolumeSource: k8s_core_v1.VolumeSource{
										HostPath: &k8s_core_v1.HostPathVolumeSource{
											Path: BackupDirectory,
										},
									},
								},
								{
									Name: "k8s-dir",
									VolumeSource: k8s_core_v1.VolumeSource{
										HostPath: &k8s_core_v1.HostPathVolumeSource{
											Path: "/etc/kubernetes",
										},
									},
								},
								{
									Name: "kubelet-run-directory",
									VolumeSource: k8s_core_v1.VolumeSource{
										HostPath: &k8s_core_v1.HostPathVolumeSource{
											Path: "/var/lib/kubelet",
										},
									},
								},
								// 节点的时区文件是一个符号连接，需要挂载完整的 tzdata 才可以使用。
								{
									Name: "zoneinfo",
									VolumeSource: k8s_core_v1.VolumeSource{
										HostPath: &k8s_core_v1.HostPathVolumeSource{
											Path: "/usr/share/zoneinfo",
										},
									},
								},
								// 备份文件的名称包含备份时间。为了让时间是节点的本地时间，所以需要挂载节点的时区文件。
								{
									Name: "localtime",
									VolumeSource: k8s_core_v1.VolumeSource{
										HostPath: &k8s_core_v1.HostPathVolumeSource{
											Path: "/etc/localtime",
										},
									},
								},
							},
							Containers: []k8s_core_v1.Container{
								{
									Name:  "backup",
									Image: backupImage(registry, BackupImageRepository, BackupImageTag),
									Command: []string{
										"proton-cs-backup",
										"backup",
										fmt.Sprintf("--limit=%d", BackupLimit),
										"--directory=/var/lib/proton-cs/backup",
										fmt.Sprintf("--prefix=%s", BackupPrefix),
									},
									VolumeMounts: []k8s_core_v1.VolumeMount{
										{
											Name:      "backup-directory",
											MountPath: "/var/lib/proton-cs/backup",
										},
										{
											Name:      "k8s-dir",
											MountPath: "/etc/kubernetes",
										},
										{
											Name:      "kubelet-run-directory",
											MountPath: "/var/lib/kubelet",
										},
										{
											Name:      "zoneinfo",
											MountPath: "/usr/share/zoneinfo",
										},
										{
											Name:      "localtime",
											MountPath: "/etc/localtime",
										},
									},
								},
							},
							RestartPolicy: k8s_core_v1.RestartPolicyOnFailure,
							NodeName:      node,
							HostNetwork:   true,
						},
					},
				},
			},
		},
	}
}

// backupImage 生成备份工具镜像的名称。 (e.g.
// registry.aishu.cn/proton/proton-cs-backup:v1.1.2)
func backupImage(registry, repository, tag string) string {
	var image string
	image = path.Join(registry, repository)
	if tag != "" {
		image = image + ":" + tag
	}
	return image
}

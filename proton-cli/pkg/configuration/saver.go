package configuration

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
)

const (
	// ProtonCLIConfigSecretName specifies in what Secret in the proton namespace the `proton-cli apply` configuration should be stored
	ProtonCLIConfigSecretName = "proton-cli-config"
	// ClusterConfigurationConfigMapKey specifies in what ConfigMap key the cluster configuration should be stored
	ClusterConfigurationConfigMapKey = "ClusterConfiguration"
)

// UploadToKubernetes 上传配置到 Kubernetes.
func UploadToKubernetes(ctx context.Context, cfg *ClusterConfig, k kubernetes.Interface, namespace ...string) error {
	d, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("unable to encode ClusterConfig: %w", err)
	}

	// If namespace is specified, use it; otherwise, use the default from file
	protonNamespace := GetProtonCliConfigNSFromFile()
	if len(namespace) > 0 && namespace[0] != "" {
		protonNamespace = namespace[0]
	}
	// TODO: 创建命名空间 proton 应该在 proton-cli 配置 CS/Kubernetes 的方法中实现，而非在上传配置这个操作中实现
	//  如果是权限问题 proton 命名空间 或已存在，则创建失败，如果有权限但是未找到则创建
	if _, err := k.CoreV1().Namespaces().Get(ctx, protonNamespace, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := k.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: protonNamespace,
				},
			}, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("unable to create proton namespace: %w", err)
			}
		} else {
			fmt.Printf("unable to get proton namespace: %v, continue create proton-cli-config secret\n", err)
		}
	}

	if _, err := k.CoreV1().Secrets(protonNamespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProtonCLIConfigSecretName,
			Namespace: protonNamespace,
		},
		Data: map[string][]byte{
			ClusterConfigurationConfigMapKey: d,
		},
	}, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("unable to create proton-cli-config secret: %w", err)
		}
		old, err := k.CoreV1().Secrets(protonNamespace).Get(ctx, ProtonCLIConfigSecretName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("unable to get old proton-cli-config secret: %w", err)
		}

		old.Data[ClusterConfigurationConfigMapKey] = d
		if _, err := k.CoreV1().Secrets(protonNamespace).Update(ctx, old, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("unable to update proton-cli-config secret: %w", err)
		}
	}
	return nil
}

func SaveToBytes(cfg *ClusterConfig) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// oldProtonCLIDir 是旧版本 proton-cli 保存配置的目录。在 1.3 版本删除已存在的目录。在 1.4 版本不再处理。
const oldProtonCLIDir = "/etc/proton-cli"

// RemoveOldProtonCLIDirIfExist 删除旧版本 proton-cli 保存配置的目录，忽略错误 fs.ErrNotExist
func RemoveOldProtonCLIDirIfExist(f files.Interface) error {
	var ctx = context.TODO()
	info, err := f.Stat(ctx, oldProtonCLIDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%v should be directory", oldProtonCLIDir)
	}

	return f.Delete(ctx, oldProtonCLIDir)
}

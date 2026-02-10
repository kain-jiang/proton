package tiller

import (
	"context"
	"fmt"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"

	"k8s.io/client-go/kubernetes"
)

const (
	KubernetesLabelNodeRoleOldControlPlane = "node-role.kubernetes.io/master"
)

const (
	// tiller 所在的命名空间
	Namespace = api_meta_v1.NamespaceSystem
	// helm 客户端访问 tiller 所用的端口
	Port = 44134
	// tiller 提供 probe 和 metrics 的端口
	HTTPPort = 44135
)

// tiller 的 kubernetes 资源的标签集
var TillerLabels = labels.Set{
	"app":  "helm",
	"name": "tiller",
}

// Reconcile 创建 tiller 的 kubernetes 资源，如果已存在则无操作
//
//  1. ServiceAccount
//  2. ClusterRoleBinding
//  3. DaemonSet
//  4. Service
func Reconcile(ctx context.Context, client kubernetes.Interface, registry string) error {
	log := logger.NewLogger()

	log.Print("create tiller service account")
	if _, err := client.CoreV1().ServiceAccounts(Namespace).Create(ctx, &ServiceAccount, api_meta_v1.CreateOptions{}); err != nil && !api_errors.IsAlreadyExists(err) {
		return fmt.Errorf("create tiller service account fail: %w", err)
	}
	log.Print("create tiller cluster role binding")
	if _, err := client.RbacV1().ClusterRoleBindings().Create(ctx, &ClusterRoleBinding, api_meta_v1.CreateOptions{}); err != nil && !api_errors.IsAlreadyExists(err) {
		return fmt.Errorf("create tiller cluster role binding fail: %w", err)
	}
	log.Print("create tiller daemon set")
	if _, err := client.AppsV1().DaemonSets(Namespace).Create(ctx, NewDaemonSet(registry), api_meta_v1.CreateOptions{}); err != nil && !api_errors.IsAlreadyExists(err) {
		return fmt.Errorf("create tiller daemon set fail: %w", err)
	}
	log.Print("create tiller service")
	if _, err := client.CoreV1().Services(Namespace).Create(ctx, &Service, api_meta_v1.CreateOptions{}); err != nil && !api_errors.IsAlreadyExists(err) {
		return fmt.Errorf("create tiller service fail: %w", err)
	}
	return nil
}

// RemoveTiller 删除 tiller 的 kubernetes 资源，如果不存在则不操作
func RemoveTiller(ctx context.Context, client kubernetes.Interface) error {
	log := logger.NewLogger()
	for _, arg := range []struct {
		op       DeleteFunc
		resource string
		name     string
	}{
		{
			op:       client.CoreV1().Services(Namespace).Delete,
			resource: "tiller service",
			name:     ServiceName,
		},
		{
			op:       client.AppsV1().DaemonSets(Namespace).Delete,
			resource: "tiller daemon set",
			name:     DaemonSetName,
		},
		{
			op:       client.RbacV1().ClusterRoleBindings().Delete,
			resource: "tiller cluster role binding",
			name:     ClusterRoleBindingName,
		},
		{
			op:       client.CoreV1().ServiceAccounts(Namespace).Delete,
			resource: "tiller service account",
			name:     ServiceAccountName,
		},
	} {
		if err := deleteKubernetesResource(ctx, arg.op, arg.resource, arg.name); err != nil {
			log.Errorf("delete %v: %v fail: %v", arg.resource, arg.name, err)
			return fmt.Errorf("delete %v: %v fail: %w", arg.resource, arg.name, err)
		}
	}
	return nil
}

type DeleteFunc func(ctx context.Context, name string, opts api_meta_v1.DeleteOptions) error

// deleteKubernetesResource 删除 kubernetes 资源,  `resource` 资源的可读的名称
func deleteKubernetesResource(ctx context.Context, fn DeleteFunc, resource, name string) error {
	log := logger.NewLogger()
	log.Infof("delete %v", resource)
	err := fn(ctx, name, api_meta_v1.DeleteOptions{})
	if api_errors.IsNotFound(err) {
		log.Debugf("%v is not found", resource)
		err = nil
	}
	return err
}

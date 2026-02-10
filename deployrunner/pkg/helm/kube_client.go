package helm

import (
	"context"
	"fmt"
	"time"

	deploymentutil "taskrunner/pkg/helm/internal"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
)

// rewrite helm client's kube client for wait check

type kubeClient struct {
	*kube.Client
}

func newkubeClient(kcli *kube.Client) *kubeClient {
	return &kubeClient{
		Client: kcli,
	}
}

// Wait waits up to the given timeout for the specified resources to be ready.
func (c *kubeClient) Wait(resources kube.ResourceList, timeout time.Duration) error {
	kcli, err := c.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}
	c.Log("beginning wait for %d resources with timeout of %v", len(resources), timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rc := ReadyChecker{
		ReadyChecker: kube.NewReadyChecker(kcli, c.Log, kube.PausedAsReady(true)),
		log:          c.Log,
		client:       kcli,
	}

	return wait.PollUntilContextCancel(ctx, 10*time.Second, true, func(_ context.Context) (done bool, err error) {
		for _, v := range resources {
			ready, err := rc.IsReady(ctx, v)
			if !ready || err != nil {
				return false, err
			}
		}
		return true, nil
	})
}

// WatchUntilReady watches the resources given and waits until it is ready.
//
// This method is mainly for hook implementations. It watches for a resource to
// hit a particular milestone. The milestone depends on the Kind.
//
// For most kinds, it checks to see if the resource is marked as Added or Modified
// by the Kubernetes event stream. For some kinds, it does more:
//
//   - Jobs: A job is marked "Ready" when it has successfully completed. This is
//     ascertained by watching the Status fields in a job's output.
//   - Pods: A pod is marked "Ready" when it has successfully completed. This is
//     ascertained by watching the status.phase field in a pod's output.
//
// Handling for other kinds will be added as necessary.
func (c *kubeClient) WatchUntilReady(resources kube.ResourceList, timeout time.Duration) error {
	// For jobs, there's also the option to do poll c.Jobs(namespace).Get():
	// https://github.com/adamreese/kubernetes/blob/master/test/e2e/job.go#L291-L300
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.LogWatchUntilReady(ctx, resources)
	err := c.Client.WatchUntilReady(resources, timeout)
	return err
}

// LogWatchUntilReady helm 原生的实现中基于k8s watcher对job和POD等对象进行状态监听，从而判断是否成功。
// 这会导致在这过程中几乎不会生成等待日志输出，从而让人迷惑是否仍在执行。
// 因此在此处增加一个粗糙的日志行为,以提示日志查看者仍有等待任务在执行
func (c *kubeClient) LogWatchUntilReady(ctx context.Context, resources kube.ResourceList) {
	msg := []string{}
	for _, info := range resources {
		kind := info.Mapping.GroupVersionKind.Kind
		msg = append(msg, fmt.Sprintf("wait resource %s, %s/%s into ready", kind, info.Namespace, info.Name))
	}

	for {
		select {
		case <-ctx.Done():
			for _, info := range resources {
				kind := info.Mapping.GroupVersionKind.Kind
				c.Log(fmt.Sprintf("wait resource %s, %s/%s  into ready finish", kind, info.Namespace, info.Name))
			}
			return
		default:
			for _, m := range msg {
				c.Log(m)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// ReadyChecker over write  kube.ReadyChecker's interface
type ReadyChecker struct {
	kube.ReadyChecker
	log    func(string, ...interface{})
	client *kubernetes.Clientset
}

func (c *ReadyChecker) IsReady(ctx context.Context, v *resource.Info) (bool, error) {
	var (
		// This defaults to true, otherwise we get to a point where
		// things will always return false unless one of the objects
		// that manages pods has been hit
		ok  = true
		err error
	)
	switch kube.AsVersioned(v).(type) {
	case *appsv1.Deployment, *appsv1beta1.Deployment, *appsv1beta2.Deployment, *extensionsv1beta1.Deployment:
		currentDeployment, err := c.client.AppsV1().Deployments(v.Namespace).Get(ctx, v.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		// If paused deployment will never be ready
		if currentDeployment.Spec.Paused {
			return true, nil
		}
		// Find RS associated with deployment
		newReplicaSet, err := deploymentutil.GetNewReplicaSet(currentDeployment, c.client.AppsV1())
		if err != nil || newReplicaSet == nil {
			return false, err
		}
		if !c.deploymentReady(newReplicaSet, currentDeployment) {
			return false, nil
		}
	case *extensionsv1beta1.DaemonSet, *appsv1.DaemonSet, *appsv1beta2.DaemonSet:
		ds, err := c.client.AppsV1().DaemonSets(v.Namespace).Get(ctx, v.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if !c.daemonSetReady(ds) {
			return false, nil
		}
	default:
		ok, err = c.ReadyChecker.IsReady(ctx, v)
	}
	if !ok || err != nil {
		return false, err
	}
	return true, nil
}

func (c *ReadyChecker) deploymentReady(rs *appsv1.ReplicaSet, dep *appsv1.Deployment) bool {
	expectedReady := *dep.Spec.Replicas
	if !(rs.Status.ReadyReplicas >= expectedReady) {
		c.log("Deployment is not ready: %s/%s. %d out of %d expected pods are ready", dep.Namespace, dep.Name, rs.Status.ReadyReplicas, expectedReady)
		return false
	}
	return true
}

func (c *ReadyChecker) daemonSetReady(ds *appsv1.DaemonSet) bool {
	// If the update strategy is not a rolling update, there will be nothing to wait for
	if ds.Spec.UpdateStrategy.Type != appsv1.RollingUpdateDaemonSetStrategyType {
		return true
	}

	// Make sure all the updated pods have been scheduled
	if ds.Status.UpdatedNumberScheduled != ds.Status.DesiredNumberScheduled {
		c.log("DaemonSet is not ready: %s/%s. %d out of %d expected pods have been scheduled", ds.Namespace, ds.Name, ds.Status.UpdatedNumberScheduled, ds.Status.DesiredNumberScheduled)
		return false
	}

	expectedReady := int(ds.Status.DesiredNumberScheduled)
	if !(int(ds.Status.NumberReady) >= expectedReady) {
		c.log("DaemonSet is not ready: %s/%s. %d out of %d expected pods are ready", ds.Namespace, ds.Name, ds.Status.NumberReady, expectedReady)
		return false
	}
	return true
}

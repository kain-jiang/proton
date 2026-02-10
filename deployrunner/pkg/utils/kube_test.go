package utils

import (
	"context"
	"testing"

	"taskrunner/trait"

	"bou.ke/monkey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewSeretRE(t *testing.T) {
	ctx := context.Background()
	g := monkey.Patch(NewKubeclient, func() (kubernetes.Interface, *trait.Error) {
		return fake.NewClientset(), nil
	})
	defer g.Unpatch()
	cli, err := NewSecretRW("qwe", "deploy-core", "deploy-core")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, _ = cli.Kcli.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cli.Namespace,
		},
	}, metav1.CreateOptions{})
	err = cli.SetContent(ctx, map[string]string{"qweqwe": "qweqw"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	v := map[string]string{}
	err = cli.GetFullConf(ctx, &v)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

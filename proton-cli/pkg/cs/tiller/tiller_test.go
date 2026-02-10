package tiller

import (
	"context"
	"errors"
	"testing"

	api_apps_v1 "k8s.io/api/apps/v1"
	api_core_v1 "k8s.io/api/core/v1"
	api_rbac_v1 "k8s.io/api/rbac/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	client_fake "k8s.io/client-go/kubernetes/fake"
	client_testing "k8s.io/client-go/testing"
)

func TestReconcile(t *testing.T) {
	type args struct {
		ctx      context.Context
		client   *client_fake.Clientset
		registry string
	}
	tests := []struct {
		name     string
		args     args
		sa       *api_core_v1.ServiceAccount
		crb      *api_rbac_v1.ClusterRoleBinding
		ds       *api_apps_v1.DaemonSet
		svc      *api_core_v1.Service
		reactors []client_testing.SimpleReactor
		wantErr  bool
	}{
		{
			name: "create",
			args: args{
				ctx:      context.Background(),
				client:   &client_fake.Clientset{},
				registry: "registry.aishu.cn:15000",
			},
		},
		{
			name: "already-exists",
			args: args{
				ctx:      context.Background(),
				client:   &client_fake.Clientset{},
				registry: "registry.aishu.cn:15000",
			},
			reactors: []client_testing.SimpleReactor{
				{
					Verb:     "create",
					Resource: "serviceaccounts",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewAlreadyExists(action.GetResource().GroupResource(), ServiceAccountName)
					},
				},
				{
					Verb:     "create",
					Resource: "clusterrolebindings",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewAlreadyExists(action.GetResource().GroupResource(), ClusterRoleBindingName)
					},
				},
				{
					Verb:     "create",
					Resource: "daemonsets",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewAlreadyExists(action.GetResource().GroupResource(), DaemonSetName)
					},
				},
				{
					Verb:     "create",
					Resource: "services",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewAlreadyExists(action.GetResource().GroupResource(), ServiceName)
					},
				},
			},
		},
		{
			name: "create-service-account-fail",
			args: args{
				ctx:    context.Background(),
				client: &client_fake.Clientset{},
			},
			reactors: []client_testing.SimpleReactor{
				{
					Verb:     "create",
					Resource: "serviceaccounts",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewInternalError(errors.New("something wrong"))
					},
				},
			},
			wantErr: true,
		},
		{
			name: "create-cluster-role-binding-fail",
			args: args{
				ctx:    context.Background(),
				client: &client_fake.Clientset{},
			},
			reactors: []client_testing.SimpleReactor{
				{
					Verb:     "create",
					Resource: "clusterrolebindings",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewInternalError(errors.New("something wrong"))
					},
				},
			},
			wantErr: true,
		},
		{
			name: "create-daemon-set-fail",
			args: args{
				ctx:    context.Background(),
				client: &client_fake.Clientset{},
			},
			reactors: []client_testing.SimpleReactor{
				{
					Verb:     "create",
					Resource: "daemonsets",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewInternalError(errors.New("something wrong"))
					},
				},
			},
			wantErr: true,
		},
		{
			name: "create-service-fail",
			args: args{
				ctx:    context.Background(),
				client: &client_fake.Clientset{},
			},
			reactors: []client_testing.SimpleReactor{
				{
					Verb:     "create",
					Resource: "services",
					Reaction: func(action client_testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, api_errors.NewInternalError(errors.New("something wrong"))
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, r := range tt.reactors {
				tt.args.client.AddReactor(r.Verb, r.Resource, r.Reaction)
			}
			if err := Reconcile(tt.args.ctx, tt.args.client, tt.args.registry); (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

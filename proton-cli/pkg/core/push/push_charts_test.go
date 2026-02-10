package push

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

func TestPushCharts(t *testing.T) {
	type args struct {
		opts ChartPushOpts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "HelmRepoSpecified",
			args: args{opts: ChartPushOpts{
				HelmRepo:  "https://test.helm.repo/chartrepo/test",
				ChartsDir: "/path/to/charts/dir",
			}},
		},
		{
			name: "K8sClientSetNil",
			args: args{opts: ChartPushOpts{
				ChartsDir: "/path/to/charts/dir",
			}},
			wantErr: true,
		},
	}
	crPushChartsPatcher := gomonkey.ApplyMethod(reflect.TypeOf(&cr.Cr{}), "PushCharts", func(_ *cr.Cr, _ string) error {
		log.Info("Patch method cr.PushCharts")
		return nil
	})
	defer crPushChartsPatcher.Reset()
	newKClientPatcher := gomonkey.ApplyFunc(client.NewK8sClient, func() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
		return nil, nil
	})
	defer newKClientPatcher.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PushCharts(tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("PushCharts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

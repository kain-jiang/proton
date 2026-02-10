package helm3

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
)

func Test_helmv3_Install(t *testing.T) {
	chartBytes, _ := testdata.ReadFile("testdata/demo-0.1.0.tgz")
	chartFile := filepath.Join(os.TempDir(), "demo-0.1.0.tgz")
	_ = os.WriteFile(chartFile, chartBytes, 0o666)
	ch, _ := loader.LoadFile(chartFile)
	defer os.Remove(chartFile)

	type fields struct {
		actionConfig *action.Configuration
		settings     *helmcli.EnvSettings
		namespace    string
		log          *logrus.Entry
	}
	type args struct {
		release string
		chart   *chart.Chart
		opts    []InstallOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1_install_success",
			fields: fields{
				actionConfig: actionConfigFixture(t),
				settings:     helmcli.New(),
				namespace:    "resource",
				log:          logrus.WithField("type", "testing"),
			},
			args: args{
				release: "demo",
				chart:   ch,
				opts: []InstallOption{
					WithInstallSkipCRDs(true),
					WithInstallValues(M{}),
					WithInstallDryRun(true),
					WithInstallWait(false, 0),
					WithInstallForce(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &helmv3{
				actionConfig: tt.fields.actionConfig,
				settings:     tt.fields.settings,
				namespace:    tt.fields.namespace,
				log:          tt.fields.log,
			}
			if _, err := c.Install(tt.args.release, tt.args.chart, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

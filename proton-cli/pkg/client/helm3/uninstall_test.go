package helm3

import (
	"testing"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	helmcli "helm.sh/helm/v3/pkg/cli"
)

func Test_helmv3_Uninstall(t *testing.T) {
	type fields struct {
		actionConfig *action.Configuration
		settings     *helmcli.EnvSettings
		namespace    string
		log          *logrus.Entry
	}
	type args struct {
		release string
		opts    []UninstallOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1_success",
			fields: fields{
				actionConfig: actionConfigFixture(t),
				settings:     helmcli.New(),
				namespace:    "resource",
				log:          logrus.WithField("type", "testing"),
			},
			args: args{
				release: "demo",
				opts: []UninstallOption{
					WithUninstallIgnoreNotFound(true),
					WithUninstallKeepHistory(false),
					WithUninstallDryRun(true),
				},
			},
		},
		{
			name: "2_exist_failed",
			fields: fields{
				actionConfig: actionConfigFixture(t),
				settings:     helmcli.New(),
				namespace:    "resource",
				log:          logrus.WithField("type", "testing"),
			},
			args: args{
				release: "demo",
				opts: []UninstallOption{
					WithUninstallIgnoreNotFound(false),
					WithUninstallKeepHistory(false),
					WithUninstallDryRun(true),
				},
			},
			wantErr: true,
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
			if err := c.Uninstall(tt.args.release, tt.args.opts...); (err != nil) != tt.wantErr {
				t.Errorf("Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package mq

import (
	"testing"

	fclient "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3/testing"

	"github.com/agiledragon/gomonkey"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestMQManager_apply(t *testing.T) {
	tests := []struct {
		name      string
		m         *MQManager
		installed map[string]string
		wantErr   bool
	}{
		{
			name: "upgrade-hosts-one-to-three",
			m: &MQManager{
				spec: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path",
				},
				registry:       "registry.example.org",
				servicePackage: "path/to/service-package",
				charts: servicepackage.Charts{
					{Metadata: chart.Metadata{
						Name:    "proton-mq-nsq",
						Version: "1.0.0",
					}},
				},
				oldConf: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
				},
			},
			installed: map[string]string{
				"proton-mq-nsq": "1.0.0",
			},
		},
		{
			name: "install-one-hosts",
			m: &MQManager{
				spec: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
					Data_path: "/data/path",
				},
				registry:       "registry.example.org",
				servicePackage: "path/to/service-package",
				charts: servicepackage.Charts{
					{Metadata: chart.Metadata{
						Name:    "proton-mq-nsq",
						Version: "1.0.0",
					}},
				},
			},
			installed: map[string]string{
				"proton-mq-nsq": "1.0.0",
			},
		},
		{
			name: "chart-not-exist",
			m: &MQManager{
				spec: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
					Data_path: "/data/path",
				},
				registry:       "registry.example.org",
				servicePackage: "path/to/service-package",
				charts: servicepackage.Charts{
					{Metadata: chart.Metadata{}},
				},
			},
			installed: map[string]string{
				"proton-mq-nsq": "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "upgrade-version",
			m: &MQManager{
				spec: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
					Data_path: "/data/path",
				},
				registry:       "registry.example.org",
				servicePackage: "path/to/service-package",
				charts: servicepackage.Charts{
					{Metadata: chart.Metadata{
						Name:    "proton-mq-nsq",
						Version: "1.0.1",
					}},
				},
				oldConf: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
					Data_path: "/data/path",
				},
			},
			installed: map[string]string{
				"proton-mq-nsq": "1.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// tt.m.Charts = tt.charts
			fhelm := fclient.New("", t.Logf)
			for n, v := range tt.installed {
				_ = fhelm.Storage.Create(&release.Release{
					Name: n,
					Info: &release.Info{},
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name:    n,
							Version: v,
						},
					},
				})
			}
			tt.m.helm3 = fhelm

			if err := tt.m.apply(); (err != nil) != tt.wantErr {
				t.Errorf("nsq.apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestMQManager_Reset(t *testing.T) {
	type fields struct {
		spec           *configuration.ProtonDataConf
		hosts          []configuration.Node
		registry       string
		helm3          helm3.Client
		servicePackage string
		charts         servicepackage.Charts
		oldConf        *configuration.ProtonDataConf
	}
	tests := []struct {
		name      string
		fields    fields
		clearData bool
		wantErr   bool
	}{
		{
			name: "successForNotClearData",
			fields: fields{
				hosts: []configuration.Node{{IP4: "1.1.1.1"}},
				spec:  &configuration.ProtonDataConf{Data_path: ""},
			},
			clearData: false,
			wantErr:   false,
		},
		{
			name: "successForClearData",
			fields: fields{
				hosts: []configuration.Node{{IP4: "1.1.1.1"}},
				spec:  &configuration.ProtonDataConf{Data_path: "ut/data/path"},
			},
			clearData: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearDataFuncPatcher := gomonkey.ApplyFunc(universal.ClearDataDir, func(host, dataPath string) error {
				log.Info("mock for func[ClearDataDir].")
				return nil
			})
			defer clearDataFuncPatcher.Reset()
			clearDataGlobalVarPatcher := gomonkey.ApplyGlobalVar(&global.ClearData, tt.clearData)
			defer clearDataGlobalVarPatcher.Reset()
			m := &MQManager{
				spec:           tt.fields.spec,
				hosts:          tt.fields.hosts,
				registry:       tt.fields.registry,
				helm3:          tt.fields.helm3,
				servicePackage: tt.fields.servicePackage,
				charts:         tt.fields.charts,
				oldConf:        tt.fields.oldConf,
			}
			if err := m.Reset(); (err != nil) != tt.wantErr {
				t.Errorf("MQManager.Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

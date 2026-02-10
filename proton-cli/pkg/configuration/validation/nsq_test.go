package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateMQNSQ(t *testing.T) {
	type args struct {
		m           *configuration.ProtonDataConf
		nodeNameSet sets.Set[string]
		fldPath     *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				m: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
		},
		{
			name: "invalid-host-undefined",
			args: args{
				m: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-x",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-too-many-hosts",
			args: args{
				m: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
						"node-3",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-data-path-missing",
			args: args{
				m: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "needn't-sort-hosts",
			args: args{
				m: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-2",
						"node-1",
						"node-0",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateMQNSQ(tt.args.m, tt.args.nodeNameSet, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}
			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateMQNSQ() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateMQNSQUpdate(t *testing.T) {
	type args struct {
		o       *configuration.ProtonDataConf
		n       *configuration.ProtonDataConf
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid-none",
			args: args{
				o: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
				n: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
			},
		},
		{
			name: "valid-hosts-expansion",
			args: args{
				o: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
					},
					Data_path: "/data/path/0",
				},
				n: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
			},
		},
		{
			name: "invalid-hosts-reduce",
			args: args{
				o: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
				n: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-1",
					},
					Data_path: "/data/path/0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-data-path",
			args: args{
				o: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
				n: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/1",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-old-host-not-front",
			args: args{
				o: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-2",
					},
					Data_path: "/data/path/0",
				},
				n: &configuration.ProtonDataConf{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid storage class name is changed",
			args: args{
				o: &configuration.ProtonDataConf{
					StorageClassName: "storage-class-old",
				},
				n: &configuration.ProtonDataConf{
					StorageClassName: "storage-class-new",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateMQNSQUpdate(tt.args.o, tt.args.n, tt.args.fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				t.Errorf("ValidateMQNSQUpdate() len(errList) = %v, wantErr %v", len(errList), tt.wantErr)
				for i, err := range errList {
					t.Errorf("ValidateMQNSQUpdate() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}

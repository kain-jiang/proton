package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateMongodb(t *testing.T) {
	type args struct {
		m           *configuration.ProtonDB
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
				m: &configuration.ProtonDB{
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
				m: &configuration.ProtonDB{
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
				m: &configuration.ProtonDB{
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
				m: &configuration.ProtonDB{
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
				m: &configuration.ProtonDB{
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
			errs := ValidateMongodb(tt.args.m, tt.args.nodeNameSet, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}

			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateMongodb() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateMongodbUpdate(t *testing.T) {
	type args struct {
		o       *configuration.ProtonDB
		n       *configuration.ProtonDB
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
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
		},
		{
			name: "valid-version",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
		},
		{
			name: "valid-hosts-expansion",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
		},
		{
			name: "invalid-hosts-reduce",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-1",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-Admin-user",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME1",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-Admin-password",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD_CHANGED",
					Data_path:    "/data/path/0",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-data-path",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
				},
				n: &configuration.ProtonDB{
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
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-1",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
			wantErr: true,
		},
		{
			name: "valid-sort-host",
			args: args{
				o: &configuration.ProtonDB{
					Hosts: []string{
						"node-2",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
				n: &configuration.ProtonDB{
					Hosts: []string{
						"node-2",
						"node-0",
						"node-1",
					},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
					Data_path:    "/data/path/0",
				},
			},
		},
		{
			name: "invalid storage class name is changed",
			args: args{
				o: &configuration.ProtonDB{
					StorageClassName: "storage-class-old",
				},
				n: &configuration.ProtonDB{
					StorageClassName: "storage-class-new",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateMongodbUpdate(tt.args.o, tt.args.n, tt.args.fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				t.Errorf("ValidateMongodbUpdate() len(errList) = %v, wantErr %v", len(errList), tt.wantErr)
				for i, err := range errList {
					t.Errorf("ValidateMongodbUpdate() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}

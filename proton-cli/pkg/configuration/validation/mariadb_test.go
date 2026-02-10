package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateMariaDB(t *testing.T) {
	type args struct {
		m           *configuration.ProtonMariaDB
		nodeNameSet sets.Set[string]
		fldPath     *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "local-data-path",
			args: args{
				m: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
					},
					Config: &configuration.ProtonMariaDBConfigs{
						Resource_requests_memory: "8G",
						Resource_limits_memory:   "9G",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
				),
			},
		},
		{
			name: "storage-class",
			args: args{
				m: &configuration.ProtonMariaDB{
					StorageClassName: "standard",
					Config: &configuration.ProtonMariaDBConfigs{
						Resource_requests_memory: "8G",
						Resource_limits_memory:   "9G",
					},
				},
			},
		},
		{
			name: "both-storage-class-and-hosts",
			args: args{
				m: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
					},
					Config: &configuration.ProtonMariaDBConfigs{
						Resource_requests_memory: "8G",
						Resource_limits_memory:   "9G",
					},
					StorageClassName: "standard",
				},
				nodeNameSet: sets.New[string](
					"node-0",
				),
			},
			wantErr: true,
		},
		{
			name: "both-storage-class-and-data-path",
			args: args{
				m: &configuration.ProtonMariaDB{
					Config: &configuration.ProtonMariaDBConfigs{
						Resource_requests_memory: "8G",
						Resource_limits_memory:   "9G",
					},
					Data_path:        "/data/path",
					StorageClassName: "standard",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if allErrs := ValidateMariaDB(tt.args.m, tt.args.nodeNameSet, tt.args.fldPath); len(allErrs) > 1 || (allErrs != nil) != tt.wantErr {
				t.Errorf("ValidateMariaDB() len(allErrs) = %v, wantErr %v", len(allErrs), tt.wantErr)
				for i, err := range allErrs {
					t.Errorf("ValidateMariaDB() allErrs[%d] = %v", i, err)
				}
			}
		})
	}
}

func TestValidateMariaDBConfig(t *testing.T) {
	type args struct {
		c       *configuration.ProtonMariaDBConfigs
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					LowerCaseTableNames:      ptr.To(0),
					Resource_requests_memory: "4G",
					Resource_limits_memory:   "5G",
				},
			},
		},
		{
			name: "valid/lower_case_table_names=nil",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					LowerCaseTableNames:      nil,
					Resource_requests_memory: "4G",
					Resource_limits_memory:   "5G",
				},
			},
		},
		{
			name: "valid/lower_case_table_names=1",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					LowerCaseTableNames:      ptr.To(1),
					Resource_requests_memory: "4G",
					Resource_limits_memory:   "5G",
				},
			},
		},
		{
			name: "valid/lower_case_table_names=2",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					LowerCaseTableNames:      ptr.To(2),
					Resource_requests_memory: "4G",
					Resource_limits_memory:   "5G",
				},
			},
		},
		{
			name: "invalid-request-memory",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					Resource_requests_memory: "1.x",
					Resource_limits_memory:   "5G",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-limit-memory",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					Resource_requests_memory: "1G",
					Resource_limits_memory:   "300si",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-lower_case_table_names=3",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					LowerCaseTableNames:      ptr.To(3),
					Resource_requests_memory: "4G",
					Resource_limits_memory:   "5G",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-lower_case_table_names=-1",
			args: args{
				c: &configuration.ProtonMariaDBConfigs{
					LowerCaseTableNames:      ptr.To(-1),
					Resource_requests_memory: "4G",
					Resource_limits_memory:   "5G",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if allErrs := ValidateMariaDBConfig(tt.args.c, tt.args.fldPath); len(allErrs) > 1 || (allErrs != nil) != tt.wantErr {
				t.Errorf("ValidateMariaDBConfig() len(allErrs) = %v, wantErr %v", len(allErrs), tt.wantErr)
				for i, err := range allErrs {
					t.Errorf("ValidateMariaDBConfig() allErrs[%d] = %v", i, err)
				}
			}
		})
	}
}

func TestValidateMariaDBUpdate(t *testing.T) {
	type args struct {
		o *configuration.ProtonMariaDB
		n *configuration.ProtonMariaDB
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "capacity-expansion",
			args: args{
				o: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
					},
				},
				n: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
				},
			},
		},
		{
			name: "update-config",
			args: args{
				o: &configuration.ProtonMariaDB{
					Config: &configuration.ProtonMariaDBConfigs{
						Innodb_buffer_pool_size:  "1G",
						Resource_requests_memory: "1G",
						Resource_limits_memory:   "1G",
					},
				},
				n: &configuration.ProtonMariaDB{
					Config: &configuration.ProtonMariaDBConfigs{
						Innodb_buffer_pool_size:  "2G",
						Resource_requests_memory: "2G",
						Resource_limits_memory:   "2G",
					},
				},
			},
		},
		{
			name: "update-data-path",
			args: args{
				o: &configuration.ProtonMariaDB{
					Data_path: "/var/lib/mariadb-old",
				},
				n: &configuration.ProtonMariaDB{
					Data_path: "/var/lib/mariadb-new",
				},
			},
			wantErr: true,
		},
		{
			name: "update-storage-class-name",
			args: args{
				o: &configuration.ProtonMariaDB{
					StorageClassName: "storage-class-name-old",
				},
				n: &configuration.ProtonMariaDB{
					StorageClassName: "storage-class-name-new",
				},
			},
			wantErr: true,
		},
		{
			name: "hosts-scale-down",
			args: args{
				o: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
				},
				n: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "old-hosts-not-in-front",
			args: args{
				o: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-1",
					},
				},
				n: &configuration.ProtonMariaDB{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAllErrs := ValidateMariaDBUpdate(tt.args.o, tt.args.n, nil); len(gotAllErrs) > 1 || (gotAllErrs != nil) != tt.wantErr {
				t.Errorf("ValidateMariaDBUpdate() len(gotAllErrs) = %v, wantErr %v", len(gotAllErrs), tt.wantErr)
				for i, err := range gotAllErrs {
					t.Errorf("ValidateMariaDBUpdate() gotAllErrs[%d] = %v", i, err)
				}
			}
		})
	}
}

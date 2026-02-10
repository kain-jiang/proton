package helm

import (
	"testing"

	"github.com/go-test/deep"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func Test_depServicesFor(t *testing.T) {
	type args struct {
		rds      *configuration.RdsInfo
		database string
	}
	tests := []struct {
		name string
		args args
		want DepServices
	}{
		{
			name: "single host",
			args: args{
				rds: &configuration.RdsInfo{
					Hosts:    "mariadb-0.example.org",
					Port:     3306,
					Username: "example-username",
					Password: "example-password",
				},
				database: "example-database",
			},
			want: DepServices{
				RDS: RDS{
					Host:     "mariadb-0.example.org",
					Port:     3306,
					Username: "example-username",
					Password: "example-password",
					Database: "example-database",
				},
			},
		},
		{
			name: "multi hosts",
			args: args{
				rds: &configuration.RdsInfo{
					Hosts:    "mariadb-0.example.org,mariadb-1.example.org,mariadb-2.example.org",
					Port:     3306,
					Username: "example-username",
					Password: "example-password",
				},
				database: "example-database",
			},
			want: DepServices{
				RDS: RDS{
					Host:     "mariadb-0.example.org",
					Port:     3306,
					Username: "example-username",
					Password: "example-password",
					Database: "example-database",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := depServicesFor(tt.args.rds, tt.args.database)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("depServicesFor() got != want: %v", d)
			}
		})
	}
}

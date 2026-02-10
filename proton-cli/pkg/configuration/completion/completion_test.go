package completion

import (
	"testing"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestCompleteClusterConfig(t *testing.T) {
	type args struct {
		c   *configuration.ClusterConfig
		pkg *servicepackage.ServicePackage
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "complete-all",
			args: args{
				c: &configuration.ClusterConfig{
					Cs:        &configuration.Cs{},
					CMS:       &configuration.CMS{},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
				pkg: &servicepackage.ServicePackage{},
			},
		},
		{
			name: "non-cms",
			args: args{
				c: &configuration.ClusterConfig{
					Cs:        &configuration.Cs{},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
				pkg: &servicepackage.ServicePackage{},
			},
		},
		{
			name: "non-installer-service",
			args: args{
				c: &configuration.ClusterConfig{
					Cs:        &configuration.Cs{},
					CMS:       &configuration.CMS{},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
				pkg: &servicepackage.ServicePackage{},
			},
		},
		{
			name: "non-kafka",
			args: args{
				c: &configuration.ClusterConfig{
					Cs:        &configuration.Cs{},
					CMS:       &configuration.CMS{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
				pkg: &servicepackage.ServicePackage{},
			},
		},
		{
			name: "non-zookeeper",
			args: args{
				c: &configuration.ClusterConfig{
					Cs:    &configuration.Cs{},
					CMS:   &configuration.CMS{},
					Kafka: &configuration.Kafka{},
				},
				pkg: &servicepackage.ServicePackage{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteClusterConfig(tt.args.c, nil, tt.args.pkg)
		})
	}
}

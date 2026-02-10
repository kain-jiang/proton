package helm

import (
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// DepServices defines helm values's depServices of proton package store.
type DepServices struct {
	RDS RDS `json:"rds"`
}

// RDS defines depServices.rds of proton package store's helm values.
type RDS struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func depServicesFor(rds *configuration.RdsInfo, database string) DepServices {
	hosts := strings.Split(rds.Hosts, ",")
	return DepServices{
		RDS: RDS{
			Host:     hosts[0],
			Port:     rds.Port,
			Username: rds.Username,
			Password: rds.Password,
			Database: database,
		},
	}
}

package testpkg

import (
	"os"
	"strconv"
	"strings"

	"taskrunner/pkg/component/resources"
)

func GetTestMysqlWithType(rdsType string) *resources.RDS {
	cfg := resources.RDS{
		Port: 3306,
	}
	rdsType = strings.ToUpper(rdsType)
	cfg.Host = os.Getenv(rdsType + "TestTaskRunnerDataBaseAddr")
	hostport := os.Getenv(rdsType + "TestTaskRunnerDataBasePort")
	if hostport == "" {
		cfg.Port = 3306
	} else {
		port, err := strconv.Atoi(hostport)
		if err != nil {
			panic(err)
		}
		cfg.Port = float64(port)
	}
	cfg.User = os.Getenv(rdsType + "TestTaskRunnerDataBaseUser")
	cfg.Password = os.Getenv(rdsType + "TestTaskRunnerDataBasePassWord")
	cfg.Type = rdsType
	cfg.DBName = os.Getenv("TestTaskRunnerDataBaseName")
	if cfg.Host == "" {
		return nil
	}
	return &cfg
}

// GetTestMysql get mysql info from env, if not exist return nil
func GetTestMysql() *resources.RDS {
	dbtype := os.Getenv("TESTDBTYPE")
	if dbtype == "" {
		dbtype = "MARIADB"
	}
	return GetTestMysqlWithType(dbtype)
}

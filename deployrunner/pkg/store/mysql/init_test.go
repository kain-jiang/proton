package store_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"taskrunner/pkg/component/resources"
	store "taskrunner/pkg/store/mysql"
	"taskrunner/trait"
)

var _DataBaseType = "MARIADB"

type (
	DBOP  = store.DBOP
	Store = trait.Store
)

const _testInitSqlDir = "../../../sql-ddl"

func NewDBOP(ctx context.Context, rds resources.RDS) (DBOP, *trait.Error) {
	return store.NewDBOP(ctx, rds)
}

func NewStore(ctx context.Context, rds resources.RDS) (trait.Store, *trait.Error) {
	return store.NewStore(ctx, rds)
}

func TestMain(m *testing.M) {
	skipList := []string{
		// resources.KDB9DBType,
		// resources.MariadbDBType,
		// resources.MysqlDBType,
		// resources.DM8DBType,
	}
	for _, ds := range store.Structors {
		_DataBaseType = strings.ToUpper(ds.DBType)
		skip := false
		for _, i := range skipList {
			if _DataBaseType == strings.ToUpper(i) {
				skip = true
				break
			}
		}
		if !skip {
			fmt.Printf("Start test database %s\n", _DataBaseType)
			m.Run()
		}
	}
}

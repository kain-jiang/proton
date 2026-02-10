package store_test

import (
	"context"

	"taskrunner/pkg/component/resources"
	driver "taskrunner/pkg/sql-driver/driver"
	store "taskrunner/pkg/store/mysql/driver/mysql"
	"taskrunner/trait"
)

var _DataBaseType = "MARIADB"

const _testInitSqlDir = "../../../../../../sql-ddl"

type (
	DBOP  = driver.DBAdminOP
	Store = *store.Store
)

func NewDBOP(ctx context.Context, rds resources.RDS) (DBOP, *trait.Error) {
	return driver.Factory.NewDBOP(ctx, rds)
}

func NewStore(ctx context.Context, rds resources.RDS) (Store, *trait.Error) {
	return store.NewStore(ctx, rds)
}

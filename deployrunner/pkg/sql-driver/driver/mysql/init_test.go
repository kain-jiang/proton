package store_test

import (
	"context"

	"taskrunner/pkg/component/resources"
	store "taskrunner/pkg/sql-driver/driver/mysql"
	"taskrunner/trait"
)

var _DataBaseType = "MARIADB"

type (
	DBOP  = *store.DBOP
	Store = *store.Store
)

// nolint:unused
const _testDDLSQLDir = ""

func NewDBOP(ctx context.Context, rds resources.RDS) (DBOP, *trait.Error) {
	return store.NewDBOP(ctx, rds)
}

func NewStore(ctx context.Context, rds resources.RDS) (Store, *trait.Error) {
	return store.NewStore(ctx, rds)
}

package store

import (
	"context"

	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/sql-driver/driver"
	store "taskrunner/pkg/sql-driver/driver/mysql"
	"taskrunner/trait"

	stmt "taskrunner/pkg/store/mysql/stmt"
)

type DBOP struct {
	*store.DBOP
}

// CreateUser create the sql instance user
// 人大金仓创建用户方式未能从官网查到,暂时无法适配,返回nil不影响其他部分使用
func (db *DBOP) CreateUser(ctx context.Context, user, passwd string) *trait.Error {
	return nil
}

// GrantUserDB grant the sql instance user with database
// 人大金仓创建用户方式未能从官网查到,暂时无法适配,返回nil不影响其他部分使用
func (db *DBOP) GrantUserDB(ctx context.Context, user, dbName string) *trait.Error {
	return nil
}

// DeleteUser grant the sql instance user with database
// 人大金仓创建用户方式未能从官网查到,暂时无法适配,返回nil不影响其他部分使用
func (db *DBOP) DeleteUser(ctx context.Context, user string) *trait.Error {
	return nil
}

func (s *DBOP) CreateDatabase(ctx context.Context, name string) *trait.Error {
	err := s.DBOP.CreateDatabase(ctx, name)
	if trait.IsInternalError(err, trait.ErrUniqueKey) {
		err = nil
	}
	return err
}

// NewDBOP create a db operator for init
func NewDBOP(ctx context.Context, rds resources.RDS) (driver.DBAdminOP, *trait.Error) {
	dbName := rds.DBName
	rds.DBName = ""
	dbop, err := store.NewDBOP(ctx, rds)
	if err != nil {
		rds.DBName = dbName
		return nil, err
	}
	rds.DBName = dbName
	dbop.WithErrorWrraper(&driver.ErrorWrraper{
		ReaderError: readerErrorWrraper,
		WriterError: writerErrorWrraper,
	})
	dmStmts := stmt.NewKDB9Stmt()
	dbop.Stmt = dmStmts.MysqlStmt
	dbop.Stmt.CreateDataBaseStmt = dmStmts.KDB9Admin.CreateDataBaseStmt
	dbop.Stmt.DeleteDataBaseStmt = dmStmts.KDB9Admin.DeleteDataBaseStmt

	return &DBOP{DBOP: dbop}, nil
}

func NewDBConn(ctx context.Context, rds resources.RDS) (driver.DBConn, *trait.Error) {
	return store.NewDBConn(ctx, rds, &driver.ErrorWrraper{
		ReaderError: readerErrorWrraper,
		WriterError: writerErrorWrraper,
	})
}

// NewStore from dsn
func NewStore(ctx context.Context, rds resources.RDS) (*store.Store, *trait.Error) {
	db, err := NewDBConn(ctx, rds)
	if err != nil {
		return nil, err
	}
	s := store.NewStoreWithDBConn(ctx, db, rds)

	return s, nil
}

func init() {
	driver.Factory.Registry(resources.KDB9DBType, NewDBConn)
	driver.Factory.RegistryDBOP(resources.KDB9DBType, NewDBOP)
}

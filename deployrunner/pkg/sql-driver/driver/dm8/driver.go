package store

import (
	"context"
	"fmt"

	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/sql-driver/driver"
	store "taskrunner/pkg/sql-driver/driver/mysql"
	"taskrunner/trait"

	stmt "taskrunner/pkg/store/mysql/stmt"
)

type DBOP struct {
	*store.DBOP
}

// GrantUserDB grant the sql instance user with database
func (db *DBOP) GrantUserDB(ctx context.Context, user, dbName string) *trait.Error {
	// return db.DBOP.GrantUserDB(ctx, user, dbName)
	if user == db.DBOP.Cfg.User {
		return nil
	}
	// fmt.Printf("GRANT %s to %s;\n", db.Stmt.RwPrivileges, user)
	_, err := db.DBOP.ExecContext(ctx, fmt.Sprintf("GRANT %s to %s;", db.Stmt.RwPrivileges, user))
	return err
}

func (db *DBOP) DeleteUser(ctx context.Context, user string) *trait.Error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("REVOKE %s FROM %s", db.Stmt.RwPrivileges, user))
	// fmt.Printf(db.Stmt.RevokeUserDB+"\n", user)
	if err != nil {
		return err
	}
	// fmt.Printf(db.Stmt.DeleteUser+"\n", user)
	_, err = db.ExecContext(ctx, fmt.Sprintf(db.Stmt.DeleteUser, user))
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
	rds.DBName = ""
	dbop.WithErrorWrraper(&driver.ErrorWrraper{
		ReaderError: readerErrorWrraper,
		WriterError: writerErrorWrraper,
	})
	dmStmts := stmt.NewDM8Stmt()
	// dbop.Stmt = dmStmts.MysqlStmt
	dbop.Stmt.CreateDataBaseStmt = dmStmts.DM8Admin.CreateDataBaseStmt
	dbop.Stmt.DeleteDataBaseStmt = dmStmts.DM8Admin.DeleteDataBaseStmt
	dbop.Stmt.CreateUser = dmStmts.DM8Admin.CreateUser
	dbop.Stmt.DeleteUser = dmStmts.DM8Admin.DeleteUser
	dbop.Stmt.UpdateUser = dmStmts.DM8Admin.UpdateUser
	dbop.Stmt.GrantUserDB = dmStmts.DM8Admin.GrantUserDB
	dbop.Stmt.RwPrivileges = dmStmts.DM8Admin.RwPrivileges

	return &DBOP{DBOP: dbop}, nil
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

func NewDBConn(ctx context.Context, rds resources.RDS) (driver.DBConn, *trait.Error) {
	return store.NewDBConn(ctx, rds, &driver.ErrorWrraper{
		ReaderError: readerErrorWrraper,
		WriterError: writerErrorWrraper,
	})
}

func init() {
	driver.Factory.Registry(resources.DM8DBType, NewDBConn)
	driver.Factory.RegistryDBOP(resources.DM8DBType, NewDBOP)
}

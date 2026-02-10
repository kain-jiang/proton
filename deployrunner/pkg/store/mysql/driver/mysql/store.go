/*
Package store use mysql database
@Title store imply trait.store interface
@Description

	this package impy trait.Store interface.
	Store user can watch trait.Store for detail interface document.
	store use mysql as inner store.
	The sql query statment may need optimize.
	basic sql execute imply in cursor struct.
*/
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"taskrunner/pkg/component/resources"
	driver "taskrunner/pkg/sql-driver"
	"taskrunner/pkg/store/mysql/upgrade/executor"
	"taskrunner/pkg/store/mysql/upgrade/operator"
	"taskrunner/trait"

	. "taskrunner/pkg/sql-driver/driver"
	stmt "taskrunner/pkg/store/mysql/stmt"

	utrait "taskrunner/pkg/store/mysql/upgrade/trait"

	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
)

// // DBOP database operator use to init database.
// // need privileges to create user and
// // create database and grant privileges on database to user
// type DBOP struct {
// 	DBConn
// 	Cfg  resources.RDS
// 	Stmt stmt.MysqlStmt
// }

// func NewDBConnWithDB(ctx context.Context, rds resources.RDS, db DBConn) *DBOP {
// 	return &DBOP{DBConn: db, Cfg: deepcopy.Copy(rds).(resources.RDS), Stmt: newStmts(rds)}
// }

// // NewDBOP create a db operator for init
// func NewDBOP(ctx context.Context, rds resources.RDS) (*DBOP, *trait.Error) {
// 	dbName := rds.DBName
// 	rds.DBName = ""
// 	db, err := driver.Factory.NewDBConn(ctx, rds)
// 	if err != nil {
// 		rds.DBName = dbName
// 		return nil, err
// 	}
// 	rds.DBName = dbName
// 	return NewDBConnWithDB(ctx, rds, db), nil
// }

// Store imply trait with mysql as database
type Store struct {
	DBConn
	Cfg resources.RDS
	SQLCursor
}

func newStmts(rds resources.RDS) stmt.MysqlStmt {
	stmts := stmt.NewMysqlStmt()

	createDBOPtions := "CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
	if strings.ToUpper(rds.Type) == "MYSQL" {
		stmts.CreateUser = stmts.CreateMysqlUser
		stmts.UpdateUser = stmts.UpdateMysqlUser
	}

	if strings.ToUpper(rds.Type) == "GOLDENDB" {
		createDBOPtions = "CHARSET=utf8mb4 COLLATE=utf8mb4_bin"
	}

	stmts.CreateDataBaseStmt = fmt.Sprintf("%s %s", stmts.CreateDataBaseStmt, createDBOPtions)
	return stmts
}

func NewStoreWithDBConn(ctx context.Context, db DBConn, rds resources.RDS) *Store {
	stmts := newStmts(rds)
	return &Store{
		SQLCursor: NewSQLCurSor(db, &stmts),
		DBConn:    db,
		Cfg:       deepcopy.Copy(rds).(resources.RDS),
	}
}

// NewStore from dsn
func NewStore(ctx context.Context, rds resources.RDS) (*Store, *trait.Error) {
	db, err := driver.Factory.NewDBConn(ctx, rds)
	if err != nil {
		return nil, err
	}
	s := NewStoreWithDBConn(ctx, db, rds)
	return s, nil
}

func NewDBOP(ctx context.Context, rds resources.RDS) (driver.DBAdminOP, *trait.Error) {
	return driver.Factory.NewDBOP(ctx, rds)
}

// Begin imply trait.store return a transaction
func (s *Store) Begin(ctx context.Context) (trait.Transaction, *trait.Error) {
	tx, err := s.begin(ctx, nil)
	return tx, err
}

// Begin imply trait, start a transaction
func (s *Store) begin(ctx context.Context, opt *sql.TxOptions) (*TX, *trait.Error) {
	tx, err := s.DBConn.BeginTx(ctx, opt)
	return s.beginWithTx(tx), err
}

func (s *Store) beginWithTx(tx driver.Transaction) *TX {
	return &TX{
		Transaction: tx,
		SQLCursor:   NewSQLCurSor(tx, s.stmt),
	}
}

func (s *Store) InitTablesFromDir(ctx context.Context, root string) *trait.Error {
	excludeSvc := []string{}
	excludeSvc = append(excludeSvc, "taskruner-multi")
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	log.SetReportCaller(true)
	return executor.ExecuteDir(ctx, root, operator.ObjectStore{
		"default": s.DBConn,
	}, s.Cfg, executor.Option{
		Stage:  utrait.InitStageInt,
		Logger: log,
	}, excludeSvc...)
}

// Close close db store
func (s *Store) Close() *trait.Error {
	return s.DBConn.Close()
}

type SQLCursor struct {
	CursorConn
	stmt           *stmt.MysqlStmt
	CreateTableSet []string
}

func NewSQLCurSor(cur CursorConn, stmts *stmt.MysqlStmt) SQLCursor {
	c := SQLCursor{
		CursorConn: cur,
		stmt:       stmts,
	}
	return c
}

// TX impl trait.Transaction
type TX struct {
	// trait.Transaction
	SQLCursor
	// *sql.Tx
	Transaction
}

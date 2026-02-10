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
	"taskrunner/trait"

	. "taskrunner/pkg/sql-driver/driver"
	stmt "taskrunner/pkg/store/mysql/stmt"

	"github.com/mohae/deepcopy"
)

// DBOP database operator use to init database.
// need privileges to create user and
// create database and grant privileges on database to user
type DBOP struct {
	DBConn
	Cfg  resources.RDS
	Stmt stmt.MysqlStmt
}

func NewDBConnWithDB(ctx context.Context, rds resources.RDS, db DBConn) *DBOP {
	return &DBOP{DBConn: db, Cfg: deepcopy.Copy(rds).(resources.RDS), Stmt: newStmts(rds)}
}

// NewDBOP create a db operator for init
func NewDBOP(ctx context.Context, rds resources.RDS) (*DBOP, *trait.Error) {
	dbName := rds.DBName
	rds.DBName = ""
	db, err := NewDBConnSimple(ctx, rds)
	if err != nil {
		rds.DBName = dbName
		return nil, err
	}
	rds.DBName = dbName
	return NewDBConnWithDB(ctx, rds, db), nil
}

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
	db, err := NewDBConnSimple(ctx, rds)
	if err != nil {
		return nil, err
	}
	s := NewStoreWithDBConn(ctx, db, rds)
	return s, nil
}

// Begin imply trait.store return a transaction
func (s *Store) Begin(ctx context.Context) (Transaction, *trait.Error) {
	tx, err := s.begin(ctx, nil)
	return tx, err
}

// Begin imply trait, start a transaction
func (s *Store) begin(ctx context.Context, opt *sql.TxOptions) (*TX, *trait.Error) {
	tx, err := s.DBConn.BeginTx(ctx, opt)
	return &TX{
		// Tx: tx,
		Transaction: tx,
		SQLCursor:   NewSQLCurSor(tx, s.stmt),
	}, err
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

// func (c *SQLCursor) InitTables(ctx context.Context) *trait.Error {
// 	// 遍历结构体的字段
// 	for _, v := range c.CreateTableSet {
// 		if _, err := c.ExecContext(ctx, v); err != nil {
// 			err.Detail = v
// 			return err
// 		}
// 	}
// 	return nil
// }

// TX impl trait.Transaction
type TX struct {
	// trait.Transaction
	SQLCursor
	// *sql.Tx
	Transaction
}

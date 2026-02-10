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
	"os"

	"taskrunner/pkg/component/resources"
	"taskrunner/trait"

	_ "github.com/AISHU-Technology/proton-rds-sdk-go/driver"

	. "taskrunner/pkg/sql-driver/driver"
)

type result struct {
	sql.Result
	*ErrorWrraper
}

func (r *result) LastInsertId() (int64, *trait.Error) {
	num, err := r.Result.LastInsertId()
	return num, r.WriterError(err)
}

func (r *result) RowsAffected() (int64, *trait.Error) {
	num, err := r.Result.RowsAffected()
	return num, r.WriterError(err)
}

type row struct {
	*sql.Row
	*ErrorWrraper
}

func (rs *row) Scan(dest ...any) *trait.Error {
	return rs.WriterError(rs.Row.Scan(dest...))
}

type rows struct {
	*sql.Rows
	*ErrorWrraper
}

func (rs *rows) Scan(dest ...any) *trait.Error {
	return rs.ReaderError(rs.Rows.Scan(dest...))
}

func (rs *rows) Close() *trait.Error {
	return rs.WriterError(rs.Rows.Close())
}

type cursorConn struct {
	RawCursorConn
	*ErrorWrraper
}

func (db *cursorConn) WithErrorWrraper(w *ErrorWrraper) {
	db.ErrorWrraper = w
}

func (db *cursorConn) ExecContext(ctx context.Context, query string, args ...any) (Result, *trait.Error) {
	res, err := db.RawCursorConn.ExecContext(ctx, query, args...)
	return &result{Result: res, ErrorWrraper: db.ErrorWrraper}, db.WriterError(err)
}

func (db *cursorConn) QueryContext(ctx context.Context, query string, args ...any) (Rows, *trait.Error) {
	rs, err := db.RawCursorConn.QueryContext(ctx, query, args...)
	return &rows{Rows: rs, ErrorWrraper: db.ErrorWrraper}, db.ReaderError(err)
}

func (db *cursorConn) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	r := db.RawCursorConn.QueryRowContext(ctx, query, args...)
	return &row{Row: r, ErrorWrraper: db.ErrorWrraper}
}

func (db *cursorConn) PrepareContext(ctx context.Context, query string) (Stmt, *trait.Error) {
	res, err := db.RawCursorConn.PrepareContext(ctx, query)
	return &SQLStmt{Stmt: res, ErrorWrraper: db.ErrorWrraper}, db.WriterError(err)
}

type transaction struct {
	CursorConn
	tx *sql.Tx
	*ErrorWrraper
}

func (t *transaction) Commit() *trait.Error {
	return t.WriterError(t.tx.Commit())
}

func (t *transaction) Rollback() *trait.Error {
	return t.WriterError(t.tx.Rollback())
}

type dbConn struct {
	CursorConn
	DB *sql.DB
	*ErrorWrraper
}

func (db *dbConn) WithErrorWrraper(w *ErrorWrraper) {
	db.ErrorWrraper = w
	db.CursorConn.WithErrorWrraper(w)
}

func (db *dbConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, *trait.Error) {
	res, err := db.DB.BeginTx(ctx, opts)
	return &transaction{tx: res, CursorConn: &cursorConn{RawCursorConn: res, ErrorWrraper: db.ErrorWrraper}, ErrorWrraper: db.ErrorWrraper}, db.WriterError(err)
}

func (db *dbConn) Close() *trait.Error {
	return db.WriterError(db.DB.Close())
}

func NewDBConnSimple(ctx context.Context, rds resources.RDS) (*dbConn, *trait.Error) {
	w := &ErrorWrraper{
		ReaderError: readerErrorWrraper,
		WriterError: writerErrorWrraper,
	}
	return NewDBConn(ctx, rds, w)
}

func NewDBConn(_ context.Context, rds resources.RDS, w *ErrorWrraper) (*dbConn, *trait.Error) {
	if rds.Port == 0 {
		rds.Port = 3306
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_bin&&multiStatements=true",
		rds.User, rds.Password, rds.Host, int(rds.Port), rds.DBName)
	os.Setenv("DB_TYPE", rds.Type)
	db, err := sql.Open("proton-rds", dsn)
	if err != nil {
		return nil, w.WriterError(err)
	}
	return &dbConn{
		DB: db,
		CursorConn: &cursorConn{
			RawCursorConn: db,
			ErrorWrraper:  w,
		},
		ErrorWrraper: w,
	}, nil
}

type SQLStmt struct {
	*sql.Stmt
	*ErrorWrraper
}

func (s *SQLStmt) ExecContext(ctx context.Context, args ...any) (Result, *trait.Error) {
	res, err := s.Stmt.ExecContext(ctx, args...)
	return &result{Result: res, ErrorWrraper: s.ErrorWrraper}, s.WriterError(err)
}

func (s *SQLStmt) Close() *trait.Error {
	return s.WriterError(s.Stmt.Close())
}

// --------------------------------------- //

func init() {
	fb := func(ctx context.Context, rds resources.RDS) (DBConn, *trait.Error) {
		return NewDBConnSimple(ctx, rds)
	}
	fbop := func(ctx context.Context, rds resources.RDS) (DBAdminOP, *trait.Error) {
		return NewDBOP(ctx, rds)
	}
	alias := []string{
		resources.MysqlDBType,
		resources.MariadbDBType,
		resources.TIDBType,
		resources.GOLDENDBType,
	}
	for _, i := range alias {
		Factory.Registry(i, fb)
		Factory.RegistryDBOP(i, fbop)
	}
}

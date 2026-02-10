package driver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"taskrunner/pkg/component/resources"
	"taskrunner/trait"

	_ "github.com/AISHU-Technology/proton-rds-sdk-go/driver"
)

type DBAdminOP interface {
	CreateUser(ctx context.Context, user, passwd string) *trait.Error
	DeleteUser(ctx context.Context, user string) *trait.Error
	GrantUserDB(ctx context.Context, user, dbName string) *trait.Error
	DeleteDatabase(ctx context.Context, dbName string) *trait.Error
	CreateDatabase(ctx context.Context, dbName string) *trait.Error
	Close() *trait.Error
	// trait.Store
	DBConn
}

// ErrorWrraper use for set db driver's special wrraper. if nil, use default
type ErrorWrraper struct {
	WriterError func(error) *trait.Error
	ReaderError func(error) *trait.Error
}

type DBConn interface {
	Close() *trait.Error
	WithErrorWrraper(*ErrorWrraper)
	BeginTx(context.Context, *sql.TxOptions) (Transaction, *trait.Error)
	CursorConn
}

type CursorConn interface {
	ExecContext(ctx context.Context, query string, args ...any) (Result, *trait.Error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, *trait.Error)
	QueryRowContext(ctx context.Context, query string, args ...any) Row
	PrepareContext(ctx context.Context, query string) (Stmt, *trait.Error)
	WithErrorWrraper(w *ErrorWrraper)
}

type Transaction interface {
	CursorConn
	Commit() *trait.Error
	Rollback() *trait.Error
}

type RawCursorConn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (Result, *trait.Error)
	Close() *trait.Error
}

type Result interface {
	// LastInsertId() (int64, *trait.Error)
	RowsAffected() (int64, *trait.Error)
}

type Row interface {
	Scan(dest ...any) *trait.Error
}

type Rows interface {
	Scan(dest ...any) *trait.Error
	Next() bool
	Close() *trait.Error
}

// ConvertDBType 将数据库类型转换为驱动类型
func ConvertDBType(dtype string) string {
	if dtype == "" {
		dtype = "MARIADB"
	}
	dtype = strings.ToUpper(dtype)
	if dtype == "MYSQL" || dtype == "GOLDENDB" || dtype == "TIDB" {
		dtype = "MARIADB"
	}
	return dtype
}

type DriverFactor struct {
	inner map[string]func(context.Context, resources.RDS) (DBConn, *trait.Error)
	dbop  map[string]func(context.Context, resources.RDS) (DBAdminOP, *trait.Error)
}

func (f *DriverFactor) Registry(name string, fb func(context.Context, resources.RDS) (DBConn, *trait.Error)) {
	name = strings.ToUpper(name)
	if _, ok := f.inner[name]; ok {
		panic(fmt.Sprintf("%s has been regsitried", name))
	}
	f.inner[name] = fb
}

func (f *DriverFactor) RegistryDBOP(name string, fb func(context.Context, resources.RDS) (DBAdminOP, *trait.Error)) {
	name = strings.ToUpper(name)
	if _, ok := f.dbop[name]; ok {
		panic(fmt.Sprintf("%s db admin op has been regsitried", name))
	}
	f.dbop[name] = fb
}

func (f *DriverFactor) NewDBConn(ctx context.Context, rds resources.RDS) (DBConn, *trait.Error) {
	name := ConvertDBType(rds.Type)
	fb := f.inner[name]
	if fb == nil {
		return nil, &trait.Error{
			Internal: trait.ErrNotFound,
			Detail:   fmt.Sprintf("the rds driver with type [%s] not found ", rds.Type),
		}
	}
	return fb(ctx, rds)
}

func (f *DriverFactor) NewDBOP(ctx context.Context, rds resources.RDS) (DBAdminOP, *trait.Error) {
	name := ConvertDBType(rds.Type)

	fb := f.dbop[name]
	if fb == nil {
		return nil, &trait.Error{
			Internal: trait.ErrNotFound,
			Detail:   fmt.Sprintf("the rds driver with type [%s] not found ", rds.Type),
		}
	}
	return fb(ctx, rds)
}

// Factory 数据库驱动实现工厂
var Factory *DriverFactor

func init() {
	Factory = &DriverFactor{
		inner: make(map[string]func(context.Context, resources.RDS) (DBConn, *trait.Error)),
		dbop:  make(map[string]func(context.Context, resources.RDS) (DBAdminOP, *trait.Error)),
	}
}

// StoreTransactionErrorMarco wrap transaction rollback or commit
func StoreTransactionErrorMarco(tx Transaction, err *trait.Error) *trait.Error {
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func StoreTransactionMarco(s DBConn, f func(Transaction) *trait.Error) *trait.Error {
	tx, err := s.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	err = f(tx)
	return StoreTransactionErrorMarco(tx, err)
}

package driver

import (
	"taskrunner/pkg/sql-driver/driver"
	_ "taskrunner/pkg/sql-driver/driver/dm8"
	_ "taskrunner/pkg/sql-driver/driver/kdb9"
	_ "taskrunner/pkg/sql-driver/driver/mysql"
)

var Factory = driver.Factory

type (
	DBAdminOP     = driver.DBAdminOP
	DBConn        = driver.DBConn
	CursorConn    = driver.CursorConn
	Transaction   = driver.Transaction
	RawCursorConn = driver.RawCursorConn
	Stmt          = driver.Stmt
	Result        = driver.Result
	Row           = driver.Row
	Rows          = driver.Rows
)

func ConvertDBType(dtype string) string {
	return driver.ConvertDBType(dtype)
}

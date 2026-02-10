package store

import (
	"database/sql"
	"fmt"

	"taskrunner/trait"

	errcode "github.com/VividCortex/mysqlerr"
	mysql "github.com/go-sql-driver/mysql"

	etrait "taskrunner/error/codes/trait"
)

func unWrapInternalError(err error) *trait.Error {
	if err0, ok := err.(*trait.Error); ok {
		return err0
	}
	return nil
}

func readerErrorWrraper(err error) *trait.Error {
	if err == nil {
		return nil
	}
	ierr := unWrapInternalError(err)
	switch {
	case ierr != nil:
		return ierr
	case err == sql.ErrNoRows:
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("not found")}
	default:
		return &trait.Error{Internal: trait.ECSQLUnknow, Err: err, Detail: "unknow error for sql database"}
	}
}

func writerErrorWrraper(err error) *trait.Error {
	if err == nil {
		return nil
	}
	switch {
	case isMysqlError(err, errcode.ER_DATA_TOO_LONG):
		return &trait.Error{
			Err:      err,
			Internal: trait.ErrParam,
			Detail:   err.Error(),
		}
	case isMysqlError(err, errcode.ER_DUP_ENTRY):
		return &trait.Error{
			Err:      err,
			Internal: trait.ErrUniqueKey,
			Detail:   err.Error(),
		}
	case isMysqlError(err, errcode.ER_DUP_FIELDNAME):
		return &trait.Error{
			Err:      err,
			Internal: etrait.ECColumnExists,
		}
	case isMysqlError(err, errcode.ER_MULTIPLE_PRI_KEY):
		return &trait.Error{
			Err:      err,
			Internal: etrait.ECPriKeyExists,
		}
	default:
		return readerErrorWrraper(err)
	}
}

func isMysqlError(err error, enum int) bool {
	is := false
	if err0, ok := err.(*mysql.MySQLError); ok {
		is = err0.Number == uint16(enum)
	}
	return is
}

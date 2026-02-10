package store

import (
	"database/sql"
	"fmt"

	etrait "taskrunner/error/codes/trait"
	"taskrunner/trait"

	"gitee.com/chunanyong/dm"
)

func readerErrorWrraper(err error) *trait.Error {
	if err == nil {
		return nil
	}
	ierr, ok := err.(*trait.Error)
	if ok {
		if ierr == nil {
			return nil
		}
	}
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
	case isDMError(err, -5403, -6169):
		return &trait.Error{
			Err:      err,
			Internal: trait.ErrParam,
			Detail:   err.Error(),
		}
	case isDMError(err, -6602, -6625):
		return &trait.Error{
			Err:      err,
			Internal: trait.ErrUniqueKey,
			Detail:   err.Error(),
		}
	case isDMError(err, -2116):
		return &trait.Error{
			Err:      err,
			Internal: etrait.ECColumnExists,
			Detail:   err.Error(),
		}
	case isDMError(err, -2654):
		return &trait.Error{
			Err:      err,
			Internal: etrait.ECPriKeyExists,
			Detail:   err.Error(),
		}
	default:
		return readerErrorWrraper(err)
	}
}

func isDMError(err error, enum ...int) bool {
	is := false
	if err0, ok := err.(*dm.DmError); ok {
		for _, v := range enum {
			if err0.ErrCode == int32(v) {
				return true
			}
		}
	}
	return is
}

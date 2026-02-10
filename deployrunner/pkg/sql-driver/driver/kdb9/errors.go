package store

import (
	"database/sql"
	"fmt"

	"taskrunner/trait"

	etrait "taskrunner/error/codes/trait"

	"github.com/AISHU-Technology/proton-rds-sdk-go/driver/kingbase/gokb"
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
	case isDMError(err, "22001"):
		return &trait.Error{
			Err:      err,
			Internal: trait.ErrParam,
			Detail:   err.Error(),
		}
	case isDMError(err, "42P06", "42P04", "23505"):
		return &trait.Error{
			Err:      err,
			Internal: trait.ErrUniqueKey,
			Detail:   err.Error(),
		}
	case isDMError(err, "42701"):
		return &trait.Error{
			Err:      err,
			Internal: etrait.ECColumnExists,
		}
	default:
		return readerErrorWrraper(err)
	}
}

func isDMError(err error, enum ...string) bool {
	is := false
	if err0, ok := err.(*gokb.Error); ok {
		for _, v := range enum {
			if err0.Code == gokb.ErrorCode(v) {
				return true
			}
		}
	}
	return is
}

package store

import (
	"bytes"
	"fmt"
	"strings"

	"taskrunner/trait"
)

// appInsFilterString compine status,sid and aname into sql statement
func (c *SQLCursor) appInsFilterString(f *trait.AppInsFilter) (stmt string, args []any) {
	args = []any{}
	buf := bytes.NewBufferString("")
	writeAnd := func(args []any) {
		if len(args) > 0 {
			buf.WriteString(" AND ")
		}
	}

	if f.Sid != -1 {
		buf.WriteString("t.sid=?")
		args = append(args, f.Sid)
	}

	if f.Name != "" {
		writeAnd(args)
		args = append(args, f.Name)
		buf.WriteString("t.aname=?")
	}
	if f.Version != "" {
		args = append(args, f.Version)
		buf.WriteString(" AND version=?")
	}

	if len(f.Status) != 0 {
		writeAnd(args)
		statusBuf := bytes.NewBufferString("t.status IN (")
		for i, s := range f.Status {
			args = append(args, s)
			if i == 0 {
				statusBuf.WriteString("?")
			} else {
				statusBuf.WriteString(",?")
			}
		}
		statusBuf.WriteString(")")
		buf.Write(statusBuf.Bytes())
	}
	if len(args) > 0 {
		return "WHERE " + buf.String(), args
	}
	return "", args
}

// appInsFilterString compine status,sid and aname into sql statement
func (c *SQLCursor) JobappInsFilterString(f *trait.AppInsFilter) (stmt string, args []any) {
	args = []any{}
	andCondition := []string{}

	if f.Sid > -1 {
		args = append(args, f.Sid)
		andCondition = append(andCondition, "a.sid=?")
	}
	if f.Name != "" {
		args = append(args, f.Name)
		andCondition = append(andCondition, "a.aname=?")
	}

	otypeCondition := make([]string, 0, len(f.Jtype))
	for _, jtype := range f.Jtype {
		if jtype >= 0 {
			args = append(args, jtype)
			if jtype == 0 {
				otypeCondition = append(otypeCondition, "(otype=? OR otype is NULL)")
			} else {
				otypeCondition = append(otypeCondition, "otype=?")
			}
		}
	}
	if len(otypeCondition) != 0 {
		andCondition = append(andCondition, strings.Join(otypeCondition, " OR "))
	}

	if len(f.Status) != 0 {
		statusBuf := bytes.NewBufferString("a.status IN (")
		for i, s := range f.Status {
			args = append(args, s)
			if i == 0 {
				statusBuf.WriteString("?")
			} else {
				statusBuf.WriteString(",?")
			}
		}
		statusBuf.WriteString(")")
		andCondition = append(andCondition, statusBuf.String())
	}
	if len(args) > 0 {
		return "WHERE " + strings.Join(andCondition, " AND "), args
	}
	return "", args
}

func (c *SQLCursor) JobLogFilter(f trait.JobLogFilter) (stmt string, args []any) {
	//
	buf := bytes.NewBuffer(nil)
	condition := make([]string, 0, 3)
	args = make([]any, 0, 3)
	if f.Sort != trait.ASCSortType {
		f.Sort = trait.DescSortType
	}

	if f.Limit == 0 {
		f.Limit = 10
	}

	if f.Offset < 0 {
		f.Offset = 0
	}

	if f.JID != -1 {
		condition = append(condition, "jid=? ")
		args = append(args, f.JID)
	}
	if f.CID != -1 {
		condition = append(condition, "cid=? ")
		args = append(args, f.CID)
	}
	if f.Timestmp > 0 {
		condition = append(condition, "ltime>? ")
		args = append(args, f.Timestmp)
	}
	if f.Timestmp < 0 {
		condition = append(condition, "ltime<? ")
		args = append(args, -f.Timestmp)
	}

	if len(condition) > 0 {
		buf.WriteString("WHERE ")
		buf.WriteString(strings.Join(condition, "AND "))
	}
	buf.WriteString(fmt.Sprintf("ORDER BY jlid %s LIMIT %d OFFSET %d", f.Sort, f.Limit, f.Offset))

	return buf.String(), args
}

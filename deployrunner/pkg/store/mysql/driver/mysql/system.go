package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"taskrunner/trait"
)

func (c *SQLCursor) DeleteSystemInfo(ctx context.Context, sid int) *trait.Error {
	works, err := c.CountWorkAppIns(ctx, &trait.AppInsFilter{
		Sid:   sid,
		Limit: 1,
	})
	if err != nil {
		return err
	}
	if works > 0 {
		return &trait.Error{
			Internal: trait.ErrApplicationStillUse,
			Detail:   works,
			Err:      fmt.Errorf("system can't delete when it has application instance"),
		}
	}
	_, err = c.ExecContext(ctx, "DELETE FROM task_system WHERE sid=?", sid)
	return err
}

func (c *SQLCursor) GetSystemInfo(ctx context.Context, sid int) (*trait.System, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetSystemInfoStmt, sid)
	s := &trait.System{
		SID: sid,
	}
	bs := []byte{}
	var des sql.NullString

	err := row.Scan(&s.NameSpace, &s.SName, &bs, &des)
	if err != nil {
		return s, err
	}
	s.Description = des.String

	if len(bs) != 0 {
		rerr := json.Unmarshal(bs, &s.Config)
		if rerr != nil {
			return nil, &trait.Error{
				Internal: trait.ECNULL,
				Err:      rerr,
				Detail:   "decode system config error after get from store",
			}
		}
	}

	return s, nil
}

func (tx *TX) InsertSystemInfo(ctx context.Context, s trait.System) (int, *trait.Error) {
	bs, err0 := json.Marshal(s.Config)
	if err0 != nil {
		return -1, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err0,
			Detail:   "decode system config error before store",
		}
	}
	_, err := tx.ExecContext(ctx, tx.stmt.InsertSystemInfoStmt, s.NameSpace, s.SName, bs, s.Description)
	if err != nil {
		return -1, err
	}
	r := tx.QueryRowContext(ctx, tx.stmt.GetInsertSystemInfoStmt, s.NameSpace, s.SName)
	sid := 0
	err = r.Scan(&sid)
	return sid, err
}

func (s *Store) InsertSystemInfo(ctx context.Context, ss trait.System) (int, *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return -1, err
	}
	jid, err := tx.InsertSystemInfo(ctx, ss)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return -1, err
		}
		return -1, err
	}
	return jid, tx.Commit()
}

func (tx *TX) UpdateSystemInfo(ctx context.Context, s trait.System) *trait.Error {
	stmt := tx.SQLCursor
	_, err := stmt.GetSystemInfo(ctx, s.SID)
	if err != nil {
		return err
	}

	return stmt.UpdateSystemInfo(ctx, s)
}

func (s *Store) UpdateSystemInfo(ctx context.Context, ss trait.System) *trait.Error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	if err := tx.UpdateSystemInfo(ctx, ss); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (c *SQLCursor) UpdateSystemInfo(ctx context.Context, s trait.System) *trait.Error {
	bs, err0 := json.Marshal(s.Config)
	if err0 != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err0,
			Detail:   "decode system config error before update",
		}
	}
	_, err := c.ExecContext(ctx, c.stmt.UpdateSystemInfoStmt, s.SName, bs, s.Description, s.SID)
	return err
}

func (c *SQLCursor) CountSystemInfo(ctx context.Context) (int, *trait.Error) {
	row := c.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM task_system`,
	)

	count := 0
	err := row.Scan(&count)
	return count, err
}

func (c *SQLCursor) ListSystemInfo(ctx context.Context, limit int, offset int) (ss []*trait.System, err *trait.Error) {
	rows, err := c.QueryContext(ctx,
		fmt.Sprintf(
			`SELECT namespace, sname, sid 
			FROM task_system ORDER BY sid LIMIT ? offset %d;`,
			offset),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := trait.System{}
		if err0 := rows.Scan(&s.NameSpace, &s.SName, &s.SID); err0 != nil {
			return nil, err0
		}
		ss = append(ss, &s)
	}
	return
}

func (c *SQLCursor) GetSystemInfoByName(ctx context.Context, name string) (*trait.System, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetSystemByNameStmt, name)
	s := &trait.System{}
	bs := []byte{}
	err := row.Scan(&s.NameSpace, &s.SName, &s.SID, &bs)
	if err != nil {
		return s, err
	}
	if len(bs) != 0 {
		if rerr := json.Unmarshal(bs, &s.Config); rerr != nil {
			return nil, &trait.Error{
				Internal: trait.ErrComponentDecodeError,
				Detail:   "decode deploy system config error",
				Err:      rerr,
			}
		}
	}

	return s, nil
}

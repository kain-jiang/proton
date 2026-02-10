package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/trait"
)

/*
+++++++++basic application instance impl++++++++++++++
*/
func (c *SQLCursor) UpdateAPPInsConfig(ctx context.Context, app trait.ApplicationInstance) *trait.Error {
	bs, err0 := json.Marshal(app.AppConfig)
	if err0 != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err0,
			Detail:   "encode application instance config before update",
		}
	}
	_, err := c.ExecContext(ctx, c.stmt.UpdateAPPInsConfigStmt, bs, app.Comment, app.ID)
	return err
}

func (c *SQLCursor) UpdateAPPInsOperateType(ctx context.Context, id int, otype int) *trait.Error {
	_, err := c.ExecContext(
		ctx,
		"UPDATE task_application_instance SET otype=? WHERE ins_id=?;",
		otype, id,
	)
	return err
}

func (c *SQLCursor) UpdateAPPInsStatus(ctx context.Context, id, status int, owner int, start, end int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UpdateAPPInsStatusStmt, status, owner, start, end, id)
	return err
}

func (c *SQLCursor) LayOffAPPIns(ctx context.Context, a *trait.ApplicationInstance) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.LayOffAPPInsStmt, a.AName, a.SID)
	return err
}

func (c *SQLCursor) LayOffAPPInsByID(ctx context.Context, id int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.LayOffAPPInsByidStmt, id)
	return err
}

func (c *SQLCursor) WorkAppIns(ctx context.Context, a *trait.ApplicationInstance) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.WorkAPPInsStmt, a.ID, a.AID, a.SID, a.Version, a.AName, a.Comment, a.CreateTime, a.StartTime, a.EndTime, a.Status)
	return err
}

func (tx *TX) WorkAppIns(ctx context.Context, a *trait.ApplicationInstance) *trait.Error {
	stmt := tx.SQLCursor
	count, err := stmt.CountWorkAppIns(ctx, &trait.AppInsFilter{
		Sid:     a.SID,
		Name:    a.AName,
		Version: a.Version,
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return stmt.WorkAppIns(ctx, a)
	}
	return nil
}

func (s *Store) WorkAppIns(ctx context.Context, a *trait.ApplicationInstance) *trait.Error {
	return driver.StoreTransactionMarco(s, func(tx driver.Transaction) *trait.Error {
		return s.beginWithTx(tx).WorkAppIns(ctx, a)
	})
}

func (c *SQLCursor) GetAPPIns(ctx context.Context, id int) (a *trait.ApplicationInstance, err *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetAPPInsStmt, id)
	appConfig := []byte{}
	a = &trait.ApplicationInstance{}
	a.ID = id
	var otype sql.NullInt64
	atraitBs := []byte{}

	if err0 := row.Scan(&a.AID, &a.SID, &a.Version, &a.AName,
		&a.Status, &appConfig, &a.Onwer, &a.Comment,
		&a.CreateTime, &a.StartTime, &a.EndTime, &otype, &atraitBs,
	); err0 != nil {
		return a, err0
	}
	if err0 := json.Unmarshal(appConfig, &a.AppConfig); err0 != nil {
		return a, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err0,
			Detail:   "decode application instance config from store",
		}
	}
	if otype.Valid {
		a.OType = int(otype.Int64)
	}
	if len(atraitBs) != 0 {
		if err0 := json.Unmarshal(atraitBs, &a.Trait); err0 != nil {
			return a, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err0,
				Detail:   "decode application instance trait from store",
			}
		}
	}

	return
}

func (c *SQLCursor) GetWorkAPPIns(ctx context.Context, name string, sid int) (id int, err *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetworkAPPInsStmt, sid, name)
	id = -1
	err0 := row.Scan(&id)
	return id, err0
}

func (c *SQLCursor) ListWorkAPPIns(ctx context.Context, filter *trait.AppInsFilter) (as []trait.ApplicationInstanceOverview, err *trait.Error) {
	condition, args := c.appInsFilterString(filter)
	args = append(args, filter.Limit)
	rows, err0 := c.QueryContext(ctx,
		fmt.Sprintf(
			`SELECT 
			t.ins_id, t.aid, t.sid, t.version, t.aname, 
			t.icomment, t.create_time, t.start_time, 
			t.end_time, t.status, s.sname
			FROM task_work_application t
			JOIN task_system s ON t.sid=s.sid
			%s ORDER BY t.ins_id DESC LIMIT ? OFFSET %d;`,
			condition, filter.Offset), args...)
	if err0 != nil {
		return as, err0
	}
	defer rows.Close()

	for rows.Next() {
		a := trait.ApplicationInstanceOverview{}
		if err0 := rows.Scan(
			&a.ID, &a.AID, &a.SID, &a.Version, &a.AName,
			&a.Comment, &a.CreateTime, &a.StartTime,
			&a.EndTime, &a.Status, &a.System.SName); err0 != nil {
			return as, err0
		}
		// a.SID = filter.Sid
		as = append(as, a)
	}
	return
}

func (c *SQLCursor) CountWorkAppIns(ctx context.Context, filter *trait.AppInsFilter) (int, *trait.Error) {
	condition, args := c.appInsFilterString(filter)
	row := c.QueryRowContext(
		ctx,
		fmt.Sprintf(
			"SELECT count(t.ins_id) FROM task_work_application t %s;",
			condition),
		args...)
	totalNum := 0
	err := row.Scan(&totalNum)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		totalNum = 0
		err = nil
	}
	return totalNum, err
}

/*
---------basic application instance impl end-----------
*/

// InsertAPPIns impl trait.APPlicationInsWriter
func (tx *TX) InsertAPPIns(ctx context.Context, a *trait.ApplicationInstance) (id int, err *trait.Error) {
	id = -1
	if len(a.Version) > 128 {
		return -1, &trait.Error{
			Internal: trait.ErrParam,
			Detail:   fmt.Sprintf("version is too long, max length is 128, input length: %d", len(a.Version)),
		}
	}

	atraitBs, rerr := json.Marshal(a.Trait)
	if rerr != nil {
		return id, &trait.Error{
			Internal: trait.ErrParam,
			Err:      rerr,
			Detail:   "encode application instance trait before store",
		}
	}

	bs, rerr := json.Marshal(a.AppConfig)
	if rerr != nil {
		return id, &trait.Error{
			Internal: trait.ErrParam,
			Err:      rerr,
			Detail:   "encode application instance config before store",
		}
	}

	_, err = tx.ExecContext(
		ctx,
		tx.stmt.InsertAPPInsStmt,
		a.AID, a.SID, a.Version, a.AName, a.Status,
		bs, 0, a.Comment, a.CreateTime, -1, -1,
		a.OType, atraitBs)
	if err != nil {
		return -1, err
	}

	// warn this may be wrong when parrelel !!!!!
	// r := tx.QueryRowContext(ctx, "SELECT ins_id FROM task_application_instance WHERE aid=? AND sid=? AND version=? AND aname=? AND status=? AND owner=? AND icomment=? AND create_time=? AND start_time=? AND end_time=?;",
	// r := tx.QueryRowContext(ctx, "SELECT ins_id FROM task_application_instance WHERE aid=? AND sid=? AND aname=? AND status=? AND owner=? AND icomment=? AND create_time=? AND start_time=? AND end_time=?;",
	// a.AID, a.SID, a.AName, a.Status, 0, a.Comment, a.CreateTime, -1, -1)
	r := tx.QueryRowContext(ctx, tx.stmt.GetInsertAPPInsStmt, a.AID, a.SID, a.Version, a.AName, a.Status, 0, a.Comment, a.CreateTime, -1, -1)
	// r := tx.QueryRowContext(ctx, tx.stmt.GetInsertAPPInsStmt)
	if err = r.Scan(&id); err != nil {
		return
	}
	a.ID = id
	stmt := tx.SQLCursor
	if err = stmt.insertAppInsComponentIns(ctx, a); err != nil {
		return
	}
	return a.ID, nil
}

// InsertAPPIns implt store. Insert application and insert component into application
func (s *Store) InsertAPPIns(ctx context.Context, a *trait.ApplicationInstance) (id int, err *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return -1, err
	}
	id, err = tx.InsertAPPIns(ctx, a)
	if err != nil {
		if err0 := tx.Rollback(); err0 != nil {
			return -1, err0
		}
		return
	}
	err = tx.Commit()
	return
}

// UpdateAPPInsConfig impl trait.APPlicationInsWriter
func (tx *TX) UpdateAPPInsConfig(ctx context.Context, a trait.ApplicationInstance) (err *trait.Error) {
	stmt := tx.SQLCursor
	err = stmt.UpdateAPPInsConfig(ctx, a)
	if err != nil {
		return
	}
	if err = stmt.UpdateAppInsComponentIns(ctx, a.Components); err != nil {
		return
	}
	return
}

// UpdateAPPInsConfig imply trait, update application instance config and component's config
func (s *Store) UpdateAPPInsConfig(ctx context.Context, a trait.ApplicationInstance) (err *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return err
	}
	err = tx.UpdateAPPInsConfig(ctx, a)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return
	}
	err = tx.Commit()
	return
}

// GetAPPIns impl trait.APPlicationInsReader
func (tx *TX) GetAPPIns(ctx context.Context, id int) (a *trait.ApplicationInstance, err *trait.Error) {
	stmt := tx.SQLCursor
	a, err = stmt.GetAPPIns(ctx, id)
	if err != nil {
		return
	}
	a.ID = id

	cis, err := stmt.getAPPInsComponentIns(ctx, id)
	if err != nil {
		return
	}
	a.Components = cis

	return a, nil
}

// GetAPPIns get app instance
func (s *Store) GetAPPIns(ctx context.Context, id int) (a *trait.ApplicationInstance, err *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return nil, err
	}
	a, err = tx.GetAPPIns(ctx, id)
	if err != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return
	}
	return a, tx.Commit()
}

// GetWorkAPPIns impl trait.ApplicationInsReader
func (tx *TX) GetWorkAPPIns(ctx context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	stmt := tx.SQLCursor
	id, err := stmt.GetWorkAPPIns(ctx, name, sid)
	if err != nil {
		return nil, err
	}
	return tx.GetAPPIns(ctx, id)
}

// GetWorkAPPIns Get work application instance
func (s *Store) GetWorkAPPIns(ctx context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return nil, err
	}
	a, err := tx.GetWorkAPPIns(ctx, name, sid)
	if err != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, err
	}
	return a, tx.Commit()
}

func (s *SQLCursor) DeleteAPPIns(ctx context.Context, id int) *trait.Error {
	_, err := s.ExecContext(ctx, s.stmt.DeleteAppinsStatusStmt, id)
	return err
}

func (s *SQLCursor) DeleteAPPInsComponents(ctx context.Context, id int) *trait.Error {
	_, err := s.ExecContext(ctx, s.stmt.DeleteAPPInsComponentInsStmt, id)
	return err
}

func (tx *TX) DeleteAPPIns(ctx context.Context, id int, force bool) *trait.Error {
	stmt := tx.SQLCursor

	cins, err := stmt.getAPPInsComponentIns(ctx, id)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return nil
	}

	for _, c := range cins {
		if err := stmt.DeleteEdgeFrom(ctx, c.CID); err != nil {
			return err
		}
	}

	for _, c := range cins {
		if force {
			if count, err := stmt.CountEdgeTo(ctx, c.CID); err != nil {
				return err
			} else if count > 0 {
				return &trait.Error{
					Internal: trait.ErrAPPlicationComponentTortuous,
					Err:      fmt.Errorf("the compoent still in use, shouldn't delete"),
					Detail:   count,
				}
			}
		} else {
			if err := stmt.deleteEdgeto(ctx, c.CID); err != nil {
				return err
			}
		}
	}
	if err := stmt.DeleteAPPIns(ctx, id); err != nil {
		return err
	}

	if err := stmt.LayOffAPPInsByID(ctx, id); err != nil {
		return err
	}

	return stmt.DeleteAPPInsComponents(ctx, id)
}

func (s *Store) DeleteAPPIns(ctx context.Context, id int, force bool) *trait.Error {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return err
	}
	if err := tx.DeleteAPPIns(ctx, id, force); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (c *SQLCursor) GetAppLock(ctx context.Context, sid int, aname string) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetAppLock, aname, sid)
	jid := -1
	err := row.Scan(&jid)
	return jid, err
}

func (c *SQLCursor) LockApp(ctx context.Context, sid int, jid int, aname string) *trait.Error {
retry:
	_, err := c.ExecContext(ctx, c.stmt.LockApp, jid, sid, aname)
	if trait.IsInternalError(err, trait.ErrUniqueKey) {
		ljid, err := c.GetAppLock(ctx, sid, aname)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			goto retry
		} else if err == nil {
			if ljid == jid {
				return nil
			}
			// TODO configurable interval
			delay := time.NewTimer(2 * time.Second)
			select {
			case <-delay.C:
				goto retry
			case <-ctx.Done():
				return ctx.Err().(*trait.Error)
			}
		} else {
			return err
		}

	}
	return err
}

func (c *SQLCursor) UnlockApp(ctx context.Context, sid, jid int, aname string) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UnlockApp, aname, sid, jid)
	return err
}

package store

import (
	"context"
	"database/sql"
	"fmt"

	"taskrunner/trait"
)

/*
+++++++++ basic job impl +++++++++++++++++++++++++
*/

func (c *SQLCursor) GetJobRecord(ctx context.Context, jid int) (int, int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetJobRecordStmt, jid)
	cur := -1
	target := -1
	err := row.Scan(&target, &cur)
	return target, cur, err
}

func (c *SQLCursor) CountJobRecord(ctx context.Context, f *trait.AppInsFilter) (int, *trait.Error) {
	condition, args := c.JobappInsFilterString(f)
	row := c.QueryRowContext(ctx, fmt.Sprintf(c.stmt.ListJobRecordCount, condition), args...)
	count := 0
	err := row.Scan(&count)
	return count, err
}

func (c *SQLCursor) ListJobRecord(ctx context.Context, filter *trait.AppInsFilter) ([]trait.JobRecord, *trait.Error) {
	condition, args := c.JobappInsFilterString(filter)
	args = append(args, filter.Limit)
	// fmt.Printf(c.stmt.ListJobRecordStmt+"\n", condition, filter.Offset)
	rows, err := c.QueryContext(ctx,
		fmt.Sprintf(c.stmt.ListJobRecordStmt, condition, filter.Offset), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]trait.JobRecord, 0)
	for rows.Next() {
		job := trait.JobRecord{
			Target: &trait.ApplicationInstance{},
		}
		curID := -1
		var otype sql.NullInt64
		if err0 := rows.Scan(&job.ID, &job.Target.ID, &curID, &job.Target.AID,
			&job.Target.AName, &job.Target.Status, &job.Target.SID,
			&job.Target.Version, &job.Target.Comment, &job.Target.CreateTime,
			&job.Target.StartTime, &job.Target.EndTime, &otype, &job.Target.SName); err0 != nil {
			return nil, err0
		}
		if curID != -1 {
			job.Current = &trait.ApplicationInstance{}
			job.Current.ID = job.ID
		}
		if otype.Valid {
			job.Target.OType = int(otype.Int64)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

/*
---------basic job impl end--------------------
*/

// GetJobRecord impl trait.JobRecordReader
func (tx *TX) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	stmt := tx.SQLCursor
	j := trait.JobRecord{
		ID: jid,
	}
	target, cur, err := stmt.GetJobRecord(ctx, jid)
	if err != nil {
		return j, err
	}
	if cur != -1 {
		a, err := stmt.GetAPPIns(ctx, cur)
		if err != nil {
			return j, err
		}
		a.ID = cur
		j.Current = a
		// // no need current application instance for this interface
		// j.Current, err = tx.GetAPPIns(ctx, cur)
		// if err != nil {
		// 	return j, err
		// }
	}
	// } else {
	// 	j.Current = &trait.ApplicationInstance{}
	// 	j.Current.ID = -1
	// }

	j.Target, err = tx.GetAPPIns(ctx, target)
	return j, err
}

// GetJobRecord impy trait.JobRecordReader
func (s *Store) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	job := trait.JobRecord{}
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return job, err
	}

	job, err = tx.GetJobRecord(ctx, jid)
	if err != nil {
		if err := tx.Commit(); err != nil {
			return job, err
		}
		return job, err
	}
	return job, tx.Commit()
}

// InsertJobRecord impy trait.JobRecordWriter
func (tx *TX) InsertJobRecord(ctx context.Context, j *trait.JobRecord) (int, *trait.Error) {
	aiid, err := tx.InsertAPPIns(ctx, j.Target)
	if err != nil {
		return -1, err
	}
	j.Target.ID = aiid

	cur := -1
	if j.Current != nil {
		cur = j.Current.ID
	}
	_, err = tx.ExecContext(ctx, tx.stmt.InsertJobRecordStmt, j.Target.ID, cur)
	if err != nil {
		return -1, err
	}
	r := tx.QueryRowContext(ctx, tx.stmt.GetInsertJobRecordStmt, j.Target.ID, cur)
	jid := 0
	err = r.Scan(&jid)
	return jid, err
}

// InsertJobRecord impl trait.Store
func (s *Store) InsertJobRecord(ctx context.Context, j *trait.JobRecord) (int, *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return -1, err
	}
	jid, err := tx.InsertJobRecord(ctx, j)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return -1, err
		}
		return -1, err
	}
	return jid, tx.Commit()
}

func (s *SQLCursor) DeleteJobRecord(ctx context.Context, jid int) *trait.Error {
	_, err := s.ExecContext(ctx, s.stmt.DeleteJobRecordStmt, jid)
	return err
}

func (tx *TX) DeleteJobRecord(ctx context.Context, jid int, force bool) *trait.Error {
	stmt := tx.SQLCursor
	target, _, err := stmt.GetJobRecord(ctx, jid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return nil
	}

	if err := tx.DeleteAPPIns(ctx, target, force); err != nil {
		return err
	}

	return stmt.DeleteJobRecord(ctx, jid)
}

func (s *Store) DeleteJobRecord(ctx context.Context, jid int, force bool) *trait.Error {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return err
	}
	if err := tx.DeleteJobRecord(ctx, jid, force); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (c *SQLCursor) InsertJobLog(ctx context.Context, log trait.JobLog) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.InsertJobLogStmt, log.JID, log.CID, log.AIID, log.Aname, log.Cname, log.Code, log.Timestamp, []byte(log.Msg))
	return err
}

func (c *SQLCursor) ListJobLog(ctx context.Context, f trait.JobLogFilter) ([]trait.JobLog, *trait.Error) {
	condition, args := c.JobLogFilter(f)
	rows, err := c.QueryContext(ctx, fmt.Sprintf(c.stmt.ListJobLogStmt, condition), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]trait.JobLog, 0)
	for rows.Next() {
		job := trait.JobLog{}
		var bs []byte
		if err0 := rows.Scan(&job.JLID, &job.JID, &job.CID, &job.AIID,
			&job.Aname, &job.Cname, &job.Code, &job.Timestamp, &bs); err0 != nil {
			return nil, err0
		}
		if len(bs) != 0 {
			job.Msg = string(bs)
		}
		items = append(items, job)
	}

	return items, nil
}

func (c *SQLCursor) CountJobLog(ctx context.Context, f trait.JobLogFilter) (int, *trait.Error) {
	condition, args := c.JobLogFilter(f)
	row := c.QueryRowContext(ctx, fmt.Sprintf(c.stmt.CountJobLogStmt, condition), args...)
	count := 0
	err := row.Scan(&count)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		err = nil
	}
	return count, err
}

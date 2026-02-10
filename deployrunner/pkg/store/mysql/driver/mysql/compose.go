package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/pkg/utils"
	"taskrunner/trait"
)

func (c *SQLCursor) InsertComposeJob(ctx context.Context, j trait.ComposeJob) *trait.Error {
	bs, rerr := json.Marshal(j.Config)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ErrParam,
			Detail:   fmt.Sprintf("encode %#v fail", j.Config),
			Err:      rerr,
		}
	}
	_, err := c.ExecContext(
		ctx, c.stmt.InsertComposeJob,
		j.Jname, j.SID, j.Status, j.Processed, j.Total,
		bs, j.Mversion, j.CreateTime, j.StartTime, j.EndTime,
		j.Description,
	)

	return err
}

func (c *SQLCursor) SetComposeJob(ctx context.Context, j trait.ComposeJob) *trait.Error {
	bs, rerr := json.Marshal(j.Config)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ErrParam,
			Detail:   fmt.Sprintf("encode %#v fail", j.Config),
			Err:      rerr,
		}
	}
	_, err := c.ExecContext(ctx, c.stmt.SetComposeJob, j.Jname, j.SID, j.Status, j.Processed, j.Total, bs, j.Jid)
	return err
}

func (c *SQLCursor) GetComposeJob(ctx context.Context, jid int) (*trait.ComposeJob, *trait.Error) {
	obj := &trait.ComposeJob{}
	bs := []byte{}
	if err := c.getComposeJob(
		ctx,
		"jid, jname, sid, status, processed, total, config, mversion, create_time, start_time, end_time, mdescription", "WHERE jid=?",
		[]any{jid}, &obj.Jid, &obj.Jname, &obj.SID, &obj.Status,
		&obj.Processed, &obj.Total, &bs, &obj.Mversion,
		&obj.CreateTime, &obj.StartTime, &obj.EndTime, &obj.Description,
	); err != nil {
		return nil, err
	}
	if rerr := json.Unmarshal(bs, &obj.Config); rerr != nil {
		return nil, &trait.Error{
			Internal: trait.ErrParam,
			Err:      rerr,
			Detail:   fmt.Sprintf("decode compose job config fail with '%s'", string(bs)),
		}
	}
	return obj, nil
}

func (c *SQLCursor) getComposeJob(ctx context.Context, fields, condition string, args []any, receiver ...any) *trait.Error {
	row := c.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM task_compose_job %s", fields, condition), args...)
	return row.Scan(receiver...)
}

func (c *SQLCursor) ListComposeJob(ctx context.Context, limit, offset int, f trait.ComposeJobFilter) ([]*trait.ComposeJobMeata, int, *trait.Error) {
	status := f.Status
	l := len(status)
	condition := ""
	buf := utils.BytesPool.Get()
	defer utils.BytesPool.Free(buf)
	writeAnd := func(args []any) {
		if len(args) > 0 {
			_, _ = buf.WriteString(" AND ")
		}
	}
	args := make([]any, 0, l+1)
	if l > 0 {
		_, _ = buf.WriteString("j.status IN (?")
		args = append(args, status[0])
		for i := 1; i < l; i++ {
			_, _ = buf.WriteString(",?")
			args = append(args, status[i])
		}
		_, _ = buf.WriteString(")")
		condition = buf.String()
	}
	if f.Name != "" {
		writeAnd(args)
		_, _ = buf.WriteString("j.jname=?")
		args = append(args, f.Name)
	}
	if f.ListType == trait.ComposeJobNormalType {
		writeAnd(args)
		_, _ = buf.WriteString("j.mversion=?")
		args = append(args, "")
	} else if f.ListType == trait.ComposeJobSuiteType {
		writeAnd(args)
		buf.WriteString("j.mversion!=?")
		args = append(args, "")
	}
	if f.SID > 0 {
		writeAnd(args)
		buf.WriteString("j.sid=?")
		args = append(args, f.SID)
	}
	if len(args) != 0 {
		condition = "WHERE " + buf.String()
	}
	rows, err := c.QueryContext(ctx,
		fmt.Sprintf(`
		SELECT j.jid, j.jname, j.sid, j.status, 
		j.processed, j.total, j.create_time, 
		j.start_time, j.end_time, j.mversion, 
		j.mdescription, s.sid, s.sname
		FROM task_compose_job j 
		JOIN task_system s ON j.sid=s.sid
		%s ORDER BY j.jid DESC LIMIT %d OFFSET %d`,
			condition, limit, offset), args...)
	if err != nil {
		return nil, -1, err
	}
	list := []*trait.ComposeJobMeata{}
	for rows.Next() {
		obj := &trait.ComposeJobMeata{}
		if err := rows.Scan(&obj.Jid, &obj.Jname, &obj.SID, &obj.Status,
			&obj.Processed, &obj.Total, &obj.CreateTime, &obj.StartTime,
			&obj.EndTime, &obj.Mversion, &obj.Description, &obj.SID, &obj.SName); err != nil {
			return nil, -1, err
		}
		list = append(list, obj)
	}

	row := c.QueryRowContext(ctx, fmt.Sprintf("SELECT count(1) FROM task_compose_job j %s", condition), args...)
	count := -1
	err = row.Scan(&count)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		count = 0
		err = nil
	}
	return list, count, err
}

func (c *SQLCursor) UpdateComposeJobProcess(ctx context.Context, jid, process int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UpdateComposeJobProcess, process, jid)
	return err
}

func (c *SQLCursor) UpdateComposeJobStatus(ctx context.Context, jid, status, startime, endtime int) *trait.Error {
	var err *trait.Error
	if startime == -2 {
		_, err = c.ExecContext(ctx, fmt.Sprintf(c.stmt.UpdateComposeJobStatus, ""), status, endtime, jid)
	} else {
		_, err = c.ExecContext(ctx, fmt.Sprintf(c.stmt.UpdateComposeJobStatus, ", start_time=?"), status, endtime, startime, jid)
	}
	return err
}

func (c *SQLCursor) GetCompoesJobTasks(ctx context.Context, jid int) ([][2]int, *trait.Error) {
	rows, err := c.QueryContext(ctx, c.stmt.GetCompoesJobTasks, jid)
	if err != nil {
		return nil, err
	}
	list := [][2]int{}
	for rows.Next() {
		index := [2]int{}
		if err := rows.Scan(&index[1], &index[0]); err != nil {
			return nil, err
		}
		list = append(list, index)
	}
	return list, nil
}

func (c *SQLCursor) InserComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.InserComposeJobTask, jid, jtindex, ajid)
	return err
}

// func (c *SQLCursor) DeleteComposeJobTask(ctx context.Context, jid, jtindex int) *trait.Error {
// 	_, err := c.ExecContext(ctx, c.stmt.DeleteComposeJobTask, jid, jtindex)
// 	return err
// }

func (c *SQLCursor) UpdateComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UpdateComposeJobTask, ajid, jid, jtindex)
	return err
}

func (c *SQLCursor) DeleteComposeJobTasks(ctx context.Context, jid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteComposeJobTasks, jid)
	return err
}

func (c *SQLCursor) GetCompoesJobTask(ctx context.Context, jid, jtindex int) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetCompoesJobTask, jid, jtindex)
	ajid := 0
	if err := row.Scan(&ajid); err != nil {
		return -1, err
	}

	return int(ajid), nil
}

func (c *SQLCursor) InsertComposeManifests(ctx context.Context, m trait.ComposeJobManifests) *trait.Error {
	bs, rerr := json.Marshal(m.Manifests)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ErrParam,
			Detail:   "compose job manifests encode fail",
			Err:      rerr,
		}
	}
	_, err := c.ExecContext(ctx, c.stmt.InsertComposeManifests, m.Name, m.Version, bs, m.Description)
	return err
}

func (c *SQLCursor) GetComposeManifests(ctx context.Context, name, version string) (*trait.ComposeJobManifests, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetComposeManifests, name, version)
	o := &trait.ComposeJobManifests{}
	bs := []byte{}
	if err := row.Scan(&o.Name, &o.Version, &o.Description, &bs); err != nil {
		return nil, err
	}
	if rerr := json.Unmarshal(bs, &o.Manifests); rerr != nil {
		return nil, &trait.Error{
			Internal: trait.ErrComponentDecodeError,
			Detail:   fmt.Sprintf("decode compose manifests %s:%s error", name, version),
			Err:      rerr,
		}
	}
	return o, nil
}

func (c *SQLCursor) ListComposeManifest(ctx context.Context, limit, offset int, filter *trait.ComposeManifestFilter) ([]*trait.ComposeJobManifestsMeta, int, *trait.Error) {
	condition := ""
	fields := "mname, mversion, mdescription"
	args := []any{}
	if filter == nil {
		filter = &trait.ComposeManifestFilter{}
	}
	if filter.NoWork {
		// if filter.Sid < 0 {
		// 	return nil, &trait.Error{
		// 		Internal: trait.ErrParam,
		// 		Detail:   "list compose manifest filter param error, when set nowork filter, it must set sid",
		// 	}
		// }
		condition = "WHERE mname not in (SELECT mname FROM task_work_compose_manifests) GROUP BY mname"
		if filter.Mname != "" {
			return nil, -1, &trait.Error{
				Internal: trait.ErrParam,
				Detail:   "must not set mname filter when nowrok is true.",
			}
		} else {
			fields = "mname"
		}
	} else {
		if filter.Mname != "" {
			condition = "WHERE mname=?"
			args = append(args, filter.Mname)
		}
	}
	args = append(args, limit)
	rows, err := c.QueryContext(ctx, fmt.Sprintf(c.stmt.ListComposeManifest, fields, condition, offset), args...)
	if err != nil {
		return nil, -1, err
	}

	list := []*trait.ComposeJobManifestsMeta{}
	for rows.Next() {
		obj := &trait.ComposeJobManifestsMeta{}
		var receiver []any
		if filter.NoWork {
			receiver = []any{&obj.Name}
		} else {
			receiver = []any{&obj.Name, &obj.Version, &obj.Description}
		}
		if err := rows.Scan(receiver...); err != nil {
			return nil, -1, err
		}
		list = append(list, obj)
	}

	fields = "count(1) "
	countQuery := fmt.Sprintf(c.stmt.ListComposeManifest, fields, condition, offset)
	if filter.NoWork {
		countQuery = fmt.Sprintf("SELECT %s FROM (%s) as namecount", fields, countQuery)
	}
	row := c.QueryRowContext(ctx, countQuery, args...)
	count := -1
	err = row.Scan(&count)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		count = 0
		err = nil
	}

	return list, count, err
}

func (c *SQLCursor) ListWorkComposeJobManifests(ctx context.Context, limit, offset int, f trait.ComposeJobFilter) ([]*trait.ComposeJobMeata, int, *trait.Error) {
	fields := "jid, mname, mversion, sid, status, mdescription"
	condition := ""
	args := []any{}
	status := f.Status
	l := len(status)
	if l > 0 {
		buf := bytes.NewBuffer([]byte{})
		_, _ = buf.WriteString("WHERE status IN (?")
		args = append(args, status[0])
		for i := 1; i < l; i++ {
			_, _ = buf.WriteString(",?")
			args = append(args, status[i])
		}
		_, _ = buf.WriteString(")")
		condition = buf.String()
	}
	if f.Name != "" {
		if len(args) != 0 {
			condition += " AND mname=?"
		} else {
			condition += "WHERE mname=?"
		}
		args = append(args, f.Name)
	}
	args = append(args, limit)
	rows, err := c.QueryContext(ctx, fmt.Sprintf(c.stmt.ListWorkComposeJobManifests, fields, condition, offset), args...)
	if err != nil {
		return nil, -1, err
	}

	list := []*trait.ComposeJobMeata{}
	for rows.Next() {
		obj := &trait.ComposeJobMeata{}

		if err := rows.Scan(&obj.Jid, &obj.Jname, &obj.Mversion, &obj.SID, &obj.Status, &obj.Description); err != nil {
			return nil, -1, err
		}
		list = append(list, obj)
	}

	fields = "count(1)"
	row := c.QueryRowContext(ctx, fmt.Sprintf(c.stmt.ListWorkComposeJobManifests, fields, condition, offset), args...)
	count := -1
	err = row.Scan(&count)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		count = 0
		err = nil
	}
	return list, count, err
}

func (c *SQLCursor) InsertWorkComposeManifests(ctx context.Context, obj trait.ComposeJobMeata) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.InsertWorkComposeManifests, obj.Jid, obj.Jname, obj.Mversion, obj.SID, obj.Status, obj.Description)
	return err
}

func (c *SQLCursor) GetWorkComposeJobManifests(ctx context.Context, obj trait.ComposeJobMeata) (*trait.ComposeJobMeata, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetWorkComposeJobManifests, obj.Jname, obj.SID)
	o := &trait.ComposeJobMeata{}
	err := row.Scan(&o.Jid, &o.Jname, &o.Mversion, &o.SID)
	return o, err
}

func (c *SQLCursor) DeleteWorkComposeJobManifests(ctx context.Context, obj trait.ComposeJobMeata) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteWorkComposeJobManifests, obj.Jname, obj.SID)
	return err
}

// ----------------- tx -------------------------------

func (tx *TX) InsertComposeJob(ctx context.Context, j trait.ComposeJob) (int, *trait.Error) {
	con := tx.SQLCursor
	if err := con.InsertComposeJob(ctx, j); err != nil {
		return -1, err
	}

	row := con.QueryRowContext(ctx, con.stmt.IDENTITY)
	jid := 0
	if err := row.Scan(&jid); err != nil {
		return -1, err
	}

	return int(jid), nil
}

func (tx *TX) SetComposeJob(ctx context.Context, j trait.ComposeJob) *trait.Error {
	stmt := tx.SQLCursor
	return stmt.SetComposeJob(ctx, j)
}

func (tx *TX) UpdateComposeJobProcess(ctx context.Context, jid, process int) *trait.Error {
	return tx.SQLCursor.UpdateComposeJobProcess(ctx, jid, process)
}

func (tx *TX) UpdateComposeJobStatus(ctx context.Context, jid, status, startime, endtime int) *trait.Error {
	return tx.SQLCursor.UpdateComposeJobStatus(ctx, jid, status, startime, endtime)
}

func (tx *TX) SetComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *trait.Error {
	stmt := tx.SQLCursor
	old, err := stmt.GetCompoesJobTask(ctx, jid, jtindex)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		err = stmt.InserComposeJobTask(ctx, jid, jtindex, ajid)
	} else if err != nil {
		return err
	} else if old != ajid {
		err = stmt.UpdateComposeJobTask(ctx, jid, jtindex, ajid)
	}
	return err
}

// func (tx *TX) DeleteComposeJobTask(ctx context.Context, jid, jtindex int) *trait.Error {
// 	return tx.SQLCursor.DeleteComposeJobTask(ctx, jid, jtindex)
// }

func (tx *TX) DeleteComposeJobTasks(ctx context.Context, jid int) *trait.Error {
	return tx.SQLCursor.DeleteComposeJobTasks(ctx, jid)
}

func (tx *TX) GetCompoesJobTask(ctx context.Context, jid, jtindex int) (int, *trait.Error) {
	return tx.SQLCursor.GetCompoesJobTask(ctx, jid, jtindex)
}

func (tx *TX) InsertWorkComposeManifests(ctx context.Context, obj trait.ComposeJobMeata) *trait.Error {
	o, err := tx.SQLCursor.GetWorkComposeJobManifests(ctx, obj)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return tx.SQLCursor.InsertWorkComposeManifests(ctx, obj)
	}
	if err != nil {
		return err
	}
	if o.Jid != obj.Jid || o.Mversion != obj.Mversion {
		if err := tx.SQLCursor.DeleteWorkComposeJobManifests(ctx, obj); err != nil {
			return err
		}
		return tx.SQLCursor.InsertWorkComposeManifests(ctx, obj)
	}
	return nil
}

//-------------------- store --------------------

func (s *Store) InsertComposeJob(ctx context.Context, j trait.ComposeJob) (int, *trait.Error) {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	jid, err := tx.InsertComposeJob(ctx, j)
	return jid, driver.StoreTransactionErrorMarco(tx, err)
}

func (s *Store) SetComposeJob(ctx context.Context, j trait.ComposeJob) *trait.Error {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return err
	}
	err = tx.SetComposeJob(ctx, j)
	return driver.StoreTransactionErrorMarco(tx, err)
}

func (s *Store) UpdateComposeJobProcess(ctx context.Context, jid, process int) *trait.Error {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return err
	}
	err = tx.UpdateComposeJobProcess(ctx, jid, process)
	return driver.StoreTransactionErrorMarco(tx, err)
}

func (s *Store) UpdateComposeJobStatus(ctx context.Context, jid, status, startime, endtime int) *trait.Error {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return err
	}
	err = tx.UpdateComposeJobStatus(ctx, jid, status, startime, endtime)
	return driver.StoreTransactionErrorMarco(tx, err)
}

func (s *Store) SetComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *trait.Error {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return err
	}
	err = tx.SetComposeJobTask(ctx, jid, jtindex, ajid)
	return driver.StoreTransactionErrorMarco(tx, err)
}

// func (s *Store) DeleteComposeJobTask(ctx context.Context, jid, jtindex int) *trait.Error {
// 	tx, err := s.begin(context.Background(), nil)
// 	if err != nil {
// 		return err
// 	}
// 	err = tx.DeleteComposeJobTask(ctx, jid, jtindex)
// 	return storeTransactionErrorMarco(tx, err)
// }

func (s *Store) DeleteComposeJobTasks(ctx context.Context, jid int) *trait.Error {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return err
	}
	err = tx.DeleteComposeJobTasks(ctx, jid)
	return driver.StoreTransactionErrorMarco(tx, err)
}

func (s *Store) GetCompoesJobTask(ctx context.Context, jid, jtindex int) (int, *trait.Error) {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	id, err := tx.SQLCursor.GetCompoesJobTask(ctx, jid, jtindex)
	return id, driver.StoreTransactionErrorMarco(tx, err)
}

func (s *Store) InsertWorkComposeManifests(ctx context.Context, obj trait.ComposeJobMeata) *trait.Error {
	return driver.StoreTransactionMarco(s, func(tx driver.Transaction) *trait.Error {
		return s.beginWithTx(tx).InsertWorkComposeManifests(ctx, obj)
	})
}

package store

import (
	"context"
	"encoding/json"
	"fmt"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/trait"
)

func (s *Store) InsertComposeJob(ctx context.Context, j trait.ComposeJob) (int, *trait.Error) {
	return InsertComposeJob(ctx, s, j)
}

func (tx *TX) InsertComposeJob(ctx context.Context, j trait.ComposeJob) (int, *trait.Error) {
	return InsertComposeJob(ctx, tx.Transaction, j)
}

func InsertComposeJob(ctx context.Context, c driver.CursorConn, j trait.ComposeJob) (int, *trait.Error) {
	bs, rerr := json.Marshal(j.Config)
	if rerr != nil {
		return -1, &trait.Error{
			Internal: trait.ErrParam,
			Detail:   fmt.Sprintf("encode %#v fail", j.Config),
			Err:      rerr,
		}
	}
	row, err := c.QueryContext(
		ctx, "INSERT INTO task_compose_job (jname, sid, status, processed, total, config, mversion, create_time, start_time, end_time, mdescription) VALUES (?,?,?,?,?,?,?,?,?,?,?) RETURNING jid",
		j.Jname, j.SID, j.Status, j.Processed, j.Total,
		bs, j.Mversion, j.CreateTime, j.StartTime, j.EndTime,
		j.Description,
	)
	if err != nil {
		return -1, err
	}
	jid := -1
	for row.Next() {
		if err := row.Scan(&jid); err != nil {
			return -1, err
		}
	}

	return jid, nil
}

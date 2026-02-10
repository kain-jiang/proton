package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/trait"
)

/*
++++++++++++ basic work component instance impl end++++++++++++++++
*/

func (c *SQLCursor) GetWorkComponentIns(ctx context.Context, sid int, com trait.ComponentNode) (cid int, err *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetWorkComponentInsStmt, sid, com.Name)
	err0 := row.Scan(&cid)
	return cid, err0
}

func (c *SQLCursor) LayoffComponentIns(ctx context.Context, cid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.LayoffComponentInsStmt, cid)
	return err
}

func (c *SQLCursor) WorkComponentIns(ctx context.Context, cins *trait.ComponentInstance) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.WorkComponentInsStmt,
		cins.CID, cins.System.SID, cins.Component.Name, cins.Component.Version,
		cins.CreateTime, cins.StartTime, cins.EndTime)
	if trait.IsInternalError(err, trait.ErrUniqueKey) {
		return nil
	}
	return err
}

func (tx *TX) WorkComponentIns(ctx context.Context, cins *trait.ComponentInstance) *trait.Error {
	stmt := tx.SQLCursor
	_, err := stmt.GetWorkComponentIns(ctx, cins.System.SID, cins.Component)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return stmt.WorkComponentIns(ctx, cins)
	}
	return nil
}

func (s *Store) WorkComponentIns(ctx context.Context, cins *trait.ComponentInstance) *trait.Error {
	return driver.StoreTransactionMarco(s, func(t driver.Transaction) *trait.Error {
		return s.beginWithTx(t).WorkComponentIns(ctx, cins)
	})
}

func (c *SQLCursor) ListWorkComponentIns(ctx context.Context, filter trait.WorkCompFilter) ([]*trait.ComponentInstanceMeta, *trait.Error) {
	if filter.Limit == 0 {
		filter.Limit = 100
	}
	rows, err := c.QueryContext(ctx,
		fmt.Sprintf(
			`SELECT 
			ins_id, ains_id, acid, sid, 
			cname, ctype, crtype, 
			version, aname, revission, 
			create_time, start_time, end_time 
			FROM task_component_instance 
			WHERE ins_id in (select ins_id from task_work_component) 
			AND aname=? 
			AND sid=? 
			LIMIT ? OFFSET %d;`,
			filter.Offset),
		filter.Aname, filter.Sid, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []*trait.ComponentInstanceMeta{}
	for rows.Next() {
		cins := &trait.ComponentInstanceMeta{}
		if err := rows.Scan(&cins.CID, &cins.AIID, &cins.Acid, &cins.System.SID, &cins.Component.Name,
			&cins.Component.ComponentDefineType, &cins.Component.Type, &cins.Component.Version,
			&cins.APPName, &cins.Revission,
			&cins.CreateTime, &cins.StartTime, &cins.EndTime); err != nil {
			return nil, err
		}
		res = append(res, cins)

	}
	return res, nil
}

/*
---------basic work component instance impl end-----------
*/

/*
+++++++++++++ basic component instance impl end ++++++++++++
*/

func (c *SQLCursor) UpdateAppInsComponentIns(ctx context.Context, coms []*trait.ComponentInstance) *trait.Error {
	stmt, err := c.PrepareContext(ctx, c.stmt.UpdateComponentInsStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, cins := range coms {
		configbs, err := json.Marshal(cins.Config)
		if err != nil {
			return &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   "encode component instance config before update",
			}
		}

		attrBS, err := json.Marshal(cins.Attribute)
		if err != nil {
			return &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   "encode component instance attribute before store",
			}
		}
		if _, err := stmt.ExecContext(ctx, cins.Status, configbs, attrBS, cins.Timeout, cins.CID); err != nil {
			return err
		}
	}
	return nil
}

func (c *SQLCursor) insertAppInsComponentIns(ctx context.Context, a *trait.ApplicationInstance) (err *trait.Error) {
	dbStmt, err := c.PrepareContext(ctx, c.stmt.InsertComponentInsStmt)
	if err != nil {
		return err
	}
	defer dbStmt.Close()
	for _, cins := range a.Components {
		configBs, err1 := json.Marshal(cins.Config)
		if err1 != nil {
			return &trait.Error{
				Internal: trait.ECNULL,
				Err:      err1,
				Detail:   "encode component instance config before store",
			}
		}
		attributeBs, err1 := json.Marshal(cins.Attribute)
		if err1 != nil {
			return &trait.Error{
				Internal: trait.ECNULL,
				Err:      err1,
				Detail:   "encode component instance attribute before store",
			}
		}
		cins.AIID = a.ID
		cins.System.SID = a.SID
		_, err = dbStmt.ExecContext(
			ctx, cins.AIID, cins.Acid, a.SID, cins.Component.Version, cins.APPName,
			cins.Component.Name, cins.Component.ComponentDefineType, cins.Component.Type,
			cins.Status, configBs, attributeBs, cins.Timeout, cins.Revission, a.CreateTime, a.StartTime, a.EndTime,
		)
		if err != nil {
			return err
		}
		r := c.QueryRowContext(
			ctx, c.stmt.GetInsertComponentInsStmt, cins.AIID, cins.Acid)
		if err = r.Scan(&cins.CID); err != nil {
			return
		}
	}
	return
}

func (c *SQLCursor) getAPPInsComponentIns(ctx context.Context, id int) (cis []*trait.ComponentInstance, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.GetAPPInsComponentInsStmt, id)
	if err0 != nil {
		return cis, err0
	}
	defer rows.Close()

	for rows.Next() {
		cins := &trait.ComponentInstance{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				AIID:      id,
				Component: trait.ComponentNode{},
			},
		}
		config := make([]byte, 0)
		attribute := make([]byte, 0)
		if err0 := rows.Scan(&cins.CID, &cins.Acid, &cins.System.SID, &cins.Component.Name,
			&cins.Component.ComponentDefineType, &cins.Component.Type, &cins.Component.Version,
			&cins.APPName, &cins.Status, &config, &attribute, &cins.Timeout, &cins.Revission,
			&cins.CreateTime, &cins.StartTime, &cins.EndTime); err0 != nil {
			return cis, err0
		}
		if err := json.Unmarshal(config, &cins.Config); err != nil {
			return nil, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err0,
				Detail:   "decode component instance config before store",
			}
		}
		if err := json.Unmarshal(attribute, &cins.Attribute); err != nil {
			return nil, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   "encode component instance attribute before store",
			}
		}
		cis = append(cis, cins)
	}
	return
}

func (c *SQLCursor) GetComponentIns(ctx context.Context, cid int) (cins *trait.ComponentInstance, err *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetComponentInsStmt, cid)
	cins = &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			Component: trait.ComponentNode{},
		},
	}
	config := make([]byte, 0)
	attribute := make([]byte, 0)
	if err0 := row.Scan(&cins.AIID, &cins.Acid, &cins.System.SID, &cins.Component.Name,
		&cins.Component.ComponentDefineType, &cins.Component.Type, &cins.Component.Version,
		&cins.APPName, &cins.Status, &config, &attribute, &cins.Timeout, &cins.Revission,
		&cins.CreateTime, &cins.StartTime, &cins.EndTime); err0 != nil {
		return nil, err0
	}
	if len(config) != 0 {
		if err0 := json.Unmarshal(config, &cins.Config); err0 != nil {
			return nil, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err0,
				Detail:   "decode component instance config from store",
			}
		}
	}

	if len(attribute) != 0 {
		if err0 := json.Unmarshal(attribute, &cins.Attribute); err0 != nil {
			return nil, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err0,
				Detail:   "encode component instance attribute from store",
			}
		}
	}

	cins.CID = cid
	return
}

/*
---------basic component instance impl end-----------
*/

// GetWorkComponentIns implt trait.Transaction
func (tx *TX) GetWorkComponentIns(ctx context.Context, sid int, cnode trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	stmt := tx.SQLCursor
	cid, err := stmt.GetWorkComponentIns(ctx, sid, cnode)
	if err != nil {
		return nil, err
	}
	return stmt.GetComponentIns(ctx, cid)
}

// GetWorkComponentIns imply trait.store
func (s *Store) GetWorkComponentIns(ctx context.Context, sid int, cnode trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	c, err := tx.GetWorkComponentIns(ctx, sid, cnode)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}
	return c, tx.Commit()
}

func (c *SQLCursor) getComponentLock(ctx context.Context, sid int, cname string) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetComponentLockStmt, sid, cname)
	jid := -1
	err := row.Scan(&jid)
	return jid, err
}

func (c *SQLCursor) LockComponent(ctx context.Context, sid int, jid int, cnode trait.ComponentNode) *trait.Error {
retry:
	_, err := c.ExecContext(ctx, c.stmt.LockComponentStmt, jid, sid, cnode.Name)
	if trait.IsInternalError(err, trait.ErrUniqueKey) {
		ljid, err := c.getComponentLock(ctx, sid, cnode.Name)
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

func (c *SQLCursor) UnlockComponent(ctx context.Context, sid, jid int, cnode trait.ComponentNode) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UnlockComponentStmt, jid, sid, cnode.Name)
	return err
}

func (c *SQLCursor) UnlockJobComponent(ctx context.Context, jid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.UnlockJobComponentStmt, jid)
	return err
}

func (tx *TX) UpdateComponentInsStatus(ctx context.Context, cid, status, revision, startTime, endTime int) *trait.Error {
	c := tx.SQLCursor
	_, err := c.ExecContext(ctx, c.stmt.UpdateComponentInsStatusStmt, status, startTime, endTime, revision+1, cid, revision)
	if err != nil {
		return err
	}
	row := c.QueryRowContext(ctx,
		`SELECT 
		status, start_time, end_time, revission 
		FROM task_component_instance 
		WHERE ins_id=?`,
		cid,
	)
	var sstatus, srevision, sstart, sendtime int
	err = row.Scan(&sstatus, &sstart, &sendtime, &srevision)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return &trait.Error{
			Internal: trait.ErrComponentInstanceRevission,
			Err:      fmt.Errorf("update component status which is nil instance"),
			Detail:   srevision,
		}
	} else if err != nil {
		return err
	}
	// cas, compare result intranction
	// 即使数据不是上一语句更改,但由于数据内容一致,操作幂等,因此也认为更改成功
	if status != sstatus ||
		startTime != sstart ||
		endTime != sendtime || revision+1 != srevision {
		return &trait.Error{
			Internal: trait.ErrComponentInstanceRevission,
			Err:      fmt.Errorf("input revision: %d, real updated revision: %d", revision, srevision),
			Detail:   srevision,
		}
	}

	return nil
}

func (s *Store) UpdateComponentInsStatus(ctx context.Context, cid, status, revision, startTime, endTime int) *trait.Error {
	return driver.StoreTransactionMarco(s, func(tx driver.Transaction) *trait.Error {
		return s.beginWithTx(tx).UpdateComponentInsStatus(ctx, cid, status, revision, startTime, endTime)
	})
}

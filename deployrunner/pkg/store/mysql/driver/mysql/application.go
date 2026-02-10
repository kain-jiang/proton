package store

import (
	"context"
	"encoding/json"
	"fmt"

	"taskrunner/trait"
)

/*
+++++++++ basic application impl +++++++++++++++++++++++++
*/

func (tx *TX) insertAPPComponents(ctx context.Context, a trait.Application) *trait.Error {
	stmt, err := tx.SQLCursor.PrepareContext(ctx, tx.stmt.InsertAPPComponentStmt)
	if err != nil {
		return err
	}

	for _, c := range a.Components() {
		meta := c.GetComponentMeta()
		_, err := stmt.ExecContext(ctx,
			a.AID, meta.Name, meta.Version, meta.ComponentDefineType,
			meta.Type, meta.DTimeout, meta.RawConfigSchema,
			meta.RawAttributeSchema, meta.Spec,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertAPP impl store insert a application
func (tx *TX) InsertAPP(ctx context.Context, a trait.Application) (int, *trait.Error) {
	if len(a.Version) > 128 {
		return -1, &trait.Error{
			Internal: trait.ErrParam,
			Detail:   fmt.Sprintf("version is too long, max length is 128, input length: %d", len(a.Version)),
		}
	}
	g, rerr := json.Marshal(a.Graph)
	if rerr != nil {
		return -1, &trait.Error{
			Internal: trait.ECNULL,
			Err:      rerr,
			Detail:   "decode application graph before store",
		}
	}
	dep, rerr := json.Marshal(a.Dependence)
	if rerr != nil {
		return -1, &trait.Error{
			Internal: trait.ECNULL,
			Err:      rerr,
			Detail:   "decode dependence before store",
		}
	}

	_, err := tx.ExecContext(
		ctx,
		tx.stmt.InsertAPPStmt, a.Type, a.Version,
		a.AName, a.ConfigSchema, g, dep)
	if err != nil {
		return -1, err
	}
	r := tx.QueryRowContext(ctx, tx.stmt.GetInsertAPPStmt, a.AName, a.Version)

	if err = r.Scan(&a.AID); err != nil {
		return -1, err
	}
	err = tx.insertAPPComponents(ctx, a)
	return a.AID, err
}

// InsertAPP impl store insert a application
func (s *Store) InsertAPP(ctx context.Context, a trait.Application) (int, *trait.Error) {
	// tx, err := s.begin(context.Background(), nil)
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	aid, err := tx.InsertAPP(ctx, a)
	if err != nil {
		if err0 := tx.Rollback(); err0 != nil {
			return -1, err0
		}
		return aid, err
	}
	return aid, tx.Commit()
}

func (c *SQLCursor) UpdateAppDependence(ctx context.Context, a trait.Application) *trait.Error {
	dep, rerr := json.Marshal(a.Dependence)
	if rerr != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      rerr,
			Detail:   "decode dependence before store",
		}
	}
	_, err := c.CursorConn.ExecContext(ctx, "UPDATE task_application SET adependence=? WHERE aname=? AND version=?", dep, a.AName, a.Version)
	return err
}

func (c *SQLCursor) DeleteAPP(ctx context.Context, aid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteAPPStmt, aid)
	return err
}

func (c *SQLCursor) DeleteAPPComponents(ctx context.Context, aid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteAPPComponentsStmt, aid)
	return err
}

func (c *SQLCursor) GetAPPComponent(ctx context.Context, cid int) (meta *trait.ComponentMeta, err *trait.Error) {
	meta = &trait.ComponentMeta{}
	var rawAttr, rawConfig []byte
	row := c.QueryRowContext(ctx, c.stmt.GetAPPComponentStmt, cid)
	err0 := row.Scan(
		&meta.CID, &meta.Name, &meta.Version,
		&meta.ComponentNode.ComponentDefineType, &meta.Type, &meta.DTimeout,
		&rawConfig, &rawAttr, &meta.Spec,
	)
	if len(rawAttr) != 0 {
		meta.RawAttributeSchema = rawAttr
	}

	if len(rawConfig) != 0 {
		meta.RawConfigSchema = rawConfig
	}

	err = err0
	return
}

func (c *SQLCursor) CountAPPRelateInsStmt(ctx context.Context, aid int) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.CountAPPRelateInsStmt, aid)
	count := 0
	err := row.Scan(&count)
	return count, err
}

// DeleteAPP impl store delete application
func (tx *TX) DeleteAPP(ctx context.Context, aid int) *trait.Error {
	count, err := tx.CountAPPRelateInsStmt(ctx, aid)
	if err != nil {
		return err
	}
	if count != 0 {
		return &trait.Error{
			Internal: trait.ErrApplicationStillUse,
			Err:      fmt.Errorf("thee application %d still relate with %d application instance, can't delete", aid, count),
			Detail:   count,
		}
	}
	if err := tx.DeleteAPPComponents(ctx, aid); err != nil {
		return err
	}
	err = tx.SQLCursor.DeleteAPP(ctx, aid)
	return err
}

// DeleteAPP impl store delete application
func (s *Store) DeleteAPP(ctx context.Context, aid int) *trait.Error {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return err
	}

	if err = tx.DeleteAPP(ctx, aid); err != nil {
		if err0 := tx.Rollback(); err0 != nil {
			return err0
		}
		return err
	}
	return tx.Commit()
}

func (c *SQLCursor) GetAPP(ctx context.Context, aid int) (a *trait.Application, err *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetAPPStmt, aid)
	var graph, schema, dep []byte
	a = &trait.Application{}
	if err := row.Scan(&a.Type, &a.Version, &a.AName, &schema, &graph, &dep); err != nil {
		return a, err
	}
	if len(schema) != 0 {
		a.ConfigSchema = schema
	}
	if rerr := json.Unmarshal(graph, &a.Graph); rerr != nil {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      rerr,
			Detail:   "decode application's graph error after get from store",
		}
	}
	if len(dep) != 0 {
		if rerr := json.Unmarshal(dep, &a.Dependence); rerr != nil {
			return nil, &trait.Error{
				Internal: trait.ECNULL,
				Err:      rerr,
				Detail:   "decode application's Dependence error after get from store",
			}
		}
	}

	a.AID = aid
	return
}

func (c *SQLCursor) getAPPComponents(ctx context.Context, aid int) (cs []*trait.ComponentMeta, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.GetAPPComponentsStmt, aid)
	if err0 != nil {
		err = err0
		return
	}
	defer rows.Close()

	for rows.Next() {
		c := &trait.ComponentMeta{}
		var rawAttr, rawConfig []byte
		if err0 := rows.Scan(
			&c.CID, &c.Name, &c.Version,
			&c.ComponentNode.ComponentDefineType, &c.Type, &c.DTimeout,
			&rawConfig, &rawAttr, &c.Spec,
		); err0 != nil {
			err = err0
			return
		}
		if len(rawAttr) != 0 {
			c.RawAttributeSchema = rawAttr
		}
		if len(rawConfig) != 0 {
			c.RawConfigSchema = rawConfig
		}
		cs = append(cs, c)
	}
	return
}

// GetAPP impl store get the app and it's component
func (tx *TX) GetAPP(ctx context.Context, aid int) (a *trait.Application, err *trait.Error) {
	a, err = tx.SQLCursor.GetAPP(ctx, aid)
	if err != nil {
		return
	}
	cs, err0 := tx.SQLCursor.getAPPComponents(ctx, aid)
	if err0 != nil {
		err = err0
		return
	}
	a.Component = cs
	return
}

// GetAPP impl store get the app and it's component
func (s *Store) GetAPP(ctx context.Context, aid int) (a *trait.Application, err *trait.Error) {
	tx, err := s.begin(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	a, err = tx.GetAPP(ctx, aid)
	if err != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return
	}
	return a, tx.Commit()
}

func (c *SQLCursor) SearchAPP(ctx context.Context, limit int, lastAID int, name string) (as []trait.ApplicationMeta, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.SearchAPPStmt, name, lastAID, limit)
	if err0 != nil {
		return as, err0
	}
	defer rows.Close()

	for rows.Next() {
		a := trait.ApplicationMeta{}
		a.AName = name
		if err0 := rows.Scan(&a.AID, &a.Version); err0 != nil {
			return as, err0
		}
		as = append(as, a)
	}
	return
}

func (c *SQLCursor) ListAPP(ctx context.Context, limit int, lastAID int) (as []trait.ApplicationMeta, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.ListAPPStmt, lastAID, limit)
	if err0 != nil {
		return as, err0
	}
	defer rows.Close()

	for rows.Next() {
		a := trait.ApplicationMeta{}
		if err0 := rows.Scan(&a.AID, &a.AName); err0 != nil {
			return as, err0
		}
		as = append(as, a)
	}
	return
}

func (c *SQLCursor) ListSystemAPPNoWorked(ctx context.Context, limit int, lastAID int, sid int) (as []trait.ApplicationMeta, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.ListSystemAPPNoWorked, lastAID, sid, limit)
	if err0 != nil {
		return as, err0
	}
	defer rows.Close()
	as = make([]trait.ApplicationMeta, 0, 5)

	for rows.Next() {
		a := trait.ApplicationMeta{}
		if err0 := rows.Scan(&a.AID, &a.AName); err0 != nil {
			return as, err0
		}
		as = append(as, a)
	}
	return
}

func (c *SQLCursor) GetAPPID(ctx context.Context, aname, aversion string) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetAPPIDStmt, aname, aversion)
	aid := -1
	err := row.Scan(&aid)
	return aid, err
}

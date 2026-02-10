package store

import (
	"context"

	"taskrunner/trait"
)

func (c *SQLCursor) AddEdge(ctx context.Context, from, to int) *trait.Error {
	// TODO ignore duplicate
	_, err := c.ExecContext(ctx, c.stmt.AddEdgeStmt, from, to)
	return err
}

func (s *Store) AddOuterChildEdge(ctx context.Context, from int, sid int, com trait.ComponentNode) *trait.Error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	err = tx.AddOuterChildEdge(ctx, from, sid, com)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (tx *TX) AddOuterChildEdge(ctx context.Context, from int, sid int, com trait.ComponentNode) *trait.Error {
	stmt := tx.SQLCursor
	to := -1
	var err *trait.Error
	// retry only onece when the work compnent may been exchange by other job
	for i := 0; i < 2; i++ {
		to, err = stmt.GetWorkComponentIns(ctx, sid, com)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			if i == 0 {
				err = nil
			} else {
				// no need to add edge which child has been deleted
				return nil
			}
		}
		if err != nil {
			return err
		}
	}
	return tx.AddEdge(ctx, from, to)
}

func (tx *TX) AddEdge(ctx context.Context, from, to int) *trait.Error {
	stmt := tx.SQLCursor
	row := stmt.QueryRowContext(ctx, stmt.stmt.GetEdgeStmt, from, to)
	count := 0
	if err := row.Scan(&count); err != nil {
		return err
	}

	// compare and insert
	if count == 0 {
		return stmt.AddEdge(ctx, from, to)
	}
	return nil
}

func (s *Store) AddEdge(ctx context.Context, from, to int) *trait.Error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	err = tx.AddEdge(ctx, from, to)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (c *SQLCursor) DeleteEdge(ctx context.Context, from, to int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteEdgeStmt, from, to)
	return err
}

func (c *SQLCursor) ChangeEdgeto(ctx context.Context, curCID, tarCID int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.ChangeEdgeToStmt, tarCID, curCID)
	// TX's ChangeEdgeTo func need raw error, err wrrapper by TX
	return err
}

func (c *SQLCursor) deleteEdgeto(ctx context.Context, cid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteEdgetoStmt, cid)
	return err
}

func (c *SQLCursor) DeleteEdgeFrom(ctx context.Context, cid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteEdgeFromStmt, cid)
	return err
}

func (c *SQLCursor) GetPointFrom(ctx context.Context, from int) (ids []int, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.GetPointFromStmt, from)
	if err0 != nil {
		err = err0
		return
	}
	for rows.Next() {
		cto := -1
		if err0 := rows.Scan(&cto); err0 != nil {
			err = err0
			return
		}
		ids = append(ids, cto)
	}
	return
}

func (c *SQLCursor) GetPointTo(ctx context.Context, to int) (ids []int, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.GetPointToStmt, to)
	if err0 != nil {
		err = err0
		return
	}
	for rows.Next() {
		cto := -1
		if err0 := rows.Scan(&cto); err0 != nil {
			err = err0
			return
		}
		ids = append(ids, cto)
	}
	return
}

func (c *SQLCursor) GetChangeEdgeToConflictStmt(ctx context.Context, cur, tar int) (ids [][2]int, err *trait.Error) {
	rows, err0 := c.QueryContext(ctx, c.stmt.GetChangeEdgeToConflictStmt, tar, cur)
	if err0 != nil {
		err = err0
		return
	}
	for rows.Next() {
		cfrom := -1
		cto := -1
		if err0 := rows.Scan(&cfrom, &cto); err0 != nil {
			err = err0
			return
		}
		ids = append(ids, [2]int{cfrom, cto})
	}
	return
}

func (c *SQLCursor) CountEdgeTo(ctx context.Context, cid int) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.CountEdgeToStmt, cid)
	count := 0
	err := row.Scan(&count)
	return count, err
}

// ChangeEdgeto change edge point to
func (tx *TX) ChangeEdgeto(ctx context.Context, curCID, tarCID int) *trait.Error {
	conflict, err := tx.GetChangeEdgeToConflictStmt(ctx, curCID, tarCID)
	if err != nil {
		return err
	}
	stmt, err := tx.SQLCursor.PrepareContext(ctx, tx.stmt.DeleteEdgeStmt)
	if err != nil {
		return err
	}
	for _, edge := range conflict {
		if _, err := stmt.ExecContext(ctx, edge[0], edge[1]); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, tx.stmt.ChangeEdgeToStmt, tarCID, curCID)
	return err
}

// ChangeEdgeFrom change edge point from
func (tx *TX) ChangeEdgeFrom(ctx context.Context, curCID, tarCID int) *trait.Error {
	_, err := tx.ExecContext(ctx, tx.stmt.ChangeEdgeFromStmt, tarCID, curCID)
	return err
}

// ChangeEdgeFrom change edge point from
func (s *Store) ChangeEdgeFrom(ctx context.Context, curCID, tarCID int) *trait.Error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	err = tx.ChangeEdgeFrom(ctx, curCID, tarCID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

// ChangeEdgeto change edge point to
func (s *Store) ChangeEdgeto(ctx context.Context, curCID, tarCID int) *trait.Error {
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	err = tx.ChangeEdgeto(ctx, curCID, tarCID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

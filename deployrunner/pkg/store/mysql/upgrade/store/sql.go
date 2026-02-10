package store

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"

	"taskrunner/pkg/component/resources"
	driver "taskrunner/pkg/sql-driver"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	ctrait "taskrunner/trait"

	"github.com/ghodss/yaml"
)

//go:embed statement
var initTablesDir embed.FS

var initStatement map[string]trait.Plan

// SQLExeuteArgs simple sql args
type SQLExeuteArgs struct {
	Statements []string `json:"statements"`
}

func init() {
	root := "statement"
	fs, err := initTablesDir.ReadDir(root)
	if err != nil {
		panic(err.Error())
	}
	initStatement = make(map[string]trait.Plan)
	for _, dbDir := range fs {
		dbtype := dbDir.Name()
		bs, err := initTablesDir.ReadFile(filepath.Join(root, dbtype, "1733196739_0_初始化升级进度表.yaml"))
		if err != nil {
			panic(err.Error())
		}
		p := trait.Plan{}
		if err := yaml.Unmarshal(bs, &p.Operators); err != nil {
			panic(err)
		}
		initStatement[driver.ConvertDBType(dbtype)] = p
	}
}

type store struct {
	driver.DBConn
}

func NewStore(ctx context.Context, rds resources.RDS) (Store, trait.Error) {
	conn, err := driver.Factory.NewDBConn(ctx, rds)
	if err != nil {
		return nil, err
	}
	s := &store{
		DBConn: conn,
	}
	return s, s.initTable(ctx, rds.Type)
}

func (s *store) initTable(ctx context.Context, dbtype string) trait.Error {
	p, ok := initStatement[driver.ConvertDBType(dbtype)]
	if !ok {
		return &ctrait.Error{
			Internal: ctrait.ErrNotFound,
			Detail:   fmt.Sprintf("%s database not support", dbtype),
		}
	}
	for _, i := range p.Operators {
		op := &SQLExeuteArgs{}
		if err := json.Unmarshal(i.Args, &op); err != nil {
			return &ctrait.Error{
				Internal: ctrait.ErrParam,
				Detail:   "升级进度表初始化脚本解码错误",
				Err:      err,
			}
		}

		for _, j := range op.Statements {
			if _, err := s.ExecContext(ctx, j); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *store) Record(ctx context.Context, p trait.PlanProcess) trait.Error {
	tx, err := s.DBConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	row := tx.QueryRowContext(
		ctx,
		"SELECT dateid from task_sqldata_upgrade WHERE dateid=? AND svcname=?;",
		p.DateID, p.ServiceName)
	did := -1
	if err = row.Scan(&did); ctrait.IsInternalError(err, ctrait.ErrNotFound) {
		err = s.insert(ctx, tx, p)
	} else if err != nil {
		return err
	} else {
		err = s.update(ctx, tx, p)
	}

	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *store) insert(ctx context.Context, cursor driver.CursorConn, p trait.PlanProcess) trait.Error {
	_, err := cursor.ExecContext(ctx,
		`INSERT INTO task_sqldata_upgrade 
		(dateid, epoch, porder, status, stage, 
		oporder, opstatus, tempdata, svcname)
		VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?);`,
		p.DateID, p.Epoch, p.Order, p.Status, p.Stage,
		p.Op.OrderID, p.Op.Status, p.Op.TempData, p.ServiceName,
	)
	return err
}

func (s *store) update(ctx context.Context, cursor driver.CursorConn, p trait.PlanProcess) trait.Error {
	_, err := cursor.ExecContext(ctx,
		`UPDATE task_sqldata_upgrade SET
		porder=?, status=?, oporder=?, opstatus=?, tempdata=? 
		WHERE dateid=? AND svcname=?;`,
		p.Order, p.Status, p.Op.OrderID, p.Op.Status, p.Op.TempData,
		p.DateID, p.ServiceName,
	)
	return err
}

func (s *store) Last(ctx context.Context, svcName string, stage int) (trait.PlanProcess, trait.Error) {
	row := s.QueryRowContext(
		ctx,
		`SELECT 
		dateid, porder from task_sqldata_upgrade 
		WHERE stage=? AND svcname=? 
		ORDER BY dateid DESC LIMIT 1;`,
		stage, svcName)
	p := trait.PlanProcess{}
	err := row.Scan(&p.DateID, &p.Order)
	return p, err
}

func (s *store) Less(ctx context.Context, svcName string, DateID, limit int, stage int) ([]trait.PlanProcess, trait.Error) {
	rows, err := s.QueryContext(
		ctx,
		`SELECT 
		dateid, porder from task_sqldata_upgrade 
		WHERE stage=? AND svcname=? AND dateid < ? 
		ORDER BY dateid DESC LIMIT ?;`,
		stage, svcName, DateID, limit)
	if err != nil {
		return nil, err
	}
	ps := []trait.PlanProcess{}
	for rows.Next() {
		p := trait.PlanProcess{}
		if err := rows.Scan(&p.DateID, &p.Order); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, err
}

func (s *store) Get(ctx context.Context, svcName string, DateID int) (*trait.PlanProcess, trait.Error) {
	row := s.QueryRowContext(
		ctx,
		`SELECT 
		dateid, epoch, porder, status, stage, 
		oporder, opstatus, tempdata, svcname 
		from task_sqldata_upgrade 
		WHERE dateid=? AND svcname=?;`,
		DateID, svcName)
	p := &trait.PlanProcess{}
	var bs []byte
	err := row.Scan(
		&p.DateID, &p.Epoch, &p.Order, &p.Status, &p.Stage,
		&p.Op.OrderID, &p.Op.Status, &bs,
		&p.ServiceName,
	)
	p.Op.TempData = bs
	return p, err
}

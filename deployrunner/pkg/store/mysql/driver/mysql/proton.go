package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"taskrunner/trait"
)

func getProtonComponentFilterStmt(cname string, ctype string, sid int) (condition string, args []any) {
	conditions := make([]string, 0, 3)

	if cname != "" {
		args = append(args, cname)
		conditions = append(conditions, "c.cname=?")
	}

	if ctype != "" {
		args = append(args, ctype)
		conditions = append(conditions, "c.ctype=?")

	}

	if sid != -1 {
		args = append(args, sid)
		conditions = append(conditions, "c.sid=?")
	}

	if len(args) != 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args
	}
	return "", nil
}

func (c *SQLCursor) CountProtonConponent(ctx context.Context, cname string, ctype string, sid int) (int, *trait.Error) {
	condition, args := getProtonComponentFilterStmt(cname, ctype, sid)
	r := c.QueryRowContext(ctx, fmt.Sprintf("SELECT count(*) FROM task_proton_component as c %s", condition), args...)
	count := 0
	if err := r.Scan(&count); err != nil {
		return -1, err
	}
	return count, nil
}

func (c *SQLCursor) ListProtonConponentWithInternal(ctx context.Context, cname string, ctype string, sid int, limit, offset int) (cs []trait.ProtonComponentMeta, err *trait.Error) {
	cs = []trait.ProtonComponentMeta{}
	condition, args := getProtonComponentFilterStmt(cname, ctype, sid)
	args = append(args, limit)
	rows, err := c.QueryContext(ctx, fmt.Sprintf("select c.ctype, c.cname, c.sid, s.sname, c.coptions from task_proton_component c left join task_system  s on c.sid=s.sid %s LIMIT ? offset %d;", condition, offset), args...)
	if err != nil {
		return cs, err
	}
	defer rows.Close()

	for rows.Next() {
		bs := []byte{}
		a := trait.ProtonComponentMeta{}
		var sname sql.NullString
		if err := rows.Scan(
			&a.Type, &a.Name, &a.SID, &sname, &bs,
		); err != nil {
			return cs, err
		}
		a.SName = sname.String
		if len(bs) != 0 {
			// info与rls的绑定关系
			// 理论上一个rls可以绑定到不同的info
			// info的不同属性与同一rls的组合向使用者提供不同形态能力，如使用同一数据库实例但隔离
			// 在此种理论关系下，rls将不与info形成1:1对应关系
			options := []string{}
			if rerr := json.Unmarshal(bs, &options); rerr != nil {
				return nil, &trait.Error{
					Internal: trait.ErrComponentDecodeError,
					Detail:   string(bs),
					Err:      rerr,
				}
			}
			if len(options) > 0 {
				a.Instance = options
			}
		}
		cs = append(cs, a)
	}
	return
}

func (c *SQLCursor) ListProtonConponent(ctx context.Context, cname string, ctype string, sid int, limit, offset int) (cs []trait.ProtonComponentMeta, err *trait.Error) {
	cs = []trait.ProtonComponentMeta{}
	condition, args := getProtonComponentFilterStmt(cname, ctype, sid)
	args = append(args, limit)
	rows, err := c.QueryContext(ctx, fmt.Sprintf("select c.ctype, c.cname, c.sid, s.sname from task_proton_component c left join task_system  s on c.sid=s.sid %s LIMIT ? offset %d;", condition, offset), args...)
	if err != nil {
		return cs, err
	}
	defer rows.Close()

	for rows.Next() {
		a := trait.ProtonComponentMeta{}
		var sname sql.NullString
		if err := rows.Scan(
			&a.Type, &a.Name, &a.SID, &sname,
		); err != nil {
			return cs, err
		}
		a.SName = sname.String
		cs = append(cs, a)
	}
	return
}

func (c *SQLCursor) InsertProtonComponent(ctx context.Context, obj trait.ProtonCompoent) *trait.Error {
	_, err := c.ExecContext(
		ctx, c.stmt.InsertProtonComponent,
		obj.Name, obj.Type, obj.SID,
		obj.Attribute, obj.Options)
	return err
}

func (c *SQLCursor) GetProtonComponent(ctx context.Context, cname string, ctype string, sid int) (*trait.ProtonCompoent, *trait.Error) {
	condition, args := getProtonComponentFilterStmt(cname, ctype, sid)
	row := c.QueryRowContext(ctx, fmt.Sprintf("SELECT c.ctype, c.cname, c.sid, c.cattribute, c.coptions FROM task_proton_component as c %s;", condition), args...)
	obj := &trait.ProtonCompoent{}
	err := row.Scan(&obj.Type, &obj.Name, &obj.SID, &obj.Attribute, &obj.Options)
	if len(obj.Attribute) == 0 {
		obj.Attribute = nil
	}
	if len(obj.Options) == 0 {
		obj.Options = nil
	}
	return obj, err
}

func (c *SQLCursor) UpdateProtonComponent(ctx context.Context, obj trait.ProtonCompoent) *trait.Error {
	_, err := c.ExecContext(
		ctx, c.stmt.UpdateProtonComponent,
		obj.Attribute, obj.Options, obj.SID, obj.Name, obj.Type)
	return err
}

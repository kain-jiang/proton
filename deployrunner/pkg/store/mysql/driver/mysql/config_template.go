package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"taskrunner/trait"

	"github.com/Masterminds/semver/v3"
)

// VersioninToNum 语义化版本号转数字
func VersioninToNum(v string) (int, *trait.Error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		// version like 7.0.5.6.1234 need to 7.5.6
		err := &trait.Error{
			Internal: trait.ErrParam,
			Err:      err,
			Detail:   fmt.Errorf("parse version '%s'", v),
		}
		vs := strings.Split(v, ".")
		if len(vs) < 4 {
			return -1, err
		}
		copy(vs[1:], vs[2:4])
		vs = vs[:3]
		v = strings.Join(vs, ".")
		sv0, err0 := semver.NewVersion(v)
		if err0 != nil {
			return -1, err
		}
		sv = sv0
	}

	return int(sv.Patch() + sv.Minor()*1000 + sv.Major()*100000), nil
}

func getVersionRange(v string) (min int, max int, err *trait.Error) {
	v = strings.Trim(v, "")
	if v == "" {
		err = &trait.Error{Internal: trait.ErrParam, Err: fmt.Errorf("aversion is invalid"), Detail: v}
		return
	}
	version := v
	sub := 0
	switch v[0] {
	case '~':
		sub = 1000 - 1
		version = v[1:]
	case '^':
		sub = 100000 - 1
		version = v[1:]
	}

	switch version[0] {
	case 'v':
		version = version[1:]
	}

	min, err = VersioninToNum(version)
	if err != nil {
		return
	}
	max = min + sub
	return
}

func getConfigTemplateFilterStmt(f trait.ApplicationConfigTemplateFilter, fields string) (string, []any, *trait.Error) {
	if f.Aname == "" {
		if f.ApplicationVersionFilter != nil || f.ApplicationLabelFilter != nil {
			return "", nil, &trait.Error{Internal: trait.ErrParam, Err: fmt.Errorf("ApplicationConfigTemplateFilter.Aname must set when use other filter")}
		}
		return "FROM task_app_config_template ct", nil, nil
	}

	// bytes缓存写字符串不需要做异常处理，当出现异常时底层会抛出panic
	buf := bytes.NewBufferString("")
	args := []any{}

	if f.ApplicationLabelFilter != nil && len(f.ApplicationLabelFilter.Labels) != 0 {
		buf.WriteString("FROM task_app_config_template_index cti JOIN ( SELECT ")
		buf.WriteString(fields)
	}

	buf.WriteString(" FROM task_app_config_template ct WHERE ")

	if f.ApplicationVersionFilter != nil {
		v, err := VersioninToNum(f.ApplicationVersionFilter.Aversion)
		if err != nil {
			return "", nil, err
		}
		switch f.ApplicationVersionFilter.Type {
		case 0, 1:
			buf.WriteString("ct.minversion=? ")
			args = append(args, v)
		case 2:
			buf.WriteString("(ct.minversion<=? AND ct.maxversion>=?) ")
			args = append(args, v, v)
		case 3:
			buf.WriteString("ct.minversion<=? ")
			args = append(args, v)
		default:
			return "", nil, &trait.Error{Internal: trait.ErrParam, Err: fmt.Errorf("ApplicationConfigTemplateFilter.versionFilter.type don't support values '%d'", f.ApplicationVersionFilter.Type)}
		}
	}

	if len(args) != 0 {
		buf.WriteString("AND ")
	}
	buf.WriteString("ct.aname=? ")
	args = append(args, f.Aname)

	if f.ApplicationLabelFilter != nil && len(f.ApplicationLabelFilter.Labels) != 0 {
		buf.WriteString(" ) ct ON ct.tid=cti.tid WHERE ")
		condition := "OR"

		buf.WriteRune('(')
		for i, l := range f.ApplicationLabelFilter.Labels {
			if i == 0 {
				buf.WriteString("cti.label=? ")
			} else {
				buf.WriteString(condition)
				buf.WriteString(" cti.label=? ")
			}
			args = append(args, l)
		}

		buf.WriteString(") GROUP BY ")
		buf.WriteString(fields)

		switch f.ApplicationLabelFilter.Condition {
		case 0, 1:
		case 2:
			buf.WriteString(fmt.Sprintf(" HAVING COUNT(cti.label)=%d", len(f.ApplicationLabelFilter.Labels)))
		default:
			return "", nil, &trait.Error{Internal: trait.ErrParam, Err: fmt.Errorf("ApplicationConfigTemplateFilter.labelFilter.Condition don't support values '%d'", f.ApplicationLabelFilter.Condition)}
		}

	}
	return buf.String(), args, nil
}

func (c *SQLCursor) CountConfigTempalte(ctx context.Context, f trait.ApplicationConfigTemplateFilter) (int, *trait.Error) {
	condition, args, err := getConfigTemplateFilterStmt(f, "ct.tid")
	if err != nil {
		return -1, err
	}

	// fmt.Printf(c.stmt.CountConfigTemplate+"\n"+"%#v\n", condition, args)
	r := c.QueryRowContext(ctx, fmt.Sprintf(c.stmt.CountConfigTemplate, condition), args...)
	count := 0
	if err := r.Scan(&count); err != nil {
		return -1, err
	}
	return count, nil
}

func (c *SQLCursor) ListConfigTemplate(ctx context.Context, f trait.ApplicationConfigTemplateFilter, limit, offset int) (cs []trait.AppliacationConfigTemplateMeta, err *trait.Error) {
	cs = []trait.AppliacationConfigTemplateMeta{}
	condition, args, err := getConfigTemplateFilterStmt(f, c.stmt.ListConfigTemplateFields)
	if err != nil {
		return nil, err
	}
	args = append(args, limit)
	// fmt.Printf(c.stmt.ListConfigTemplate+"\n"+"%#v\n", c.stmt.ListConfigTemplateFields, condition, offset, args)
	rows, err := c.QueryContext(ctx, fmt.Sprintf(c.stmt.ListConfigTemplate, c.stmt.ListConfigTemplateFields, condition, offset), args...)
	if err != nil {
		return cs, err
	}
	defer rows.Close()

	for rows.Next() {
		bs := []byte{}
		a := trait.AppliacationConfigTemplateMeta{}
		if err := rows.Scan(
			&a.Tid, &a.Tversion, &a.Tname,
			&a.Tdescription, &a.Aname, &a.Aversion, &bs,
		); err != nil {
			return cs, err
		}
		if rerr := json.Unmarshal(bs, &a.Labels); rerr != nil {
			err = &trait.Error{Internal: trait.ErrParam, Err: rerr, Detail: "labels decode error"}
			return
		}
		cs = append(cs, a)
	}
	return
}

func (c *SQLCursor) InsertConfigTempalte(ctx context.Context, cfg trait.AppliacationConfigTemplate) *trait.Error {
	meta := cfg.AppliacationConfigTemplateMeta
	config, err0 := json.Marshal(cfg.Config)
	if err0 != nil {
		return &trait.Error{Internal: trait.ErrParam, Err: err0, Detail: "config encode error"}
	}
	labels, err0 := json.Marshal(cfg.Labels)
	if err0 != nil {
		return &trait.Error{Internal: trait.ErrParam, Err: err0, Detail: "labels encode error"}
	}

	min, max, err := getVersionRange(meta.Aversion)
	if err != nil {
		return err
	}

	_, err = c.ExecContext(
		ctx, c.stmt.InsertConfigTemplate,
		meta.Tversion, meta.Tname, meta.Tdescription,
		meta.Aname, meta.Aversion, config, labels, min, max)
	return err
}

func (c *SQLCursor) GetConfigTemplateID(ctx context.Context, tversion, tname, aname string) (tid int, err *trait.Error) {
	row := c.QueryRowContext(ctx, c.stmt.GetConfigTemplateID, tversion, tname, aname)
	err = row.Scan(&tid)
	return
}

func (c *SQLCursor) UpdateConfigTemplate(ctx context.Context, cfg trait.AppliacationConfigTemplate) *trait.Error {
	meta := cfg.AppliacationConfigTemplateMeta
	config, err0 := json.Marshal(cfg.Config)
	if err0 != nil {
		return &trait.Error{Internal: trait.ErrParam, Err: err0, Detail: "config encode error"}
	}
	labels, err0 := json.Marshal(cfg.Labels)
	if err0 != nil {
		return &trait.Error{Internal: trait.ErrParam, Err: err0, Detail: "labels encode error"}
	}

	min, max, err := getVersionRange(meta.Aversion)
	if err != nil {
		return err
	}

	_, err = c.ExecContext(
		ctx, c.stmt.UpdateConfigTemplate,
		meta.Tdescription, meta.Aversion,
		config, labels, min, max,
		meta.Tid)
	return err
}

func (c *SQLCursor) DeleteConfigTemplate(ctx context.Context, tid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteConfigTemplate, tid)
	return err
}

func (c *SQLCursor) GetConfigTemplate(ctx context.Context, tid int) (cfg *trait.AppliacationConfigTemplate, err *trait.Error) {
	cfg = &trait.AppliacationConfigTemplate{}
	meta := cfg.AppliacationConfigTemplateMeta
	var config, labels []byte
	row := c.QueryRowContext(ctx, c.stmt.GetConfigTemplate, tid)
	err = row.Scan(
		&meta.Tversion, &meta.Tname, &meta.Tdescription,
		&meta.Aname, &meta.Aversion, &config, &labels,
	)
	if err != nil {
		return
	}
	if err0 := json.Unmarshal(config, &cfg.Config); err0 != nil {
		err = &trait.Error{Internal: trait.ErrParam, Err: err0, Detail: "config decode error"}
		return
	}

	if err0 := json.Unmarshal(labels, &meta.Labels); err0 != nil {
		err = &trait.Error{Internal: trait.ErrParam, Err: err0, Detail: "labels decode error"}
		return
	}
	cfg.AppliacationConfigTemplateMeta = meta
	return
}

func (c *SQLCursor) InsertConfigTemplateIndex(ctx context.Context, tid int, labels []string) *trait.Error {
	stmt, err := c.PrepareContext(ctx, c.stmt.InsertConfigTemplateIndex)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, l := range labels {
		if _, err := stmt.ExecContext(ctx, tid, l); err != nil {
			return err
		}
	}
	return stmt.Close()
}

func (c *SQLCursor) DeleteConfigTemplateIndex(ctx context.Context, tid int) *trait.Error {
	_, err := c.ExecContext(ctx, c.stmt.DeleteConfigTemplateIndex, tid)
	return err
}

func (tx *TX) insertConfigTempalte(ctx context.Context, cfg trait.AppliacationConfigTemplate) (id int, err *trait.Error) {
	stmt := tx.SQLCursor
	err = stmt.InsertConfigTempalte(ctx, cfg)
	if err != nil {
		return -1, err
	}
	id, err = stmt.GetConfigTemplateID(ctx, cfg.Tversion, cfg.Tname, cfg.Aname)
	if err != nil {
		return id, err
	}

	cfg.Tid = id
	err = tx.updateConfigTemplateIndex(ctx, cfg)
	return
}

func (tx *TX) updateConfigTemplateIndex(ctx context.Context, cfg trait.AppliacationConfigTemplate) *trait.Error {
	stmt := tx.SQLCursor
	if err := stmt.DeleteConfigTemplateIndex(ctx, cfg.Tid); err != nil {
		return err
	}
	return stmt.InsertConfigTemplateIndex(ctx, cfg.Tid, cfg.Labels)
}

func (s *Store) InsertConfigTempalte(ctx context.Context, cfg trait.AppliacationConfigTemplate) (id int, err *trait.Error) {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return -1, err
	}
	jid, err := tx.InsertConfigTempalte(ctx, cfg)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return -1, err
		}
		return -1, err
	}
	return jid, tx.Commit()
}

// InsertAPPIns impl trait.APPlicationInsWriter
func (tx *TX) InsertConfigTempalte(ctx context.Context, cfg trait.AppliacationConfigTemplate) (id int, err *trait.Error) {
	stmt := tx.SQLCursor
	tid, err := stmt.GetConfigTemplateID(ctx, cfg.Tversion, cfg.Tname, cfg.Aname)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return tx.insertConfigTempalte(ctx, cfg)
	}
	if err != nil {
		return -1, err
	}
	cfg.Tid = tid
	return tid, tx.UpdateConfigTemplate(ctx, cfg)
}

func (s *Store) UpdateConfigTemplate(ctx context.Context, cfg trait.AppliacationConfigTemplate) *trait.Error {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return err
	}
	err = tx.UpdateConfigTemplate(ctx, cfg)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (tx *TX) UpdateConfigTemplate(ctx context.Context, cfg trait.AppliacationConfigTemplate) *trait.Error {
	stmt := tx.SQLCursor
	if err := stmt.UpdateConfigTemplate(ctx, cfg); err != nil {
		return err
	}
	return tx.updateConfigTemplateIndex(ctx, cfg)
}

func (s *Store) DeleteConfigTemplate(ctx context.Context, tid int) *trait.Error {
	tx, err := s.begin(ctx, nil)
	if err != nil {
		return err
	}
	err = tx.DeleteConfigTemplate(ctx, tid)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (tx *TX) DeleteConfigTemplate(ctx context.Context, tid int) *trait.Error {
	stmt := tx.SQLCursor
	if err := stmt.DeleteConfigTemplateIndex(ctx, tid); err != nil {
		return err
	}
	return stmt.DeleteConfigTemplate(ctx, tid)
}

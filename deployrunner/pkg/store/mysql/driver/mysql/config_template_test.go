package store_test

import (
	"context"
	"fmt"
	"testing"

	"taskrunner/test"
	"taskrunner/trait"

	store "taskrunner/pkg/store/mysql/driver/mysql"
)

func TestConfigTempalte(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	cfg := trait.AppliacationConfigTemplate{
		AppliacationConfigTemplateMeta: trait.AppliacationConfigTemplateMeta{
			Aversion: "~2.12.0-123",
			Aname:    "test",
		},
	}
	// insert
	tid, err := s.InsertConfigTempalte(ctx, cfg)
	tt.AssertNil(err)

	// insert into update when conflict
	cfg.Tid = tid
	tid0, err := s.InsertConfigTempalte(ctx, cfg)
	tt.AssertNil(err)
	tt.Assert(tid, tid0)

	// update
	cfg.Aversion = "^2.1.12-qwe"
	cfg.Tdescription = "qwe79123"
	cfg.Config.AppConfig = map[string]interface{}{"test": float64(123)}
	cfg.Labels = append(cfg.Labels, "test0", "test1")
	err = s.UpdateConfigTemplate(ctx, cfg)
	tt.AssertNil(err)

	// get
	cfg0, err := s.GetConfigTemplate(ctx, tid)
	tt.AssertNil(err)
	for i, j := range cfg0.Labels {
		tt.Assert(cfg.Labels[i], j)
	}
	tt.Assert(cfg.Tdescription, cfg0.Tdescription)
	tt.Assert(cfg.Aversion, cfg0.Aversion)
	tt.Assert(cfg.Config.AppConfig["test"], cfg0.Config.AppConfig["test"])

	// delete
	err = s.DeleteConfigTemplate(ctx, tid)
	tt.AssertNil(err)
}

func TestVersionParse(t *testing.T) {
	tt := test.TestingT{T: t}
	versions := []string{
		"v2.17.0",
		"2.17.0",
		"v2.17.0-alpha",
		"2.17.0-alpha",
		"v2.17.0-alpha.0.7+12321.2024",
		"2.17.0-alpha.0.7+12321.2024",
		"v7.0.5.6",
		"7.0.5.6",
		"v7.0.5.6.123.456",
	}
	for _, v := range versions {
		_, err := store.VersioninToNum(v)
		tt.AssertNil(err)
	}
}

func TestConfigTemplateFilter(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	cfg := trait.AppliacationConfigTemplate{
		AppliacationConfigTemplateMeta: trait.AppliacationConfigTemplateMeta{
			Aversion: "~2.12.1-123-qwe",
			Aname:    "test",
		},
	}
	cfg.Tdescription = "qwe79123"
	cfg.Config.AppConfig = map[string]interface{}{"test": 123}
	cfg.Labels = append(cfg.Labels, "test0", "test1")

	cfg0 := trait.AppliacationConfigTemplate{
		AppliacationConfigTemplateMeta: trait.AppliacationConfigTemplateMeta{
			Aversion: "~2.12.0-123-qwe",
			Aname:    "test",
			Tname:    "test0",
		},
	}
	cfg0.Tdescription = "qwe79123"
	cfg0.Config.AppConfig = map[string]interface{}{"test": 123}
	cfg0.Labels = append(cfg0.Labels, "test0", "test2")
	cfg1 := trait.AppliacationConfigTemplate{
		AppliacationConfigTemplateMeta: trait.AppliacationConfigTemplateMeta{
			Aversion: "~2.13.0-123-qwe",
			Aname:    "test",
			Tname:    "test0",
			Tversion: "v2.13.0",
		},
	}
	cfg1.Labels = cfg0.Labels
	// insert
	_, err := s.InsertConfigTempalte(ctx, cfg)
	tt.AssertNil(err)
	_, err = s.InsertConfigTempalte(ctx, cfg0)
	tt.AssertNil(err)
	_, err = s.InsertConfigTempalte(ctx, cfg1)
	tt.AssertNil(err)

	{
		// filter
		f := trait.ApplicationConfigTemplateFilter{}
		tss := []struct {
			count   int
			wantErr bool
			errNum  int
			preSet  func()
		}{
			// {
			// 	count: 3,
			// },
			// {
			// 	count: 3,
			// 	preSet: func() {
			// 		f.Aname = cfg.Aname
			// 	},
			// },
			// {
			// 	count: 0,
			// 	preSet: func() {
			// 		f.Aname = f.Aname + f.Aname
			// 	},
			// },
			{
				count: 3,
				preSet: func() {
					f.Aname = cfg.Aname
					f.ApplicationLabelFilter = &trait.ApplicationLabelFilter{
						Labels:    cfg.Labels,
						Condition: 1,
					}
				},
			},
			{
				count: 1,
				preSet: func() {
					f.ApplicationLabelFilter.Condition = 2
				},
			},
			{
				count: 1,
				preSet: func() {
					f.ApplicationVersionFilter = &trait.ApplicationVersionFilter{
						Aversion: cfg.Aversion[1:],
					}
					f.ApplicationLabelFilter.Condition = 0
				},
			},
			{
				count: 2,
				preSet: func() {
					f.ApplicationVersionFilter.Aversion = cfg.Aversion[1:]
					f.ApplicationVersionFilter.Type = 2
				},
			},
			{
				count: 3,
				preSet: func() {
					f.ApplicationVersionFilter.Aversion = cfg1.Aversion[1:]
					f.ApplicationVersionFilter.Type = 3
				},
			},
		}
		for i, ts := range tss {
			t.Run(fmt.Sprintf("testcaset-%d", i), func(t *testing.T) {
				tt := test.TestingT{T: t}
				if ts.preSet != nil {
					ts.preSet()
				}
				cs, err0 := s.ListConfigTemplate(ctx, f, 100, 0)
				c, err := s.CountConfigTempalte(ctx, f)
				if ts.wantErr {
					tt.AssertError(ts.errNum, err)
					tt.AssertError(ts.errNum, err0)
				} else {
					tt.AssertNil(err)
					tt.AssertNil(err0)
					tt.Assert(ts.count, c)
					tt.Assert(ts.count, len(cs))
				}
			})
		}
	}
}

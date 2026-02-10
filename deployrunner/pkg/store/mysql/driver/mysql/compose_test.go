package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"taskrunner/test"
	"taskrunner/trait"

	"github.com/mohae/deepcopy"
)

func TestComposeJob(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	ss := trait.System{}
	sid, err := s.InsertSystemInfo(ctx, ss)
	tt.AssertNil(err)
	j := trait.ComposeJob{
		ComposeJobMeata: trait.ComposeJobMeata{
			Jname:       "test",
			Status:      1,
			Processed:   1,
			Total:       2,
			Description: "test",
		},
		Config: trait.ComposeJobConfig{
			ProtonComponent: []json.RawMessage{},
			AppConfig: []*trait.ApplicationInstance{
				{
					AppConfig: map[string]interface{}{
						"test": "test",
					},
					Components: []*trait.ComponentInstance{
						{
							Config: map[string]interface{}{"test": "test"},
						},
					},
				},
			},
		},
	}
	j.SID = sid
	jid, err := s.InsertComposeJob(ctx, j)
	tt.AssertNil(err)
	j.Jid = jid
	get, err := s.GetComposeJob(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(&j, get)
	tt.Assert(j.Description, get.Description)

	// j.SID = 2
	err = s.SetComposeJob(ctx, j)
	tt.AssertNil(err)
	get, err = s.GetComposeJob(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(&j, get)

	err = s.UpdateComposeJobProcess(ctx, jid, 3)
	tt.AssertNil(err)
	err = s.UpdateComposeJobStatus(ctx, jid, 4, -1, -1)
	tt.AssertNil(err)

	err = s.SetComposeJobTask(ctx, jid, 1, 1)
	tt.AssertNil(err)
	err = s.SetComposeJobTask(ctx, jid, 1, 2)
	tt.AssertNil(err)

	aid, err := s.GetCompoesJobTask(ctx, jid, 1)
	tt.AssertNil(err)
	tt.Assert(aid, 2)

	err = s.SetComposeJobTask(ctx, jid, 3, 2)
	tt.AssertNil(err)

	aids, err := s.GetCompoesJobTasks(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(2, len(aids))

	err = s.DeleteComposeJobTasks(ctx, jid)
	tt.AssertNil(err)

	ls, count, err := s.ListComposeJob(ctx, 100, 0, trait.ComposeJobFilter{
		Name: j.Jname,
	})
	tt.AssertNil(err)
	tt.Assert(1, len(ls))
	tt.Assert(1, count)
	tt.Assert(ls[0].Processed, 3)
	tt.Assert(ls[0].Status, 4)

	{
		err := s.UpdateComposeJobStatus(ctx, jid, -1, 100, 100)
		tt.AssertNil(err)
		ls, _, err := s.ListComposeJob(ctx, 100, 0, trait.ComposeJobFilter{
			Status: []int{4},
		})
		tt.AssertNil(err)
		tt.Assert(0, len(ls))

		ls, _, err = s.ListComposeJob(ctx, 100, 0, trait.ComposeJobFilter{
			Status: []int{-1, 4},
		})
		tt.AssertNil(err)
		tt.Assert(1, len(ls))
	}

	{
		j.Mversion = "test"
		_, err := s.InsertComposeJob(ctx, j)
		tt.AssertNil(err)
		{
			ls, _, err := s.ListComposeJob(ctx, 100, 0, trait.ComposeJobFilter{
				ListType: trait.ComposeJobSuiteType,
			})
			tt.AssertNil(err)
			tt.Assert(1, len(ls))
		}
		{
			ls, _, err := s.ListComposeJob(ctx, 100, 0, trait.ComposeJobFilter{
				ListType: trait.ComposeJobNormalType,
			})
			tt.AssertNil(err)
			tt.Assert(1, len(ls))
		}
		{
			ls, _, err := s.ListComposeJob(ctx, 100, 0, trait.ComposeJobFilter{
				ListType: trait.ComposeJobAllType,
			})
			tt.AssertNil(err)
			tt.Assert(2, len(ls))
		}

	}
}

func TestComposeManifests(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	ss := trait.System{}
	sid, err := s.InsertSystemInfo(ctx, ss)
	tt.AssertNil(err)
	m := trait.ComposeJobManifests{
		ComposeJobManifestsMeta: trait.ComposeJobManifestsMeta{
			Name:    "test",
			Version: "test",
		},
	}

	{
		filter := trait.ComposeManifestFilter{
			NoWork: true,
			Sid:    -1,
			Mname:  "",
		}
		{
			_, _, err := s.ListComposeManifest(ctx, 10, 0, &filter)
			tt.AssertNil(err)
		}
	}

	tt.AssertNil(s.InsertComposeManifests(ctx, m))
	mm, err := s.GetComposeManifests(ctx, m.Name, m.Version)
	tt.AssertNil(err)
	tt.Assert(m.Name, mm.Name)
	tt.Assert(m.Version, mm.Version)

	j := trait.ComposeJob{}
	j.SID = sid
	j.Jid = 10
	j.Jname = m.Name
	j.Mversion = m.Version

	tt.AssertNil(s.InsertWorkComposeManifests(ctx, j.ComposeJobMeata))
	tt.AssertNil(s.InsertWorkComposeManifests(ctx, j.ComposeJobMeata))
	// mj, err := s.GetWorkComposeJobManifests(ctx, j.ComposeJobMeata)
	// tt.AssertNil(err)
	// tt.Assert(j.ComposeJobMeata.Jid, mj.Jid)
	// tt.Assert(j.ComposeJobMeata.Mversion, mj.Mversion)
	j.ComposeJobMeata.Jid--
	tt.AssertNil(s.InsertWorkComposeManifests(ctx, j.ComposeJobMeata))

	ljm, _, err := s.ListWorkComposeJobManifests(ctx, 10, 0, trait.ComposeJobFilter{})
	tt.AssertNil(err)
	tt.Assert(1, len(ljm))

	{
		filter := trait.ComposeManifestFilter{
			NoWork: true,
			Sid:    -1,
		}
		{
			filter.Sid = 1
			filter.Mname = j.Jname
			_, _, err := s.ListComposeManifest(ctx, 10, 0, &filter)
			tt.AssertError(trait.ErrParam, err)
		}

		{
			tt.AssertNil(s.DeleteWorkComposeJobManifests(ctx, j.ComposeJobMeata))
			filter.Mname = ""
			lm, _, err := s.ListComposeManifest(ctx, 10, 0, &filter)
			tt.AssertNil(err)
			tt.Assert(1, len(lm))
			tt.Assert(m.Name, lm[0].Name)
			tt.Assert(m.Description, lm[0].Description)
			// tt.Assert(m.Version, lm[0].Version)
		}

		{
			filter.NoWork = true
			lm, _, err := s.ListComposeManifest(ctx, 10, 0, &filter)
			tt.AssertNil(err)
			tt.Assert(1, len(lm))
		}

		{
			m0 := deepcopy.Copy(m).(trait.ComposeJobManifests)
			m0.Version += "0"
			tt.AssertNil(s.InsertComposeManifests(ctx, m0))

			filter.Mname = j.Jname
			filter.NoWork = false
			lm, count, err := s.ListComposeManifest(ctx, 10, 0, &filter)
			tt.AssertNil(err)
			tt.Assert(2, len(lm))
			tt.Assert(2, count)
			tt.Assert(m.Name, lm[0].Name)
			tt.Assert(m.Description, lm[0].Description)
			tt.Assert(m.Version, lm[0].Version)
		}

		{
			lm, _, err := s.ListComposeManifest(ctx, 10, 0, nil)
			tt.AssertNil(err)
			tt.Assert(2, len(lm))
		}

	}
}

package store_test

import (
	"context"
	"testing"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestJob(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	ss := &trait.System{}
	sid, err := s.InsertSystemInfo(ctx, *ss)
	tt.AssertNil(err)
	job := &trait.JobRecord{
		Target: &trait.ApplicationInstance{
			ApplicationinstanceMeta: trait.ApplicationinstanceMeta{ID: 1},
			Application: trait.Application{
				ApplicationMeta: trait.ApplicationMeta{
					AID: 1,
				},
			},
		},
		// Current: &trait.ApplicationInstance{ApplicationinstanceMeta: trait.ApplicationinstanceMeta{ID: 1}},

	}
	job.Target.SID = sid

	jid, err := s.InsertJobRecord(ctx, job)
	tt.AssertNil(err)
	job0, err := s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(jid, job0.Target.ID)
	tt.Assert(1, job0.Target.AID)

	c, err := s.CountJobRecord(ctx, &trait.AppInsFilter{
		Sid:  -1,
		Name: "",
	})
	tt.AssertNil(err)
	tt.Assert(1, c)

	id := job.Target.ID
	err = s.UpdateAPPInsStatus(ctx, id, trait.AppDoingStatus, -1, -1, -1)
	tt.AssertNil(err)
	// js, err := s.ListJobRecordExecuting(ctx, -1, 10)
	// tt.AssertNil(err)
	// tt.Assert(1, len(js))
	// tt.Assert(js[0][0], jid)
	// tt.Assert(js[0][1], id)
	// tt.Assert(js[0][2], -1)

	// js, err = s.ListJobRecordExecuting(ctx, jid, 10)
	// tt.AssertNil(err)
	// tt.Assert(0, len(js))

	job0, err = s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(jid, job0.Target.ID)
	tt.Assert(1, job0.Target.AID)

	jobs, err := s.ListJobRecord(ctx, &trait.AppInsFilter{
		Offset: 0,
		Limit:  10,
		Sid:    -1,
	})
	tt.AssertNil(err)
	tt.Assert(1, len(jobs))
	job1 := jobs[0]
	tt.Assert(jid, job1.ID)
	tt.Assert(job0.Target.ID, job1.Target.ID)
	tt.Assert(job0.Target.SID, job1.Target.SID)
	tt.Assert((*trait.ApplicationInstance)(nil), job1.Current)
	/* this is interesting，inline '!=' is ok.
	but interface{} will convert job1.Current into (*ins, nil).
	so (*ins,nil )!= (nil)
	*/
	// tt.Assert(nil, job1.Current)

	jobs, err = s.ListJobRecord(ctx, &trait.AppInsFilter{
		Offset: 1,
		Limit:  10,
		Sid:    -1,
	})
	tt.AssertNil(err)
	tt.Assert(0, len(jobs))

	job0.Current = job0.Target
	jid, err = s.InsertJobRecord(ctx, &job0)
	tt.AssertNil(err)
	job1, err = s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(job0.Current.ID, job1.Current.ID)

	jobs, err = s.ListJobRecord(ctx, &trait.AppInsFilter{
		Offset: 0,
		Limit:  10,
		Sid:    -1,
	})
	tt.AssertNil(err)
	tt.Assert(2, len(jobs))
	job1 = jobs[0]
	tt.Assert(jid, job1.ID)
	tt.Assert(job0.Current.ID, job1.Current.ID)
}

func TestJobLog(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	f := trait.JobLogFilter{
		JID: -1,
		CID: -1,
	}

	count, err := s.CountJobLog(ctx, f)
	tt.AssertNil(err)
	tt.Assert(0, count)
	log := trait.JobLog{}
	err = s.InsertJobLog(ctx, log)
	tt.AssertNil(err)

	ls, err := s.ListJobLog(ctx, f)
	tt.AssertNil(err)
	tt.Assert(1, len(ls))
	log0 := ls[0]
	tt.Assert("", log0.Msg)

	{
		testfilter := func(n int) {
			ls, err := s.ListJobLog(ctx, f)
			tt.AssertNil(err)
			tt.Assert(n, len(ls))
		}
		log.JID = 1
		tt.AssertNil(s.InsertJobLog(ctx, log))

		log.CID = 1
		tt.AssertNil(s.InsertJobLog(ctx, log))

		log.JID = 0
		tt.AssertNil(s.InsertJobLog(ctx, log))

		f.JID = 1
		testfilter(2)
		count, err = s.CountJobLog(ctx, f)
		tt.AssertNil(err)
		tt.Assert(2, count)

		f.CID = 1
		testfilter(1)

		f.JID = 0
		testfilter(1)

		f.CID = 0
		testfilter(1)

		f.CID = -1
		f.JID = -1
		testfilter(4)

		f.Offset = 2
		testfilter(2)

		f.Offset = -1
		log.Timestamp = 123
		tt.AssertNil(s.InsertJobLog(ctx, log))

		f.Timestmp = -1
		testfilter(4)

		f.Timestmp = 1234
		testfilter(0)

		f.Timestmp = 122
		testfilter(1)

		f.Offset = 100
		testfilter(0)
		count, err = s.CountJobLog(ctx, f)
		tt.AssertNil(err)
		tt.Assert(0, count)
	}

	{
		log.Msg = "test line feed\n\ttest '\\t'\n test \\\n "
		tt.AssertNil(s.InsertJobLog(ctx, log))
	}
}

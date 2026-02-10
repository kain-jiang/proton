package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"taskrunner/pkg/app/builder"
	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	"taskrunner/test"
	testdata "taskrunner/test"
	testchart "taskrunner/test/charts"
	"taskrunner/trait"

	"helm.sh/helm/v3/pkg/action"
)

// SystemFaker system store faker
type SystemFaker struct {
	WorkAPP               map[string]*trait.ApplicationInstance
	workComponent         map[string]*trait.ComponentInstance
	system                *trait.System
	componentInstanceLock map[string]int
	appLock               map[string]int
}

// DbStoreFaker store faker
// TODB instead with sqlite
type DbStoreFaker struct {
	trait.Store
	WantErr      *trait.Error
	AppLangCache map[string]string
	AppCache     []*trait.Application
	AppComs      []*trait.ComponentMeta
	SystemCache  []*SystemFaker
	AppInsCache  []*trait.ApplicationInstance
	ComponentIns []*trait.ComponentInstance
	jobInsCache  []*trait.JobRecord
	jobLogs      []trait.JobLog
	edges        [][2]int
	ErrMap       map[string]*trait.Error
	lock         sync.Mutex
}

func (s *DbStoreFaker) wantFuncErr(f string) *trait.Error {
	if err, ok := s.ErrMap[f]; err != nil && ok {
		return err
	}
	return s.WantErr
}

// LayOffAPPIns impl store
func (s *DbStoreFaker) LayOffAPPIns(ctx context.Context, a *trait.ApplicationInstance) *trait.Error {
	id := a.ID
	if id >= len(s.AppInsCache) || id < 0 {
		return &trait.Error{
			Internal: trait.ErrNotFound,
			Err:      fmt.Errorf("layOffAppIns"),
		}
	}
	c := s.AppInsCache[id]
	delete(s.SystemCache[c.System.SID].WorkAPP, a.AName)

	return s.wantFuncErr("LayOffAPPIns")
}

// edge

// GetPointTo get from component instance id
func (s *DbStoreFaker) GetPointTo(ctx context.Context, to int) (poits []int, err *trait.Error) {
	for _, e := range s.edges {
		if to == e[1] {
			poits = append(poits, e[0])
		}
	}
	err = s.wantFuncErr("GetPointTo")
	return
}

// GetAPPComponent get application component defined
func (s *DbStoreFaker) GetAPPComponent(ctx context.Context, acid int) (com *trait.ComponentMeta, err *trait.Error) {
	if acid < 0 || acid >= len(s.AppComs) {
		err = &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("GetAPPComponent mock")}
		return
	}

	com = s.AppComs[acid]
	err = s.wantFuncErr("GetAPPComponent")
	return
}

// GetPointFrom get to component instance id
func (s *DbStoreFaker) GetPointFrom(ctx context.Context, from int) (poits []int, err *trait.Error) {
	for _, e := range s.edges {
		if from == e[0] {
			poits = append(poits, e[1])
		}
	}
	err = s.wantFuncErr("GetPointTo")
	return
}

func (s *DbStoreFaker) ChangeEdgeFrom(ctx context.Context, curID, tarID int) *trait.Error {
	newEdges := map[string][2]int{}

	for _, e := range s.edges {
		if curID == e[0] {
			e[0] = tarID
		}
		k := fmt.Sprintf("%#v", e)
		newEdges[k] = e
	}
	edges := make([][2]int, 0, len(newEdges))
	for _, e := range newEdges {
		edges = append(edges, e)
	}
	s.edges = edges

	err := s.wantFuncErr("ChangeEdgeFrom")
	return err
}

// CountEdgeTo impl store
func (s *DbStoreFaker) CountEdgeTo(ctx context.Context, cid int) (int, *trait.Error) {
	count := 0
	for _, e := range s.edges {
		if cid == e[1] {
			count++
		}
	}
	return count, s.wantFuncErr("CountEdgeTo")
}

// ChangeEdgeto impl store
func (s *DbStoreFaker) ChangeEdgeto(ctx context.Context, cur, tatget int) *trait.Error {
	for _, e := range s.edges {
		if cur == e[1] {
			e[1] = tatget
		}
	}

	return s.wantFuncErr("ChangeEdgeto")
}

// componentIns

// WorkComponentIns impl store
func (s *DbStoreFaker) WorkComponentIns(ctx context.Context, cins *trait.ComponentInstance) *trait.Error {
	err := s.wantFuncErr("WorkComponentIns")
	if err != nil {
		return err
	}
	cid := cins.CID
	if cid >= len(s.ComponentIns) || cid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("workComponentIns")}
	}
	c := s.ComponentIns[cid]

	s.SystemCache[c.System.SID].workComponent[c.Component.Name] = c
	return err
}

func (s *DbStoreFaker) ListWorkComponentIns(ctx context.Context, filter trait.WorkCompFilter) ([]*trait.ComponentInstanceMeta, *trait.Error) {
	if filter.Sid > len(s.SystemCache) {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("system notfound")}
	}
	offset := 0
	res := make([]*trait.ComponentInstanceMeta, 0, filter.Limit)
	for _, c := range s.SystemCache[filter.Sid].workComponent {
		if offset >= filter.Offset {
			res = append(res, &c.ComponentInstanceMeta)
		}
	}
	return res, s.wantFuncErr("ListWorkComponentIns")
}

// LockComponent impl store
func (s *DbStoreFaker) LockComponent(ctx context.Context, sid int, jid int, cnode trait.ComponentNode) *trait.Error {
retry:
	s.lock.Lock()
	if sid >= len(s.SystemCache) || sid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("lockComponent mock")}
	}
	sc := s.SystemCache[sid]
	lockOwn, ok := sc.componentInstanceLock[cnode.Name]
	if !ok {
		sc.componentInstanceLock[cnode.Name] = jid
		defer s.lock.Unlock()
		return s.wantFuncErr("LockComponent")
	} else if lockOwn == jid {
		defer s.lock.Unlock()
		return s.wantFuncErr("LockComponent")
	} else {
		s.lock.Unlock()
		delay := time.NewTimer(2 * time.Second)
		select {
		case <-delay.C:
			goto retry
		case <-ctx.Done():
			return ctx.Err().(*trait.Error)
		}
	}
}

// UnlockComponent impl store
func (s *DbStoreFaker) UnlockComponent(ctx context.Context, sid int, jid int, cnode trait.ComponentNode) *trait.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sid >= len(s.SystemCache) || sid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("unlockComponent")}
	}
	if err := s.wantFuncErr("UnlockComponent"); err != nil {
		return err
	}
	sc := s.SystemCache[sid]
	delete(sc.componentInstanceLock, cnode.Name)
	return nil
}

// LockComponent impl store
func (s *DbStoreFaker) LockApp(ctx context.Context, sid int, jid int, aname string) *trait.Error {
retry:
	s.lock.Lock()
	if sid >= len(s.SystemCache) || sid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("lockComponent mock")}
	}
	sc := s.SystemCache[sid]
	lockOwn, ok := sc.appLock[aname]
	if !ok {
		defer s.lock.Unlock()
		sc.appLock[aname] = jid
		return s.wantFuncErr("LockApp")
	} else if lockOwn == jid {
		defer s.lock.Unlock()
		return s.wantFuncErr("LockApp")
	} else {
		s.lock.Unlock()
		delay := time.NewTimer(2 * time.Second)
		select {
		case <-delay.C:
			goto retry
		case <-ctx.Done():
			return ctx.Err().(*trait.Error)
		}
	}
}

// UnlockComponent impl store
func (s *DbStoreFaker) UnlockApp(ctx context.Context, sid int, jid int, aname string) *trait.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if sid >= len(s.SystemCache) || sid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("unlockComponent")}
	}
	if err := s.wantFuncErr("UnlockApp"); err != nil {
		return err
	}
	sc := s.SystemCache[sid]
	delete(sc.appLock, aname)
	return nil
}

// UnlockJobComponent unlock job hold lock
func (s *DbStoreFaker) UnlockJobComponent(ctx context.Context, jid int) *trait.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, sc := range s.SystemCache {
		needUnlock := []string{}
		for name, jid0 := range sc.componentInstanceLock {
			if jid0 == jid {
				needUnlock = append(needUnlock, name)
			}
		}
		for _, name := range needUnlock {
			delete(sc.componentInstanceLock, name)
		}
	}
	return nil
}

// GetComponentIns impl store
func (s *DbStoreFaker) GetComponentIns(ctx context.Context, cid int) (cins *trait.ComponentInstance, err *trait.Error) {
	if cid >= len(s.ComponentIns) || cid < 0 {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("getcomponentins mock")}
	}
	cins = s.ComponentIns[cid]
	err = s.wantFuncErr("GetComponentIns")
	return
}

// UpdateComponentInsStatus impl store
func (s *DbStoreFaker) UpdateComponentInsStatus(ctx context.Context, cid, status, revission, startTime, endTime int) *trait.Error {
	if cid >= len(s.ComponentIns) || cid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("updateComponentInsStatus")}
	}
	err := s.wantFuncErr("UpdateComponentInsStatus")
	if err != nil {
		return err
	}
	s.ComponentIns[cid].Status = status
	s.ComponentIns[cid].Revission = revission + 1
	s.ComponentIns[cid].StartTime = startTime
	s.ComponentIns[cid].EndTime = endTime

	return nil
}

// InsertComponentIns imple store
func (s *DbStoreFaker) InsertComponentIns(ctx context.Context, com *trait.ComponentInstance) (int, *trait.Error) {
	s.ComponentIns = append(s.ComponentIns, com)
	return len(s.ComponentIns) - 1, s.wantFuncErr("InsertComponentIns")
}

// GetWorkComponentIns impl store
func (s *DbStoreFaker) GetWorkComponentIns(ctx context.Context, sid int, component trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	err := s.wantFuncErr("GetWorkComponentIns")
	if err != nil {
		return nil, err
	}
	if sid >= len(s.SystemCache) || sid < 0 {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("getworkcomponentins mock")}
	}
	system := s.SystemCache[sid]
	c := system.workComponent[component.Name]
	if c == nil {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("getworkcomponentins mock")}
	}
	return c, s.wantFuncErr("GetWorkComponentIns")
}

// LayoffComponentIns impl store
func (s *DbStoreFaker) LayoffComponentIns(ctx context.Context, cid int) *trait.Error {
	if cid >= len(s.ComponentIns) || cid < 0 {
		return s.WantErr
	}
	err := s.wantFuncErr("LayoffComponentIns")
	if err != nil {
		return err
	}
	c := s.ComponentIns[cid]
	delete(s.SystemCache[c.System.SID].workComponent, c.Component.Name)
	return nil
}

// DeleteEdgeFrom impl store
func (s *DbStoreFaker) DeleteEdgeFrom(ctx context.Context, cid int) *trait.Error {
	err := s.wantFuncErr("DeleteEdgeFrom")
	if err != nil {
		return err
	}
	edges := s.edges
	for i := 0; i < len(edges); i++ {
		e := edges[i]
		if e[1] == cid {
			if i < len(edges)-1 {
				copy(edges[i+1:], edges[i:])
			}
			edges = edges[:len(edges)-1]
		}
	}
	s.edges = edges
	return nil
}

// GetAPPIns impl store
func (s *DbStoreFaker) GetAPPIns(ctx context.Context, id int) (*trait.ApplicationInstance, *trait.Error) {
	if id >= len(s.AppInsCache) || id < 0 {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("GetAPPIns mock")}
	}
	return s.AppInsCache[id], s.wantFuncErr("GetAPPIns")
}

// GetWorkAPPIns impl store
func (s *DbStoreFaker) GetWorkAPPIns(ctx context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	if sid >= len(s.SystemCache) || sid < 0 {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("getworkappins mock")}
	}
	sysgroup := s.SystemCache[sid]

	if insp := sysgroup.WorkAPP[name]; insp != nil {
		return insp, s.wantFuncErr("GetWorkAPPIns")
	}

	return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("getworkappins mock")}
}

// UpdateAPPInsConfig impl store
func (s *DbStoreFaker) UpdateAPPInsConfig(ctx context.Context, ins trait.ApplicationInstance) *trait.Error {
	id := ins.ID
	if id >= len(s.AppInsCache) || id < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("updateappinsconfig mock")}
	}
	old := s.AppInsCache[id]
	old.Status = ins.Status
	old.AppConfig = ins.AppConfig
	target := old
	for _, c := range ins.Components {
		cfg := target.ComponentInsExistedOrCreate(c.Component)
		cfg.Status = c.Status
		cfg.Config = c.Config
		cfg.Attribute = c.Attribute
		cfg.Timeout = c.Timeout
	}

	return s.wantFuncErr("UpdateAPPIns")
}

// UpdateAPPInsStatus impl store
func (s *DbStoreFaker) UpdateAPPInsStatus(ctx context.Context, id int, status int, owner int, startTime, endTime int) *trait.Error {
	err := s.wantFuncErr("UpdateAPPInsStatus")
	if err != nil {
		return err
	}
	if id >= len(s.AppInsCache) || id < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("app ins not found mock")}
	}
	s.AppInsCache[id].Status = status
	s.AppInsCache[id].Onwer = owner
	s.AppInsCache[id].StartTime = startTime
	s.AppInsCache[id].EndTime = endTime
	return err
}

// InsertAPPIns impl store
func (s *DbStoreFaker) InsertAPPIns(ctx context.Context, ins *trait.ApplicationInstance) (int, *trait.Error) {
	s.AppInsCache = append(s.AppInsCache, ins)
	for _, c := range ins.Components {
		cid, err := s.InsertComponentIns(ctx, c)
		if err != nil {
			return -1, err
		}
		c.CID = cid
	}
	return len(s.AppInsCache) - 1, s.wantFuncErr("InsertAPPIns")
}

// WorkAppIns impl store
func (s *DbStoreFaker) WorkAppIns(ctx context.Context, app *trait.ApplicationInstance) *trait.Error {
	id := app.ID
	if id >= len(s.AppInsCache) || id < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("work app ins mock")}
	}
	a := s.AppInsCache[id]
	sid := a.SID
	s.SystemCache[sid].WorkAPP[a.Application.AName] = a
	for _, c := range a.Components {
		if err := s.WorkComponentIns(ctx, c); err != nil {
			return err
		}
	}
	return s.wantFuncErr("WorkAppIns")
}

// AddEdge impl store
func (s *DbStoreFaker) AddEdge(ctx context.Context, from, to int) *trait.Error {
	err := s.wantFuncErr("AddEdge")
	if err != nil {
		return err
	}
	for _, e := range s.edges {
		if e[0] == from && e[1] == to {
			return err
		}
	}
	s.edges = append(s.edges, [2]int{from, to})
	return err
}

// AddEdge impl store
func (s *DbStoreFaker) AddOuterChildEdge(ctx context.Context, from int, sid int, com trait.ComponentNode) *trait.Error {
	err := s.wantFuncErr("AddOuterChildEdge")
	if err != nil {
		return err
	}
	cins, err := s.GetWorkComponentIns(ctx, sid, com)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, e := range s.edges {
		if e[0] == from && e[1] == cins.CID {
			return nil
		}
	}
	s.edges = append(s.edges, [2]int{from, cins.CID})
	return err
}

// job

// InsertJobRecord impl store
func (s *DbStoreFaker) InsertJobRecord(ctx context.Context, job *trait.JobRecord) (int, *trait.Error) {
	id, err := s.InsertAPPIns(ctx, job.Target)
	if err != nil {
		return -1, err
	}
	job.Target.ID = id
	s.jobInsCache = append(s.jobInsCache, job)
	return len(s.jobInsCache) - 1, s.wantFuncErr("InsertJobRecord")
}

// GetJobRecord impl store
func (s *DbStoreFaker) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	jb := trait.JobRecord{}
	if jid > len(s.jobInsCache) || jid < 0 {
		return jb, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("get job record mock")}
	}
	return *s.jobInsCache[jid], s.wantFuncErr("GetJobRecord")
}

func (s *DbStoreFaker) InsertConfigTempalte(ctx context.Context, cfg trait.AppliacationConfigTemplate) (int, *trait.Error) {
	return -1, s.wantFuncErr("GetJobRecord")
}

// ListJobRecord impl store interface
func (s *DbStoreFaker) ListJobRecord(ctx context.Context, f *trait.AppInsFilter) ([]trait.JobRecord, *trait.Error) {
	limit := f.Limit
	offset := f.Offset
	res := make([]trait.JobRecord, 0, f.Limit)
	count := 0
	for _, job := range s.jobInsCache {
		if count == limit {
			break
		}
		if f.Sid >= 0 && job.Target.SID != f.Sid {
			continue
		}
		if f.Name != "" && f.Name != job.Target.AName {
			continue
		}

		for _, status := range f.Status {
			if job.Target.Status == status && count >= offset {
				res = append(res, *job)
			}
		}
		if len(f.Status) == 0 && count >= offset {
			res = append(res, *job)
		}
	}
	return res, s.wantFuncErr("ListJobRecordExecuting")
}

func (s *DbStoreFaker) InsertJobLog(ctx context.Context, j trait.JobLog) *trait.Error {
	j.JLID = len(s.jobLogs)
	s.jobLogs = append(s.jobLogs, j)
	return s.wantFuncErr("InsertJobLog")
}

// application

// InsertAPP  impl store
func (s *DbStoreFaker) InsertAPP(ctx context.Context, app trait.Application) (int, *trait.Error) {
	app.AID = len(s.AppCache)
	s.AppCache = append(s.AppCache, &app)
	for _, com := range app.Component {
		s.AppComs = append(s.AppComs, com)
		com.CID = len(s.AppComs) - 1
	}
	return app.AID, s.wantFuncErr("InsertAPP")
}

// GetAPP impl store
func (s *DbStoreFaker) GetAPP(ctx context.Context, aid int) (*trait.Application, *trait.Error) {
	if aid >= len(s.AppCache) || aid < 0 {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("get app mock")}
	}
	return s.AppCache[aid], s.wantFuncErr("GetAPP")
}

// system

// InsertSystemInfo impl
func (s *DbStoreFaker) InsertSystemInfo(ctx context.Context, ss trait.System) (int, *trait.Error) {
	s.SystemCache = append(s.SystemCache, &SystemFaker{
		system:                &ss,
		WorkAPP:               make(map[string]*trait.ApplicationInstance),
		workComponent:         make(map[string]*trait.ComponentInstance),
		componentInstanceLock: make(map[string]int),
		appLock:               make(map[string]int),
	})
	return len(s.SystemCache) - 1, s.wantFuncErr("InsertSystemInfo")
}

// UpdateSystemInfo impl
func (s *DbStoreFaker) UpdateSystemInfo(ctx context.Context, ss trait.System) *trait.Error {
	sid := ss.SID
	if sid >= len(s.SystemCache) || sid < 0 {
		return &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("system not found mock")}
	}
	s.SystemCache[sid].system = &ss
	return s.wantFuncErr("UpdateSystemInfo")
}

// GetSystemInfo impl
func (s *DbStoreFaker) GetSystemInfo(ctx context.Context, sid int) (*trait.System, *trait.Error) {
	if sid >= len(s.SystemCache) || sid < 0 {
		return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("not found mock")}
	}
	return s.SystemCache[sid].system, s.wantFuncErr("GetSystemInfo")
}

// tx

// Begin impl
func (s *DbStoreFaker) Begin(ctx context.Context) (trait.Transaction, *trait.Error) {
	return s, s.wantFuncErr("Begin")
}

// Commit impl
func (s *DbStoreFaker) Commit() *trait.Error {
	return s.wantFuncErr("Commit")
}

// Rollback impl
func (s *DbStoreFaker) Rollback() *trait.Error {
	return s.wantFuncErr("Rollback")
}

// InsertAppLang insert lang and update cache
func (s *DbStoreFaker) InsertAppLang(ctx context.Context, lang, aname, alias, zone string) *trait.Error {
	s.AppLangCache[zone+lang+aname] = alias
	return s.wantFuncErr("InsertAppLang")
}

func (s *DbStoreFaker) GetAppLang(ctx context.Context, lang, aname, zone string) (string, *trait.Error) {
	a, ok := s.AppLangCache[zone+lang+aname]
	if !ok {
		s.AppLangCache[zone+lang+aname] = aname
		a = aname
	}
	return a, nil
}

// HelmRepoMock faker helm repo
type HelmRepoMock struct {
	Chart    map[string][]byte
	RepoName string
	WantErr  *trait.Error
}

// Store store chart
func (r *HelmRepoMock) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	r.Chart[chart.Name+chart.Version] = data
	return r.WantErr
}

// Name return repo name
func (r *HelmRepoMock) Name() string {
	return r.RepoName
}

// Fetch fetch chart from repo
func (r *HelmRepoMock) Fetch(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	if bs, ok := r.Chart[c.Name+c.Version]; ok {
		return bs, r.WantErr
	}
	return nil, &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("helm fetch mock")}
}

// HelmCliMock mock helm cli
type HelmCliMock struct {
	Err *trait.Error
}

// Values impl helm cli
func (c *HelmCliMock) Values(ctx context.Context, name, ns string) (map[string]interface{}, *trait.Error) {
	panic("no imply")
}

// Uninstall impl helm cli
func (c *HelmCliMock) Uninstall(ctx context.Context, name, ns string, timeout int, log action.DebugLog) *trait.Error {
	return c.Err
}

// Install impl helm cli
func (c *HelmCliMock) Install(ctx context.Context, name, ns string, chart *helm.Chart, cfg map[string]interface{}, timeout int, log action.DebugLog) *trait.Error {
	return c.Err
}

// GetTestAppliationBytes get applicationbyte from test case
func GetTestAppliationBytes(t *testing.T) []byte {
	// apf := testdata.TestAPP
	// testWantError(t, nil, err)
	// defer apf.Close()
	apf := bytes.NewReader(testdata.TestAPP)
	buf := bytes.NewBuffer(nil)
	tt := test.TestingT{T: t}

	cfg, err := builder.LoadConfiguration(apf)
	tt.AssertNil(err)

	repo := &testchart.MemoryHelmRepoMock{
		RepoName: "test",
		FS:       testdata.TestCharts,
	}

	f, err0 := os.CreateTemp(os.TempDir(), "test-config-tempalte-*.json")
	tt.AssertNil(err0)
	defer f.Close()
	defer os.Remove(f.Name())
	cfgt := trait.AppliacationConfigTemplate{
		AppliacationConfigTemplateMeta: trait.AppliacationConfigTemplateMeta{
			Aname:    "test",
			Aversion: "~v2.12.0-123",
			Tname:    "test0",
			Tversion: "qwe",
		},
		Config: trait.ApplicationConfigSet{
			AppConfig: map[string]interface{}{
				"qwe": 123,
			},
		},
	}
	bs0, err0 := json.Marshal(cfgt)
	tt.AssertNil(err0)
	_, err0 = f.Write(bs0)
	tt.AssertNil(err0)
	tt.AssertNil(f.Sync())

	b, err := builder.NewApplicationBuilder(&cfg, buf, io.Discard, repo)
	tt.AssertNil(err)
	b.ConfigTemplatePath = f.Name()
	err = b.Build(context.Background())
	tt.AssertNil(err)
	return buf.Bytes()
}

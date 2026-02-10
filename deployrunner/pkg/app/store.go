package app

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"taskrunner/pkg/app/builder"
	"taskrunner/pkg/helm"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/sirupsen/logrus"
)

type ApplangCacheStore struct {
	trait.Store
	cache sync.Map
	index sync.Map
}

type LangSetter interface {
	SetLang(string)
	LangKey() string
}

func NewAppLangCacheStore(s trait.Store) *ApplangCacheStore {
	return &ApplangCacheStore{
		Store: s,
		cache: sync.Map{},
		index: sync.Map{},
	}
}

func (s *ApplangCacheStore) GetAname(lang, alias, zone string) string {
	aname, ok := s.index.Load(zone + lang + alias)
	if !ok {
		return ""
	}
	return aname.(string)
}

// InsertAppLang insert lang and update cache
func (s *ApplangCacheStore) InsertAppLang(ctx context.Context, lang, aname, alias, zone string) *trait.Error {
	s.cacheAppLang(lang, aname, alias, zone)
	return s.Store.InsertAppLang(ctx, lang, aname, alias, zone)
}

func (s *ApplangCacheStore) cacheAppLang(lang, aname, alias, zone string) {
	s.cache.Store(zone+lang+aname, alias)
	s.index.Store(zone+lang+alias, aname)
}

func (s *ApplangCacheStore) GetAppLang(ctx context.Context, lang, aname string, zone string) (string, *trait.Error) {
	a, ok := s.cache.Load(zone + lang + aname)
	if !ok {
		alias := aname
		err := utils.RetryN(ctx, func() (bool, *trait.Error) {
			aliasTemp, err := s.Store.GetAppLang(ctx, lang, aname, zone)
			alias = aliasTemp

			if trait.IsInternalError(err, trait.ErrNotFound) {
				alias = aname
				s.cacheAppLang(lang, aname, alias, zone)
				return false, nil
			}
			if err == nil {
				s.cacheAppLang(lang, aname, alias, zone)
			}
			return err != nil, err
		}, 5, 3*time.Second)
		return alias, err
	}
	return a.(string), nil
}

// Store application system manager
type Store struct {
	*ApplangCacheStore
	Log *logrus.Logger
	// helm.Repos
	HelmRepo helm.Repo
}

// NewStore return a Store
func NewStore(log *logrus.Logger, s trait.Store, repo helm.Repo) *Store {
	return &Store{
		Log:               log,
		ApplangCacheStore: NewAppLangCacheStore(s),
		HelmRepo:          repo,
	}
}

func (s *Store) rollbackWithLog(tx trait.Transaction) func() {
	return func() {
		if err := tx.Rollback(); err != nil {
			s.Log.Errorf("rollback error: %s", err.Error())
		}
	}
}

func (s *Store) MergeJobConfigs(job *trait.JobRecord, items ...*trait.ApplicationInstance) {
	for _, ins := range items {
		s.MergeJobConfigCompose(job, ins, true)
	}
}

func (s *Store) MergeJobConfigCompose(job *trait.JobRecord, ins *trait.ApplicationInstance, composeMerge bool) {
	if ins == nil {
		return
	}
	target := job.Target
	for _, c := range ins.Components {
		cfg := target.ComponentInsExisted(c.Component)
		if cfg == nil {
			continue
		}
		switch c.Status {
		case trait.AppIgnoreStatus:
			// 仅允许人工设置为成功或忽略,创建套件任务和配置已有任务会使用该能力
			cfg.Status = c.Status
		case trait.AppinitStatus:
			if !composeMerge {
				cfg.Status = c.Status
			}
		default:

		}

		if composeMerge {
			if c.Config != nil {
				cfg.Config = utils.MergeMaps(cfg.Config, c.Config)
			}
			if c.Attribute != nil {
				cfg.Attribute = utils.MergeMaps(cfg.Attribute, c.Attribute)
			}
			if c.Timeout != 0 {
				cfg.Timeout = c.Timeout
			}
		} else {
			cfg.Config = c.Config
			cfg.Attribute = c.Attribute
			cfg.Timeout = c.Timeout
			// c.Status = cfg.Status
			// cfg.ComponentInstanceTrait = c.ComponentInstanceTrait
		}
	}

	target.Comment = ins.Comment
	if composeMerge {
		if ins.AppConfig != nil {
			target.AppConfig = utils.MergeMaps(target.AppConfig, ins.AppConfig)
		}
	} else {
		target.AppConfig = ins.AppConfig
	}
	// target.Trait.UpgradeParent = ins.Trait.UpgradeParent
	target.Trait = ins.Trait
}

// MergeJobConfig merge the job.target instance config with input instance
func (s *Store) MergeJobConfig(job *trait.JobRecord, ins *trait.ApplicationInstance) {
	s.MergeJobConfigCompose(job, ins, false)
}

// MarkJobDelete 标记删除
func (s *Store) MarkJobDelete(ctx context.Context, id int) *trait.Error {
	tx, err := s.Store.Begin(ctx)
	if err != nil {
		s.Log.Errorf("start a trasaction error, please contact env maintainer: %s", err.Error())
		return err
	}
	job, err := tx.GetJobRecord(ctx, id)
	if err != nil {
		s.Log.Errorf("get the jobrecord error %s", err.Error())
		s.rollbackWithLog(tx)()
		return err
	}

	if job.Target.Onwer != 0 {
		s.Log.Errorf("job %d control by executor %d , couldn't change config", job.ID, job.Target.Onwer)
		defer s.rollbackWithLog(tx)()
		return &trait.Error{
			Internal: trait.ErrJobExecuting,
			Err:      fmt.Errorf("can't set job config which is running"),
			Detail:   fmt.Sprintf("job %d is execute by executor %d", job.ID, job.Target.Onwer),
		}
	}

	for _, doingStatus := range trait.JobDoingStauts {
		if job.Target.Status == doingStatus {
			s.Log.Errorf("job %d is executing, couldn't change config", job.ID)
			defer s.rollbackWithLog(tx)()
			return &trait.Error{
				Internal: trait.ErrJobExecuting,
				Err:      fmt.Errorf("can't set job config which is running"),
				Detail:   fmt.Sprintf("job %d is execute by executor %d with status %d", job.ID, job.Target.Onwer, job.Target.Status),
			}
		}
	}
	if err := tx.UpdateAPPInsOperateType(ctx, job.Target.ID, trait.JobDeleteOType); err != nil {
		s.Log.Errorf("mark job %d to delete job error: %s", job.ID, err.Error())
		defer s.rollbackWithLog(tx)()
		return err
	}

	return tx.Commit()
}

// SetJobConfig check application instance config and set
func (s *Store) SetJobConfig(ctx context.Context, id int, ins *trait.ApplicationInstance) *trait.Error {
	tx, err := s.Store.Begin(ctx)
	if err != nil {
		s.Log.Errorf("start a trasaction error, please contact env maintainer: %s", err.Error())
		return err
	}
	job, err := tx.GetJobRecord(ctx, id)
	if err != nil {
		s.Log.Errorf("get the jobrecord error %s", err.Error())
		s.rollbackWithLog(tx)()
		return err
	}

	if job.Target.Onwer != 0 {
		s.Log.Errorf("job %d control by executor %d , couldn't change config", job.ID, job.Target.Onwer)
		defer s.rollbackWithLog(tx)()
		return &trait.Error{
			Internal: trait.ErrJobExecuting,
			Err:      fmt.Errorf("can't set job config which is running"),
			Detail:   fmt.Sprintf("job %d is execute by executor %d", job.ID, job.Target.Onwer),
		}
	}

	for _, doingStatus := range trait.JobDoingStauts {
		if job.Target.Status == doingStatus {
			s.Log.Errorf("job %d is executing, couldn't change config", job.ID)
			defer s.rollbackWithLog(tx)()
			return &trait.Error{
				Internal: trait.ErrJobExecuting,
				Err:      fmt.Errorf("can't set job config which is running"),
				Detail:   fmt.Sprintf("job %d is execute by executor %d with status %d", job.ID, job.Target.Onwer, job.Target.Status),
			}
		}
	}

	// aid := job.Target.AID
	// app, err := tx.GetAPP(ctx, aid)
	// if err != nil {
	// 	s.Log.Errorf("get application from store error: %s", err.Error())
	// 	defer s.rollbackWithLog(tx)()
	// 	return err
	// }
	// job.Target.Application = *app
	s.MergeJobConfig(&job, ins)
	target := job.Target
	s.Log.Debugf("get the application %d", target.AID)
	app, err := tx.GetAPP(ctx, target.AID)
	if err != nil {
		s.Log.Errorf("get the application error: %s", err.Error())
		defer s.rollbackWithLog(tx)()
		return err
	}
	target.Application = *app
	s.Log.Debugf("application config %#v", target.AppConfig)
	if err := target.Validate(); err != nil {
		s.Log.Debugf("application config schame %s", target.ConfigSchema)
		s.Log.Errorf("validate the input config error: %s", err.Error())
		defer s.rollbackWithLog(tx)()
		return err
	}
	target.Status = trait.AppConfirmedStatus
	if err := tx.UpdateAPPInsConfig(ctx, *target); err != nil {
		s.Log.Errorf("update job config error: %s", err.Error())
		defer s.rollbackWithLog(tx)()
		return err
	}

	if err := tx.UpdateAPPInsStatus(ctx, target.ID, target.Status, 0, -1, -1); err != nil {
		s.Log.Errorf("update job status error: %s", err.Error())
		defer s.rollbackWithLog(tx)()
		return err
	}

	err = tx.Commit()
	return err
}

// NewFakeJobRecord create a jobrecord by current instance in system, but don't store it
func (s *Store) NewFakeJobRecord(ctx context.Context, aid, sid int) (*trait.JobRecord, *trait.Error) {
	// TODO
	tx, err := s.Store.Begin(ctx)
	if err != nil {
		s.Log.Errorf("start a trasaction error, please contact env maintainer: %s", err.Error())
		return nil, err
	}
	defer s.rollbackWithLog(tx)()
	return s.GetApplicationJobSnapshot(ctx, tx, aid, sid)
}

// GetApplicationJobSnapshot get current application instance as job config
func (s *Store) GetApplicationJobSnapshot(ctx context.Context, tx trait.Transaction, aid, sid int) (*trait.JobRecord, *trait.Error) {
	// TODO needn't get system if not resource in system
	system, err := tx.GetSystemInfo(ctx, sid)
	if err != nil {
		err.Detail = fmt.Sprintf("system id: %d", sid)
		s.Log.Errorf("get the system info error:%s", err.Error())
		return nil, err
	}

	app, err := tx.GetAPP(ctx, aid)
	if err != nil {
		s.Log.Errorf("get application from store error: %s", err.Error())
		return nil, err
	}

	cur, err := tx.GetWorkAPPIns(ctx, app.AName, sid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		// install job
		cur = nil
		// return s.install(ctx, cfg, aid, sid, tx)
	} else if err != nil {
		s.Log.Errorf("get current application instance error:%s", err.Error())
		return nil, err
	} else {
		cur.System = *system
	}

	appIns, err := NewAPPIns(cur, *system, *app)
	if err != nil {
		s.Log.Errorf("cur config couldn't use in target application [%s:%s],error:%s", app.AName, app.Version, err.Error())
		return nil, err
	}
	appIns.AID = aid
	appIns.CreateTime = int(time.Now().Unix())
	s.Log.Tracef("application instance with %d components", len(appIns.Components))

	j := &trait.JobRecord{
		ID:      -1,
		Current: cur,
		Target:  appIns,
	}

	return j, nil
}

// func (s *Store)

// NewJobRecord create a jobrecord for install/update
func (s *Store) NewJobRecord(ctx context.Context, aid int, sid int) (int, *trait.Error) {
	return s.NewJobRecordType(ctx, aid, sid, trait.JobUpgradeOType, trait.AppinitStatus)
}

// NewJobRecord create a jobrecord for install/update
func (s *Store) NewJobRecordType(ctx context.Context, aid int, sid int, otype int, status int) (int, *trait.Error) {
	// TODO
	tx, err := s.Store.Begin(ctx)
	if err != nil {
		s.Log.Errorf("start a trasaction error, please contact env maintainer: %s", err.Error())
		return -1, err
	}

	j, err := s.GetApplicationJobSnapshot(ctx, tx, aid, sid)
	if err != nil {
		defer s.rollbackWithLog(tx)()
		return -1, err
	}
	if otype != trait.JobDeleteOType {
		if j.Current == nil {
			otype = trait.JobInstallOType
		}
	}
	j.Target.OType = otype
	j.Target.Status = status

	// start job
	jid, err := tx.InsertJobRecord(ctx, j)
	if err != nil {
		s.Log.Errorf("create job record error: %s", err.Error())
		s.rollbackWithLog(tx)()
		return -1, err
	}
	j.ID = jid

	if err = tx.Commit(); err != nil {
		s.Log.Errorf("commit transaction error: %s", err.Error())
	}

	return jid, nil
}

// NewAPPIns create a application instance from application and current application.
// warn this will change the component instance in cur. don't use the cur after this function.
func NewAPPIns(cur *trait.ApplicationInstance, ss trait.System, target trait.Application) (*trait.ApplicationInstance, *trait.Error) {
	ains := &trait.ApplicationInstance{
		Application: target,
		ApplicationinstanceMeta: trait.ApplicationinstanceMeta{
			System: ss,
			Status: trait.AppinitStatus,
		},
	}

	coms := target.Components()
	ins := make([]*trait.ComponentInstance, 0, len(coms))

	for _, c := range coms {
		cins := new(trait.ComponentInstance)
		cn := c.GetComponentMeta().ComponentNode
		cins.Component = cn
		cins.Component.Name = cn.Name
		if cur != nil {
			cins = cur.ComponentInsExistedOrCreate(cn)
			// if err := c.Validate(cins.Config, cins.Attribute); err != nil {
			// 	return nil, err
			// }
		}
		// reset from new component
		cins.Acid = c.CID
		cins.Component.Version = cn.Version
		cins.Status = trait.AppinitStatus
		cins.APPName = target.AName
		cins.System = ss
		cins.Component.Version = cn.Version
		cins.Component.ComponentDefineType = cn.ComponentDefineType
		if cins.Timeout == 0 {
			cins.Timeout = c.Timeout()
		}
		ins = append(ins, cins)
	}
	ains.Components = ins
	if cur != nil {
		ains.AppConfig = cur.AppConfig
	}
	ains.Trait.UpgradeParent = false
	return ains, nil
}

// UploadApplicationPackage upload a application packge
func (s *Store) UploadApplicationPackage(ctx context.Context, r io.Reader) (int, *trait.Error) {
	app, tfs, err0 := builder.ParseApplication(r)
	if err0 != nil {
		s.Log.Errorf("parse log error: %s", err0.Error())
		err := &trait.Error{
			Err:      err0,
			Internal: trait.ErrApplicationFile,
		}
		return -1, err
	}
	if err := s.storeAddtionnalFiles(ctx, tfs); err != nil {
		return -1, err
	}

	if app.LangNames != nil {
		for k, v := range app.LangNames {
			if err := s.InsertAppLang(ctx, k, app.AName, v, "app"); err != nil {
				return -1, err
			}
		}
	}

	aid, err := s.Store.InsertAPP(ctx, app)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrUniqueKey) {
			// 更新依赖项
			err0 := s.Store.UpdateAppDependence(ctx, app)
			if err0 != nil {
				return -1, err0
			}
		}
		s.Log.Errorf("store application into store error: %s", err.Error())
		return -1, err
	}

	return aid, err
}

// storeAddtionnalFiles store ch into chart repo and upload config template
func (s *Store) storeAddtionnalFiles(ctx context.Context, fs map[*tar.Header][]byte) *trait.Error {
	for f, bs := range fs {
		if c := builder.ParseHelmChartMeta(f.Name); c != nil {
			if err := s.HelmRepo.Store(ctx, c, bs); err != nil {
				if trait.IsInternalError(err, trait.ErrHelmRepoNoFound) {
					s.Log.Errorf("the repo %s is not found", c.Repository)
				} else {
					s.Log.Errorf("store chart %s:%s error: %s", c.Name, c.Version, err)
				}
				return err
			}
		} else if strings.HasPrefix(f.Name, trait.ConfigTemplateDir) {
			cfg := &trait.AppliacationConfigTemplate{}
			if err := json.Unmarshal(bs, &cfg); err != nil {
				s.Log.Warnf("the file '%s' isn't config template file ignore it", f.Name)
				continue
			}
			if _, err := s.Store.InsertConfigTempalte(ctx, *cfg); err != nil {
				s.Log.Errorf("upload the config template file errpr: %s", err.Err)
				return err
			}
		}
	}
	return nil
}

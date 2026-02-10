package trait

import (
	"context"
	"fmt"
)

type ctx = context.Context

// Store is Application  and application instance db interface
type Store interface {
	Begin(c ctx) (Transaction, *Error)
	Close() *Error
	InitTablesFromDir(ctx context.Context, root string) *Error
	ApplicationWriter
	ApplicationInsWriter
	JobRecordWriter
	SystemWriter
	VerifyRecordReader
	ConfigTemplateWriter
	ProtonComponentWriter
	ComposeJobWriter
}

// Transaction deal things atomic
type Transaction interface {
	Commit() *Error
	Rollback() *Error
	ApplicationWriter
	ApplicationInsWriter
	JobRecordWriter
	SystemWriter
	ConfigTemplateWriter
	ComposeJobWriter
}

// ApplicationReader read application info
type ApplicationReader interface {
	ListAPP(c ctx, limit int, lastAID int) ([]ApplicationMeta, *Error)
	GetAPPID(c ctx, aname, aversion string) (int, *Error)
	SearchAPP(c ctx, limit int, lastAID int, name string) ([]ApplicationMeta, *Error)
	// warn!!! this is cross query with work application instance table
	ListSystemAPPNoWorked(ctx ctx, limit int, lastAID int, sid int) (as []ApplicationMeta, err *Error)
	GetAPP(c ctx, aid int) (*Application, *Error)
	GetAPPComponent(c ctx, acid int) (*ComponentMeta, *Error)
	LangReader
}

type LangReader interface {
	GetAppLang(ctx ctx, lang, aname string, zone string) (string, *Error)
	// // GetAname shouldn't imply for basic store. This is imply by cache. TODO move interface from here
	// GetAname(lang, alias, zone string) string
}

// ApplicationWriter write application info
type ApplicationWriter interface {
	ApplicationReader
	InsertAPP(ctx, Application) (int, *Error)
	InsertAppLang(ctx ctx, lang, aname, alias, zone string) *Error
	UpdateAppDependence(ctx context.Context, a Application) *Error
	DeleteAPP(ctx, int) *Error
}

// ApplicationInsReader read application info
type ApplicationInsReader interface {
	GetAPPIns(c ctx, id int) (*ApplicationInstance, *Error)
	GetWorkAPPIns(c ctx, name string, sid int) (*ApplicationInstance, *Error)
	// GetInsConfig(c ctx, id int) *ApplicationInstance
	// ListWorkAPPIns(c ctx, id, sid, limit int) ([]ApplicationInstanceOverview, *WrapperInternalError)
	CountWorkAppIns(c ctx, filter *AppInsFilter) (int, *Error)
	ListWorkAPPIns(c ctx, filter *AppInsFilter) ([]ApplicationInstanceOverview, *Error)
}

// AppInsFilter filter condition
type AppInsFilter struct {
	Sid     int
	Limit   int
	Offset  int
	Status  []int
	Name    string
	Jtype   []int
	Version string
}

// ApplicationInsWriter write application info
type ApplicationInsWriter interface {
	ApplicationInsReader
	ComponentInsWriter
	UpdateAPPInsConfig(c ctx, app ApplicationInstance) *Error
	// InsertAPPIns will insert componentIns, it must set the cid in appIns
	InsertAPPIns(c ctx, app *ApplicationInstance) (int, *Error)
	// UpdateConfig(c ctx, id int, config *ApplicationInstance) *WrapperInternalError
	UpdateAPPInsStatus(c ctx, id int, status int, owner int, startTime, endTime int) *Error
	UpdateAPPInsOperateType(ctx context.Context, id int, otype int) *Error

	LayOffAPPIns(c ctx, app *ApplicationInstance) *Error
	WorkAppIns(c ctx, app *ApplicationInstance) *Error
	LockApp(ctx ctx, sid int, jid int, aname string) *Error
	UnlockApp(ctx ctx, sid, jid int, aname string) *Error
}

// EdgeReader edge reader
type EdgeReader interface {
	CountEdgeTo(c ctx, cid int) (int, *Error)
}

// EdgeWrtier control relation with component node
type EdgeWrtier interface {
	EdgeReader
	DeleteEdgeFrom(c ctx, curCID int) *Error
	ChangeEdgeto(c ctx, curCID, tarCID int) *Error
	ChangeEdgeFrom(c ctx, curCID, tarCID int) *Error
	GetPointTo(c ctx, to int) ([]int, *Error)
	GetPointFrom(c ctx, from int) ([]int, *Error)
	DeleteEdge(c ctx, from, to int) *Error
	AddEdge(c ctx, from, to int) *Error
	AddOuterChildEdge(ctx ctx, from int, sid int, com ComponentNode) *Error
}

// ComponentInsReader read component ins info
type ComponentInsReader interface {
	GetWorkComponentIns(c ctx, sid int, com ComponentNode) (*ComponentInstance, *Error)
	GetComponentIns(c ctx, cid int) (*ComponentInstance, *Error)
	// GetComponentInsInApp(c ctx, cid int) ([]*ComponentInstance, *WrapperInternalError)
	ListWorkComponentIns(c ctx, filter WorkCompFilter) ([]*ComponentInstanceMeta, *Error)
}

// ComponentInsWriter write component info
type ComponentInsWriter interface {
	ComponentInsReader
	EdgeWrtier
	LockComponent(ctx ctx, sid int, jid int, cnode ComponentNode) *Error
	// warn: locker owner is the job.
	// if the job cross execute the same component task and job interupt abnormal without unlock,
	// the locker will be held by the job until the job next reexeute
	UnlockComponent(ctx ctx, sid, jid int, cnode ComponentNode) *Error
	UnlockJobComponent(ctx ctx, jid int) *Error
	LayoffComponentIns(c ctx, cid int) *Error
	// WorkComponentIns will add work instance id into obj
	WorkComponentIns(c ctx, cins *ComponentInstance) *Error
	UpdateComponentInsStatus(c ctx, cid, status, revision, startTime, endTime int) *Error

	// InserComponentIns must done by InsertAPPIns, so don't defined in interface
	// InsertComponentIns(c ctx, com *ComponentInstance) (int, *WrapperInternalError)

	// // DeleteComponentInsInApp only run when delete a application instance
	// DeleteComponentInsInApp(c ctx, id int) *WrapperInternalError
}

// JobRecordReader read job info
type JobRecordReader interface {
	GetJobRecord(c ctx, id int) (JobRecord, *Error)
	ListJobRecord(c ctx, f *AppInsFilter) ([]JobRecord, *Error)
	CountJobRecord(c ctx, f *AppInsFilter) (int, *Error)
	ListJobLog(c ctx, f JobLogFilter) ([]JobLog, *Error)
	CountJobLog(c ctx, f JobLogFilter) (int, *Error)
	/*
		ListJobRecordExecuting return job's link overview
		ListJobRecordExecuting should using when recovery from interupt
		@jobs is the job's id array
		@job[0] job id
		@job[1] job taget appliation instance id
		@job[2] job current appliation instance id
	*/
	// ListJobRecordExecuting(c ctx, id, limit int) (jobs [][3]int, err *WrapperInternalError)
}

// JobRecordWriter job record writer
type JobRecordWriter interface {
	JobRecordReader
	InsertJobRecord(c ctx, j *JobRecord) (int, *Error)
	InsertJobLog(ctx, JobLog) *Error
}

// SystemReader read system info
type SystemReader interface {
	GetSystemInfo(c ctx, sid int) (*System, *Error)
	DeleteSystemInfo(ctx context.Context, sid int) *Error
	ListSystemInfo(c ctx, limit int, offset int) ([]*System, *Error)
	CountSystemInfo(c ctx) (int, *Error)
	// ListSystemWithNameInfo(c ctx, name string, limit int, last int) ([]*System, *WrapperInternalError)
	GetSystemInfoByName(ctx ctx, name string) (*System, *Error)
}

// SystemWriter system config writer
type SystemWriter interface {
	SystemReader
	InsertSystemInfo(c ctx, s System) (int, *Error)
	UpdateSystemInfo(c ctx, s System) *Error
}

// VerifyRecordReader read verify record
type VerifyRecordReader interface {
	GetVerifyRecord(c ctx, jid int) (VerifyRecord, *Error)
	GetDataTestEntries(ctx ctx, did int, limit int, offset int) ([]DataTestEntry, *Error)
	CountDataTestEntries(ctx ctx, did int) (int, *Error)
	GetFunctionTestEntries(ctx ctx, fid int, limit int, offset int) ([]FunctionTestEntry, *Error)
	CountFunctionTestEntries(ctx ctx, fid int) (int, *Error)
}

// ConfigTemplateWriter application config template
type ConfigTemplateWriter interface {
	InsertConfigTempalte(ctx ctx, cfg AppliacationConfigTemplate) (int, *Error)
	UpdateConfigTemplate(ctx ctx, cfg AppliacationConfigTemplate) *Error
	DeleteConfigTemplate(ctx ctx, tid int) *Error
	ConfigReaderTemplateReader
}

// ConfigReaderTemplate read config template
type ConfigReaderTemplateReader interface {
	CountConfigTempalte(ctx ctx, f ApplicationConfigTemplateFilter) (int, *Error)
	ListConfigTemplate(ctx ctx, f ApplicationConfigTemplateFilter, limit, offset int) (cs []AppliacationConfigTemplateMeta, err *Error)
	GetConfigTemplate(ctx ctx, tid int) (cfg *AppliacationConfigTemplate, err *Error)
}

type ProtonComponentWriter interface {
	UpdateProtonComponent(ctx ctx, obj ProtonCompoent) *Error
	InsertProtonComponent(ctx ctx, obj ProtonCompoent) *Error
	ProtonComponentReader
}

type ProtonComponentReader interface {
	CountProtonConponent(ctx ctx, cname string, ctype string, sid int) (int, *Error)
	ListProtonConponent(ctx ctx, cname string, ctype string, sid int, limit, offset int) (cs []ProtonComponentMeta, err *Error)
	ListProtonConponentWithInternal(ctx context.Context, cname string, ctype string, sid int, limit, offset int) (cs []ProtonComponentMeta, err *Error)
	GetProtonComponent(ctx ctx, cname string, ctype string, sid int) (*ProtonCompoent, *Error)
}

// AppliacationConfigTemplateMeta config template meta data
type AppliacationConfigTemplateMeta struct {
	// 生产存储中，模板唯一ID，不需要填写
	Tid int `json:"tid"`
	// 模板名
	Tname string `json:"tname"`
	// 模板版本
	Tversion string `json:"tversion"`
	// 模板描述
	Tdescription string `json:"tdescription"`
	// 使用模板的应用名
	Aname string `json:"aname"`
	// 模板适用的应用版本
	Aversion string `json:"aversion"`
	// 模板场景标签
	Labels []string `json:"labels,omitempty"`
}

func checkStringFieldIsEmpty(filedName []string, values ...string) string {
	for i, fn := range filedName {
		if values[i] == "" {
			return fmt.Sprintf("filed '%s' is empty", fn)
		}
	}
	return ""
}

func (meta *AppliacationConfigTemplateMeta) Validate() *Error {
	if fe := checkStringFieldIsEmpty(
		[]string{"aname", "aversion", "tname", "tversion"},
		meta.Aname, meta.Aversion, meta.Tname, meta.Tversion); fe != "" {
		return &Error{
			Err:      fmt.Errorf("%s, the field must set", fe),
			Internal: ErrParam,
		}
	}
	return nil
}

// AppliacationConfigTemplate 应用配置模板
type AppliacationConfigTemplate struct {
	// 应用配置模板元信息
	AppliacationConfigTemplateMeta `json:",inline"`
	// 配置模板具体配置
	Config ApplicationConfigSet `json:"config"`
}

// ApplicationConfigSet 应用配置模板集，可以不含完整配置而仅含需要的配置
type ApplicationConfigSet struct {
	// 应用实例的应用级配置,每次修改后需要重新执行绑定的任务后生效
	AppConfig map[string]interface{} `json:"appConfig,omitempty"`
	// 字典格式的应用实例内的组件配置,字典key为组件名,值为组件配置
	Components map[string]*ComponentInstance `json:"components,omitempty"`
}

// ApplicationConfigTemplateFilter 应用配置模板过滤器
type ApplicationConfigTemplateFilter struct {
	// 应用包名称,启用其它任意过滤器时必填
	Aname string `json:"aname"`
	// 基于标签的选择器，不设置该字段参数时表示不筛选标签，选填。
	ApplicationLabelFilter *ApplicationLabelFilter `json:"labelFilter,omitempty"`
	// 基于版本兼容的选择器，不设置该字段参数时表示不筛选标签，选填
	ApplicationVersionFilter *ApplicationVersionFilter `json:"versionFilter,omitempty"`
}

// ApplicationVersionFilter 版本兼容条件过滤器
type ApplicationVersionFilter struct {
	// 应用包版本
	Aversion string `json:"aversion"`
	// 0或1时视为精确匹配应用版本号，2时为模糊予以版本兼容匹配
	Type int `json:"type"`
}

// ApplicationLabelFilter 应用配置模板标签过滤器
type ApplicationLabelFilter struct {
	// 标签字符串数组，数组元素关系由condition决定
	Labels []string `json:"labels"`
	// 标签数组元素关系，0和1时表示"或"关系，2表示"与"关系
	Condition int `json:"condition"`
}

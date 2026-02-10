package trait

import (
	"context"
	"encoding/json"
	"fmt"
)

type ComposeJobReader interface {
	ListComposeJob(ctx context.Context, limit, offset int, f ComposeJobFilter) ([]*ComposeJobMeata, int, *Error)
	GetCompoesJobTask(ctx context.Context, jid, jtindex int) (int, *Error)
	GetComposeJob(ctx context.Context, jid int) (*ComposeJob, *Error)
	GetCompoesJobTasks(ctx context.Context, jid int) ([][2]int, *Error)

	GetComposeManifests(ctx context.Context, name, version string) (*ComposeJobManifests, *Error)
	ListComposeManifest(ctx context.Context, limit, offset int, filter *ComposeManifestFilter) ([]*ComposeJobManifestsMeta, int, *Error)
	ListWorkComposeJobManifests(ctx context.Context, limit, offset int, filter ComposeJobFilter) ([]*ComposeJobMeata, int, *Error)
}

type ComposeJobWriter interface {
	InsertComposeJob(ctx context.Context, j ComposeJob) (int, *Error)
	SetComposeJob(ctx context.Context, j ComposeJob) *Error
	UpdateComposeJobProcess(ctx context.Context, jid, process int) *Error
	UpdateComposeJobStatus(ctx context.Context, jid, status, starttime, endtime int) *Error
	// InserComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *Error
	SetComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *Error
	// UpdateComposeJobTask(ctx context.Context, jid, jtindex, ajid int) *Error
	DeleteComposeJobTasks(ctx context.Context, jid int) *Error

	InsertComposeManifests(ctx context.Context, m ComposeJobManifests) *Error
	InsertWorkComposeManifests(ctx context.Context, obj ComposeJobMeata) *Error
	DeleteWorkComposeJobManifests(ctx, ComposeJobMeata) *Error
	ComposeJobReader

	GetWorkAPPIns(c ctx, name string, sid int) (*ApplicationInstance, *Error)
	GetAPP(c ctx, aid int) (*Application, *Error)

	// ApplicationInsReader
	// ApplicationReader
	// ComponentInsReader
}

type ComposeJobFilter struct {
	Status   []int
	Name     string
	SID      int
	ListType int
}

const (
	ComposeJobAllType    = 0
	ComposeJobNormalType = 1
	ComposeJobSuiteType  = 2
)

// ComposeJob 组合任务用于表示上层应用与基础资源组件的操作安装清单
type ComposeJob struct {
	ComposeJobMeata `json:",inline"`
	// 基础组件和应用任务配置清单
	Config ComposeJobConfig `json:"config"`
}

// ComposeJobMeata 组合任务资源清单列表
// 系统名与命名空间为第一次必填,后续调用则复用且不会更更改
type ComposeJobMeata struct {
	System `json:",inline"`
	// 任务名称,可为空
	Jname       string `json:"jname"`
	Title       string `json:"title"` // nolint:unused
	Mversion    string `json:"mversion,omitempty"`
	Description string `json:"description"`
	// 任务ID,该值自动生成
	Jid int `json:"jid"`
	// 任务状态, 4为成功,5为失败
	Status int `json:"status"`
	// 当前进度
	Processed int `json:"processed"`
	// 总任务数
	Total int `json:"total"`

	CreateTime int `json:"createTime"`
	StartTime  int `json:"startTime"`
	EndTime    int `json:"endTime"`
}

// LangKey impl LangSetter, no thread safe
func (c *ComposeJobMeata) LangKey() string {
	return c.Jname
}

// SetLang impl LangSetter, no thread safe
func (c *ComposeJobMeata) SetLang(l string) {
	c.Title = l
}

type ComposeJobConfig struct {
	// 访问地址配置,用于设置实例的默认入口网关配置
	AccessInfo AccessInfo `json:"access_info,omitempty"`
	// proton有状态基础资源组件配置
	ProtonComponent []json.RawMessage `json:"pcomponents"`
	// 上层应用安装任务配置,其中应用名称与版本为必填项
	AppConfig []*ApplicationInstance `json:"apps"`
}

type AccessInfo struct {
	Addr      string `json:"addr"`
	HttpsPort int    `json:"https_port"`
	HttpPort  int    `json:"http_port"`
}

func (i *AccessInfo) ToBytes() []byte {
	s := fmt.Sprintf(`access_addr: https://%s:%d
access_type: external
cert_download_feature: true
devicespec.conf: |+
  [DeviceSpec]
  HardwareType = AS10000

enable_http: false
ingress_network_mode: ""
language: zh_CN
mode: standard`, i.Addr, i.HttpsPort)
	return []byte(s)
}

type ComposeJobManifestsMeta struct {
	Name        string            `json:"mname"`
	Version     string            `json:"mversion"`
	Description string            `json:"description"`
	Title       string            `json:"title,omitempty"`
	LangNames   map[string]string `json:"langNames,omitempty"`
}

// LangKey impl LangSetter, no thread safe
func (c *ComposeJobManifestsMeta) LangKey() string {
	return c.Name
}

// SetLang impl LangSetter, no thread safe
func (c *ComposeJobManifestsMeta) SetLang(l string) {
	c.Title = l
}

type ComposeJobManifests struct {
	ComposeJobManifestsMeta `json:",inline"`
	Manifests               ComposeJobConfig `json:"config"`
}

type ComposeManifestFilter struct {
	NoWork bool   `json:"nowork"`
	Sid    int    `json:"sid"`
	Mname  string `json:"mname"`
}

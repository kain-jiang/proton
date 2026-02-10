package trait

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// ConfigSchema config doc and schema
type ConfigSchema struct {
	RawSchame json.RawMessage `json:"configSchema,omitempty"`
	// schema    *openapi3.Schema `json:"-"`
	schema gojsonschema.JSONLoader `json:"-"`
}

// Validate validate the obj
func (s *ConfigSchema) Validate(cfg map[string]interface{}) error {
	loader := gojsonschema.NewBytesLoader(s.RawSchame)
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	cfgLoader := gojsonschema.NewGoLoader(cfg)
	result, err := gojsonschema.Validate(loader, cfgLoader)
	if err != nil {
		return err
	}
	if !result.Valid() {
		buf := bytes.NewBufferString("")
		for _, err := range result.Errors() {
			buf.WriteString(err.String())
			buf.WriteString("\n")
		}
		return fmt.Errorf("validate error: %s", buf.String())
	}
	return nil
}

func newConfigSchema(raw json.RawMessage) (ConfigSchema, *Error) {
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	c := ConfigSchema{
		RawSchame: raw,
		schema:    gojsonschema.NewGoLoader(raw),
	}
	return c, nil
}

// Application store a package
type Application struct {
	ApplicationMeta `json:",inline"`
	// 应用包配置文档json格式对象
	ConfigSchema json.RawMessage `json:"configSchema,omitempty" swaggertype:"object"`
	configSchema *ConfigSchema   `json:"-"`
	// 应用包内组件数组
	Component []*ComponentMeta `json:"components"`
	// 应用包内组件依赖图
	Graph []Edge `json:"graph"`
	// uiSChema 控制前端部分组件样式渲染
	UISchema json.RawMessage `json:"uiSchema,omitempty" swaggertype:"object"`
	// Dependence 应用依赖
	Dependence []AppDepMeta `json:"dependence,omitempty"`
}

type AppDepMeta struct {
	// 应用包版本
	Version string `json:"version"`
	// 应用包名称
	AName string `json:"name"`
}

// NewUISChema create a new uischema
func (a *Application) NewUISChema() (sch ApplicationUISchema, err error) {
	err = json.Unmarshal(a.UISchema, &sch)
	return
}

// GetComponent retuen the compoent
func (a *Application) GetComponent(name string) *ComponentMeta {
	for _, c := range a.Component {
		if c.GetComponentMeta().Name == name {
			return c
		}
	}
	return nil
}

// Components return the component in application
func (a *Application) Components() []*ComponentMeta {
	return a.Component[:]
}

// Edge is component graph edge
type Edge struct {
	// 依赖起点组件
	From ComponentNode `json:"From"`
	// 依赖终点组件
	To ComponentNode `json:"To"`
}

// ApplicationMeta meta data
// this is a application overview
type ApplicationMeta struct {
	// 应用定义类型与版本,为未来预留
	Type string `json:"appDefineType"`
	// 应用包ID
	AID int `json:"aid"`
	// 应用包版本
	Version string `json:"version"`
	// 应用包名称
	AName string `json:"name"`
	// 应用包人类阅读名称,跟随国际化参数变化,无国际化映射时显示aname
	Alias string `json:"title"`
	// 应用包国际化名称映射
	LangNames map[string]string `json:"langNames,omitempty"`
}

// LangKey impl LangSetter, no thread safe
func (c *ApplicationMeta) LangKey() string {
	return c.AName
}

// SetLang impl LangSetter, no thread safe
func (c *ApplicationMeta) SetLang(l string) {
	c.Alias = l
}

// ApplicationUISchema ui schema for application
type ApplicationUISchema struct {
	AppConfig  map[string]interface{}        `json:"appConfig,omitempty"`
	Components map[string]*ComponentUISChema `json:"components,omitempty"`
}

type ApplicationTrait struct {
	UpgradeParent bool `json:"UpgradeParent"`
	RetainOrder   bool `json:"RetainOrder"`
}

var ApplicationTraitSchema = json.RawMessage(`{
    "title": "应用任务控制参数",
    "description": "应用任务控制参数用于控制应用任务整体行为",
    "type": "object",
    "properties": {
        "UpgradeParent": {
            "title": "是否执行父组件更新阶段",
            "description": "执行父组件更新阶段会更新所有依赖该应用包内组件的其他上层应用组件，默认不执行",
            "type": "boolean"
        },
		"RetainOrder": {
            "title": "是否包持旧版本的升级流程先升级组件",
            "description": "执行移除组件的时机，默认先移除，再升级",
            "type": "boolean"
        }
    }
}`)

// ApplicationInstance a application instance
type ApplicationInstance struct {
	Application             `json:"application"`
	ApplicationinstanceMeta `json:",inline"`
	Trait                   ApplicationTrait `json:"trait,omitempty"`
	// 应用实例的应用级配置,每次修改后需要重新执行绑定的任务后生效
	AppConfig map[string]interface{} `json:"appConfig,omitempty"`
	// 应用实例内的组件实例信息
	Components []*ComponentInstance `json:"components,omitempty"`
}

// ComponentInsExistedOrCreate return a exist ins or create a new then return
func (a *ApplicationInstance) ComponentInsExistedOrCreate(c ComponentNode) *ComponentInstance {
	ci := a.ComponentInsExisted(c)

	if ci == nil {
		ci = &ComponentInstance{
			ComponentInstanceMeta: ComponentInstanceMeta{
				Component: c,
			},
		}
		a.Components = append(a.Components, ci)
	}

	return ci
}

// ComponentInsExistedOrCreate return a exist ins or create a new then return
func (a *ApplicationInstance) ComponentInsExisted(c ComponentNode) *ComponentInstance {
	for _, ci := range a.Components {
		// if c.Name == ci.Component.Name && c.Version == ci.Component.Version {
		if c.Name == ci.Component.Name {
			return ci
		}
	}
	return nil
}

// Validate validate ins config
func (a *ApplicationInstance) Validate() *Error {
	if a.configSchema == nil {
		schema, err := newConfigSchema(a.ConfigSchema)
		if err != nil {
			err0 := fmt.Errorf("init config schame error: %s", err.Error())
			return &Error{
				Err:      err0,
				Internal: ErrApplicationFile,
				Detail:   err,
			}
		}
		a.configSchema = &schema
	}
	if err := a.configSchema.Validate(a.AppConfig); err != nil {
		err0 := fmt.Errorf("validate application config schame error: %s", err.Error())
		return &Error{
			Err:      err0,
			Internal: ErrConfigValidate,
			Detail:   err,
		}
	}
	for _, component := range a.Components {
		c := a.Application.GetComponent(component.Component.Name)
		if c == nil {
			continue
		}

		if err := c.Validate(component.Config, component.Attribute); err != nil {
			bs, _ := json.Marshal(component.Attribute)
			fmt.Printf("%s\n", bs)
			err0 := fmt.Errorf("validate component %s config schame error: %s", c.Name, err.Error())
			return &Error{
				Err:      err0,
				Internal: ErrConfigValidate,
				Detail:   err,
			}
		}
	}
	return nil
}

// ApplicationinstanceMeta ins meta
type ApplicationinstanceMeta struct {
	// 执行应用实例安装/升级的执行器的ID,正整数时表示在集群执行器内执行,为0表示未占有执行器,为负数时代表为命令行同步执行器执行
	Onwer int `json:"owner"`
	// 任务实例执行阶段与状态,
	// 0: 初始化;
	// 1：配置已确认;
	// 2: 任务等待执行;
	// 3: 任务正在执行;
	// 4: 任务执行成功;
	// 5: 任务失败;
	// 6: 任务已暂停;
	// 7: 任务停止中;
	// 8: 任务缺少依赖组件,失败;
	// 9: 任务卸载上一版本组件失败;
	// 11: 删除上一版本组件中;
	// 12: 目标版本服务更新完毕;
	// 13: 父组件更新阶段;
	// 14: 父组件更新失败;
	Status int `json:"status"`
	System `json:",inline"`
	// 应用实例ID
	ID int `json:"id"`
	// 备注,长度不可超过128字符
	Comment string `json:"comment"`
	// 创建时间
	CreateTime int `json:"createTime"`
	// 任务开始时间
	StartTime int `json:"startTime"`
	// 结束时间
	EndTime int `json:"endTime"`
	// 操作类型.
	// 值0,1和2分别对应更新,安装和回滚,这三个值仅用于使用者对任务进行标记,内部逻辑无差异.
	// 3 为删除操作
	OType int `json:"operateType"`
}

// ApplicationInstanceOverview ins overview
type ApplicationInstanceOverview struct {
	ApplicationMeta         `json:",inline"`
	ApplicationinstanceMeta `json:",inline"`
}

// ComponentInstanceMeta instance metadata
type ComponentInstanceMeta struct {
	// 组件实例所属的应用实例ID
	AIID int `json:"aiid"`
	// 组件实例ID
	CID int `json:"cid"`
	// 应用定义内组件定义的ID
	Acid int `json:"acid"`
	// 组件实例更新配置版本,当该版本变更时意味着组件由于更新或依赖更新的需要进行了一次变更,用于多任务信息同步和一致性管理
	Revission int `json:"-"`
	// 应用实例所属系统信息
	System System `json:"system"`
	// 组件定义基础信息
	Component ComponentNode `json:"component"`
	// 组件所属应用名称
	APPName string `json:"appName"`
	// 创建时间
	CreateTime int `json:"createTime"`
	// 任务开始时间
	StartTime int `json:"startTime"`
	// 结束时间
	EndTime int `json:"endTime"`
}

// ComponentInstance component instance info
type ComponentInstance struct {
	ComponentInstanceMeta  `json:",inline"`
	ComponentInstanceTrait `json:"trait"`
	// 组件配置,每次更改需要触发其任务执行
	Config map[string]interface{} `json:"config,omitempty"`
	// 组件属性配置,每次更改需要触发其任务执行
	Attribute map[string]interface{} `json:"attribute,omitempty"`
	AppConfig map[string]interface{} `json:"-"`
}

func (c *ComponentInstance) GetDeployTrait() map[string]interface{} {
	m := c.GetMiniTrait()
	m["aiid"] = c.AIID
	m["trait"] = map[string]interface{}{
		"status":  c.Status,
		"timeout": c.Timeout,
	}
	return map[string]interface{}{
		"deployTrait": m,
	}
}

func (c *ComponentInstance) GetMiniTrait() map[string]interface{} {
	return map[string]interface{}{
		"version": c.Component.Version,
	}
}

// ComponentUISChema ui schema for config
type ComponentUISChema struct {
	Config    map[string]interface{} `json:"config,omitempty"`
	Attribute map[string]interface{} `json:"attribute,omitempty"`
}

// ComponentInstanceTrait component instance common traits
type ComponentInstanceTrait struct {
	// 组件实例状态;1->初始化;4->成功;5->失败;6->暂停;7->停止中;8->缺失依赖组件而失败;10->忽略安装;
	Status int `json:"status"`
	// 任务执行超时时间设置
	Timeout int `json:"timeout"`
}

// JobRecord a install or upgrade job record
type JobRecord struct {
	// 任务ID
	ID int `json:"jid"`
	// 任务创建时的应用实例信息,在安装任务中为空
	Current *ApplicationInstance `json:"current"`
	// 任务创建时选择安装或升级的目标应用实例信息与配置
	Target *ApplicationInstance `json:"target"`
}

// System system info
type System struct {
	// 系统ID
	SID int `json:"sid,omitempty"`
	// 系统运行的k8s命名空间
	NameSpace string `json:"namespace,omitempty"`
	// 系统可视化名称,长度不大于20字符
	SName string `json:"systemName,omitempty"`
	// 系统级别配置,以字典对象表示。系统内应用与组件可见,优先级仅高于默认配置最低,会被其他层级配置覆盖
	Config map[string]interface{} `json:"config,omitempty" swaggertype:"object"`
	// 描述
	Description string `json:"description,omitempty"`
}

// // ComponentRuntimeConfig a instance runtime config
// type ComponentRuntimeConfig struct {
// 	ComponentInsData *ComponentInstance
// 	Topology         []*ComponentInstance
// 	System           *System
// 	AppConfig        []byte
// }

type JobLog struct {
	// 日志自增ID,主键
	JLID int `json:"jlid"`
	// 任务ID
	JID int `json:"jid"`
	// 组件实例ID
	CID int `json:"cid"`
	// 应用实例ID
	AIID int `json:"aiid"`
	// 应用名称
	Aname string `json:"aname"`
	// 国际化应用名称显示
	Alias string `json:"title"`
	// 组件名称
	Cname string `json:"cname"`
	// 日志分类
	Code int `json:"code"`
	// 日志信息
	Msg string `json:"msg"`
	// 日志记录日期时间戳
	Timestamp int `json:"time"`
}

type SortType string

const (
	// 降序排序
	DescSortType SortType = "DESC"
	// 升序排序
	ASCSortType SortType = "ASC"
)

type JobLogFilter struct {
	// 限制返回最大行数
	Limit int
	// 相同过滤条件下的偏移量
	Offset int
	// 排序枚举
	Sort SortType
	// 按job ID过滤,为-1时表示不过滤
	JID int
	// 按组件实例ID过滤,为-1时表示不过滤
	CID int
	// 按秒级时间戳过滤,0表示不过滤,负数表示为时间戳以前,正数表示时间戳以后
	Timestmp int
}

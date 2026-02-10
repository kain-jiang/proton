package trait

import (
	"encoding/json"
	"fmt"

	"taskrunner/pkg/log"
)

type ListParam struct {
	Limit  int
	Offset int
	// Order  int
}

type WorkCompFilter struct {
	ListParam
	Aname string
	Sid   int
}

// Component component imply ins
// TODO move to pkg/graph/task
type Component interface {
	// GetCompoentConfigSchema() (*ConfigSchema, *WrapperInternalError)
	// GetCompoentAttributeSchema() (*ConfigSchema, *WrapperInternalError)
	// GetConfig() map[string]interface{}
	// SetConfig(map[string]interface{})
	Validate(config, attribute map[string]interface{}) *Error
	GetComponentType() string
	GetComponentMeta() ComponentMeta
	Timeout() int
}

// Task component task runner
type Task interface {
	// SetJson(ctx, []byte) *WrapperInternalError
	WithLog(*log.TaskLogger)
	Install(ctx) *Error
	Uninstall(ctx) *Error
	Component() *ComponentNode
	Attribute() map[string]interface{}
	ComponentIns() *ComponentInstance
	// OldComponentIns() *ComponentInstance
	SetTopology(cs []*ComponentInstance)
	SetComponentIns(cins *ComponentInstance) *Error
}

// ComponentNode a node  in a application
type ComponentNode struct {
	// 组件名称
	Name string `json:"name"`
	// 组件版本
	Version string `json:"version"`
	// 组件定义方式
	ComponentDefineType string `json:"componentDefineType"`
	// 特殊可复用组件定义下的组件类型
	Type string `json:"type,omitempty"`
}

// GetComponentType impl component interfac, return the component type
func (c *ComponentNode) GetComponentType() string {
	return c.ComponentDefineType
}

// ComponentMeta mata
type ComponentMeta struct {
	ComponentNode `json:",inline"`
	CID           int `json:"-"`
	// 组件配置文档json格式对象
	RawConfigSchema json.RawMessage `json:"configSchema,omitempty" swaggertype:"object"`
	ConfigSchema    *ConfigSchema   `json:"-"`
	// 组件属性配置文档json格式对象
	RawAttributeSchema json.RawMessage `json:"attributeSchema,omitempty" swaggertype:"object"`
	AttributeSchema    *ConfigSchema   `json:"-"`
	// 本组件依赖的其他组件列表
	Deps []ComponentNode `json:"dependence"`
	// 组件定义特殊定义项,内部使用
	Spec []byte `json:"spec,omitempty" swaggerignore:"true"`
	// images explicit declaration
	// 组件任务执行超时时间
	DTimeout int `json:"timeout,omitempty"`
}

// AddEdgeInto add edge into input
func (c *ComponentMeta) AddEdgeInto(es []Edge) ([]Edge, *Error) {
	for _, dep := range c.Deps {
		if dep.Name == "" {
			return es, &Error{
				Internal: ECNULL,
				Err:      fmt.Errorf("component %s's deps is empty string", c.Name),
				Detail:   c.Name,
			}
		}
		es = append(es, Edge{
			From: ComponentNode{
				Name:    c.Name,
				Version: c.Version,
			},
			To: dep,
		})
	}
	return es, nil
}

// Timeout return component default timeout
func (c *ComponentMeta) Timeout() int {
	return c.DTimeout
}

// GetComponentMeta basic Component meta
func (c *ComponentMeta) GetComponentMeta() ComponentMeta {
	return *c
}

// GetCompoentAttributeSchema impl component interface
func (c *ComponentMeta) GetCompoentAttributeSchema() (*ConfigSchema, *Error) {
	if c.AttributeSchema == nil {
		sc, err := newConfigSchema(c.RawAttributeSchema)
		if err != nil {
			return nil, err
		}
		c.AttributeSchema = &sc
	}
	return c.AttributeSchema, nil
}

// GetCompoentConfigSchema impl component interface
func (c *ComponentMeta) GetCompoentConfigSchema() (*ConfigSchema, *Error) {
	if c.ConfigSchema == nil {
		sc, err := newConfigSchema(c.RawConfigSchema)
		if err != nil {
			return nil, err
		}
		c.ConfigSchema = &sc
	}
	return c.ConfigSchema, nil
}

// ValidateAttribute validate attribute
func (c *ComponentMeta) ValidateAttribute(cfg map[string]interface{}) *Error {
	sc, err := c.GetCompoentAttributeSchema()
	if err != nil {
		return err
	}
	if err := sc.Validate(cfg); err != nil {
		return &Error{
			Internal: ErrParam,
			Err:      err,
			Detail:   fmt.Sprintf("Component %s:%s", c.Name, c.Version),
		}
	}
	return nil
}

// ValidateConfig validate config
func (c *ComponentMeta) ValidateConfig(cfg map[string]interface{}) *Error {
	sc, err := c.GetCompoentConfigSchema()
	if err != nil {
		return err
	}
	if err := sc.Validate(cfg); err != nil {
		return &Error{
			Internal: ErrParam,
			Err:      err,
			Detail:   fmt.Sprintf("Component %s:%s", c.Name, c.Version),
		}
	}
	return nil
}

// Validate validate config and attribute
func (c *ComponentMeta) Validate(cfg, attr map[string]interface{}) *Error {
	if err := c.ValidateConfig(cfg); err != nil {
		return err
	}
	return c.ValidateAttribute(attr)
}

type ProtonComponentMeta struct {
	System `json:",inline"`
	Name   string `json:"name"`
	// Instance 为内部使用字段,不进行序列化
	Instance []string `json:"-"`
	Type     string   `json:"type,omitempty"`
}

type ProtonCompoent struct {
	ProtonComponentMeta `json:",inline"`
	Options             []byte
	Attribute           []byte
}

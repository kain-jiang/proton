package rest

import (
	"encoding/json"
	"fmt"
	"strconv"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

var componentTraits = json.RawMessage(`{
    "title": "组件通用控制参数",
    "description": "组件任务通用控制参数",
    "type": "object",
    "properties": {
        "timeout": {
            "title": "任务超时时间",
            "description": "任务执行超时时间,为0或不设置时使用默认超时时间",
            "type": "integer",
            "minimum": 0
        },
        "status": {
            "title": "状态",
            "description": "可以直接设置状态为成功或忽略,当状态为该两项时将会忽略该组件任务的执行",
            "type": "integer",
            "oneOf": [
                {
                    "const": 0,
                    "title": "初始化"
                },
                {
                    "const": 4,
                    "title": "成功"
                },
                {
                    "const": 5,
                    "title": "失败"
                },
                {
                    "const": 8,
                    "title": "依赖组件缺失,失败"
                },
                {
                    "const": 10,
                    "title": "忽略执行"
                }
            ]
        }
    }
}`)

// SchemaMeta json schema meta
type SchemaMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type appSChema struct {
	SchemaMeta `json:",inline"`
	Properties appProperties `json:"properties"`
}

type appProperties struct {
	// 应用任务控制参数
	Trait interface{} `json:"trait"`
	// 应用级配置文档
	App interface{} `json:"appConfig,omitempty"`
	// 组件级配置文档
	Components componentsSchema `json:"components"`
}

type componentsSchema struct {
	SchemaMeta `json:",inline"`
	// 以各个组件名称为properties的key值，values为对应组件schema文档
	Components map[string]componentSchema `json:"properties"`
}

type componentSchema struct {
	SchemaMeta `json:",inline"`
	Properties componentProperties `json:"properties"`
}

func newComponentSchema(c *trait.ComponentMeta) componentSchema {
	sc := componentSchema{
		SchemaMeta: SchemaMeta{
			Title:       c.Name,
			Description: "组件级配置,会覆盖应用级同名覆盖",
			Type:        "object",
		},
		Properties: componentProperties{
			// Config:    c.RawConfigSchema,
			// Attribute: c.RawAttributeSchema,
			Trait: componentTraits,
		},
	}

	if !jsonRawIsnull(c.RawAttributeSchema) {
		sc.Properties.Attribute = c.RawAttributeSchema
	}
	if !jsonRawIsnull(c.RawConfigSchema) {
		sc.Properties.Config = c.RawConfigSchema
	}
	return sc
}

type componentProperties struct {
	// 组件配置schema
	Config interface{} `json:"config,omitempty"`
	// 属性配置schema
	Attribute interface{} `json:"attribute,omitempty"`
	// 组件任务控制通用特性配置schema,详情见ComponentIntanceTrait对象文档
	Trait interface{} `json:"trait"`
}

func newSchemaFromApplication(a *trait.Application) appSChema {
	components := make(map[string]componentSchema, len(a.Component))
	for _, c := range a.Component {
		components[c.Name] = newComponentSchema(c)
	}

	var aSchema interface{}
	if !jsonRawIsnull(a.ConfigSchema) {
		aSchema = a.ConfigSchema
	}
	cSchema := componentsSchema{
		SchemaMeta: SchemaMeta{
			Title:       "组件配置",
			Description: "当组件配置非空时,同名组件配置项会覆盖应用级配置",
			Type:        "object",
		},
		Components: components,
	}

	return appSChema{
		SchemaMeta: SchemaMeta{
			Title:       a.AName,
			Description: "应用包配置",
			Type:        "object",
		},
		Properties: appProperties{
			App:        aSchema,
			Components: cSchema,
			Trait:      trait.ApplicationTraitSchema,
		},
	}
}

func parseIntFromQuery(ctx *gin.Context, params ...string) (res []int, err *HTTPError) {
	for _, p := range params {
		pStr := ctx.Query(p)
		if pStr == "" {
			return res, ParamError.From(fmt.Sprintf("%s must set a int, it is empty now", p))
		}
		pInt, err := strconv.Atoi(pStr)
		if err != nil {
			return res, ParamError.From(fmt.Sprintf("%s must set a int, it is '%s', convert error: %s", p, pStr, err.Error()))
		}
		res = append(res, pInt)
	}
	return
}

func ParseIntFromQueryWithDefault(ctx *gin.Context, params []string, dvs ...string) (res []int, err *HTTPError) {
	for i, p := range params {
		pstr := ctx.DefaultQuery(p, dvs[i])
		pInt, err := strconv.Atoi(pstr)
		if err != nil {
			return res, ParamError.From(fmt.Sprintf("%s must set a int, it is '%s', convert error: %s", p, pstr, err.Error()))
		}
		res = append(res, pInt)
	}
	return
}

func ConvertStringToIntArray(inputs ...string) (res []int, err error) {
	for _, s := range inputs {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		res = append(res, i)
	}
	return
}

var _nullJSONByte = []byte("null")

func jsonRawIsnull(raw json.RawMessage) bool {
	if raw == nil {
		return true
	}
	bs := []byte(raw)
	if len(bs) != len(_nullJSONByte) {
		return false
	}

	for i, j := range bs {
		if _nullJSONByte[i] != j {
			return false
		}
	}
	return true
}

func CheckStringFieldIsEmpty(filedName []string, values ...string) string {
	for i, fn := range filedName {
		if values[i] == "" {
			return fmt.Sprintf("filed '%s' is empty", fn)
		}
	}
	return ""
}

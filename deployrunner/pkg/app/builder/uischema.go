package builder

import (
	"encoding/json"
	"fmt"

	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/utils"
	"taskrunner/trait"
)

var _SchemaComposition = []string{"allOf", "anyOf", "oneOf"}

func parseUISchema(schema map[string]interface{}) (map[string]interface{}, error) {
	ctype, ok := schema["type"]
	if !ok {
		// return nil, fmt.Errorf("type 字段必须声明,且使用oneOf,allOf,anyOf等组合时类型必须为object")
		ctype = "object"
	}

	res := map[string]interface{}{}
	switch ctype {
	case "string":
		if v, ok := schema["dpsensitive"]; ok {
			isSensitive, ok := v.(bool)
			if ok && isSensitive {
				res = map[string]interface{}{"ui:widget": "password"}
			}
		}
	case "object":
		v, ok := schema["properties"]
		if ok {
			properties, ok := v.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("properties定义输入错误")
			}
			for k, v := range properties {
				attrSchema, ok := v.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("%s字段定义错误, 必须为properties属性字段格式, 目前为%#v", k, v)
				}
				uischema, err := parseUISchema(attrSchema)
				if err != nil {
					return nil, fmt.Errorf("%s 字段解析错误：%s", k, err.Error())
				}
				if len(uischema) != 0 {
					res[k] = uischema
				}
			}
		} else {
			state := 0
			var compositionSchema []map[string]interface{}
			for _, k := range _SchemaComposition {
				if v, ok := schema[k]; ok {
					state++
					list, ok := v.([]interface{})
					if !ok {
						return nil, fmt.Errorf("%s 组合定义错误, 必须为jsonschema对象数组,当前为 %#v", k, v)
					}
					for j, i := range list {
						sch, ok := i.(map[string]interface{})
						if !ok {
							return nil, fmt.Errorf("%s 组合第%d个对象定义错误, 必须为jsonschema对象数组,当前为 %#v", k, j, i)
						}
						compositionSchema = append(compositionSchema, sch)
					}
				}
			}
			if state == 0 {
				return nil, nil
			}
			if state != 1 || compositionSchema == nil {
				return nil, fmt.Errorf("未定义properties时,必须有且定义一个oneOf,anyOf或allOf")
			}
			for i, sc := range compositionSchema {
				uischema, err := parseUISchema(sc)
				if err != nil {
					return nil, fmt.Errorf("组合定义第%d个定义错误: %s", i+1, err.Error())
				}
				if len(uischema) != 0 {
					res = utils.MergeMapsIgnoreNil(res, uischema)
				}
			}

		}

	case "array":
		items, ok := schema["items"]
		if ok {
			itemsSchema, ok := items.(map[string]interface{})
			if ok {
				uischema, err := parseUISchema(itemsSchema)
				if err != nil {
					return nil, fmt.Errorf("数组类items定义错误: %s", err.Error())
				}
				if len(uischema) != 0 {
					res = map[string]interface{}{"items": uischema}
				}
			}

		}
	}

	if len(res) == 0 {
		res = nil
	}
	return res, nil
}

func getUIShemaObj(raw json.RawMessage) (map[string]interface{}, error) {
	if len(raw) == 0 || raw == nil {
		return nil, nil
	}
	schema := make(map[string]interface{})
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}
	if len(schema) == 0 {
		return nil, nil
	}

	return parseUISchema(schema)
}

func parseAPPUISchema(app trait.Application) (*trait.ApplicationUISchema, error) {
	appConf, err := getUIShemaObj(app.ConfigSchema)
	if err != nil {
		err = fmt.Errorf("解析应用config uischema错误: %s", err.Error())
		return nil, err
	}
	componentSchemas := map[string]*trait.ComponentUISChema{}
	for _, c := range app.Component {
		oldSchema := c.RawConfigSchema
		oldAttrSchema := c.RawAttributeSchema
		if err := resources.ReplaceApplicationComponentSchame(c); err != nil {
			err0 := fmt.Errorf("替换基础资源组件%s的jsonschema配置文档错误: %s", c.Type, err.Error())
			return nil, err0
		}
		attr, err := getUIShemaObj(c.RawAttributeSchema)
		if err != nil {
			err = fmt.Errorf("解析组件%s attribute uischema错误: %s", c.Name, err.Error())
			return nil, err
		}

		conf, err := getUIShemaObj(c.RawConfigSchema)
		if err != nil {
			err = fmt.Errorf("解析组件%s config uischema错误: %s", c.Name, err.Error())
			return nil, err
		}
		if attr != nil || conf != nil {
			component := &trait.ComponentUISChema{}
			component.Attribute = attr
			component.Config = conf
			componentSchemas[c.Name] = component
		}
		// recovery schema document
		c.RawConfigSchema = oldSchema
		c.RawAttributeSchema = oldAttrSchema
	}
	if len(componentSchemas) == 0 {
		componentSchemas = nil
	}

	uiSchema := &trait.ApplicationUISchema{
		AppConfig:  appConf,
		Components: componentSchemas,
	}
	return uiSchema, nil
}

// SetAPPUISchema set uischema from application
// TODO move this code
func SetAPPUISchema(app *trait.Application) *trait.Error {
	uischem, err := parseAPPUISchema(*app)
	if err != nil {
		return &trait.Error{
			Internal: trait.ErrApplicationFile,
			Err:      err,
			Detail:   "parse application ui schema from schema",
		}
	}
	bs, err := json.Marshal(uischem)
	if err != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "encode application ui schema",
		}
	}
	app.UISchema = bs
	return nil
}

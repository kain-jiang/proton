package trait

import (
	"encoding/json"
	"testing"
)

func TestComponentValidate(t *testing.T) {
	rawAttr := `{
		"title": "第三方OCR服务",
		"description": "第三方OCR服务",
		"type": "object",
		"oneOf": [
			{
				"title": "合合OCR",
				"description": "合合OCR",
				"type": "object",
				"properties": {
					"source_type": {
						"type": "string",
						"description": "服务标识",
						"const": "hehe",
						"default": "hehe"
					},
					"document": {
						"description": "document",
						"properties": {
							"switch": {
								"type": "string",
								"enum": [
									"true",
									"false"
								],
								"description": "服务使能开关"
							},
							"host": {
								"type": "string",
								"description": "服务器"
							},
							"port": {
								"type": "string",
								"description": "端口"
							},
							"protocol": {
								"type": "string",
								"description": "服务连接协议",
								"enum": [
									"https",
									"http"
								]
							}
						}
					}
				}
			},
			{
				"title": "第四范式OCR",
				"description": "第四范式OCR",
				"type": "object",
				"properties": {
					"source_type": {
						"type": "string",
						"description": "服务标识",
						"const": "t4th",
						"default": "t4th"
					},
					"document": {
						"description": "document",
						"properties": {
							"switch": {
								"type": "string",
								"enum": [
									"true",
									"false"
								],
								"description": "服务使能开关"
							},
							"host": {
								"type": "string",
								"description": "服务器"
							},
							"port": {
								"type": "string",
								"description": "端口"
							},
							"connectType": {
								"type": "string",
								"enum": [
									"ocr",
									"fileReader"
								],
								"description": "连接第四范式ORC的方式"
							},
							"protocol": {
								"type": "string",
								"description": "服务连接协议",
								"enum": [
									"https",
									"http"
								]
							}
						}
					}
				}
			}
		]
	}`
	attr := `{
		"source_type": "hehe",
		"document": {
			"switch": "true",
			"host": "123.123.123..123",
			"port": "123",
			"connectType": "fileReader"
		}
	}`
	c := &ComponentMeta{
		RawAttributeSchema: []byte(rawAttr),
	}
	attrMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(attr), &attrMap)
	if err := c.ValidateAttribute(attrMap); err != nil {
		t.Error(err)
		t.Errorf("%#v\n", attrMap)
	}

	doc := `{
		"type": "object",
		"properties": {
			"addr": {
				"type": "string",
				"format": "uri"
			}
		}
	}`
	schema, rerr := newConfigSchema(json.RawMessage(doc))
	if rerr != nil {
		t.Fatal(rerr.Error())
	}
	if err := schema.Validate(map[string]interface{}{
		"addr": "q2310987",
	}); err == nil {
		t.Fatal("must error")
	} else {
		t.Log(err.Error())
	}

	if err := schema.Validate(map[string]interface{}{
		"addr": "q2310987",
	}); err == nil {
		t.Fatal("must error")
	} else {
		t.Log(err.Error())
	}
}

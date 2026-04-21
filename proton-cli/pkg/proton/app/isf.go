package app

// StripISFRdsDatabase 返回一份 values 的深拷贝，并删除 depServices.rds.database 字段。
//
// ISF 的各 chart 自己会指定具体的 MariaDB database 名，不应由顶层 values 下发一个
// 统一的 rds.database，否则会把连接串错误地指向同一个库。该行为与
// deploy/scripts/services/isf.sh 中安装 ISF 前用 `sed '/database:/d'` 剔除
// rds.database 字段的逻辑保持一致。
//
// 非 ISF 产品（kweaver-core / kweaver-dip 等）仍保留 database 字段。
func StripISFRdsDatabase(values map[string]any) map[string]any {
	out := deepCopyValues(values)
	dep, ok := out["depServices"].(map[string]any)
	if !ok {
		return out
	}
	rds, ok := dep["rds"].(map[string]any)
	if !ok {
		return out
	}
	delete(rds, "database")
	return out
}

// deepCopyValues 对 values map 做递归深拷贝，仅处理 map[string]interface{} 嵌套；
// 其它类型（slice/基本类型）原样引用，对本场景足够（values 不会在后续被原地改写）。
func deepCopyValues(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		if m, ok := v.(map[string]any); ok {
			out[k] = deepCopyValues(m)
		} else {
			out[k] = v
		}
	}
	return out
}

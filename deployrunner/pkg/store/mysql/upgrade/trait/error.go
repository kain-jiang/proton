package trait

import "errors"

var (
	// ErrUnknowFile 未知文件
	ErrUnknowFile = errors.New("未知文件格式")

	// ErrOverPlan 进度超出总计划进度
	ErrOverPlan = errors.New("进度超出总计划进度")

	// ErrNotFound 数据为找到
	ErrNotFound = errors.New("数据未找到")

	// ErrOrder 计划进度与程序不匹配
	ErrOrder = errors.New("计划进度序号不可能大于总计划进度,请确认代码合并情况")

	// ErrNotExpect 非期望对象
	ErrNotExpect = errors.New("非期望对象")

	// ErrParam 参数错误
	ErrParam = errors.New("参数错误")
)

package trait

import (
	"encoding/json"
	"sort"

	ttrait "taskrunner/trait"

	"github.com/sirupsen/logrus"
)

// WorkEnv 计划执行过程中可用的全局环境和工具等
type WorkEnv struct {
	Log *logrus.Logger
}

// Error 通用error别名
type Error = *ttrait.Error

// OperatorMeta 算子定义元数据
type OperatorMeta struct {
	// 算子名字，一般算子不需要名字，用于存在依赖的算子
	Name string `json:"name"`
	// 算子类型
	Command string `json:"command"`
}

// Operator 算子表示
type Operator struct {
	OperatorMeta `json:",inline"`
	// 各个算子实现提供的参数配置.
	Args json.RawMessage `json:"args"`
	// 包装的组件
	Compose []Operator `json:"compose"`
}

// ExecuteOP execute算子定义与执行实现
type ExecuteOP struct {
	//    执行语句
	Statements []string
	// //   开启事务
	// transaction bool
}

// PlanMeta  执行计划元数据
type PlanMeta struct {
	// 从文件名中自动提取的执行计划文件日期，必须精确到秒
	DateID int
	// 从文件名中自动提取的执行计划代数，一般跟随版本单调递增
	Epoch int
	// 服务名称，可用于隔离不同服务的执行
	ServiceName string
	// 执行阶段，即pre或post
	Stage int
	// 子总计划中的排序
	Order int `json:"-"`
}

// Plan 执行计划文件表示
type Plan struct {
	PlanMeta
	// 执行计划文件内的表示
	Operators []Operator
}

// PlanProcess 执行计划记录, 持久化在生产环境存储中
type PlanProcess struct {
	PlanMeta
	// 执行计划状态
	Status int
	// 当前执行的算子记录
	Op OperatorProcess
}

// OperatorProcess 执行计划中某个算子的执行记录
type OperatorProcess struct {
	PlanMeta
	// 算子在执行计划中的排序，因执行计划文件不可更改，因此可直接使用排序作为唯一id
	OrderID int
	// 执行状态
	Status int
	// 过程中间数据，由对应算子定义、序列化、反序列化和各种使用。如用于内部进度控制等
	TempData json.RawMessage
}

// StagePlan 一个服务的数据库升级各个阶段计划集合
type StagePlan struct {
	SvcName string
	// 0: pre
	// 1: post
	// 3: init
	Pre [3][]Plan
}

// Sort 基于dateID排序
func (p *StagePlan) Sort() {
	for _, k := range p.Pre {
		sort.Slice(k, func(i, j int) bool {
			return k[i].DateID < k[j].DateID
		})
	}
}

// Plans 对执行计划目录遍历获得的总体执行计划，与执行程序绑定。基于环境执行计划会重新排序与执行。
type Plans struct {
	Plans map[string]StagePlan
}

const (
	// PreStage pre
	PreStage = "pre"
	// PostStage post
	PostStage = "post"
	// InitStage init
	InitStage = "init"
)

const (
	// PreStageInt pre
	PreStageInt = 0
	// PostStageInt post
	PostStageInt = 1
	// InitStageInt init
	InitStageInt = 2
)

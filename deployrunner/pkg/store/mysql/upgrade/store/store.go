package store

import (
	"context"

	"taskrunner/pkg/store/mysql/upgrade/trait"
)

// Store 存储执行计划进度
type Store interface {
	// Record 用于存储进度与更新进度
	Record(context.Context, trait.PlanProcess) trait.Error
	// Last 用于获取最新的存储进度
	Last(ctx context.Context, svcName string, stage int) (trait.PlanProcess, trait.Error)
	// Less 用于进度定位与获取进度列表
	Less(ctx context.Context, svcName string, DateID, limit int, stage int) ([]trait.PlanProcess, trait.Error)
	// Greate 用于进度定位与获取进度列表
	// Great(ctx context.Context, svcName string, DateID, limit int, stage int) ([]trait.OperatorProcess, trait.Error)
	// Get 用于获取具体的进度信息
	Get(ctx context.Context, svcName string, DateID int) (*trait.PlanProcess, trait.Error)
}

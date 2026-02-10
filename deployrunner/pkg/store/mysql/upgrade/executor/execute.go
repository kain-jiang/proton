package executor

import (
	"context"
	"sort"

	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/store/mysql/upgrade/operator"
	"taskrunner/pkg/store/mysql/upgrade/store"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	ttrait "taskrunner/trait"

	"github.com/sirupsen/logrus"
)

// Option 执行可选参数
type Option struct {
	Stage       int
	Logger      *logrus.Logger
	AtLeastOnce bool
}

func ExecuteDir(ctx context.Context, root string, objs operator.ObjectStore, rds resources.RDS, opts Option, exclude ...string) (err trait.Error) {
	s, err := store.NewStore(ctx, rds)
	if err != nil {
		return err
	}
	ps, err := BuildMultiSvcFromDir(root, rds.Type, exclude...)
	if err != nil {
		return err
	}
	for _, p := range ps.Plans {
		if err := Execute(ctx, objs, s, p, opts); err != nil {
			return err
		}
	}
	return nil
}

// Execute 执行一个服务的整体计划
func Execute(ctx context.Context, objs operator.ObjectStore, s store.Store, p trait.StagePlan, opts Option) (err trait.Error) {
	if opts.Logger == nil {
		log := logrus.New()
		log.SetReportCaller(true)
		opts.Logger = log
	}
	log := opts.Logger
	w := &trait.WorkEnv{
		Log: log,
	}
	stage := opts.Stage

	if stage != trait.PostStageInt {
		// 只在pre阶段或混合模式下运行

		// init阶段必会运行不跳过
		once := opts.AtLeastOnce
		opts.AtLeastOnce = true
		initPlan, gerr := getPlans(ctx, s, p, trait.InitStageInt, opts)
		if gerr != nil {
			return gerr
		}
		log.Infof("执行模块%s init阶段, 待执行计划长度%d", p.SvcName, len(initPlan))
		if err := execute(ctx, objs, s, initPlan, w, opts); err != nil {
			return err
		}
		opts.AtLeastOnce = once
	}

	var plans []trait.Plan

	// make plan order by stage
	if stage == trait.PreStageInt || stage == trait.PostStageInt {
		plans, err = getPlans(ctx, s, p, stage, opts)
		if err != nil {
			return err
		}
		log.Infof("执行模块%s stage: %d, 待执行计划长度%d", p.SvcName, stage, len(plans))
	} else if stage == trait.InitStageInt {
		// only init stage
		return nil
	} else {
		// mix stag
		pres, gerr := getPlans(ctx, s, p, trait.PreStageInt, opts)
		if gerr != nil {
			return gerr
		}
		log.Infof("执行模块%s stage: pre, 待执行计划长度%d", p.SvcName, len(pres))
		posts, gerr := getPlans(ctx, s, p, trait.PostStageInt, opts)
		if gerr != nil {
			return nil
		}
		log.Infof("执行模块%s stage: post, 待执行计划长度%d", p.SvcName, len(posts))
		i := 0
		j := 0
		lenPre := len(pres)
		lenPost := len(posts)
		for i < lenPre && j < lenPost {
			if pres[i].Epoch <= posts[j].Epoch {
				plans = append(plans, pres[i])
				i++
			} else {
				plans = append(plans, posts[j])
				j++
			}
		}
		for i < lenPre {
			plans = append(plans, pres[i])
			i++
		}
		for j < lenPost {
			plans = append(plans, posts[j])
			j++
		}
	}

	return execute(ctx, objs, s, plans, w, opts)
}

func planExecuteErrorHandler(ctx context.Context, s store.Store, pr trait.PlanProcess) {
	pr.Status = ttrait.AppFailStatus
	if err := s.Record(ctx, pr); err != nil {
		// TODO log
		panic(err)
	}
}

func execute(ctx context.Context, objs operator.ObjectStore, s store.Store, p []trait.Plan, w *trait.WorkEnv, opts Option) trait.Error {
	log := w.Log
	for _, pl := range p {
		process, err := s.Get(ctx, pl.ServiceName, pl.DateID)
		if err != nil {
			if ttrait.IsInternalError(err, ttrait.ErrNotFound) {
				process = nil
			} else {
				return err
			}
		}
		if process == nil {
			process = &trait.PlanProcess{
				PlanMeta: pl.PlanMeta,
				Status:   ttrait.AppDoingStatus,
				Op: trait.OperatorProcess{
					PlanMeta: pl.PlanMeta,
					Status:   ttrait.AppDoingStatus,
					OrderID:  0,
				},
			}
		}

		if process.Status == ttrait.AppSucessStatus {
			// dateID:    0 1 3 4   --> 0 1 2 3
			// orderID:   0 1 2 3 4 --> 0 1 2 2 3 runtime
			// orderID:   0 1 2 3 4 --> 0 1 2 3 3 runtime
			// Last() runtime can get right result
			// Lest() runtime can get right result
			// the wrong order won't cause error
			// orderID:   0 1 2 3 4 --> 0 1 2 3 4
			if process.Order != pl.Order {
				process.Order = pl.Order
				log.Infof("%s计划%d已成功, 仅调整order", process.ServiceName, process.DateID)
				if err := s.Record(ctx, *process); err != nil {
					return err
				}
			}
			continue
		}

		if err := s.Record(ctx, *process); err != nil {
			log.Infof(
				"%s计划%d执行状态记录失败, error: %s",
				process.ServiceName, process.DateID,
				err.Error(),
			)
			return err
		}
		lastOrder := process.Op.OrderID
		ops := pl.Operators[process.Op.OrderID:]
		for j, op := range ops {
			e, err := operator.NewExecutor(objs, op)
			if err != nil {
				planExecuteErrorHandler(ctx, s, *process)
				return err
			}
			process.Op = trait.OperatorProcess{
				PlanMeta: pl.PlanMeta,
				OrderID:  j + lastOrder,
				Status:   ttrait.AppDoingStatus,
			}
			if err := e.Execute(ctx, w, s, *process); err != nil {
				process.Op.Status = ttrait.AppFailStatus
				log.Errorf(
					"执行%s计划%d的第%d个算子失败, error: %s",
					process.ServiceName, process.DateID, j, err.Error(),
				)
				planExecuteErrorHandler(ctx, s, *process)
				return err
			}

		}
		process.Status = ttrait.AppSucessStatus
		if err := s.Record(ctx, *process); err != nil {
			log.Infof(
				"%s计划%d执行状态记录失败, error: %s",
				process.ServiceName, process.DateID,
				err.Error(),
			)
			return err
		}
		log.Infof("%s计划%d执行成功", process.ServiceName, process.DateID)

	}
	return nil
}

func getPlans(ctx context.Context, s store.Store, p trait.StagePlan, stage int, opts Option) ([]trait.Plan, trait.Error) {
	log := opts.Logger
	log.Debugf("stage: %d, get plan length %d", stage, len(p.Pre[stage]))
	if len(p.Pre[stage]) == 0 {
		return nil, nil
	}
	sort.Slice(p.Pre[stage], func(i, j int) bool {
		return p.Pre[stage][i].DateID < p.Pre[stage][j].DateID
	})
	for i, j := range p.Pre[stage] {
		j.Order = i
	}
	index, err := IndexPlan(ctx, s, p, stage)
	if err != nil {
		return nil, err
	}
	i := 0
	if index != nil {
		i = index.Order
	} else if !opts.AtLeastOnce {
		// 至少运行一次
		// 主要原因为init阶段的计划是可变更的,从而导致pre和post阶段在安装时是不必要且可能出错的
		// 用于已有系统接入系统时需要进行升级而由于未接入不存在进度时
		// 或模块从未升级不存在进度时
		// 因此该配置在"升级"场景应该设置为true,"安装"场景设置为false以避免无效脚本
		// 后续已有环境再升级由于存在进度该配置将会无效
		lastOrder := len(p.Pre[stage]) - 1
		last := p.Pre[stage][lastOrder]
		// 记录进度以跳过执行
		if err := s.Record(ctx, trait.PlanProcess{
			PlanMeta: last.PlanMeta,
			Status:   ttrait.AppDoingStatus,
		}); err != nil {
			return nil, err
		}
		return p.Pre[stage][lastOrder:], nil
	}
	return p.Pre[stage][i:], nil
}

// IndexPlan 计划定位
func IndexPlan(ctx context.Context, s store.Store, p trait.StagePlan, stage int) (*trait.PlanProcess, trait.Error) {
	i, err := s.Last(ctx, p.SvcName, stage)
	if ttrait.IsInternalError(err, ttrait.ErrNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ps := p.Pre[stage]
	if len(ps) <= i.Order {
		return nil, &ttrait.Error{
			Internal: ttrait.ErrParam,
			Err:      trait.ErrOverPlan,
			Detail:   "index upgrade plan error",
		}
	}

	for {
		if ps[i.Order].DateID == i.DateID {
			return &i, nil
		} else if ps[i.Order].DateID < i.DateID {
			l, err := s.Less(ctx, p.SvcName, ps[i.Order].DateID, 100, stage)
			if err != nil {
				return nil, err
			}
			if len(l) == 0 {
				// 无公共父节点意味着从计划头部开始完整执行
				return &trait.PlanProcess{}, nil
			}
			// search from plan slices
			for j := len(l) - 1; j >= 0; j-- {
				i = l[j]
				if ps[i.Order].DateID == i.DateID {
					return &i, nil
				}
			}
		} else {
			return nil, &ttrait.Error{
				Internal: ttrait.ErrParam,
				Err:      trait.ErrOrder,
				Detail:   "index upgrade plan order error",
			}
		}
	}
}

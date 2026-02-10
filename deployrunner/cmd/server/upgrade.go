package main

import (
	"context"
	"errors"
	"fmt"
	"math"

	// "taskrunner/cmd/utils"
	"path/filepath"

	cutils "taskrunner/cmd/utils"
	"taskrunner/pkg/component/resources"
	driver "taskrunner/pkg/sql-driver"
	"taskrunner/pkg/store/mysql/upgrade/executor"
	"taskrunner/pkg/store/mysql/upgrade/store"
	"taskrunner/trait"

	utrait "taskrunner/pkg/store/mysql/upgrade/trait"

	"github.com/spf13/cobra"
)

var (
	svcDir      = ""
	allDir      = ""
	stage       = ""
	excludeSvc  = []string{}
	skipUpgrade = false
)

func newDBConnWithInit(ctx context.Context, rds resources.RDS) (driver.DBConn, *trait.Error) {
	if err := cutils.InitDatabase(ctx, rds); err != nil {
		return nil, err
	}
	conn, err := driver.Factory.NewDBConn(ctx, rds)
	return conn, err
}

func newUpgradeCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use: "data",
		Long: `
		运行指定目录下服务或目录下所有服务。
		通过stage参数指定运行模式:
		- init: 仅运行安装初始化init阶段,常用于第一次安装,避免无意义的pre和post阶段.
		- pre: 运行所有未运行的int和pre阶段任务,常用于临近版本更新升级.
		- post: 运行所有未运行的post阶段任务,常用于临近版本更新升级.
		- mix: 混合执行pre和post任务,epoch交替执行, 同一epoch下先执行所有pre再执行post. 常用于跨版本升级`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := icfg.NewLogger()

			var stageInt int
			if stage == utrait.PreStage {
				stageInt = utrait.PreStageInt
			} else if stage == utrait.PostStage {
				stageInt = utrait.PostStageInt
			} else if stage == "init" {
				stageInt = utrait.InitStageInt
			} else if stage == "mix" {
				// 运行时参数无该配置项，用于标识混合执行
				stageInt = math.MaxInt
			} else {
				return fmt.Errorf("unsport stage %s", stage)
			}

			opts := executor.Option{
				Stage:       stageInt,
				Logger:      log,
				AtLeastOnce: !skipUpgrade,
			}

			cfg, err := icfg.LoadFromYamlFile(ctx)
			if err != nil {
				return err
			}
			rds := cfg.Rds
			db, err := newDBConnWithInit(ctx, rds)
			if err != nil {
				return err
			}
			objStore := map[string]any{
				"default": db,
			}

			if svcDir != "" {
				// 指定服务运行
				s, err := store.NewStore(ctx, rds)
				if err != nil {
					return err
				}
				svcName := filepath.Base(svcDir)
				p, err := executor.BuildSvc(svcDir, svcName, rds.Type)
				if err != nil {
					return err
				}
				if err := executor.Execute(cmd.Context(), objStore, s, p, opts); err != nil {
					return err
				}

			} else if allDir != "" {
				excludeSvc = append(excludeSvc, "taskrunner-multi")
				if err := executor.ExecuteDir(ctx, allDir, objStore, rds, opts, excludeSvc...); err != nil {
					return err
				}
				return nil
			} else {
				return errors.New("至少需要以serivce或all参数选择运行任务")
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&svcDir, "service", "", "服务升级包目录,该参数优先级高于all")
	cmd.Flags().StringVar(&allDir, "all", "", "所有服务升级包目录")
	cmd.Flags().StringVarP(&stage, "stage", "s", "", "运行模式,仅支持pre, post, init和mix三种模式")
	cmd.Flags().StringArrayVarP(&excludeSvc, "exclude", "e", nil, "排除模块列表,通过重复指定传入多个")
	cmd.Flags().BoolVar(&skipUpgrade, "skip", false, "用于在计划进度记录的场景进行安装时跳过各个阶段执行,仅执行初始化阶段,命令行工具常用于升级因此默认不条跳过")
	if err := cmd.MarkFlagRequired("stage"); err != nil {
		panic(err.Error())
	}
	if err := cmd.MarkFlagRequired("skip"); err != nil {
		panic(err.Error())
	}

	icfg.AddStoreFlags(cmd)

	return cmd
}

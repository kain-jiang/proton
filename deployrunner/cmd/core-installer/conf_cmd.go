package main

import (
	"fmt"

	"taskrunner/cmd/version"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	pnamespace         = "proton"
	inputConfigExample = `
# 该文件一般用于命令行参数未开放配置的配置,如各个服务的pod安全上下文
# 同时如果有命令行参数参与控制的配置,命令行控制配置优先级最高。如副本数。
#
# 应用级配置,会被下文的组件级配置覆盖,如下文调整整体resources.requests
# 一场景该配置使用默认值即可
appConfig:
  replicaCount: 1
  resources:
    requests:
      cpu: 10m
      memory: 10Mi

# 组件级配置,会覆盖上文的配置
# 同时如果有命令行参数参与控制的,命令行控制配置优先级最高,如下文的mode参数
components:
  # 组件名,以指定组件的特定配置
  deploy-installer:
    replicacount: 1
    resources:
      requests:
        cpu: 1m
        memory: 1Mi
`
)

func newConfExampleCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "example",
		Short:   "配置文件样例",
		Long:    "该命令主要用于需要额外配置时的配置文件写法简易介绍。如果已安装建议通过get命令获取已有配置再修改。",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(inputConfigExample)
			return nil
		},
	}
}

func newConfCmd() *cobra.Command {
	ConfCmd := &cobra.Command{
		Use:     "conf",
		Short:   "配置文件便捷获取子命令",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	ConfCmd.AddCommand(newConfGetCmd())
	ConfCmd.AddCommand(newConfExampleCmd())

	return ConfCmd
}

func newConfGetCmd() *cobra.Command {
	ConfGetCmd := &cobra.Command{
		Use:     "get",
		Short:   "获取当前配置子命令",
		Long:    "使用配置文件进行配置前建议通过该命令获取当前,然后再向内填写需要增加或修改的配置",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New()
			log.SetReportCaller(true)

			// 获取已有配置
			ctx := cmd.Context()
			oldInputCfg := &InputConfig{}
			scli, err := utils.NewSecretRW(pnamespace, "deploy-core", "deploy-core")
			if err != nil {
				log.Errorf("加载k8s 客户端失败, error: %s", err.Error())
				return err
			}
			if err := scli.GetFullConf(ctx, &oldInputCfg); trait.IsInternalError(err, trait.ErrComponentNotFound) {
				log.Errorf("配置文件不存在, error: %s", err.Error())
				return err
			} else if err != nil {
				log.Errorf("加载已有配置失败")
				return err
			}
			bs, rerr := yaml.Marshal(oldInputCfg)
			if rerr != nil {
				log.Errorf("convert config into string error: %s", rerr.Error())
				return rerr
			}
			fmt.Printf("\n%s\n", string(bs))
			return nil
		},
	}

	flags := ConfGetCmd.Flags()
	flags.StringVarP(&pnamespace, "pnamespace", "p", pnamespace, "proton cli conf命名空间")
	return ConfGetCmd
}

package main

/*
配置管理
*/

import (
	"context"
	"fmt"

	cutils "taskrunner/cmd/utils"
	"taskrunner/pkg/store/proton/system"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/spf13/cobra"
)

var (
	defaultNamespace = "anyshare"
	serviceAccount   = ""

	sucessDoc = `安装/升级 最小化安装服务成功。
	如果已进行过deploystudio包的安装,此时可以通过正常的部署工作台页面进行后续的安装/升级。
	如果未安装deploystudio包,此时由于未安装入口网关,无法通过正常渠道访问可视化安装页面,可以通过以下方式进行deploystudio包的安装。
	1. 可以通过本工具的子命令代理从本机最小化安装页面后从页面继续安装deploymentstudio包。例如命令: "core-installer proxy -n %s"。子命令可通过命令"core-installer proxy --help"查看子命令说明文档
	2. 或进入deploy-installer对应的POD内,通过容器内命令行工具对deploystudio包进行安装(建议使用场景专属配置)
	3. 多实例模式可直接进入多实例页面进行后续应用安装
	
	注意: 
	1. 通过最小化安装的deploy-installer pod内部命令行工具intaller job 相关命令对deploystudio包进行安装的任务,不会被计入deploy-installer灾难恢复过程计划任务,需要重新执行的任务,由操作者通过命令行进行操作。
	2. 不通过当前工具进行的配置,仅对helm层级配置进行复用,会被当前工具输入配置覆盖。
	3. 使用core-installer proxy命令时默认监听本地8888端口以复用proton-cli serve的端口。如果需要其他端口可通过参数'-p 1234'进行设置。
	4. 使用core-installer proxy监听的端口,工具程序不会自动打开防火墙,需要人工开启,所以默认使用8888端口, 复用proton-cli serve已开放的防火墙的端口。
	5. core-installer proxy 默认代理的服务命名空间为anyshare, 当前安装命名空间为%s,如果安装在不同命名空间下需要额外参数指定。
	6. core-installer proxy 在某些托管云k8s环境中可能由于网络问题无法直连集群内服务,此时可以通过参数"-m apiserver"设置为以k8s api-server中转的模式,api-server模式并不稳定,一般并不使用`
)

// CmdConfig 命令行级配置,可用于一般性应用配置以及行为控制参数的配置
// 目前为兼容考虑，仍保留deploy-installer组件配置，后续其他组件配置不应该增加至此
// 此处未来不能再添加组件配置，只能增加应用配置
type CmdConfig struct {
	cutils.Config
	Replicacount int    `json:"replicaCounts"`
	Timeout      int    `json:"timeout"`
	ForceCfg     string `json:"forceConfig"`
	ChartPath    string `json:"chart"`
	Conf         string `json:"inputConfig"`
	Database     string
	PNamespace   string
	cutils.CommonConfig
	cutils.HelmConfig
}

func newDefaultCOnfig() *CmdConfig {
	return &CmdConfig{
		// Replicacount: 1,
		// Interactive:  false,
		CommonConfig: cutils.NewDefaultCommonConfig(),
		HelmConfig:   cutils.NewDefaultHelmConfig(),
	}
}

func (c *CmdConfig) addFlags(cmd *cobra.Command) {
	// set default namesapce
	c.HelmSeting.SetNamespace(defaultNamespace)
	c.CommonConfig.AddFlags(cmd)
	c.HelmConfig.HelmSeting.AddFlags(cmd.Flags())
	cmd.Flags().StringVar(&c.PNamespace, "pnamespace", "proton", "proton-cli configuration namespace")
	cmd.Flags().StringVar(&c.ForceCfg, "force", "", "helm force resource updates through a replacement strategy. default nil, set false into disable")
	cmd.Flags().StringVar(&c.Database, "database", "deploy", "deploy-intaller will use this database, now must set deploy")
	cmd.Flags().IntVarP(&c.Replicacount, "replica", "r", 0, "应用级副本数设置")
	cmd.Flags().IntVarP(&c.Timeout, "timeout", "t", 600, "helm任务执行超时时间,默认600秒")
	cmd.Flags().IntVarP(&c.Parallel, "parallel", "p", 0, "deploy-installer服务任务池大小,默认10")
	// cmd.Flags().BoolVarP(&c.Interactive, "interactive", "i", false, "当值为true时,通过命令行交互式收集安装配置;否则可以通过其他命令行参数传递配置;默认为false")
	cmd.Flags().StringVarP(&c.ChartPath, "chart", "c", "./helm_charts", "应用包chart路径")
	cmd.Flags().StringVarP(&c.System.SName, "sname", "s", "", "默认创建系统名,默认为aishu,升级时该配置无效,如特殊情况需要可通过helm upgrade命令或其他方式调整")
	cmd.Flags().StringVarP(&c.Conf, "file", "f", "", "服务配置文件,用于对未在命令行暴露的配置,优先级低于命令行参数")

	// cmd.Flags().StringVar(&serviceAccount, "serviceaccont", "", "k8s服务账户,用于部署应用服务,需要满足相关权限。设置后使用传入的外置serviceaccount,不再自建账户与角色等")
	// cmd.MarkFlagRequired("chart")
}

func (c *CmdConfig) init() *trait.Error {
	c.System.NameSpace = c.HelmSeting.Namespace()
	return c.GetProtonCliConf()
}

func (c *CmdConfig) GetProtonCliConf() *trait.Error {
	pcli, err0 := cutils.GetDefaultProtonCli(c.Pcfg)
	if err0 != nil {
		return err0
	}
	pcfg, err0 := pcli.GetConf(context.TODO())
	if err0 != nil {
		return err0
	}
	c.ImageRepo = &pcfg.ToCRComponent().ImageRepo
	c.Rds = *pcfg.Resources.Rds
	return nil
}

func (c *CmdConfig) ToInputConfig(upgrade bool) (InputConfig, *trait.Error) {
	rds, _, err := c.Rds.ToMap()
	core := system.CoreConfig{
		Database:  c.Database,
		Namespace: c.System.NameSpace,
	}
	cfg := InputConfig{
		AppConfig: map[string]any{
			"image":     c.ImageRepo.ToMap(),
			"namespace": c.System.NameSpace,
		},
		Components: map[string]any{
			"deploy-installer": c.ToDeployInstallerConfig(upgrade),
			"rds":              rds,
			"deploy-core":      core.ToMapValues(),
		},
	}
	if c.Replicacount > 0 {
		cfg.AppConfig["replicaCount"] = c.Replicacount
	}
	return cfg, err
}

func (c *CmdConfig) ToDeployInstallerConfig(upgrade bool) map[string]any {
	dv := map[string]any{
		"log_level": c.LogLevel,
	}
	if c.Parallel > 0 {
		dv["parallel"] = c.Parallel
	}

	if upgrade {
		switch c.ForceCfg {
		case "false", "False":
			dv["force"] = "false"
		case "true", "True":
			dv["force"] = "true"
		}

		if serviceAccount == "" {
			dv["createnamespace"] = "true"
		} else {
			dv["createnamespace"] = "false"
		}
	}
	dv["system_name"] = c.System.SName
	dv["protonConf"] = map[string]interface{}{
		"namespace": c.PNamespace,
	}

	return dv
}

// InputConfig 使用者自定义输入配置文件,用于细粒度服务配置。
// 分为应用级配置与组件级配置
// 该配置项支持一下数据源，优先级依次降低:
// 1. 从命令行传入，即CmdConfig初始化
// 2. 命令行指定配置文件
// 3. 已有实例配置(主要用于升级时减少配置重新填写)
type InputConfig struct {
	AppConfig  map[string]any `json:"appConfig,omitempty"`
	Components map[string]any `json:"components,omitempty"`
}

func (c *InputConfig) ToValues(cname string) (map[string]any, *trait.Error) {
	v := map[string]any{
		"depServices": c.Components,
	}
	v = utils.MergeMaps(v, c.AppConfig)
	if c.Components != nil {
		if cv, ok := c.Components[cname]; ok {
			cvmap, ok := cv.(map[string]any)
			if !ok {
				return nil, &trait.Error{
					Internal: trait.ErrParam,
					Detail:   cname,
					Err:      fmt.Errorf("component %s config isn't a map", cname),
				}
			}
			v = utils.MergeMaps(v, cvmap)
		}
	}
	return v, nil
}

func MergeInputConfig(cfgs ...*InputConfig) *InputConfig {
	res := &InputConfig{}
	for _, i := range cfgs {
		res.AppConfig = utils.MergeMaps(res.AppConfig, i.AppConfig)
		res.Components = utils.MergeMaps(res.Components, i.Components)
	}
	return res
}

// ServiceConfig 单个服务安装时的配置项,参照当前应用包格式进行设计。
// 应用级配置与组件级配置合并提供。
// InputConfig转换后的单个组件可看到的配置项
type ServiceConfig struct {
	// 服务依赖配置
	Config     map[string]any
	Depservice map[string]any
}

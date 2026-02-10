package main

/*
	部署工作台最小化核心包部署工具，该工具将会以最简化的方式安装deployCore
*/

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cutils "taskrunner/cmd/utils"
	"taskrunner/cmd/version"
	"taskrunner/pkg/helm"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	cfg := newDefaultCOnfig()
	cmd := &cobra.Command{
		Use:     "core",
		Short:   `安装最小化安装服务,提供最基础安装能力`,
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return miniInstallRun(cmd.Context(), cfg)
		},
	}

	cfg.addFlags(cmd)
	cmd.AddCommand(proxyCmd)
	cmd.AddCommand(newConfCmd())

	if err := cmd.ExecuteContext(cmd.Context()); err != nil {
		os.Exit(1)
	}
}

func getChart(log *logrus.Logger, fpath string) (*helm.Chart, error) {
	fin, err0 := os.Open(fpath)
	if err0 != nil {
		log.Errorf("%s chart打开错误: %s", fpath, err0)
		return nil, err0
	}
	defer fin.Close()

	chart, err := helm.ParseChartFromTGZ(fin, "v2")
	if err != nil {
		log.Errorf("解析helm chart包错误: %s", err.Error())
		return nil, err
	}
	return chart, nil
}

func miniInstallRun(ctx context.Context, cfg *CmdConfig) error {
	cfg.System.NameSpace = cfg.HelmSeting.Namespace()
	log := cfg.CommonConfig.NewLogger()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	log.WithField("cmdconfig", *cfg).Debug("debug cmd configuration")

	cfg.Pcfg.Namespace = cfg.PNamespace
	log.WithField("pcfg", cfg.Pcfg).WithField("pnamespace", cfg.PNamespace).Debug("debug proton configuration")
	pcli, err := cutils.GetDefaultProtonCli(cfg.Pcfg)
	if err != nil {
		log.Errorf("创建proton client错误: %s", err.Error())
		return err
	}
	pcfg, err := pcli.GetConf(ctx)
	if err != nil {
		log.Errorf("获取proton cli conf错误: %s", err)
		return err
	}
	cfg.ImageRepo = &pcfg.ToCRComponent().ImageRepo
	log.Trace(pcfg.DeployConf.Serviceaccount)
	serviceAccount = pcfg.DeployConf.Serviceaccount

	if err := cfg.init(); err != nil {
		return err
	}

	hcli := helm.NewHelm3Client(log, cfg.HelmSeting)
	if serviceAccount != "" {
		hcli.WithCreateNamespace(false)
	}
	hcli.WithWait(false)

	// 获取已有配置
	oldInputCfg := &InputConfig{}
	isUpgrade := false
	scli, err := utils.NewSecretRW(cfg.PNamespace, "deploy-core", "deploy-core")
	if err != nil {
		log.Errorf("加载k8s 客户端失败, error: %s", err.Error())
		return err
	}
	if err := scli.GetFullConf(ctx, &oldInputCfg); trait.IsInternalError(err, trait.ErrComponentNotFound) {
		isUpgrade = false
	} else if err != nil {
		log.Errorf("加载已有配置失败")
		return err
	}

	// 计算命令参数获得 命令传入配置
	cmdInput, err := cfg.ToInputConfig(isUpgrade)
	if err != nil {
		log.Errorf("加载配置错误, error: %s", err.Error())
		return err
	}
	// 解析命令指定的配置文件
	inputCfg := &InputConfig{}
	if cfg.Conf != "" {
		if rerr := utils.ReadYamlFromFile(cfg.Conf, inputCfg); rerr != nil {
			log.Errorf("读取输入配置错误,请确认是否需要配置,文件路径是否正确以及格式是否正确: %s", rerr.Error())
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      rerr,
			}
		}
	}
	// 优先级自低向高为, 已有配置,配置文件配置和命令行配置
	finalInputCfg := MergeInputConfig(oldInputCfg, inputCfg, &cmdInput)
	if err := scli.SetContent(ctx, finalInputCfg); err != nil {
		log.Errorf("存储输入配置错误, error: %s", err)
		return err
	}

	log.Debugf("scan components dir: %s ", cfg.ChartPath)

	allComponents := make(map[string]*helm.Chart, 0)
	if rerr := utils.TraverseDirectory(cfg.ChartPath, func(path string, info os.FileInfo, _ error) error {
		log.Debugf("walk file %s", path)
		c, rerr := getChart(log, path)
		if rerr != nil {
			log.Errorf("自%s文件中加载chart失败, error: %s", path, rerr.Error())
			return rerr
		}

		allComponents[c.Name()] = c
		return nil
	}); rerr != nil {
		return &trait.Error{
			Internal: trait.ErrParam,
			Err:      rerr,
			Detail:   "遍历组件失败",
		}
	}

	needComponents := make([]*helm.Chart, 0)

	needComponents = append(needComponents, allComponents["deploy-installer"])

	hcli.WithWait(true)
	for _, c := range needComponents {
		v, err := finalInputCfg.ToValues(c.Name())
		if err != nil {
			log.Errorf("组件配置初始化错误, error: %s", err.Error())
			return err
		}
		oldValues, err := hcli.Values(ctx, c.Name(), cfg.System.NameSpace)
		if !trait.IsInternalError(err, trait.ErrNotFound) && err != nil {
			log.Errorf("获取 %s 历史配置错误: %s", c.Name(), err.Error())
			return err
		}

		values := utils.MergeMaps(oldValues, v)
		if err = hcli.Install(ctx, c.Name(), cfg.HelmSeting.Namespace(), c, values, cfg.Timeout, log.Debugf); err != nil {
			log.Errorf("安装 %s chart 错误: %s", c.Name(), err.Error())
			return err
		}
	}

	log.Infof(sucessDoc, cfg.System.NameSpace, cfg.System.NameSpace)
	if cfg.System.NameSpace != defaultNamespace {
		log.Warnf("服务安装命名空间非默认命名空间,后续操作注意指定相关命名空间参数为: '%s'", cfg.System.NameSpace)
	}
	log.Info("\033[安装成功\033[m")
	return nil
}

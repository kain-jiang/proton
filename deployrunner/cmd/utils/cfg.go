package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"taskrunner/pkg/app"
	"taskrunner/pkg/app/executor"
	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/helm"
	"taskrunner/pkg/store/proton/system"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	helmrepo "taskrunner/pkg/helm/repos"
	store "taskrunner/pkg/store/mysql"
	pstore "taskrunner/pkg/store/proton"

	"github.com/ghodss/yaml"
	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli"
)

// InstallerConfig start config
type InstallerConfig struct {
	CommonConfig
	HelmConfig
	ConfigPath    string
	JobConfigPath string
}

// CommonConfig basic config for helm
type CommonConfig struct {
	LogLevel string
}

// NewDefaultCommonConfig create a default common config
func NewDefaultCommonConfig() CommonConfig {
	return CommonConfig{
		LogLevel: "info",
	}
}

// NewLogger create a new log by the config
func (c *CommonConfig) NewLogger() *logrus.Logger {
	return NewLogger(c.LogLevel)
}

// AddFlags add helm flags
func (c *CommonConfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.LogLevel, "log_level", "v", "debug", "log filter level")
}

// HelmConfig helm config for helm install/upgrade
type HelmConfig struct {
	HelmSeting *helm.EnvSettings
}

// AddFlags add helm flags
func (c *HelmConfig) AddFlags(cmd *cobra.Command) {
	c.HelmSeting.AddFlags(cmd.Flags())
	cmd.Flags().BoolVar(&c.HelmSeting.Force, "force", true, "helm force resource updates through a replacement strategy")
	cmd.Flags().BoolVar(&c.HelmSeting.CreateNamespace, "create", true, "helm create namespace")
}

// NewDefaultHelmConfig create a default helm config
func NewDefaultHelmConfig() HelmConfig {
	return HelmConfig{
		HelmSeting: &helm.EnvSettings{
			EnvSettings:     cli.New(),
			Force:           true,
			CreateNamespace: true,
		},
	}
}

// NewDefaultInstallerConfig return a default config
func NewDefaultInstallerConfig() *InstallerConfig {
	return &InstallerConfig{
		CommonConfig: NewDefaultCommonConfig(),
		HelmConfig:   NewDefaultHelmConfig(),
	}
}

// AddStoreFlags add store config flag into command
func (c *InstallerConfig) AddStoreFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.ConfigPath, "config", "a", "config.devel", "config")
	c.CommonConfig.AddFlags(cmd)
	// cmd.MarkFlagRequired("config")
}

// AddJobConfigFlags add job config flag into command
func (c *InstallerConfig) AddJobConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.JobConfigPath, "job_config", "j", "", "the application  config in the job")
	if err := cmd.MarkFlagRequired("job_config"); err != nil {
		panic(err)
	}
}

// AddEngineFlags add executor engine flags
func (c *InstallerConfig) AddEngineFlags(cmd *cobra.Command) {
	c.AddStoreFlags(cmd)
	c.HelmConfig.AddFlags(cmd)
}

// NewLogger return a logger init by flagsets
func NewLogger(loglevel string) *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)
	if level >= logrus.DebugLevel {
		log.SetReportCaller(true)
	}
	return log
}

// Config  engine config
type Config struct {
	HelmRepo  []helmrepo.RepoConf `json:"HelmRepo,omitempty"`
	Rds       resources.RDS       `json:"rds,omitempty"`
	Parallel  int                 `json:"parallel"`
	ImageRepo *cluster.ImageRepo  `json:"ImageRepo,omitempty"`
	ID        int                 `json:"id,omitempty"`
	System    trait.System        `json:"system,omitempty"`
	Pcfg      ProtonCli           `json:"protonConf"`
}

// GetDefaultProtonCli get a default proton cli with default config
func GetDefaultProtonCli(cfg ProtonCli) (*pstore.ProtonClient, *trait.Error) {
	if cfg.Namespace == "" {
		cfg.Namespace = "proton"
	}
	if cfg.ConfName == "" {
		cfg.ConfName = "proton-cli-config"
	}
	if cfg.ConfKey == "" {
		cfg.ConfKey = "ClusterConfiguration"
	}
	return GetProtonCli(&cfg)
}

// LoadFromYamlFile load config from file
func (c *InstallerConfig) LoadFromYamlFile(ctx context.Context) (*Config, *trait.Error) {
	fpath := c.ConfigPath
	fin, err := os.Open(fpath)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "open config file",
		}
	}
	defer fin.Close()
	bs, err := io.ReadAll(fin)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "read bytes from  config file",
		}
	}
	cfg := &Config{}
	err = yaml.Unmarshal(bs, cfg)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "decode config from yaml bytes",
		}
	}

	if cfg.ID == 0 {
		nodeName := os.Getenv("DEPLOY_POD_NAME")
		parts := strings.Split(nodeName, "-")
		l := len(parts) - 1
		if l <= 1 {
			cfg.ID = -1
		} else {
			id, err := strconv.Atoi(parts[l-1])
			if err != nil {
				id = -1
			}
			cfg.ID = id
		}

	}

	{
		// 初始化默认system信息
		if cfg.System.SName == "" {
			cfg.System.SName = "aishu"
		}
		if cfg.System.NameSpace == "" {
			cfg.System.NameSpace = "anyshare"
		}
	}

	{
		// 初始化默认proton cli配置
		if cfg.Pcfg.Namespace == "" {
			cfg.Pcfg.Namespace = "proton"
		}
		if cfg.Pcfg.ConfName == "" {
			cfg.Pcfg.ConfName = "proton-cli-config"
		}
		if cfg.Pcfg.ConfKey == "" {
			cfg.Pcfg.ConfKey = "ClusterConfiguration"
		}
	}

	return cfg, c.MergeProtonConf(ctx, cfg)
}

// / 合并 proton-cli 配置文件
func (c *InstallerConfig) MergeProtonConf(ctx context.Context, cfg *Config) *trait.Error {
	log := c.NewLogger()

	pcli, err := GetDefaultProtonCli(cfg.Pcfg)
	if err != nil {
		log.Errorf("init proton client impl error: %s", err.Error())
		return err
	}
	pcfg, err := pcli.GetConf(ctx)
	if err != nil {
		log.Errorf("load proton cli conf error: %s", err)
		return err
	}

	log.Tracef("database config %#v\n", cfg.Rds)
	// 配置数据库，用户名，密码，来自于应用配置
	if cfg.Rds.Host == "" { // always true
		sqlCfg := cfg.Rds
		rds, err := pcfg.GetRDSComponent()
		if err != nil {
			log.Errorf("get rds from proton error: %s", err.Error())
			return err
		}
		if sqlCfg.Password != "" {
			rds.Password = sqlCfg.Password
		}
		if sqlCfg.User != "" {
			rds.User = sqlCfg.User
		}
		if sqlCfg.DBName != "" {
			rds.DBName = sqlCfg.DBName
		}
		cfg.Rds = *rds
	}

	if cfg.Rds.DBName == "" {
		// 必须设置数据库名称
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      errors.New("database name is empty"),
			Detail:   "database name is empty",
		}
		// cfg.Rds.DBName = "deploy"
	}
	log.Tracef("database merge config %#v\n", cfg.Rds)

	cr := pcfg.ToCRComponent()
	if cfg.ImageRepo == nil {
		cfg.ImageRepo = &cr.ImageRepo
	}
	if cfg.HelmRepo == nil {
		cfg.HelmRepo = cr.HelmRepo
	}

	return nil
}

// NewRunnerEngineAndConfig create a runner engine and return the config for reading
func (c *InstallerConfig) NewRunnerEngineAndConfig(ctx context.Context, log *logrus.Logger) (e *executor.Executor, cfg *Config, err *trait.Error) {
	cfg, err = c.LoadFromYamlFile(ctx)
	if err != nil {
		log.Errorf("load config file error: %s", err.Error())
		return
	}
	s, err := c.NewStore(ctx, log, cfg)
	if err != nil {
		log.Errorf("create store error: %s", err.Error())
		return nil, nil, err
	}
	hcli := helm.NewHelm3Client(s.Log, c.HelmSeting)
	// kcli初始化放在此处并不一定合适，应根据config进行配置加载，目前未有需求，暂置此处。
	kcli, err := utils.NewKubeclient()
	if err != nil {
		return nil, nil, err
	}
	// proton-cli初始化
	pcli, err := GetDefaultProtonCli(cfg.Pcfg)
	if err != nil {
		return nil, nil, err
	}
	e = executor.NewExecutor(s, cfg.Parallel, hcli, *cfg.ImageRepo, cfg.ID, kcli, pcli)

	return e, cfg, err
}

// NewRunnerEngine create a runner engine
func (c *InstallerConfig) NewRunnerEngine(ctx context.Context, log *logrus.Logger) (e *executor.Executor, err *trait.Error) {
	e, _, err = c.NewRunnerEngineAndConfig(ctx, log)
	return
}

// NewStore return a store operator
func (c *InstallerConfig) NewStore(ctx context.Context, log *logrus.Logger, cfg *Config) (*app.Store, *trait.Error) {

	rds := cfg.Rds
	if err := InitDatabase(ctx, rds); err != nil {
		log.Errorf("create database or user error: %s", err.Error())
		return nil, err
	}

	db, err := store.NewStore(ctx, rds)
	if err != nil {
		log.Errorf("init database error: %s", err.Error())
		return nil, err
	}
	if cfg.HelmRepo == nil {
		log.Error("helm repo must config, it is nil now.")
	}
	coreCfg := &system.CoreConfig{
		Namespace: cfg.System.NameSpace,
		Database:  cfg.Rds.DBName,
	}
	var storeWrapper trait.Store
	{
		ps, err := pstore.NewStore(
			cfg.Pcfg.Namespace, cfg.Pcfg.ConfName,
			cfg.Pcfg.ConfKey, *coreCfg, db)
		if err != nil {
			return nil, err
		}
		storeWrapper = ps
	}

	relHelmRepo, err := helmrepo.NewHaMutilUploadRepo(cfg.HelmRepo)
	if err != nil {
		return nil, err
	}
	s := app.NewStore(log, storeWrapper, relHelmRepo)
	return s, err
}

// InitDatabase init the database if container adminKey
// TODO abstract with multil db type
func InitDatabase(ctx context.Context, rds resources.RDS) *trait.Error {
	if rds.AdminKey == "" {
		return nil
	}
	adminKeydec, err0 := base64.StdEncoding.DecodeString(rds.AdminKey)
	if err0 != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err0,
			Detail:   "decode bas64 string when parse admin key",
		}
	}
	index := bytes.IndexRune(adminKeydec, ':')
	if index == -1 || index >= len(adminKeydec)-1 {
		return nil
	}
	user := adminKeydec[:index]
	passwd := adminKeydec[index+1:]
	adminRds := deepcopy.Copy(rds).(resources.RDS)
	adminRds.User = string(user)
	adminRds.Password = string(passwd)
	db, err := store.NewDBOP(ctx, adminRds)
	if err != nil {
		return err
	}
	defer db.Close()
	dataUser := rds.User

	if err := db.CreateUser(ctx, dataUser, rds.Password); err != nil {
		return err
	}
	if err := db.CreateDatabase(ctx, rds.DBName); err != nil {
		return err
	}
	if err := db.GrantUserDB(ctx, rds.User, rds.DBName); err != nil {
		fmt.Printf("create db eror: %s\n", err.Error())
		return err
	}
	return err
}

// CreateDefaultSystem if the input system is not config, create the system with default config
func CreateDefaultSystem(ctx context.Context, e trait.Store, s trait.System) (trait.System, *trait.Error) {
	res := deepcopy.Copy(s).(trait.System)
	if res.SName == "" {
		res.SName = "aishu"
	}
	if res.NameSpace == "" {
		res.NameSpace = "anyshare"
	}

	sid, err := e.InsertSystemInfo(ctx, res)
	if trait.IsInternalError(err, trait.ErrUniqueKey) {
		res0, err0 := e.GetSystemInfoByName(ctx, res.SName)
		res = *res0
		err = err0
	} else {
		res.SID = sid
	}
	return res, err
}

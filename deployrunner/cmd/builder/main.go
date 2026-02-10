package main

import (
	"os"

	"taskrunner/cmd/version"
	"taskrunner/pkg/app/builder"
	"taskrunner/pkg/helm"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	applicationPath = ""
	dst             = ""
	imgDst          = ""
	configFilePath  = ""
	logLevel        = "trace"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "srv",
		Long: `task runner manager use to manage application system.
			manager can create new job record but don't execute.
			manager can has multi replica.`,
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := NewLogger()
			log.SetReportCaller(true)

			fin, err0 := os.Open(applicationPath)
			if err0 != nil {
				log.Fatalf("application文件打开失败: %s", err0)
				return err0
			}
			defer fin.Close()
			cfg, err := builder.LoadConfiguration(fin)
			if err != nil {
				log.Fatalf("application 文件解析失败: %s", err)
				return err
			}

			appOut, err0 := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
			if err0 != nil {
				log.Fatalf("打开包存档文件失败: %s", err0)
				return err0
			}
			defer appOut.Close()

			imgOut, err0 := os.OpenFile(imgDst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
			if err0 != nil {
				log.Fatalf("打开镜像名称列表存档文件失败: %s", err0)
				return err
			}

			defer imgOut.Close()

			repos := make([]helm.Repo, 0, len(cfg.HelmRepos))
			for _, repo := range cfg.HelmRepos {
				repos = append(repos, repo)
			}

			// add local repo
			if cfg.LocalRepo != nil {
				repos = append(repos, cfg.LocalRepo)
			}

			b, err := builder.NewApplicationBuilder(&cfg, appOut, imgOut, repos...)
			if err != nil {
				log.Fatalf("创建appliation包失败: %s", err)
				return err
			}
			b.Log = log
			b.ConfigTemplatePath = configFilePath
			err = b.Build(cmd.Context())
			if err != nil {
				log.Fatalf("构建application包失败: %s", err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&applicationPath, "app", "a", "", "application file path")
	if err := cmd.MarkFlagRequired("app"); err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&dst, "dst", "d", "", "application package file path")
	if err := cmd.MarkFlagRequired("dst"); err != nil {
		panic(err)
	}
	cmd.Flags().StringVarP(&imgDst, "images", "i", "", "images file path")
	if err := cmd.MarkFlagRequired("images"); err != nil {
		panic(err)
	}
	cmd.Flags().StringVarP(&logLevel, "log_level", "v", "info", "log filter level")
	cmd.Flags().StringVarP(&configFilePath, "config_template", "c", "", "config template file path or dir path")
	return cmd
}

// NewLogger return a logger init by flagsets
func NewLogger() *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)
	return log
}

func main() {
	cmd := newBuildCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

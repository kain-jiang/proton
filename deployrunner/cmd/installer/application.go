package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"taskrunner/cmd/version"

	cutils "taskrunner/cmd/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewUploadCmd return a application package upload command
func NewUploadCmd() *cobra.Command {
	cfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "upload appfile",
		Short:   `upload the application file`,
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := cfg.NewLogger()
			ctx := cmd.Context()

			scfg, err := cfg.LoadFromYamlFile(ctx)
			if err != nil {
				logrus.Fatalf("load config file error: %s", err.Error())
				return err
			}

			s, err := cfg.NewStore(cmd.Context(), log, scfg)
			if err != nil {
				return err
			}

			ain, err0 := os.Open(args[0])
			if err0 != nil {
				logrus.Fatalf("open application package error: %s", err0.Error())
				return err
			}
			defer ain.Close()

			aid, err := s.UploadApplicationPackage(cmd.Context(), ain)
			if err != nil {
				logrus.Fatalf("upload application error: %s", err.Error())
				return err
			}
			logrus.Infof("upload application sucess, application id: %d", aid)
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
	cfg.AddStoreFlags(cmd)

	return cmd
}

// NewAppGetCmd return get application package command
func NewAppGetCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "get aid",
		Long:    `get the application detail info`,
		Short:   " get application package detai info",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := icfg.NewLogger()
			ctx := cmd.Context()
			cfg, err := icfg.LoadFromYamlFile(ctx)
			if err != nil {
				logrus.Fatalf("load config file error: %s", err.Error())
				return err
			}

			aid, err0 := strconv.Atoi(args[0])
			if err0 != nil {
				log.Fatalf("aid muse is a interger, parse error: %s", err0.Error())
				return err
			}

			s, err := icfg.NewStore(cmd.Context(), log, cfg)
			if err != nil {
				return err
			}

			a, err := s.GetAPP(ctx, aid)
			if err != nil {
				logrus.Fatalf("get application error: %s", err.Error())
				return err
			}

			bs, err0 := json.MarshalIndent(a, "", "  ")
			if err0 != nil {
				logrus.Fatalf("please check command version, encode application error: %s", err0.Error())
				return err
			}
			fmt.Println(string(bs))
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
	icfg.AddStoreFlags(cmd)

	return cmd
}

// NewAppSearchCmd return search application package command
func NewAppSearchCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "search aname",
		Long:    `search the application with the application name. dispay the application overview`,
		Short:   "search application package overview",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := icfg.NewLogger()
			log.SetReportCaller(true)
			ctx := cmd.Context()

			cfg, err := icfg.LoadFromYamlFile(ctx)
			if err != nil {
				logrus.Fatalf("load config file error: %s", err.Error())
				return err
			}

			aname := args[0]

			s, err := icfg.NewStore(cmd.Context(), log, cfg)
			if err != nil {
				return err
			}

			as, err := s.SearchAPP(ctx, limit, lastid, aname)
			if err != nil {
				logrus.Fatalf("get application error: %s", err.Error())
				return err
			}

			bs, err0 := json.MarshalIndent(as, "", "  ")
			if err0 != nil {
				logrus.Fatalf("please check command version, encode application error: %s", err0.Error())
				return err0
			}
			fmt.Println(string(bs))
			return nil
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().IntVar(&lastid, "lastid", -1, "last app id")
	cmd.Flags().IntVarP(&lastid, "limit", "l", 3, "result num")
	icfg.AddStoreFlags(cmd)

	return cmd
}

func newApplicationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app",
		Long:    `control application.`,
		Short:   "application package command",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewUploadCmd(),
		NewAppGetCmd(),
		NewAppSearchCmd(),
	)

	return cmd
}

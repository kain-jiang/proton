package main

import (
	"encoding/json"
	"fmt"

	"taskrunner/cmd/version"
	"taskrunner/trait"

	cutils "taskrunner/cmd/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	namespace  = ""
	systeMName = ""
	limit      = 3
	lastid     = 0
	systemID   = -1
)

// NewCreateSystemCmd return a application package upload command
func NewCreateSystemCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "create a system for application instance.",
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

			s, err := icfg.NewStore(cmd.Context(), log, cfg)
			if err != nil {
				return err
			}

			si := trait.System{
				NameSpace: namespace,
				SName:     systeMName,
			}
			si.SID, err = s.InsertSystemInfo(ctx, si)
			if err != nil {
				logrus.Fatalf("create system error: %s", err.Error())
				return err
			}

			bs, _ := json.MarshalIndent(si, "", "  ")
			fmt.Println(string(bs))

			return nil
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "anyshare", "system namespace")
	cmd.Flags().StringVarP(&systeMName, "sname", "s", "", "system name")
	if err := cmd.MarkFlagRequired("sname"); err != nil {
		panic(err)
	}
	icfg.AddStoreFlags(cmd)

	return cmd
}

// NewListSystemCmd return list system  command
func NewListSystemCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "list",
		Long:    "list system, if you want to walk the next page, record the last one sid, then use lastid flag",
		Short:   "list system",
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

			s, err := icfg.NewStore(cmd.Context(), log, cfg)
			if err != nil {
				return err
			}
			var ss []*trait.System
			ss, err = s.ListSystemInfo(ctx, limit, lastid)
			if err != nil {
				logrus.Fatalf("list system error: %s", err.Error())
				return err
			}

			bs, _ := json.MarshalIndent(ss, "", "  ")
			fmt.Println(string(bs))

			return nil
		},
	}

	cmd.Flags().IntVar(&lastid, "lastid", 0, "分页偏移量,为兼容某些场景用法参数含义变更，但不更改参数名")
	cmd.Flags().IntVarP(&limit, "limit", "l", 3, "result set num")
	cmd.Flags().StringVarP(&systeMName, "sname", "s", "", "system name")
	icfg.AddStoreFlags(cmd)
	return cmd
}

// NewSystemCmd return sysmtem cmd
func NewSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "system",
		Long: `system is a logic group for application instance. 
		a system only has one applicaton instance with same type.
		a system can has mutil application instance with defferrent type.`,
		Short:   "create a system space for application instance",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewCreateSystemCmd(),
		NewListSystemCmd(),
	)

	return cmd
}

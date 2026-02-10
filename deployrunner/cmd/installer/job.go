package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"taskrunner/cmd/version"
	"taskrunner/trait"

	cutils "taskrunner/cmd/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func dumpJSONApplicationIns(appIns *trait.ApplicationInstance, w io.Writer) error {
	bs, err := json.MarshalIndent(appIns, "", "  ")
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	return err
}

// NewJobConfigGetCmd return get job config command
func NewJobConfigGetCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "cget jobid",
		Short:   `get application config in the job then write to job config file`,
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

			jid, err0 := strconv.Atoi(args[0])
			if err0 != nil {
				log.Fatalf("parse jobID error: %s", err0.Error())
				return err0
			}

			job, err := s.GetJobRecord(ctx, jid)
			if err != nil {
				log.Fatalf("get application instance config error: %s", err.Error())
				return err
			}

			fout, err0 := os.OpenFile(icfg.JobConfigPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
			if err0 != nil {
				log.Fatalf("you can retry this command. open output job config file error: %s", err0.Error())
				return err0

			}
			ains := job.Target
			log.Tracef("get application instance config: %#v", ains)
			if err0 = dumpJSONApplicationIns(ains, fout); err0 != nil {
				log.Fatalf("write config error: %s", err0.Error())
			}

			return nil
		},
		Args: cobra.ExactArgs(1),
	}

	icfg.AddJobConfigFlags(cmd)
	icfg.AddStoreFlags(cmd)

	return cmd
}

// NewJobListCmd return list jobrecord  command
func NewJobListCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "list",
		Long:    `list jobs overview`,
		Short:   "list job",
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
			var ss []trait.JobRecord
			ss, err = s.ListJobRecord(ctx, &trait.AppInsFilter{
				Offset: lastid,
				Limit:  limit,
				Sid:    systemID,
			})
			if err != nil {
				logrus.Fatalf("list job error: %s", err.Error())
				return err
			}

			bs, _ := json.MarshalIndent(ss, "", "  ")
			fmt.Println(string(bs))

			return nil
		},
	}

	cmd.Flags().IntVar(&lastid, "offset", 0, "resultset offset=pageNum * pageLimit + pageRows")
	cmd.Flags().IntVar(&systemID, "sid", -1, "system id. if sid <0, don't filter with system info")

	cmd.Flags().IntVarP(&limit, "limit", "l", 3, "result set max number")
	icfg.AddStoreFlags(cmd)

	return cmd
}

// NewJobCreateCmd return a install command
func NewJobCreateCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use: "create applicationID systemID",
		Long: `create a job for install/upgrade the application in the system. 
		command will output application config to job_config file for next step.
		user must set job config before execute the job.
		if systemID not set, it will use the default system`,
		Short:   "create a job in the system to install/upgrade application ",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := icfg.NewLogger()
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

			aid, err0 := strconv.Atoi(args[0])
			if err0 != nil {
				log.Fatalf("parse aid error %s", err0.Error())
			}
			var sid int
			if len(args) == 2 {
				sid, err0 = strconv.Atoi(args[1])
				if err0 != nil {
					log.Fatalf("parse aid error %s", err0.Error())
				}
			} else {
				ss, err := cutils.CreateDefaultSystem(ctx, s, cfg.System)
				if err != nil {
					log.Errorf("create default system error: %s", err.Error())
					return err
				}
				sid = ss.SID
			}

			jid, err := s.NewJobRecord(ctx, aid, sid)
			if err != nil {
				logrus.Fatalf("create new job record error: %s", err.Error())
				return nil
			}

			job, err := s.GetJobRecord(ctx, jid)
			if err != nil {
				logrus.Fatalf("please get config by cget command, jobId is %d. process get application instance config error: %s", jid, err.Error())
				return err
			}

			fout, err0 := os.OpenFile(icfg.JobConfigPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
			if err0 != nil {
				logrus.Fatalf("you can try get config by get command, jobId is %d.. open output job config file error: %s", jid, err0.Error())
				return err0

			}
			defer fout.Close()
			appIns := job.Target
			if err0 = dumpJSONApplicationIns(appIns, fout); err0 != nil {
				logrus.Fatalf("please check command version then get config by cget command, jobdID is %d. error: %s", jid, err0)
				return err0
			}

			fmt.Printf("{\"jobID\": %d}\n", jid)

			return nil
		},
		Args: cobra.RangeArgs(1, 2),
	}
	icfg.AddJobConfigFlags(cmd)
	icfg.AddStoreFlags(cmd)

	return cmd
}

// NewJobSetCmd return a install command
func NewJobSetCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use: "set jobid",
		Long: `set the application config in the job. 
		client must do this step before start job, this mean client has comfir the application config`,
		Short:   `set the application config in the job`,
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

			jid, err0 := strconv.Atoi(args[0])
			if err0 != nil {
				log.Fatalf("parse aid error %s", err0.Error())
			}

			userConfig := &trait.ApplicationInstance{}
			if icfg.JobConfigPath != "" {
				bs, err := os.ReadFile(icfg.JobConfigPath)
				if err != nil {
					logrus.Fatalf("load job config file error: %s", err.Error())
					return err
				}
				if err = json.Unmarshal(bs, userConfig); err != nil {
					logrus.Fatalf("decode job config error: %s", err.Error())
					return err
				}
				log.Tracef("load config %#v", userConfig)
			}

			s, err := icfg.NewStore(cmd.Context(), log, cfg)
			if err != nil {
				return err
			}

			err = s.SetJobConfig(ctx, jid, userConfig)
			if err != nil {
				logrus.Fatalf("set job record error: %s", err.Error())
				return err
			}

			fmt.Printf("set jobrecord %d success\n", jid)

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
	// job config isn't must when set
	icfg.AddStoreFlags(cmd)
	icfg.AddJobConfigFlags(cmd)

	return cmd
}

// NewJobExecutemd return a install command
func NewJobExecutemd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use:     "execute jobid",
		Short:   `this command will execute the job `,
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := icfg.NewLogger()
			ctx := cmd.Context()

			jid, err0 := strconv.Atoi(args[0])
			if err0 != nil {
				log.Fatalf("parse aid error %s", err0.Error())
			}

			e, err := icfg.NewRunnerEngine(ctx, log)
			if err != nil {
				return err
			}

			log.Tracef("start job %d", jid)
			if err := e.StartJob(ctx, jid); err != nil {
				logrus.Fatalf("start the job %d error: %s", jid, err.Error())
				return err
			}

			log.Tracef("execute job %d", jid)
			err = e.ExecuteJob(ctx)
			log.Tracef("execute job result: %#v", err)
			if trait.IsInternalError(err, trait.ECJobCancel) || trait.IsInternalError(err, trait.ECExit) {
				log.Fatalf("process interup job %d, recevive error: %s, cancel job", jid, err.Error())
				ctx0, cancel := trait.WithTimeoutCauseContext(context.Background(), 3*time.Second, nil)
				defer cancel()
				err = e.CancelJob(ctx0, jid)
				if err != nil {
					log.Fatalf("receive interupt signal, cancel job error: %s", err.Error())
				}

				logrus.Fatalf("receive interupt signal, cancel job finish")
				return err
			}

			if err != nil {
				logrus.Fatalf("execute job %d error: %s", jid, err.Error())
				return err
			}
			fmt.Printf("execute job %d success\n", jid)

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
	icfg.AddEngineFlags(cmd)

	return cmd
}

// var null string

func newjobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "job",
		Long: `install or upgrade the  application instances. 
		A job will install create a application instance instead the application instance has the same type in system.
		1. upload application that you need  by application upload cmd
		2. create a job by create cmd.
		3. set the application in the job by set cmd
		4. execute the job by execute cmd
		5. if execute job fail, you can retry step 4 or step 3`,
		Short:   "job operate a applicatiton instance need install/upgrade to a new instance",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewJobListCmd(),
		NewJobCreateCmd(),
		NewJobSetCmd(),
		NewJobExecutemd(),
		NewJobConfigGetCmd(),
	)

	return cmd
}

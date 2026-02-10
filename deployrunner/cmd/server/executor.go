package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"taskrunner/api/rest"
	"taskrunner/api/rest/proton_component"
	"taskrunner/api/rest/proton_component/compose"
	"taskrunner/pkg/utils"

	// "taskrunner/cmd/utils"
	cutils "taskrunner/cmd/utils"
	"taskrunner/cmd/version"
	"taskrunner/trait"

	"github.com/spf13/cobra"
)

func newExecutorCmd() *cobra.Command {
	icfg := cutils.NewDefaultInstallerConfig()
	cmd := &cobra.Command{
		Use: "executor",
		Long: `task runner executor use to execute job..
			Executor can start/stop the job.
			Executor is a statefulset, it must be one replica in current version`,
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := icfg.NewLogger()
			ctx := cmd.Context()

			ctx0, cancel := trait.WithCancelCauesContext(ctx)
			defer cancel(&trait.Error{
				Internal: trait.ECExit,
				Err:      context.Canceled,
				Detail:   "executor main routine exit",
			})
			go func() {
				ch := make(chan os.Signal, 1)
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
				<-ch
				close(ch)
				cancel(&trait.Error{
					Internal: trait.ECExit,
					Err:      context.Canceled,
					Detail:   "executor main routine exit",
				})
			}()

			s, cfg, err := icfg.NewRunnerEngineAndConfig(ctx, log)
			if err != nil {
				return err
			}
			defer s.Close()
			defer log.Info("process safe exit")

			// 获取proton-cli conf代码段移动到此以后将会导致与proton-cli配置强绑定，后续需要考虑进行合适的方式调整
			pcli, err := cutils.GetDefaultProtonCli(cfg.Pcfg)
			if err != nil {
				log.Errorf("init proton conf client error: %s", err.Error())
				return err
			}
			pcfg, err := pcli.GetConf(ctx)
			if err != nil {
				log.Errorf("读取proton-cli conf配置错误, error: %s", err.Error())
				return err
			}

			var rs *rest.ExecutorEngine
			var ps proton_component.GinServer
			{
				// TODO reactor
				ss, err := cutils.CreateDefaultSystem(ctx, s, cfg.System)
				if err != nil {
					log.Errorf("create default system error: %s", err.Error())
					return err
				}

				rs, err = rest.NewExecutorEngine(s, ss)
				if err != nil {
					log.Errorf("create rest api server error: %s", err.Error())
					return err
				}

				// pcli, err := cutils.GetDefaultProtonCli(cfg.Pcfg)
				// if err != nil {
				// 	log.Errorf("init proton conf client error: %s", err.Error())
				// 	return err
				// }
				ps0, err := proton_component.NewServer(pcli, ss, pcfg.DeployConf.Namespace)
				if err != nil {
					log.Errorf("init proton component conf proxy server error:%s", err.Error())
					return err
				}
				ps = ps0
			}

			{

				// compose job handler registry
				kcli, err := utils.NewKubeclient()
				if err != nil {
					log.Errorf("init k8s client error: %s", err.Error())
					return err
				}

				composeHandler := compose.NewServer(ctx0, s.Store.Store, ps, rs, kcli, rs.Log)
				for _, r := range rs.RouterGroups() {
					ps.RegistryHandler(r)
					composeHandler.RegistryHandler(r)
				}
				if err := composeHandler.Recovery(ctx0); err != nil {
					return err
				}
			}

			srv := &http.Server{
				Addr:    ":9090",
				Handler: rs,
			}

			wg, err := rs.Recover(ctx0)
			if err != nil {
				log.Errorf("start recover routine error: %s", err.Error())
			}
			wg.Add(2)
			go func() {
				defer wg.Done()
				s.Run(ctx0)
			}()

			go func() {
				defer wg.Done()
				defer cancel(&trait.Error{
					Internal: trait.ECExit,
					Err:      context.Canceled,
					Detail:   "executor main routine exit",
				})
				_ = startHTTPServer(ctx0, srv, log)
			}()
			wg.Wait()
			return nil
		},
	}
	icfg.AddEngineFlags(cmd)
	return cmd
}

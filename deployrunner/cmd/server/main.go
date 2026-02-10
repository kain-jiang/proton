package main

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

//	@title			安装/更新执行器HTTP API文档
//	@version		0.1.0
//	@description	安装/更新执行器HTTP API接口文档

//	@contact.name	API Support
//	@contact.email	tiga.gan@aishu.cn

//	@BasePath	/api/deploy-installer/v1

// startHTTPServer start
func startHTTPServer(ctx context.Context, srv *http.Server, log *logrus.Logger) (err error) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Info("http server running")
		if err = srv.ListenAndServe(); err != nil {
			log.Errorf("http server runnig error: [%s], exit...", err.Error())
		}
	}()

	go func() {
		<-ctx.Done()
		log.Info("receive stop signal, try stopping http server...")
		_ = srv.Shutdown(context.Background())
	}()
	wg.Wait()

	log.Info("server exit")
	return
}

func main() {
	cmd := newExecutorCmd()
	cmd.AddCommand(newUpgradeCmd())
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

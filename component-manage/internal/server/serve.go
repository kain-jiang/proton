package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/routers"

	"github.com/gin-gonic/gin"
)

type Serve struct {
	server *http.Server
}

// RunHttpServe 运行HTTP服务
func (s *Serve) RunHttpServe() {
	engine := gin.New()

	ginLogger := gin.Logger()
	if global.Config.Log.DisableHealth {
		ginLogger = gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/health/ready", "/health/alive"},
		})
	}

	// 使用默认中间件
	engine.Use(
		ginLogger,
		gin.Recovery(),
	)

	// 注册路由
	err := routers.RegistryRouter(engine)
	if err != nil {
		global.Logger.WithError(err).Fatalf("registry router failed")
	}

	s.server = &http.Server{
		Addr:    global.Config.ServerHost(),
		Handler: engine,
	}

	err = s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		global.Logger.WithError(err).Fatal("run server failed.")
	}
}

// ShutDownHttpServe 服务优雅退出
func (s *Serve) WaitShutDownHttpServe() error {
	global.Logger.Info("start wait shutdown server ...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	global.Logger.Info("shutdown server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		global.Logger.WithError(err).Error("server shutdown failed.")
		return err
	}

	global.Logger.Info("server exiting ...")
	return nil
}

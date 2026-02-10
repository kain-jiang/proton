package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration/validation"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/apply"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

const durationBeforeShutdownHTTPServer = 30 * time.Second

//go:embed web
var res embed.FS

var log = logger.NewLogger()

// serverCmd represents the server command
var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "proton-cli server --port=?",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.NewLogger().Debugf("%#v", version.Get())

			html, err := fs.Sub(res, "web")
			if err != nil {
				return err
			}
			http.Handle("/", http.FileServer(http.FS(html)))

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			ih := &initialHandler{cancel: cancel}
			http.Handle("/init", ih)

			u := &url.URL{
				Scheme: "http",
				// Golang 默认合并处理 IPv4 和 IPv6 请求
				// host 使用 0.0.0.0 或 [::] 都可以
				Host: net.JoinHostPort("::", port),
			}
			s := http.Server{Addr: u.Host}
			log.Infof("Please access web service %s to complete Proton Runtime service deploy.", u)

			http.Handle(httpPatternResult, &resultHandler{initialHandler: ih})

			go func() {
				if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error(err)
				}
			}()

			<-ctx.Done()
			log.Printf("shutdown http server")
			if err := s.Shutdown(ctx); errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}
	port string
)

type initialHandler struct {
	sync.RWMutex

	// 用于通知停止 http server
	cancel context.CancelFunc

	// initialHandler 是否在执行
	running bool

	// initialHandler 是否执行过
	completed bool

	// apply.Apply 返回的 error
	err error
}

func (h *initialHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.NewLogger()

	if !h.run() {
		log.Println("conflict")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "/init cannot be called concurrently")
		return
	}
	defer h.stop()

	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Content Type is not application/json")
		log.Error("Content Type is not application/json")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
	}

	// 记录接收到的请求
	log.Debugf("%v %v from %v", r.Method, r.URL.String(), r.RemoteAddr)
	log.Debugf("Request Headers:")
	for k, values := range r.Header {
		for _, v := range values {
			log.Debugf("    %v: %v", k, v)
		}
	}
	log.Debugf("Request Body: %s", string(body))
	_ = os.WriteFile("/tmp/proton-cli.json", body, 0666)

	conf, err := configuration.LoadFromBytes(body)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Errorf("write response body fail: %v", err)
		}
		return
	}

	if conf.Deploy != nil {
		if err := configuration.UpdateProtonCliEnvConfig(conf.Deploy.Namespace); err != nil {
			log.Errorf("unable to update proton-cli.yaml: %v", err)
			return
		}
	}

	var errInvalidClusterConfig = new(validation.InvalidError)
	err = apply.Apply(conf)
	if errors.As(err, &errInvalidClusterConfig) {
		w.WriteHeader(http.StatusBadRequest)
		for _, e := range errInvalidClusterConfig.ErrorList {
			fmt.Fprintf(w, "%v\n", e)
		}
		return
	}
	h.setResult(true, err)
	if err != nil {
		log.Errorf("initial cluster fail: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "initial cluster fail: %v", err)
		return
	}

	fmt.Fprintf(w, "initial cluster end")
	log.Info("initial cluster end")

	w.WriteHeader(http.StatusOK)

	go func() {
		log.Printf("Wait %v before stopping the http server for the front-end to query the initialized results", durationBeforeShutdownHTTPServer)
		time.Sleep(durationBeforeShutdownHTTPServer)
		log.Print("Shutdown http server")
		h.cancel()
	}()
}

// 设置状态为运行中。并返回 true。如果已经处于运行中或已经成功执行过则返回 false。
func (h *initialHandler) run() bool {
	h.Lock()
	defer h.Unlock()
	if h.running || (h.completed && h.err == nil) {
		return false
	}
	h.running = true
	return true
}

func (h *initialHandler) stop() {
	h.Lock()
	defer h.Unlock()
	h.running = false
}

func (h *initialHandler) setResult(isCompleted bool, err error) {
	h.Lock()
	defer h.Unlock()
	h.completed, h.err = isCompleted, err
}

// result 返回 initialHandler 是否处于运行中，是否已经完成，以及执行结果的异常
func (h *initialHandler) getResult() (isRunning, isCompleted bool, err error) {
	h.RLock()
	defer h.RUnlock()
	return h.running, h.completed, h.err
}

const httpPatternResult = "/alpha/result"

type resultHandler struct {
	initialHandler *initialHandler
}

func (h *resultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	running, completed, err := h.initialHandler.getResult()
	switch {
	case running:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "The initialization is running.")
	case !completed:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "No initialization has been executed.")
	case completed && err == nil:
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Success")
	case completed && err != nil:
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "initial cluster fail: %v", err)
	}
}

func init() {
	// 添加命令
	rootCmd.AddCommand(serverCmd)
	// 接收参数port
	serverCmd.Flags().StringVar(&port, "port", "8888", "端口号")
}

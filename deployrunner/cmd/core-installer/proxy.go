package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"taskrunner/cmd/version"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	cutils "taskrunner/cmd/utils"

	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
)

type backend struct {
	name      string
	namespace string
	host      string
	port      int
	scheme    string
	lg        *logrus.Logger
	prefix    string
	kcli      kubernetes.Interface
	kHTTPCli  *http.Client
}

func (b *backend) apiServerProxyPath(rawUrl string) *url.URL {
	return b.kcli.CoreV1().RESTClient().Get().
		Namespace(b.namespace).Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort(b.scheme, b.name, strconv.Itoa(b.port))).
		Suffix(b.prefix, rawUrl).URL()
}

func (b *backend) proxyByApiServer(w http.ResponseWriter, r *http.Request) {
	backendUrl := deepcopy.Copy(r.URL).(*url.URL)
	newUrl := b.apiServerProxyPath(backendUrl.Path)
	backendUrl.Scheme = newUrl.Scheme
	backendUrl.Host = newUrl.Host
	backendUrl.Path = newUrl.Path
	b.do(b.kHTTPCli, w, r, backendUrl.String())
}

func (b *backend) do(client *http.Client, w http.ResponseWriter, r *http.Request, newUrl string) {
	// 创建新的请求，将请求体作为 io.Reader 传递
	start := time.Now()
	var err1 error
	statusCode := -1
	defer func() {
		cost := time.Since(start).Seconds()
		if err1 != nil {
			b.lg.Errorf("Method: %s, RawUrl: %s, backendUrl: %s, statusCode: %d, cost: %fs, error:%s",
				r.Method, r.URL.String(), newUrl, statusCode, cost, err1.Error())
		} else {
			b.lg.Infof("Method: %s, RawUrl: %s, backendUrl: %s, statusCode: %d, cost: %fs",
				r.Method, r.URL.String(), newUrl, statusCode, cost)
		}
	}()
	req, err := http.NewRequest(r.Method, newUrl, r.Body)
	if err != nil {
		err1 = fmt.Errorf("create request url %s error: %s", newUrl, err.Error())
		b.lg.Error(err1.Error())
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		err1 = fmt.Errorf("error sending request url error: %s", err.Error())
		b.lg.Error(err1.Error())
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 复制响应信息
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	statusCode = resp.StatusCode
	w.WriteHeader(resp.StatusCode)

	// 将响应体返回给客户端
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		err1 = fmt.Errorf("read from backend to client error: %s", err.Error())
		b.lg.Error(err1.Error())
	}
}

func (b *backend) proxy(w http.ResponseWriter, r *http.Request) {
	// 创建一个新的 HTTP 客户端
	backendUrl := deepcopy.Copy(r.URL).(*url.URL)
	backendUrl.Scheme = b.scheme
	backendUrl.Host = b.host
	backendUrl.Path = b.prefix + backendUrl.Path
	newUrl := backendUrl.String()
	cli := &http.Client{}
	b.do(cli, w, r, newUrl)
}

func NewBackend(ctx context.Context, ns string, svcName string, scheme string, port int, prefix string, log *logrus.Logger) (*backend, error) {
	k, err := utils.NewKubeclient()
	if err != nil {
		return nil, err
	}
	khcli, err := utils.NewKubeHTTPClient()
	if err != nil {
		return nil, err
	}
	svc, err0 := k.CoreV1().Services(ns).Get(ctx, svcName, v1.GetOptions{})
	client := &backend{
		prefix:    prefix,
		namespace: ns,
		name:      svcName,
		scheme:    scheme,
		lg:        log,
		port:      port,
		kcli:      k,
		kHTTPCli:  khcli,
	}
	if kerrors.IsNotFound(err0) {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err0,
			Detail:   fmt.Sprintf("get %s service info from namespace %s", svcName, ns),
		}
	} else if err0 != nil {
		return nil, &trait.Error{
			Internal: trait.ECK8sUnknow,
			Err:      err0,
			Detail:   fmt.Sprintf("get %s service info from namespace %s", svcName, ns),
		}
	}
	client.host = fmt.Sprintf("%s:%d", svc.Spec.ClusterIP, port)
	return client, nil
}

var (
	namespace = "anyshare"
	logLevel  = "debug"
	addr      = ""
	port      = 8888
	mode      = "apiserver"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: `从当前节点代理应用部署最小化页面和后端`,
	Long: `使用mini-deploy proxy 命令可以通过当前节点的指定端口(默认8888)代理最小化部署页面。
本地代理采用http协议, 通过http://ip:port可以进入最小化安装页面`,
	Version: version.Version,
	Run: func(cmd *cobra.Command, args []string) {
		if proxyCmdRun(cmd.Context()) != nil {
			os.Exit(1)
		}
	},
}

func proxyCmdRun(ctx context.Context) error {
	log := cutils.NewLogger(logLevel)
	route := http.NewServeMux()
	ctx, cancel := trait.WithCancelCauesContext(ctx)
	defer cancel(&trait.Error{
		Internal: trait.ECExit,
		Err:      context.Canceled,
		Detail:   "executor main routine exit",
	})
	deployInstaler, err := NewBackend(ctx, namespace, "deploy-installer", "http", 9090, "/internal", log)
	if err != nil {
		log.Errorf("Please check deploy-installer has been installed. Get deploy installer error: %s", err.Error())
		return err
	}
	// deployWeb, err := NewBackend(ctx, namespace, "deploy-web-core-static", "http", 18800, "", log)
	// if err != nil {
	// 	log.Errorf("Please check deploy-mini-web has been installed. Get deploy installer error: %s", err.Error())
	// 	return err
	// }
	route.HandleFunc("/", redirectRoot)
	// route.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
	// })

	if mode == "apiserver" {
		// route.HandleFunc("/mini/deploy/", deployWeb.proxyByApiServer)
		// route.HandleFunc("/mini/deploy/service-management/", deployWeb.proxyByApiServer)
		route.HandleFunc("/api/deploy-installer/", deployInstaler.proxyByApiServer)
	} else {
		// route.HandleFunc("/mini/deploy/", deployWeb.proxy)
		// route.HandleFunc("/mini/deploy/service-management/", deployWeb.proxy)
		route.HandleFunc("/api/deploy-installer/", deployInstaler.proxy)
	}

	spa := spaHandler{staticFS: staticFiles, staticPath: "static", indexPath: "index.html"}
	route.Handle("/mini/deploy/", spa)
	// route.Handle("/libs/", spa)
	// route.Handle("/static/", spa)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: route,
	}
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case <-ch:
		}
		log.Info("receive stop signal, try stopping http server...")

		close(ch)
		_ = srv.Shutdown(context.Background())
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	locker := &sync.Mutex{}
	var err0 error
	go func() {
		defer wg.Done()
		log.Info("start proxy")
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("http server running error: %s", err.Error())
				locker.Lock()
				defer locker.Unlock()
				err0 = err
			}
		}
	}()
	wg.Wait()
	locker.Lock()
	defer locker.Unlock()
	log.Info("exit")
	return err0
}

func redirectRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h := w.Header()
		h.Set("Location", "./mini/deploy/service-management/service-deploy")
		w.WriteHeader(http.StatusFound)
	} else {
		http.NotFound(w, r)
	}
}

func init() {
	flags := proxyCmd.Flags()
	flags.StringVarP(&logLevel, "log_level", "v", "debug", "log filter level")
	flags.StringVarP(&namespace, "namespace", "n", namespace, "namespace scope for this request")
	flags.StringVarP(&addr, "addr", "a", addr, "the proxy local addr, default listen all")
	flags.IntVarP(&port, "port", "p", port, "the proxy local port")
	flags.StringVarP(&mode, "mode", "m", mode, "the proxy mode; may need 'apiserver' mod in some cloud env. other value will be 'direct' mode")
}

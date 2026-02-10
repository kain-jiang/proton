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
	"strings"
	"sync"
	"syscall"
	"time"

	"taskrunner/api/rest"
	"taskrunner/cmd/version"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	cutils "taskrunner/cmd/utils"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Rules struct {
	Rules            []Rule `json:"rules"`
	CoreRules        []Rule `json:"coreRules"`
	AllRoleCoreRules []Rule `json:"allRoleCoreRules"`
}

// Rule 代理规则
type Rule struct {
	SvcName string   `json:"svc"`
	Port    int      `json:"port"`
	Path    []string `json:"path"`
	// 去除路由前缀，用于新路由到旧路由的转换
	RemovePrefix string `json:"removePrefix"`
	// 为路由增加前缀,在removePrefix操作之后增加
	Prefix string `json:"prefix"`
	// 理论上不进行认证的只有敏感信息无关的GET与POST接口
	// 且不该在多实例转发单实例行为中
	NotAuthMethod []string `json:"noAuthMehod"`
}

func ShouldAuth(ctx *gin.Context, r Rule) bool {
	method := ctx.Request.Method
	for _, m := range r.NotAuthMethod {
		mu := strings.ToUpper(m)
		if method == mu {
			return false
		}
	}
	return true
}

type Handler struct {
	Rule
	auth OauthMiddleware
	log  logrus.Entry
	cli  *http.Client
}

func NewHandler(r Rule, log logrus.Entry, cli *http.Client, auth OauthMiddleware) *Handler {
	return &Handler{
		auth: auth,
		Rule: r,
		log:  log,
		cli:  cli,
	}
}

func (h *Handler) Proxy(ctx *gin.Context) {
	if ShouldAuth(ctx, h.Rule) {
		h.auth.Authentication(ctx)
		if ctx.IsAborted() {
			return
		}
	}
	proxy(ctx, h.cli, fmt.Sprintf("%s:%d", h.Rule.SvcName, h.Rule.Port), h.Prefix, h.RemovePrefix, h.log)
}

type MultiHandler struct {
	m   *MutiIdentificationMiddleware
	log logrus.Entry
	Rule
	*SystemManager
}

func NewMultiHandler(log *logrus.Entry, rule Rule, tp *TokenParser, s *SystemManager) *MultiHandler {
	return &MultiHandler{
		m:             NewMutiIdentificationMiddleware(log, tp),
		log:           *log,
		Rule:          rule,
		SystemManager: s,
	}
}

func (h *MultiHandler) Proxy(ctx *gin.Context) {
	log := h.log
	sid := 0
	if ShouldAuth(ctx, h.Rule) {
		ti := h.m.GetUserInfo(ctx)
		if ctx.IsAborted() {
			return
		}
		if ti.Sid <= 0 {
			logError(ctx, &h.log, "get system id from token for proxy fail")
			rest.AuthenticationError.From(
				fmt.Sprintf(
					"user must bind a system, user role: [%s], system id: [%d]",
					ti.Role, ti.Sid),
			).AbortGin(ctx)
			return
		}
		sid = ti.Sid
	} else {
		idStr := ctx.Request.Header.Get("Systemid")
		sidInt, rerr := strconv.Atoi(idStr)
		if rerr != nil {
			logError(ctx, &h.log, "get Systemid from header for proxy fail")
			rest.NotFoundError.From("must set Systemid header for proxy").AbortGin(ctx)
			return
		}
		sid = sidInt
	}

	s, err := h.getSystem(ctx, sid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		logError(ctx, &h.log, fmt.Sprintf("system [%d] not found", sid))
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, err)
	} else if err != nil {
		logError(ctx, &h.log, err.Error())
		log.Error(err.Error())
	}

	proxy(ctx, h.client, fmt.Sprintf("%s.%s:%d", h.Rule.SvcName, s.NameSpace, h.Rule.Port), h.Prefix, h.RemovePrefix, h.log)
}

func proxy(ctx *gin.Context, client *http.Client, host string, prefix string, removePrefix string, log logrus.Entry) {
	// 创建新的请求，将请求体作为 io.Reader 传递
	start := time.Now()
	var err1 error
	r := ctx.Request
	w := ctx.Writer
	statusCode := -1
	url := deepcopy.Copy(ctx.Request.URL).(*url.URL)
	url.Scheme = "http"
	url.Host = host
	url.Path = prefix + strings.TrimPrefix(url.Path, removePrefix)
	newUrl := url.String()
	defer func() {
		cost := time.Since(start).Seconds()
		if err1 != nil {
			log.Errorf("Method: %s, RawUrl: %s, backendUrl: %s, statusCode: %d, cost: %fs, error:%s",
				r.Method, r.URL.String(), newUrl, statusCode, cost, err1.Error())
		} else {
			log.Infof("Method: %s, RawUrl: %s, backendUrl: %s, statusCode: %d, cost: %fs",
				r.Method, r.URL.String(), newUrl, statusCode, cost)
		}
	}()

	req, rerr := http.NewRequest(ctx.Request.Method, newUrl, ctx.Request.Body)
	if rerr != nil {
		err1 = rerr
		rest.UnknownError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	req.Header = r.Header

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		err1 = fmt.Errorf("error sending request url error: %s", err.Error())
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
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
	}
}

type TokenParser struct {
	cli *http.Client `json:"-"`
	url string
	// tokenCache
}

func NewTokenParser(cli *http.Client, url string) *TokenParser {
	return &TokenParser{
		cli: cli,
		url: url,
	}
}

type TokenHttpResp struct {
	TokenInfo `json:"data"`
}

type TokenInfo struct {
	Role string `json:"role"`
	Sid  int    `json:"systemId"`
}

// GetUserinfo cache token and user info
func (t *TokenParser) GetUserinfo(ctx context.Context, token string) (*TokenInfo, *trait.Error) {
	return t.getUserinfo(ctx, token)
}

func (t *TokenParser) getUserinfo(_ context.Context, token string) (*TokenInfo, *trait.Error) {
	u := &TokenHttpResp{}
	err := utils.DoJsonHTTP(t.cli, http.MethodGet, t.url, nil, map[string]string{
		"Authorization": token,
	}, u, 200)
	return &u.TokenInfo, err
}

type SystemManager struct {
	lock   *sync.RWMutex
	cache  map[int]trait.System
	Url    string `json:"url"`
	log    *logrus.Entry
	client *http.Client
}

func NewSstemManager(url string, cli *http.Client, log *logrus.Entry) *SystemManager {
	return &SystemManager{
		lock:   &sync.RWMutex{},
		cache:  make(map[int]trait.System),
		Url:    url,
		log:    log,
		client: cli,
	}
}

func (t *SystemManager) cacheSystem(s trait.System) {
	// no cahce config, free the bytes object
	s.Config = nil
	t.lock.Lock()
	defer t.lock.Unlock()
	t.cache[s.SID] = s
}

func (t *SystemManager) GetSystem(ctx context.Context, sid int) (trait.System, *trait.Error) {
	t.lock.RLock()
	s, ok := t.cache[sid]
	t.lock.RUnlock()
	if !ok {
		s, err := t.getSystem(ctx, sid)
		if err != nil {
			t.cacheSystem(s)
		}
		return s, err
	}
	return s, nil
}

func (t *SystemManager) getSystem(_ context.Context, sid int) (trait.System, *trait.Error) {
	s := trait.System{}
	err := utils.DoJsonHTTP(t.client, http.MethodGet, fmt.Sprintf("%s/%d", t.Url, sid), nil, nil, &s, 200)
	return s, err
}

var (
	namespace = "test"
	logLevel  = "debug"
	addr      = ""
	port      = 8888
	rulePath  = ""
)

var proxyCmd = &cobra.Command{
	Use:     "proxy",
	Short:   `代理请求到对应实例`,
	Long:    `基于多实例或单实例模式,结合token将请求代理到对应实例.实现请求路由与转发.`,
	Version: version.Version,
	Run: func(cmd *cobra.Command, args []string) {
		if proxyCmdRun(cmd.Context()) != nil {
			os.Exit(1)
		}
	},
}

func proxyCmdRun(ctx context.Context) error {
	log := cutils.NewLogger(logLevel)
	groute := gin.New()
	route := NewRouter()
	groute.Any("*Any", route.Handler)
	bs, err := os.ReadFile(rulePath)
	if err != nil {
		log.Errorf("read rule file [%s] error: %s", rulePath, err.Error())
		return err
	}
	rs := &Rules{}
	if err := yaml.Unmarshal(bs, rs); err != nil {
		log.Errorf("decode rule file [%s], error: %s", rulePath, err.Error())
		return err
	}
	cli := &http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var coreAuth OauthMiddleware
	var coreAllRoleAuth OauthMiddleware
	var coreToSupport OauthMiddleware

	// TODO 修正为core鉴权
	kcli, err := utils.NewKubeclient()
	if err != nil {
		log.Errorf("load k8s clinet error: %s", err.Error())
		return err
	}
	// 单实例下不需要基于多实例角色进行鉴权，因此认证鉴权方式一致
	coreAuth = newOauthMiddleware(log, namespace, kcli)
	coreAllRoleAuth = coreAuth
	coreToSupport = coreAuth

	{
		registryRoute := func(rules []Rule, m OauthMiddleware, log logrus.Entry) {
			for _, r := range rules {
				handler := NewHandler(r, log, cli, m)
				for _, p := range r.Path {
					route.AddRoute(p, handler.Proxy)
				}
			}
		}

		// 核心包转发至support实例路由，仅实例管理员可调用
		registryRoute(rs.Rules, coreToSupport, *log.WithField("module", "mproxy"))

		// 核心包超级管理员鉴权路由
		registryRoute(rs.CoreRules, coreAuth, *log.WithField("module", "coreAdmin"))
		// 核心包路由，仅认证token不进行鉴权。
		registryRoute(rs.AllRoleCoreRules, coreAllRoleAuth, *log.WithField("module", "coreAllRole"))

	}
	ctx, cancel := trait.WithCancelCauesContext(ctx)
	defer cancel(&trait.Error{
		Internal: trait.ECExit,
		Err:      context.Canceled,
		Detail:   "executor main routine exit",
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: groute,
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

func init() {
	flags := proxyCmd.Flags()
	flags.StringVarP(&logLevel, "log_level", "v", "debug", "log filter level")
	flags.StringVarP(&addr, "addr", "a", addr, "the proxy local addr, default listen all")
	flags.IntVarP(&port, "port", "p", port, "the proxy local port")
	flags.StringVarP(&rulePath, "rule", "r", rulePath, `代理规则文件路径,仅支持yaml格式`)
	flags.StringVarP(&namespace, "namespace", "n", "", "单实例模式时延迟加载的auth配置命名空间")

	if err := proxyCmd.MarkFlagRequired("rule"); err != nil {
		panic(err)
	}
}

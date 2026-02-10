package main

import (
	"context"
	"fmt"
	"net/http"

	"taskrunner/api/rest"
	"taskrunner/pkg/oauth"
	"taskrunner/trait"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	_roleRegular string = "platform-regular"
	_roleAdmin   string = "platform-admin"
)

type OauthMiddleware interface {
	Authentication(ctx *gin.Context)
}
type oauthMiddleware struct {
	*logrus.Logger
	RealAuthentication gin.HandlerFunc
	kcli               kubernetes.Interface
	namespace          string
	// clients            map[string]oauth.Client
}

func newOauthMiddleware(log *logrus.Logger, namespace string, kcli kubernetes.Interface) *oauthMiddleware {
	m := oauthMiddleware{
		Logger: log,
		// clients:   make(map[string]oauth.Client),
		namespace: namespace,
	}
	m.RealAuthentication = m.PassAuthentication
	m.kcli = kcli
	return &m
}

func (o *oauthMiddleware) Authentication(ctx *gin.Context) {
	o.RealAuthentication(ctx)
}

func (o *oauthMiddleware) PassAuthentication(ctx *gin.Context) {
	ocli, err0 := o.loadOauthClient(ctx)
	if err0 != nil {
		o.Logger.Error(err0)
		rest.UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}
	if ocli != nil {
		realAuthentication := func(ctx *gin.Context) {
			// why deploy-web set token in cookies rather then authorization header?
			token, err0 := ctx.Request.Cookie("deploy.oauth2_token")
			if err0 == http.ErrNoCookie {
				rest.InvalidAuthorized.From("please set token in cookies").AbortGin(ctx)
				return
			}

			uid, err := ocli.GetUserID(ctx, token.Value)
			if err != nil {
				if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
					err1 := rest.HTTPError{
						StatusCode: err.Detail.(int),
						ErrorCode:  err.Detail.(int),
						Detail:     err.Error(),
					}
					err1.AbortGin(ctx)
				} else {
					o.Logger.Error(err)
					rest.UnknownError.From(err.Error()).AbortGin(ctx)
				}
				return
			}
			roles, err := ocli.GetUserRole(ctx, uid)
			if err != nil {
				if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
					err1 := rest.HTTPError{
						StatusCode: err.Detail.(int),
						ErrorCode:  err.Detail.(int),
						Detail:     err.Error(),
					}
					err1.AbortGin(ctx)
				} else {
					o.Logger.Error(err)
					rest.UnknownError.From(err.Error()).AbortGin(ctx)
				}
				return
			}
			allowPass := false
			o.Logger.Tracef("roles: %#v", roles)
			for _, role := range roles {
				allowPass = allowPass || role == "super_admin"
				allowPass = allowPass || role == "sys_admin"
				allowPass = allowPass || role == "sec_admin"
				allowPass = allowPass || role == "audit_admin"
				allowPass = allowPass || role == "org_manager"
				allowPass = allowPass || role == "org_audit"
			}
			if !allowPass {
				rest.AuthenticationError.AbortGin(ctx)
				return
			}
		}
		o.RealAuthentication = realAuthentication
		o.RealAuthentication(ctx)
	}
}

func (o *oauthMiddleware) loadOauthClient(ctx context.Context) (cli *oauth.Client, err error) {
	se, err := o.kcli.CoreV1().Secrets(o.namespace).Get(ctx, "cms-release-config-oauth-registry-info", v1.GetOptions{})
	if errors.IsNotFound(err) {
		// no found, no load
		err = nil
		return
	} else if err != nil {
		return
	}
	body, ok := se.Data["default.yaml"]
	if !ok {
		return
	}

	clients := make(map[string]oauth.Client)

	err = yaml.Unmarshal(body, &clients)
	if err != nil {
		err = fmt.Errorf("load oatuh clients from secret error: %s, resp: %s", err, body)
		o.Logger.Error(err)
		return
	}

	ocli, ok := clients["deploy-web"]
	if !ok {
		// no found, no load
		return
	}

	ocli = oauth.NewClient(o.Logger, ocli.HydraOauthClientID, ocli.HydraOauthClientSecret)
	clients["deploy-web"] = ocli
	cli = &ocli
	return
}

type MutiOauthMiddleware struct {
	m *MutiIdentificationMiddleware
	// Rule
	*TokenParser
}

func NewMutiOauthMiddleware(log *logrus.Entry, tp *TokenParser) *MutiOauthMiddleware {
	return &MutiOauthMiddleware{
		m:           NewMutiIdentificationMiddleware(log, tp),
		TokenParser: tp,
	}
}

func (h *MutiOauthMiddleware) Authentication(ctx *gin.Context) {
	ti := h.m.GetUserInfo(ctx)
	if ctx.IsAborted() {
		return
	}

	if ti.Role != _roleAdmin {
		rest.AuthenticationError.From(
			"The user isn't allow to operate system resource",
		).AbortGin(ctx)
		return
	}
}

type DirectOauthMiddleware struct{}

func (h *DirectOauthMiddleware) Authentication(ctx *gin.Context) {}

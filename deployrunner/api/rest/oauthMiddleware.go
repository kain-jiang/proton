package rest

import (
	"context"
	"fmt"
	"net/http"

	"taskrunner/pkg/oauth"
	"taskrunner/trait"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}
	if ocli != nil {
		realAuthentication := func(ctx *gin.Context) {
			// why deploy-web set token in cookies rather then authorization header?
			token, err0 := ctx.Request.Cookie("deploy.oauth2_token")
			if err0 == http.ErrNoCookie {
				InvalidAuthorized.From("please set token in cookies").AbortGin(ctx)
				return
			}

			uid, err := ocli.GetUserID(ctx, token.Value)
			if err != nil {
				if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
					err1 := HTTPError{
						StatusCode: err.Detail.(int),
						ErrorCode:  err.Detail.(int),
						Detail:     err.Error(),
					}
					err1.AbortGin(ctx)
				} else {
					o.Logger.Error(err)
					UnknownError.From(err.Error()).AbortGin(ctx)
				}
				return
			}
			roles, err := ocli.GetUserRole(ctx, uid)
			if err != nil {
				if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
					err1 := HTTPError{
						StatusCode: err.Detail.(int),
						ErrorCode:  err.Detail.(int),
						Detail:     err.Error(),
					}
					err1.AbortGin(ctx)
				} else {
					o.Logger.Error(err)
					UnknownError.From(err.Error()).AbortGin(ctx)
				}
				return
			}
			isSuperOrSys := false
			o.Logger.Tracef("roles: %#v", roles)
			for _, role := range roles {
				isSuperOrSys = isSuperOrSys || role == "super_admin" || role == "sys_admin"
			}
			if !isSuperOrSys {
				AuthenticationError.AbortGin(ctx)
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

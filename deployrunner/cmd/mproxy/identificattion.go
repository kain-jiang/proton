package main

import (
	"net/http"

	"taskrunner/api/rest"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MutiIdentificationMiddleware struct {
	log logrus.Entry
	// Rule
	*TokenParser
}

func NewMutiIdentificationMiddleware(log *logrus.Entry, tp *TokenParser) *MutiIdentificationMiddleware {
	return &MutiIdentificationMiddleware{
		log: *log,
		// Rule:          rule,
		TokenParser: tp,
	}
}

func logError(ctx *gin.Context, log *logrus.Entry, msg string) {
	log.Errorf("RawUrl: %s, error: %s", ctx.Request.URL, msg)
}

func (h *MutiIdentificationMiddleware) GetUserInfo(ctx *gin.Context) *TokenInfo {
	token := ""
	// why deploy-web set token in cookies rather then authorization header?
	tokenCookie, rerr := ctx.Request.Cookie("deploy.tenant_token")
	if rerr == http.ErrNoCookie {
		logError(ctx, &h.log, "not set token header or cookie")
		rest.InvalidAuthorized.From("please set token in cookies").AbortGin(ctx)
		return nil
	}

	if tokenCookie.Value == "" {
		token = ctx.Request.Header.Get("Authorization")
	} else {
		token = "Bearer " + tokenCookie.Value
	}

	ti, err := h.GetUserinfo(ctx, token)
	if trait.IsInternalError(err, trait.ErrNotFound) ||
		trait.IsInternalError(err, trait.ECNoAuthorized) {
		logError(ctx, &h.log, err.Error())
		rest.InvalidAuthorized.From(err.Error()).AbortGin(ctx)
		return nil
	} else if trait.IsInternalError(err, trait.ECInvalidAuthorized) {
		logError(ctx, &h.log, err.Error())
		rest.InvalidAuthorized.From(err.Error()).AbortGin(ctx)
		return nil
	} else if err != nil {
		logError(ctx, &h.log, err.Error())
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return nil
	}
	return ti
}

func (h *MutiIdentificationMiddleware) Authentication(ctx *gin.Context) {
	_ = h.GetUserInfo(ctx)
	if ctx.IsAborted() {
		return
	}
	// if ti.Role != _roleAdmin {
	// 	rest.AuthenticationError.From(
	// 		"The user isn't allow to operate system resource",
	// 	).AbortGin(ctx)
	// 	return
	// }
}

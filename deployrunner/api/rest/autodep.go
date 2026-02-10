package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

type AppChoosed struct {
	AName    string `json:"name"`
	AVersion string `json:"version"`
	Select   bool   `json:"select"`
}

type AppDependencyItem struct {
	AName  string `json:"name"`
	ATitle string `json:"title"`
}

type AppChoosedResult struct {
	AID          int                 `json:"aid"`
	AName        string              `json:"name"`
	ATitle       string              `json:"title"`
	AVersion     string              `json:"version"`
	Versions     []string            `json:"versions"`
	Select       bool                `json:"select"`
	Installed    bool                `json:"installed"`
	Dependencies []AppDependencyItem `json:"dependencies"`
}

// GenerateAutoDependencies parse applications' depedence then generate result to select
//
//	@Summary		解析请求中的应用数组对应应用包的依赖信息,实际应该的部署、更新内容
//	@Description	解析请求中的应用数组对应应用包的依赖信息,实际应该的部署、更新内容
//	@Tags			application
//	@Accept			json
//	@Produce		json
//	@Param			sid		query		string			true	"SystemID"
//	@Param			apps	body		[]AppChoosed	true	"应用包选择列表"
//	@Success		200		{object}	map[string]AppChoosedResult
//	@Failure		400		{object}	HTTPError
//	@Failure		404		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/application/autodependence [POST]
func (e *ExecutorEngine) GenerateAutoDependencies(ctx *gin.Context) {
	var apps []AppChoosed
	systemID := ctx.Query("sid")
	sid, err := strconv.Atoi(systemID)
	if systemID == "" || err != nil {
		ParamError.From("must provide integer system id").AbortGin(ctx)
		return
	}

	lang := ctx.Query("lang")
	if lang == "" {
		ParamError.From("must provide string lang").AbortGin(ctx)
		return
	}

	if rerr := ctx.BindJSON(&apps); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// 过滤掉未被选择的APP，其他APP将重新生成
	var selectApps []AppChoosed
	for _, app := range apps {
		if app.Select {
			selectApps = append(selectApps, app)
		}
	}

	// 解析依赖树
	acs, err := e.autoParseDependency(ctx, selectApps)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
			return
		}
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	// 补全依赖关系
	result, terr := e.autoCompleteDependencies(ctx, acs, sid, lang)
	if terr != nil {
		if trait.IsInternalError(terr, trait.ErrApplicationNotFound) {
			ApplicationNotfoundError.From(terr.ToJson()).AbortGin(ctx)
			return
		}
		UnknownError.From(terr.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (e *ExecutorEngine) autoParseDependency(ctx *gin.Context, selectApps []AppChoosed) (map[string]AppChoosed, error) {
	result := make(map[string]AppChoosed)
	for _, a := range selectApps {
		aid, err := e.Store.GetAPPID(ctx, a.AName, a.AVersion)
		if err != nil {
			return nil, fmt.Errorf("get app id failed: %w", err)
		}
		app, err := e.Store.GetAPP(ctx, aid)
		if err != nil {
			return nil, fmt.Errorf("get app failed: %w", err)
		}

		// 添加依赖
		for _, d := range app.Dependence {
			if ac, ok := result[d.AName]; !ok {
				// 不在结果内(版本选择其实无效)
				result[d.AName] = AppChoosed{
					AName:    d.AName,
					AVersion: d.Version,
					Select:   false,
				}
			} else {
				// 除非select，不然要变更版本
				// 或者不变更任何逻辑
				v := d.Version
				if ac.Select {
					v = ac.AVersion
				}
				result[d.AName] = AppChoosed{
					AName:    d.AName,
					AVersion: v,
					Select:   ac.Select,
				}
			}
		}

		// 添加自己， 一定是选中的，版本已经固定
		result[app.AName] = AppChoosed{
			AName:    app.AName,
			AVersion: app.Version,
			Select:   true, // 自己一定是被选择的
		}
	}
	return result, nil
}

func (e *ExecutorEngine) autoCompleteDependencies(ctx *gin.Context, acs map[string]AppChoosed, sid int, lang string) (map[string]AppChoosedResult, *trait.Error) {
	result := make(map[string]AppChoosedResult)
	for an, ac := range acs {
		var versions []string
		apps, err := e.autoGetAllApps(ctx, ac.AName)
		if err != nil {
			e.Log.WithField("app", ac.AName).Errorln("auto get all app versions failed")
			return nil, err
		}

		// 查询失败，不存在
		if len(apps) == 0 {
			return nil, &trait.Error{
				Internal: trait.ErrApplicationNotFound,
				Detail:   an,
				Err:      fmt.Errorf("cannot find any uploaded's app for %s", an),
			}
		}

		// 所有版本
		for _, app := range apps {
			versions = append(versions, app.Version)
		}

		// 自动选择最新版本
		selectVersion := ac.AVersion
		if !ac.Select {
			selectVersion = versions[len(versions)-1]
		}

		// 单个结果自动选中
		_select := ac.Select
		if len(versions) == 1 {
			_select = true
		}

		// 判断是否安装
		installed, err := e.autoAppInstalled(ctx, an, sid)
		if err != nil {
			e.Log.WithField("app", an).Errorln("auto check app installed failed")
			return nil, err
		}

		// 当前依赖
		a, err := e.autoGetAnApp(ctx, ac.AName, selectVersion)
		if err != nil {
			e.Log.WithField("app", ac.AName).WithField("version", selectVersion).Errorln("auto get app version failed")
			return nil, err
		}
		deps := make([]AppDependencyItem, 0, len(a.Dependence))
		for _, dep := range a.Dependence {
			_atitle, err := e.autoGetAnAppTitle(ctx, dep.AName, lang)
			if err != nil {
				e.Log.WithField("app", an).Errorln("auto get app lang failed on dep")
				return nil, err
			}
			deps = append(deps, AppDependencyItem{
				AName:  dep.AName,
				ATitle: _atitle,
			})
		}

		// 获取别名
		atitle, err := e.autoGetAnAppTitle(ctx, ac.AName, lang)
		if err != nil {
			e.Log.WithField("app", ac.AName).Errorln("auto get app lang failed")
			return nil, err
		}

		result[an] = AppChoosedResult{
			AID:          a.AID,
			ATitle:       atitle,
			AName:        ac.AName,
			AVersion:     selectVersion,
			Versions:     versions,
			Select:       _select,
			Installed:    installed,
			Dependencies: deps,
		}

	}
	return result, nil
}

func (e *ExecutorEngine) autoGetAnAppTitle(ctx *gin.Context, name string, lang string) (string, *trait.Error) {
	return e.Store.GetAppLang(ctx, lang, name, AppZone)
}

func (e *ExecutorEngine) autoGetAllApps(ctx *gin.Context, aname string) ([]trait.ApplicationMeta, *trait.Error) {
	lastAid := -1
	limit := 50

	var result []trait.ApplicationMeta
	for {
		searchResult, err := e.Store.SearchAPP(ctx, limit, lastAid, aname)
		if err != nil {
			return nil, err
		}
		result = append(result, searchResult...)
		if len(searchResult) < limit {
			break
		}
		lastAid = searchResult[len(searchResult)-1].AID
	}
	return result, nil
}

func (e *ExecutorEngine) autoGetAnApp(ctx *gin.Context, aname, aversion string) (*trait.Application, *trait.Error) {
	aid, err := e.Store.GetAPPID(ctx, aname, aversion)
	if err != nil {
		return nil, err
	}
	app, err := e.Store.GetAPP(ctx, aid)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (e *ExecutorEngine) autoAppInstalled(ctx *gin.Context, aname string, sid int) (bool, *trait.Error) {
	_, err := e.Store.GetWorkAPPIns(ctx, aname, sid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

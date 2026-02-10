package rest

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

const AppZone = "app"

// GetAPP get application detail from store
//
//	@Summary		获取应用详细信息
//	@Description	通过应用包名与版本获取应用详细信息
//	@Tags			application
//	@Accept			json
//	@Produce		json
//	@Param			lang	query		string	false	"语言参数"
//	@Param			name	path		string	true	"应用包名"
//	@Param			version	path		string	true	"应用包版本"
//	@Success		200		{object}	trait.Application
//	@Failure		400		{object}	HTTPError
//	@Failure		404		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/application/name/{name}/{version} [get]
func (e *ExecutorEngine) GetAPPWithName(ctx *gin.Context) {
	a := ctx.Param("name")
	v := ctx.Param("version")
	aid, err := e.GetAPPID(ctx, a, v)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
	} else if err == nil {
		e.getAPP(ctx, aid)
	} else {
		UnknownError.From(err.Error()).AbortGin(ctx)
	}
}

// InstallSort 需要解析排序的应用包列表
type InstallSort struct {
	// 排序后的应用列表列表
	Sorted []AppMetaWithLang `json:"sorted"`
	// 依赖但未在请求中的应用列表
	Outer []*AppMetaWithLang `json:"outer"`
}

// AppMetaWithLang 附带语言展示的应用信息
type AppMetaWithLang struct {
	trait.AppDepMeta `json:",inline"`
	// 应用包人类阅读名称,跟随国际化参数变化,无国际化映射时显示aname
	Alias string `json:"title"`
}

func (l *AppMetaWithLang) LangKey() string {
	return l.AName
}

func (l *AppMetaWithLang) SetLang(lang string) {
	l.Alias = lang
}

// GetApplicationDependenceSort parse applications' depedence then return installer sort
//
//	@Summary		解析请求中的应用数组对应应用包的依赖信息,返回安装行为顺序
//	@Description	解析请求中的应用数组对应应用包的依赖信息,返回安装行为顺序.当各个应用未完整填写自身依赖信息时,获得的顺序可能有误
//	@Tags			application
//	@Accept			json
//	@Produce		json
//	@Param			lang			query		string				false	"语言参数"
//	@Param			check_system	query		bool				false	"是否基于当前系统进行缺失依赖检查,默认为true"
//	@Param			apps			body		[]AppMetaWithLang	true	"应用包列表"
//	@Success		200				{object}	InstallSort
//	@Failure		400				{object}	HTTPError
//	@Failure		404				{object}	HTTPError
//	@Failure		500				{object}	HTTPError
//	@Router			/application/dependencesort [PUT]
func (e *ExecutorEngine) GetApplicationDependenceSort(ctx *gin.Context) {
	apps := []AppMetaWithLang{}
	if rerr := ctx.BindJSON(&apps); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	appInfos := []*trait.Application{}
	for _, a := range apps {
		aid, err := e.Store.GetAPPID(ctx, a.AName, a.Version)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			err.Detail = fmt.Sprintf("application %s:%s not found, please upload.", a.AName, a.Version)
			NotFoundError.From(err.Error()).AbortGin(ctx)
			return
		} else if err != nil {
			err.Detail = fmt.Sprintf("get application %s:%s error", a.AName, a.Version)
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		app, err := e.Store.GetAPP(ctx, aid)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			err.Detail = fmt.Sprintf("application %s:%s not found, please upload.", a.AName, a.Version)
			NotFoundError.From(err.Error()).AbortGin(ctx)
			return
		} else if err != nil {
			err.Detail = fmt.Sprintf("get application %s:%s error", a.AName, a.Version)
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		appInfos = append(appInfos, app)
	}
	res := InstallSort{}

	// 获取系统当前实例列表
	var exists []trait.ApplicationInstanceOverview
	if check := ctx.Query("check_system"); check != "false" {
		sid := 0
		if e.SID >= 0 {
			sid = e.SID
		} else {
			queryInt, err := parseIntFromQuery(ctx, "sid")
			if err != nil {
				err.AbortGin(ctx)
				return
			}
			sid = queryInt[0]
		}
		f := &trait.AppInsFilter{
			Sid:    sid,
			Offset: 0,
			Limit:  100,
		}
		for {
			metas, err := e.ListWorkAPPIns(ctx, f)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			f.Offset += len(metas)
			exists = append(exists, metas...)
			if len(metas) == 0 {
				break
			}
		}
	}

	// 排序与外部节点计算
	res.Outer, res.Sorted = parseApplicationDeps(apps, appInfos, exists)
	// 语言获取
	if lang := ctx.Query("lang"); lang != "" {
		s := e.Store
		err := trait.ConvertLangs(ctx, s, res.Outer, lang, AppZone)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
	}
	ctx.JSON(http.StatusOK, res)
}

func parseApplicationDeps(input []AppMetaWithLang, apps []*trait.Application, exists []trait.ApplicationInstanceOverview) (external []*AppMetaWithLang, sorted []AppMetaWithLang) {
	index := make(map[string]bool, len(exists))
	for _, a := range exists {
		index[a.AName] = true
	}
	appIndex := make(map[string]*trait.Application, len(input))
	for _, a := range apps {
		appIndex[a.AName] = a
		index[a.AName] = true
	}
	cache := map[string]bool{}
	for _, a := range apps {
		for _, i := range a.Dependence {
			if o := index[i.AName]; !o {
				// 外部依赖节点
				if _, ok := cache[i.AName]; !ok {
					// 避免重复添加
					external = append(external, &AppMetaWithLang{
						AppDepMeta: i,
					})
				}
			}
		}
	}

	sort.Slice(input, func(i, j int) bool {
		io := appIndex[input[i].AName]
		jo := appIndex[input[j].AName]
		for _, k := range jo.Dependence {
			if io.AName == k.AName {
				return true
			}
		}
		return false
	})

	return external, input
}

// GetAPP get application detail from store
//
//	@Summary		获取应用详细信息
//	@Description	通过应用包ID获取应用详细信息
//	@Tags			application
//	@Accept			json
//	@Produce		json
//	@Param			lang	query		string	false	"语言参数"
//	@Param			id		path		int		true	"应用包ID"
//	@Success		200		{object}	trait.Application
//	@Failure		400		{object}	HTTPError
//	@Failure		404		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/application/{id} [get]
func (e *ExecutorEngine) GetAPP(ctx *gin.Context) {
	id := ctx.Param("aid")
	aid, err0 := strconv.Atoi(id)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the application is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	e.getAPP(ctx, aid)
}

func (e *ExecutorEngine) getAPP(ctx *gin.Context, aid int) {
	a, err := e.Store.GetAPP(ctx, aid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
	} else {
		if lang := ctx.Query("lang"); lang != "" {
			alias, err := e.GetAppLang(ctx, lang, a.AName, AppZone)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			a.Alias = alias
		}
		ctx.JSON(http.StatusOK, a)
	}
}

// ListAPP list application overview from store
//
//	@Summary		获取已上传的应用包列表
//	@Description	获取特定应用包列表,并提供分页滚动查询能力,当分页内为空时表示无下一分页。
//	@Description	当name参数为空时分页内为不同名称的应用包,不含相同名称的不同版本信息。
//	@Descriptio		当name参数非空时,分页内为特定包的不同版本列表。
//
//	@Tags			application
//	@Accept			json
//	@Produce		json
//
//	@Param			lang	query		string	false	"语言参数"
//	@Param			offset	query		int		false	"分页内最后一个应用包ID或以-1起始分页查询，默认-1"
//	@Param			limit	query		int		false	"分页大小,默认20"
//	@Param			sid		query		int		false	"系统ID，多实例模式必填"
//	@Param			name	query		string	false	"应用包名"
//	@Param			nowork	query		bool	false	"是否过滤已安装的应用"
//
//	@Success		200		{object}	[]trait.ApplicationMeta
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/application [get]
func (e *ExecutorEngine) ListAPP(ctx *gin.Context) {
	id := ctx.Query("offset")
	limit := ctx.Query("limit")
	aname := ctx.Query("name")
	nowork := ctx.Query("nowork")
	filterwork := nowork == "" || nowork == "false"
	lastAid := -1
	limitInt := 20
	if id != "" {
		aid, err := strconv.Atoi(id)
		if err != nil {
			ParamError.From(fmt.Sprintf("the offset is not a int, error: %s", err.Error())).AbortGin(ctx)
			return
		}
		lastAid = aid
	}

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			ParamError.From(fmt.Sprintf("the limit is not a int, error: %s", err.Error())).AbortGin(ctx)
			return
		}
		limitInt = l
	}

	var as []trait.ApplicationMeta
	var err *trait.Error
	if aname != "" {
		as, err = e.Store.SearchAPP(ctx, limitInt, lastAid, aname)
	} else if !filterwork {
		sid := 0
		if e.SID >= 0 {
			sid = e.SID
		} else {
			queryInt, err := parseIntFromQuery(ctx, "sid")
			if err != nil {
				err.AbortGin(ctx)
				return
			}
			sid = queryInt[0]
		}
		as, err = e.Store.ListSystemAPPNoWorked(ctx, limitInt, lastAid, sid)
	} else {
		as, err = e.Store.ListAPP(ctx, limitInt, lastAid)
	}
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
	} else {
		if lang := ctx.Query("lang"); lang != "" {
			for i, j := range as {
				alias, err := e.GetAppLang(ctx, lang, j.AName, AppZone)
				if err != nil {
					UnknownError.From(err.Error()).AbortGin(ctx)
					return
				}
				as[i].Alias = alias
			}
		}
		ctx.JSON(http.StatusOK, as)
	}
}

// UploadApplicationPackage upload application package
//
//	@Summary		上传应用包
//	@Description	上传应用包文件,接口将会解析包内容并保存
//
//	@Tags			application
//	@Accept			multipart/form-data
//	@Accept			octet-stream
//	@Produce		json
//
//	@Param			file			formData	file		false	"File to upload content-type is multipart/form-data"	@Accept	multipart/form-data
//	@Param			file			body		[]byte		false	"File to upload when content-type is octet-stream"		@Accept	octet-stream
//	@Success		200				{object}	int			"返回应用包ID"
//	@Failure		400				{object}	HTTPError	"客户端请求参数错误"
//	@Failure		409				{object}	HTTPError	"客户端重复上传应用包"
//	@Failure		500				{object}	HTTPError	"系统内部错误"
//	@Router			/application 	[post]
func (e *ExecutorEngine) UploadApplicationPackage(ctx *gin.Context) {
	reader := ctx.Request.Body
	switch ctx.ContentType() {
	case "multipart/form-data":
		h, err0 := ctx.FormFile("file")
		if err0 != nil {
			ParamError.From(err0.Error()).AbortGin(ctx)
			return
		}
		f, err0 := h.Open()
		if err0 != nil {
			ParamError.From(err0.Error()).AbortGin(ctx)
		}
		defer f.Close()
		reader = f
	default:
	}

	aid, err := e.Executor.UploadApplicationPackage(ctx, reader)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrApplicationFile) {
			ParamError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrUniqueKey) {
			UniqueKeyError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, aid)
	}
}

// CreateSystemInfo create system
//
//	@Summary		创建一个运行服务的系统环境
//	@Description	创建一个运行服务的系统环境，使用系统承载各个应用的运行。
//	@Description	一个系统内不会出现两个相同的应用或组件，但几个系统间可以。
//
//	@Tags			system
//	@Accept			json
//	@Produce		json
//
//	@Param			systemInfo	body		trait.System	true	"系统信息与相关配置"
//
//	@Success		200			{object}	int				"返回系统ID"
//	@Failure		400			{object}	HTTPError		"客户端请求参数错误"
//	@Failure		409			{object}	HTTPError		"客户端请求重复创建"
//	@Failure		500			{object}	HTTPError		"系统内部错误"
//	@Router			/system 	[post]
func (e *ExecutorEngine) CreateSystemInfo(ctx *gin.Context) {
	s := &trait.System{}
	if err0 := ctx.BindJSON(s); err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	id, err := e.Store.InsertSystemInfo(ctx, *s)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrUniqueKey) {
			UniqueKeyError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, id)
	}
}

// UpdateSystemInfo update system
//
//	@Summary		更新一个已存在的系统配置
//	@Description	更新一个已存在的系统配置但仅更新其最外层配置,不会除非在其上应用与组件更新
//	@Description	一个系统内不会出现两个相同的应用或组件，但几个系统间可以。
//
//	@Tags			system
//	@Accept			json
//	@Produce		json
//
//	@Param			systemInfo	body		trait.System	true	"系统信息与相关配置"
//
//	@Success		200			{object}	int				"返回为空"
//	@Failure		400			{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError		"对象不存在，不允许更新"
//	@Failure		500			{object}	HTTPError		"系统内部错误"
//	@Router			/system 	[PUT]
func (e *ExecutorEngine) UpdateSystemInfo(ctx *gin.Context) {
	s := &trait.System{}
	if err0 := ctx.BindJSON(s); err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	err := e.Store.UpdateSystemInfo(ctx, *s)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}

// GetSystemInfo get the system info
//
//	@Summary		获取运行服务的系统环境
//	@Description	获取运行服务的系统环境。
//	@Description	一个系统内不会出现两个相同的应用或组件，但几个系统间可以。
//
//	@Tags			system
//	@Accept			json
//	@Produce		json
//
//	@Param			sid				path		int				true	"系统ID"
//
//	@Success		200				{object}	trait.System	"系统信息"
//	@Failure		400				{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404				{object}	HTTPError		"对象不存在"
//	@Failure		500				{object}	HTTPError		"系统内部错误"
//	@Router			/system/{sid}	[get]
func (e *ExecutorEngine) GetSystemInfo(ctx *gin.Context) {
	id := ctx.Param("sid")
	sid, err0 := strconv.Atoi(id)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the application is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	s, err := e.Store.GetSystemInfo(ctx, sid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
	} else {
		ctx.JSON(http.StatusOK, s)
	}
}

// DeleteSystemInfo delete the system info
//
//	@Summary		删除system信息
//	@Description	删除system信息
//	@Description	注意该接口目前仅轻量删除system对象，而不会清理system承载的其他对象
//
//	@Tags			system
//	@Accept			json
//	@Produce		json
//
//	@Param			sid				path		int			true	"系统ID"
//
//	@Success		200				{object}	nil			"系统信息"
//	@Failure		400				{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404				{object}	HTTPError	"对象不存在"
//	@Failure		500				{object}	HTTPError	"系统内部错误"
//	@Router			/system/{sid}	[delete]
func (e *ExecutorEngine) DeleteSystemInfo(ctx *gin.Context) {
	id := ctx.Param("sid")
	sid, err0 := strconv.Atoi(id)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the application is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	err := e.Store.DeleteSystemInfo(ctx, sid)
	if trait.IsInternalError(err, trait.ErrApplicationStillUse) {
		ApplicationStillUseError.From(err.Error()).AbortGin(ctx)
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
	} else {
		ctx.JSON(http.StatusOK, sid)
	}
}

type SystemList struct {
	Data  []*trait.System `json:"data"`
	Total int             `json:"totalNum"`
}

// ListSystemInfo list system
//
//	@Summary		获取当前系统列表
//	@Description	获取当前系统列表,并提供分页滚动查询能力,当分页内为空时表示无下一分页。
//
//	@Tags			system
//	@Accept			json
//	@Produce		json
//
//	@Param			offset	query		int		false	"偏移量,默认0"
//	@Param			limit	query		int		false	"分页大小,默认20"
//	@Param			page	query		bool	false	"是否获取总数,设置该参数时结果返回格式为'{"data": []trait.System, "totalNum": int}'，为兼容旧格式与使用默认为false"
//	@Param			mode	query		bool	false	"是否根据系统单或多实例模式策略返回数据,默认否"
//
//	@Success		200		{object}	[]trait.System
//	@Failure		400		{object}	HTTPError
//	@Failure		500		{object}	HTTPError
//	@Router			/system [get]
func (e *ExecutorEngine) ListSystemInfo(ctx *gin.Context) {
	if e.SID >= 0 && ctx.Query("mode") == "true" {
		s, err := e.Store.GetSystemInfo(ctx, e.SID)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
		ss := []*trait.System{s}

		if ctx.Query("page") == "true" {
			ctx.JSON(http.StatusOK, SystemList{
				Data:  ss,
				Total: 1,
			})
		} else {
			ctx.JSON(http.StatusOK, ss)
		}
		return
	}

	id := ctx.Query("offset")
	limit := ctx.Query("limit")
	lastid := 0
	limitInt := 20
	if id != "" {
		aid, err0 := strconv.Atoi(id)
		if err0 != nil {
			ParamError.From(fmt.Sprintf("the offset is not a int, error: %s", err0.Error())).AbortGin(ctx)
			return
		}
		lastid = aid
	}

	if limit != "" {
		l, err0 := strconv.Atoi(limit)
		if err0 != nil {
			ParamError.From(fmt.Sprintf("the limit is not a int, error: %s", err0.Error())).AbortGin(ctx)
			return
		}
		if l < 0 {
			l = 0
		}
		limitInt = l
	}
	ss, err := e.Store.ListSystemInfo(ctx, limitInt, lastid)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	if ss == nil {
		ss = []*trait.System{}
	}
	if ctx.Query("page") == "true" {
		count, err := e.Store.CountSystemInfo(ctx)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		ctx.JSON(http.StatusOK, SystemList{
			Data:  ss,
			Total: count,
		})
	} else {
		ctx.JSON(http.StatusOK, ss)
	}
}

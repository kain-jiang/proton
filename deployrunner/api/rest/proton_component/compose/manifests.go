package compose

import (
	"encoding/json"
	"fmt"
	"net/http"

	"taskrunner/api/rest"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

const manifestZone = "manifests"

type ManifestsJob struct {
	trait.ComposeJobMeata `json:",inline"`
	// 基础组件和应用任务配置清单
	ComposeJobSchema `json:"config"`
}

func (mj *ManifestsJob) ComposeJob() *trait.ComposeJob {
	return &trait.ComposeJob{
		ComposeJobMeata: mj.ComposeJobMeata,
		Config:          mj.ComposeJobSchema.ComposeJobConfig(),
	}
}

func ComposeJob2ManifestsJob(j *trait.ComposeJob) *ManifestsJob {
	return &ManifestsJob{
		ComposeJobMeata:  j.ComposeJobMeata,
		ComposeJobSchema: ComposejobConfig2ComposeJobsSchema(j.Config),
	}
}

// Manifests 应用与资源组件资源清单
type Manifests struct {
	trait.ComposeJobManifestsMeta `json:",inline"`
	Manifests                     ComposeJobSchema `json:"config"`
}

func ComposeManifests2Manifests(mm *trait.ComposeJobManifests) *Manifests {
	m := &Manifests{
		ComposeJobManifestsMeta: mm.ComposeJobManifestsMeta,
		Manifests:               ComposejobConfig2ComposeJobsSchema(mm.Manifests),
	}
	return m
}

func (m *Manifests) ComposeManifest() *trait.ComposeJobManifests {
	return &trait.ComposeJobManifests{
		ComposeJobManifestsMeta: m.ComposeJobManifestsMeta,
		Manifests:               m.Manifests.ComposeJobConfig(),
	}
}

// ComposeJobSchema 是ComposeJobManifests的转换，用于前端接口参数一致
type ComposeJobSchema struct {
	// 访问地址配置,用于设置实例的默认入口网关配置
	AccessInfo trait.AccessInfo `json:"-"`
	// proton有状态基础资源组件配置
	ProtonComponent []json.RawMessage `json:"pcomponents"`
	// 上层应用安装任务配置,其中应用名称与版本为必填项
	AppConfig []rest.JobShcema `json:"apps"`
}

func ComposejobConfig2ComposeJobsSchema(cj trait.ComposeJobConfig) ComposeJobSchema {
	apps := cj.AppConfig
	ajs := make([]rest.JobShcema, 0, len(apps))
	for _, i := range apps {
		ains := rest.JobShcema{
			ApplicationInstanceOverview: trait.ApplicationInstanceOverview{
				ApplicationMeta:         i.ApplicationMeta,
				ApplicationinstanceMeta: i.ApplicationinstanceMeta,
			},
			FromData: rest.AppInstanceConfig{
				AppConfig: i.AppConfig,
			},
		}
		coms := make(map[string]*trait.ComponentInstance, len(i.Components))
		for _, c := range i.Components {
			coms[c.Component.Name] = c
		}
		ains.FromData.Components = coms
		ajs = append(ajs, ains)
	}

	return ComposeJobSchema{
		AccessInfo:      cj.AccessInfo,
		ProtonComponent: cj.ProtonComponent,
		AppConfig:       ajs,
	}
}

func (jc *ComposeJobSchema) ComposeJobConfig() trait.ComposeJobConfig {
	ps := make([]*trait.ApplicationInstance, 0, len(jc.AppConfig))
	for _, i := range jc.AppConfig {
		ps = append(ps, i.ToAppIns())
	}
	return trait.ComposeJobConfig{
		AccessInfo:      jc.AccessInfo,
		ProtonComponent: jc.ProtonComponent,
		AppConfig:       ps,
	}
}

// UploadManifests 上传套件任务配置清单
//
//	@Summary		上传套件任务配置清单
//	@Description	上传套件任务配置清单，用于后续查看套件和创建套件任务
//
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//
//	@Param			config	body		Manifests	true	"套件配置"
//
//	@Success		200		{object}	int			"返回为空"
//	@Failure		400		{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404		{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		412		{object}	HTTPError	"不允许设置任务配置"
//	@Failure		500		{object}	HTTPError	"系统内部错误"
//	@Router			/manifests [post]
func (s *Server) UploadManifests(ctx *gin.Context) {
	obj := &Manifests{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		rest.ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	objj := obj.ComposeManifest()
	if objj.LangNames != nil {
		for k, v := range objj.LangNames {
			if err := s.InsertAppLang(ctx, k, objj.Name, v, manifestZone); err != nil {
				rest.UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
		}
	}

	if err := s.InsertComposeManifests(ctx, *objj); err != nil {
		if trait.IsInternalError(err, trait.ErrUniqueKey) {
			rest.UniqueKeyError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrParam) {
			rest.ParamError.From(err.Error()).AbortGin(ctx)
		} else {
			rest.UnknownError.From(err.Error()).AbortGin(ctx)
		}
		return
	}
}

// GetManifests get manifests detail from store
//
//	@Summary		获取套件配置清单
//	@Description	通过名称与版本获取套件配置清单
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Param			lang	query		string		false	"语言参数"
//	@Param			name	path		string		true	"套件名"
//	@Param			version	path		string		true	"套件版本"
//	@Success		200		{object}	Manifests	"返回套件配置清单"
//	@Failure		400		{object}	HTTPError	"参数错误"
//	@Failure		404		{object}	HTTPError	"对象不存在"
//	@Failure		500		{object}	HTTPError	"系统异常或内部错误"
//	@Router			/manifests/{name}/{version} [get]
func (s *Server) GetManifests(ctx *gin.Context) {
	name := ctx.Param("name")
	version := ctx.Param("version")
	obj, err := s.GetComposeManifests(ctx, name, version)
	if err == nil {

		if lang := ctx.Query("lang"); lang != "" {
			if err := trait.ConvertLangs(ctx, s, obj.Manifests.AppConfig, lang, rest.AppZone); err != nil {
				rest.UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			alias, err := s.GetAppLang(ctx, lang, obj.Name, manifestZone)
			if err != nil {
				rest.UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			obj.Title = alias
		}
		m := ComposeManifests2Manifests(obj)
		ctx.JSON(http.StatusOK, m)
	} else if trait.IsInternalError(err, trait.ErrNotFound) {
		rest.NotFoundError.From(err.Error()).AbortGin(ctx)
	} else {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
	}
}

type ManifestMetaList struct {
	Data  []*trait.ComposeJobManifestsMeta `json:"data"`
	Total int                              `json:"totalNum"`
}

// GetManifests get manifests detail from store
//
//	@Summary		获取套件配置清单
//	@Description	通过名称与版本获取套件配置清单
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Param			lang	query		string				false	"语言参数"
//	@Param			name	query		string				true	"套件名"
//	@Param			title	query		string				false	"对应语言的包名"
//	@param			limit	query		int					false	"分页数"
//	@param			offset	query		int					false	"偏移量"
//	@param			nowork	query		bool				false	"是否过滤已安装"
//	@param			sid		query		int					false	"系统空间ID"
//
//	@Success		200		{object}	ManifestMetaList	"返回套件配置清单"
//	@Failure		400		{object}	HTTPError			"参数错误"
//	@Failure		404		{object}	HTTPError			"对象不存在"
//	@Failure		500		{object}	HTTPError			"系统异常或内部错误"
//	@Router			/manifests [get]
func (s *Server) ListManifests(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		lang := ctx.Query("lang")
		alias := ctx.Query("title")
		aname := s.Store.GetAname(lang, alias, manifestZone)
		if aname == "" {
			aname = alias
		}
		name = aname
	}
	querys, herr := rest.ParseIntFromQueryWithDefault(ctx, []string{"limit", "offset", "sid"}, "10", "0", "-1")
	if herr != nil {
		herr.AbortGin(ctx)
		return
	}
	nowork := ctx.Query("nowork")
	filterwork := nowork == "" || nowork == "false"
	f := &trait.ComposeManifestFilter{
		NoWork: !filterwork && name == "",
		Mname:  name,
		Sid:    querys[2],
	}
	objs, total, err := s.ListComposeManifest(ctx, querys[0], querys[1], f)
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	if lang := ctx.Query("lang"); lang != "" {
		err := trait.ConvertLangs(ctx, s, objs, lang, manifestZone)
		if err != nil {
			rest.UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
	}

	ctx.JSON(http.StatusOK, ManifestMetaList{
		Data:  objs,
		Total: total,
	})
}

type ComposeJobMetaList struct {
	Data  []*trait.ComposeJobMeata `json:"data"`
	Total int                      `json:"totalNum"`
}

// ListWorkManifests get work manifest detail from store
//
//	@Summary		已安装套件列表
//	@Description	已安装套件列表
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Param			lang	query		string				false	"语言参数"
//	@Param			name	query		string				true	"套件名"
//	@Param			title	query		string				false	"对应语言的包名"
//	@Param			status	query		[]int				false	"状态过滤器设置,多个状态间关系为'或关系'"
//	@param			limit	query		int					false	"分页数"
//	@param			offset	query		int					false	"偏移量"
//	@param			sid		query		int					false	"系统空间ID"
//
//	@Success		200		{object}	ComposeJobMetaList	"返回套件配置清单"
//	@Failure		400		{object}	HTTPError			"参数错误"
//	@Failure		404		{object}	HTTPError			"对象不存在"
//	@Failure		500		{object}	HTTPError			"系统异常或内部错误"
//	@Router			/manifests/work [get]
func (s *Server) ListWorkManifests(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		lang := ctx.Query("lang")
		alias := ctx.Query("title")
		aname := s.Store.GetAname(lang, alias, manifestZone)
		if aname == "" {
			aname = alias
		}
		name = aname
	}
	status, err1 := rest.ConvertStringToIntArray(ctx.QueryArray("status")...)
	if err1 != nil {
		rest.ParamError.From(fmt.Sprintf("status must set with int array, convert error: %s", err1.Error())).AbortGin(ctx)
		return
	}
	querys, herr := rest.ParseIntFromQueryWithDefault(ctx, []string{"limit", "offset", "sid"}, "10", "0", "-1")
	if herr != nil {
		herr.AbortGin(ctx)
		return
	}

	objs, total, err := s.ListWorkComposeJobManifests(ctx, querys[0], querys[1], trait.ComposeJobFilter{
		Name:   name,
		Status: status,
	})
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if lang := ctx.Query("lang"); lang != "" {
		err := trait.ConvertLangs(ctx, s, objs, lang, manifestZone)
		if err != nil {
			rest.UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
	}

	ctx.JSON(http.StatusOK, ComposeJobMetaList{
		Data:  objs,
		Total: total,
	})
}

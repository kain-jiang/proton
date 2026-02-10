package proton_component

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"component-manage/pkg/models/types"
	"component-manage/pkg/models/types/components"
	"taskrunner/api/rest"
	"taskrunner/api/rest/proton_component/cmp"
	"taskrunner/pkg/component/resources"
	pstore "taskrunner/pkg/store/proton"
	"taskrunner/trait"

	pcfg "taskrunner/pkg/store/proton/configuration"

	"github.com/gin-gonic/gin"
)

const (
	kafkaRlsName    = "kafka"
	zkRlkName       = "zookeeper"
	opensearRlsName = "opensearch"
	redisRlsName    = "proton-redis"
	etcdRlsName     = "proton-etcd"
	pleRlsName      = "proton-policy-engine"
	mariadbRlsName  = "mariadb"
	mongdbRlsName   = "mongodb"
	nebulaRlsName   = "nebula"
)

const (
	rdsInfoName        = "rds"
	mqInfoName         = "mq"
	redisInfoName      = "redis"
	etcdInfoName       = "proton-etcd"
	opensearchInfoName = "opensearch"
	mongoInfoName      = "mongodb"
	opaInfoName        = "proton-policy-engine"
)

type GinServer interface {
	RegistryHandler(r *gin.RouterGroup)
	GetProtonKafka(ctx *gin.Context)
	UpdateKafka(ctx *gin.Context)
	UpdateMQInfo(ctx *gin.Context)
	GetMQInfo(ctx *gin.Context)
	GetProtonZookeeper(ctx *gin.Context)
	UpdateZookeeper(ctx *gin.Context)
	GetProtonOpensearch(ctx *gin.Context)
	UpdateProtonOpensearch(ctx *gin.Context)
	UpdateOpensearchInfo(ctx *gin.Context)
	GetOpensearchInfo(ctx *gin.Context)
	GetProtonMongoDB(ctx *gin.Context)
	UpdateProtonMongoDB(ctx *gin.Context)
	UpdateMongoDBInfo(ctx *gin.Context)
	GetMongoDBInfo(ctx *gin.Context)
	GetProtonNebula(ctx *gin.Context)
	UpdateProtonNebula(ctx *gin.Context)
	GetProtonMariaDB(ctx *gin.Context)
	UpdateProtonMariaDB(ctx *gin.Context)
	UpdateRdsInfo(ctx *gin.Context)
	GetRdsInfo(ctx *gin.Context)
	GetProtonRedis(ctx *gin.Context)
	UpdateProtonRedis(ctx *gin.Context)
	UpdateRedisInfo(ctx *gin.Context)
	GetRedisInfo(ctx *gin.Context)
	GetProtonPolicyEngine(ctx *gin.Context)
	UpdateProtonPolicyEngine(ctx *gin.Context)
	UpdatePolicyEngineInfo(ctx *gin.Context)
	GetPolicyEngineInfo(ctx *gin.Context)
	GetProtonEtcd(ctx *gin.Context)
	UpdateProtonEtcd(ctx *gin.Context)
	UpdateEtcdInfo(ctx *gin.Context)
	GetEtcdInfo(ctx *gin.Context)
}

// Server provide proton-cli-config and component-management proxy
type Server struct {
	system trait.System
	ccli   *cmp.Client
	pcli   *pstore.ProtonClient
}

func NewServer(pcli *pstore.ProtonClient, ss trait.System, ns string) (*Server, *trait.Error) {
	return &Server{
		system: ss,
		ccli:   cmp.NewClient(ns),
		pcli:   pcli,
	}, nil
}

type InfoMeta = trait.ProtonComponentMeta

func (s *Server) RegistryHandler(r *gin.RouterGroup) {
	r.GET("/components/release", s.ListRelease)
	r.GET("/components/info", s.ListInfo)
	r.GET("/components/release/kafka/:name", s.GetProtonKafka)
	r.PUT("/components/release/kafka", s.UpdateKafka)

	r.PUT("/components/info/mq", s.UpdateMQInfo)
	// r.GET("/components/info/mq/:name", s.GetMQInfo)
	r.GET("/components/info/mq/*any", s.GetMQInfo)

	r.GET("/components/release/zookeeper/:name", s.GetProtonZookeeper)
	r.PUT("/components/release/zookeeper", s.UpdateZookeeper)

	r.GET("/components/release/opensearch/:name", s.GetProtonOpensearch)
	r.PUT("/components/release/opensearch", s.UpdateProtonOpensearch)

	r.PUT("/components/info/opensearch", s.UpdateOpensearchInfo)
	// r.GET("/components/info/opensearch/:name", s.GetOpensearchInfo)
	r.GET("/components/info/opensearch/*any", s.GetOpensearchInfo)

	r.GET("/components/release/mongodb/:name", s.GetProtonMongoDB)
	r.PUT("/components/release/mongodb", s.UpdateProtonMongoDB)

	r.PUT("/components/info/mongodb", s.UpdateMongoDBInfo)
	// r.GET("/components/info/mongodb/:name", s.GetMongoDBInfo)
	r.GET("/components/info/mongodb/*any", s.GetMongoDBInfo)

	r.GET("/components/release/nebula/:name", s.GetProtonNebula)
	r.PUT("/components/release/nebula", s.UpdateProtonNebula)

	r.GET("/components/release/mariadb/:name", s.GetProtonMariaDB)
	r.PUT("/components/release/mariadb", s.UpdateProtonMariaDB)

	r.PUT("/components/info/rds", s.UpdateRdsInfo)
	// r.GET("/components/info/rds/:name", s.GetRdsInfo)
	r.GET("/components/info/rds/*any", s.GetRdsInfo)

	r.GET("/components/release/redis/:name", s.GetProtonRedis)
	r.PUT("/components/release/redis", s.UpdateProtonRedis)

	r.PUT("/components/info/redis", s.UpdateRedisInfo)
	// r.GET("/components/info/redis/:name", s.GetRedisInfo)
	r.GET("/components/info/redis/*any", s.GetRedisInfo)

	r.GET("/components/release/policyengine/:name", s.GetProtonPolicyEngine)
	r.PUT("/components/release/policyengine", s.UpdateProtonPolicyEngine)

	r.PUT("/components/info/policyengine", s.UpdatePolicyEngineInfo)
	// r.GET("/components/info/policyengine/:name", s.GetPolicyEngineInfo)
	r.GET("/components/info/policyengine/*any", s.GetPolicyEngineInfo)

	r.GET("/components/release/etcd/:name", s.GetProtonEtcd)
	r.PUT("/components/release/etcd", s.UpdateProtonEtcd)

	r.PUT("/components/info/etcd", s.UpdateEtcdInfo)
	// r.GET("/components/info/etcd/:name", s.GetEtcdInfo)
	r.GET("/components/info/etcd/*any", s.GetEtcdInfo)
}

func GetProtonReleaseMarco[T cmp.ComponentGeneric](ctx *gin.Context, s *Server, rls *cmp.ComponentInstance[T]) {
	err := cmp.Get(ctx, s.ccli, rls)
	if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
		herr := HTTPError{
			StatusCode: err.Detail.(int),
			ErrorCode:  err.Detail.(int),
			Detail:     err.Error(),
		}
		herr.AbortGin(ctx)
		return
	}
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.AbortGin(ctx)
		return
	}
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}

func UpdateProtonReleaseMarco[T cmp.ComponentGeneric](ctx *gin.Context, s *Server, rls cmp.ComponentInstance[T]) {
	err := cmp.Update(ctx, s.ccli, &rls)
	if trait.IsInternalError(err, trait.ErrParam) {
		ParamError.From(err.Error()).AbortGin(ctx)
		return
	}
	if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
		herr := HTTPError{
			StatusCode: err.Detail.(int),
			ErrorCode:  err.Detail.(int),
			Detail:     "provider: component-manage, resp:" + err.Error(),
		}
		herr.AbortGin(ctx)
		return
	}
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}

func UpdateInfoBindReleaseMarco[T cmp.ComponentGeneric](ctx *gin.Context, s *Server, rls *cmp.ComponentInstance[T]) {
	if err := cmp.Update(ctx, s.ccli, rls); err != nil {
		if trait.IsInternalError(err, trait.ErrParam) {
			ParamError.From(err.Error()).AbortGin(ctx)
			return
		}
		if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
			herr := HTTPError{
				StatusCode: err.Detail.(int),
				ErrorCode:  err.Detail.(int),
				Detail:     "provider: component-manage, resp:" + err.Error(),
			}
			herr.AbortGin(ctx)
			return
		}
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}

type infoList struct {
	Data  []InfoMeta `json:"data"`
	Total int        `json:"totalNum"`
}

// ListInfo list connect infos
//
//	@Summary	获取连接信息
//
//	@Tags		proton,ProtonCompoment
//	@Accept		json
//	@Produce	json
//
//	@Param		offset				query		int			false	"分页偏移量"
//	@Param		limit				query		int			false	"分页大小"
//	@Param		sid					query		int			false	"系统ID"
//
//	@Success	200					{object}	infoList	"连接对象元信息"
//	@Failure	400					{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404					{object}	HTTPError	"对象不存在"
//	@Failure	500					{object}	HTTPError	"系统内部错误"
//	@Router		/components/info	[get]
func (s *Server) ListInfo(ctx *gin.Context) {
	param, rerr := rest.ParseIntFromQueryWithDefault(ctx, []string{"limit", "offset"}, "10", "0")
	if rerr != nil {
		rest.ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	ctype := ctx.Query("type")
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	resp := infoList{
		Data: make([]InfoMeta, 0, param[0]),
	}

	conns := conf.ResourceConnectInfo

	if conns.Mq != nil && (ctype == "" || ctype == resources.MQType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   mqInfoName,
			Type:   resources.MQType,
			System: s.system,
		})
	}

	if conns.OpenSearch != nil && (ctype == "" || ctype == resources.OpensearchType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   opensearchInfoName,
			Type:   resources.OpensearchType,
			System: s.system,
		})
	}

	if conns.Mongodb != nil && (ctype == "" || ctype == resources.MongodbType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   mongoInfoName,
			Type:   resources.MongodbType,
			System: s.system,
		})
	}

	if conns.Rds != nil && (ctype == "" || ctype == resources.RDSType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   rdsInfoName,
			Type:   resources.RDSType,
			System: s.system,
		})
	}

	if conns.Redis != nil && (ctype == "" || ctype == resources.REDISType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   redisInfoName,
			Type:   resources.REDISType,
			System: s.system,
		})
	}

	if conns.PolicyEngine != nil && (ctype == "" || ctype == resources.POAType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   opaInfoName,
			Type:   resources.POAType,
			System: s.system,
		})
	}

	if conns.Etcd != nil && (ctype == "" || ctype == resources.EtcdType) {
		resp.Data = append(resp.Data, InfoMeta{
			Name:   etcdInfoName,
			Type:   resources.EtcdType,
			System: s.system,
		})
	}
	resp.Total = len(resp.Data)
	ctx.JSON(http.StatusOK, resp)
}

type releaseList struct {
	Data  []cmp.ComponentInstanceMeta `json:"data"`
	Total int                         `json:"totalNum"`
}

// ListRelease list proton release
//
//	@Summary	获取内置组件实例列表
//
//	@Tags		proton,ProtonCompoment,zookeeper
//	@Accept		json
//	@Produce	json
//
//	@Param		offset				query		int			false	"分页偏移量"
//	@Param		limit				query		int			false	"分页大小"
//	@Param		type				query		[]string	false	"内置组件类型字段"
//	@Param		sid					query		int			false	"系统ID"
//
//	@Success	200					{object}	releaseList	"内置组件实例元数据"
//	@Failure	400					{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404					{object}	HTTPError	"对象不存在"
//	@Failure	500					{object}	HTTPError	"系统内部错误"
//	@Router		/components/release	[get]
func (s *Server) ListRelease(ctx *gin.Context) {
	ls, err := cmp.All(ctx, s.ccli)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	param, rerr := rest.ParseIntFromQueryWithDefault(ctx, []string{"limit", "offset"}, "10", "0")
	if rerr != nil {
		rest.ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	ctype := ctx.QueryArray("type")
	resp := releaseList{}
	for i := range ls {
		ls[i].System = s.system
	}
	resp.Data, resp.Total = filterRelease(ls, param[1], param[0], ctype...)
	ctx.JSON(http.StatusOK, resp)
}

func filterRelease(ls []cmp.ComponentInstanceMeta, offset, limit int, ctypes ...string) (res []cmp.ComponentInstanceMeta, total int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	if len(ctypes) == 0 {
		length := len(ls)
		if len(ls) <= offset {
			return res, length
		}
		if len(ls) < offset+limit {
			return ls[offset:], length
		}
		return ls[offset : offset+limit], length
	}
	index := map[string]bool{}
	for _, i := range ctypes {
		index[i] = true
	}
	cur := 0
	res = make([]cmp.ComponentInstanceMeta, 0, limit)
	for _, j := range ls {
		if _, ok := index[j.Type]; ok {
			cur++
			if cur > offset && cur <= limit {
				res = append(res, j)
			}
		}
	}
	total = cur
	return
}

// GetProtonZookeeper get proton zookeeper
//
//	@Summary	获取proton的zookeeper实例
//
//	@Tags		proton,ProtonCompoment,zookeeper
//	@Accept		json
//	@Produce	json
//
//	@Param		name									path		string						true	"zookeeper实例名称"
//
//	@Success	200										{object}	types.ComponentZookeeper	"zookeeper实例数据"
//	@Failure	400										{object}	HTTPError					"客户端请求参数错误"
//	@Failure	404										{object}	HTTPError					"对象不存在"
//	@Failure	500										{object}	HTTPError					"系统内部错误"
//	@Router		/components/release/zookeeper/{name}	[get]
func (s *Server) GetProtonZookeeper(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ParamError.From("must set zookeeper realse name").AbortGin(ctx)
		return
	}

	rls := &cmp.ComponentInstance[types.ComponentZookeeper]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "zookeeper",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		raw := ctx.DefaultQuery("raw", "")
		if raw == "" || raw == "false" {
			rls.Instance.Info.Sasl.Password = ""
		}
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateZookeeper update proton zookeeper instance
//
//	@Summary		更新proton zookeeper实例配置
//	@Description	更新proton zookeeper实例配置
//
//	@Tags			proton,ProtonCompoment,zookeeper
//	@Accept			json
//	@Produce		json
//
//	@Param			obj								body		types.ComponentZookeeper	true	"zookeeper配置信息"
//
//	@Success		200								{object}	int							"返回为空"
//	@Failure		400								{object}	HTTPError					"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError					"对象不存在，不允许更新"
//	@Failure		500								{object}	HTTPError					"系统内部错误"
//	@Router			/components/release/zookeeper 	[PUT]
func (s *Server) UpdateZookeeper(ctx *gin.Context) {
	obj := &types.ComponentZookeeper{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	if obj.Name == "" {
		ParamError.From("name must be not empty").AbortGin(ctx)
		return
	}

	rls := cmp.ComponentInstance[types.ComponentZookeeper]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "zookeeper",
		},
		Instance: *obj,
	}

	UpdateProtonReleaseMarco(ctx, s, rls)
	if ctx.IsAborted() {
		return
	}

	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	conf.ZooKeeper = obj.Params

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// GetProtonKafka get proton kafka
//
//	@Summary	获取proton的kafka实例
//
//	@Tags		proton,ProtonCompoment,kafka
//	@Accept		json
//	@Produce	json
//
//	@Param		name								path		string					true	"Kafka"
//
//	@Success	200									{object}	types.ComponentKafka	"Kafka"
//	@Failure	400									{object}	HTTPError				"客户端请求参数错误"
//	@Failure	404									{object}	HTTPError				"对象不存在"
//	@Failure	500									{object}	HTTPError				"系统内部错误"
//	@Router		/components/release/kafka/{name}	[get]
func (s *Server) GetProtonKafka(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ParamError.From("must set kafka realse name").AbortGin(ctx)
		return
	}

	rls := &cmp.ComponentInstance[types.ComponentKafka]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "kafka",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		raw := ctx.DefaultQuery("raw", "")
		if raw == "" || raw == "false" {
			rls.Instance.Info.Auth.Password = ""
		}
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateKafka update proton kafka instance
//
//	@Summary		更新proton kafka实例配置
//	@Description	更新proton kafka实例配置
//
//	@Tags			proton,ProtonCompoment,kafka
//	@Accept			json
//	@Produce		json
//
//	@Param			obj							body		types.ComponentKafka	true	"zookeeper配置信息"
//
//	@Success		200							{object}	int						"返回为空"
//	@Failure		400							{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError				"对象不存在，不允许更新"
//	@Failure		500							{object}	HTTPError				"系统内部错误"
//	@Router			/components/release/kafka 	[PUT]
func (s *Server) UpdateKafka(ctx *gin.Context) {
	obj := &types.ComponentKafka{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	if obj.Name == "" {
		ParamError.From("must set kafka realse name").AbortGin(ctx)
		return
	}
	if obj.Dependencies == nil || obj.Dependencies.Zookeeper == "" {
		obj.Dependencies = &components.KafkaComponentDependencies{
			Zookeeper: zkRlkName,
		}
	}

	rls := &cmp.ComponentInstance[types.ComponentKafka]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "kafka",
		},
		Instance: *obj,
	}
	UpdateProtonReleaseMarco(ctx, s, *rls)
	if ctx.IsAborted() {
		return
	}
	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	conf.Kafka = obj.Params
	if conf.ResourceConnectInfo.Mq != nil {
		if conf.ResourceConnectInfo.Mq.SourceType == "internal" &&
			conf.ResourceConnectInfo.Mq.MQType == "kafka" {

			rls := &cmp.ComponentInstance[types.ComponentKafka]{
				ComponentInstanceMeta: cmp.ComponentInstanceMeta{
					Name: obj.Name,
					Type: "kafka",
				},
			}
			err := cmp.Get(ctx, s.ccli, rls)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			conf.ResourceConnectInfo.Mq.FromKafkaInfo(*rls.Instance.Info)
			conf.Kafka = rls.Instance.Params
		}
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type MqFull struct {
	InfoMeta `json:",inline"`
	// mq连接信息
	Info pcfg.MqInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance json.RawMessage `json:"instance,omitempty"`
	// 使用内建kafka时，kafka绑定的内建zookeeper配置
	ZK json.RawMessage `json:"zookeeper,omitempty"`
}

// GetMQInfo 获取mq类型示例信息
//
//	@Summary	获取mq类型示例信息
//
//	@Tags		proton,connect,mq
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name						path		string		true	"mq"
//
//	@Success	200							{object}	MqFull		"mq信息"
//	@Failure	400							{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404							{object}	HTTPError	"对象不存在"
//	@Failure	500							{object}	HTTPError	"系统内部错误"
//	@Router		/components/info/mq/{name}	[get]
func (s *Server) GetMQInfo(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("release name must not be empty").abortGin(ctx)
	// 	return
	// }
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if conf.ResourceConnectInfo.Mq == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	mq := conf.ResourceConnectInfo.Mq
	res := &MqFull{
		InfoMeta: InfoMeta{
			Name: "mq",
		},
		Info: *mq,
	}
	if mq.SourceType == "internal" {
		if mq.MQType == "kafka" {
			// TODO remove
			rls := &cmp.ComponentInstance[types.ComponentKafka]{
				ComponentInstanceMeta: cmp.ComponentInstanceMeta{
					Name: kafkaRlsName,
					Type: "kafka",
				},
			}
			err := cmp.Get(ctx, s.ccli, rls)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			raw := ctx.DefaultQuery("raw", "")
			if raw == "" || raw == "false" {
				rls.Instance.Info.Auth.Password = ""
			}

			kins := rls.Instance
			bs, rerr := json.Marshal(kins)
			if rerr != nil {
				UnknownError.From("decode kafka param error: " + rerr.Error()).AbortGin(ctx)
				return
			}
			res.Instance = bs

			zRls := &cmp.ComponentInstance[types.ComponentZookeeper]{
				ComponentInstanceMeta: cmp.ComponentInstanceMeta{
					Name: zkRlkName,
					Type: "zookeeper",
				},
			}
			err = cmp.Get(ctx, s.ccli, zRls)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}

			if raw == "" || raw == "false" {
				zRls.Instance.Info.Sasl.Password = ""
			}

			zins := zRls.Instance
			bs, rerr = json.Marshal(zins)
			if rerr != nil {
				UnknownError.From("decode kafka param error: " + rerr.Error()).AbortGin(ctx)
				return
			}
			res.ZK = bs
		}
		// nsq don't support
	}
	raw := ctx.DefaultQuery("raw", "")
	if raw == "" || raw == "false" {
		if res.Info.Auth != nil {
			res.Info.Auth.Password = ""
		}
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdateKafka update mq info
//
//	@Summary		更新mq连接信息
//	@Description	更新mq连接信息
//
//	@Tags			proton,connect,mq
//	@Accept			json
//	@Produce		json
//
//	@Param			obj						body		MqFull		true	"mq"
//
//	@Success		200						{object}	int			"返回为空"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/components/info/mq 	[PUT]
func (s *Server) UpdateMQInfo(ctx *gin.Context) {
	body := &MqFull{}
	if rerr := ctx.BindJSON(body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "mq"
	obj := body.Info

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		if obj.MQType == "kafka" {
			zRls := &cmp.ComponentInstance[types.ComponentZookeeper]{
				ComponentInstanceMeta: cmp.ComponentInstanceMeta{
					Name: zkRlkName,
					Type: "zookeeper",
				},
			}
			zupdate := true
			err = cmp.Get(ctx, s.ccli, zRls)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			zins := &zRls.Instance

			if body.ZK == nil {
				zupdate = false
			} else {
				zInput := &types.ComponentZookeeper{}
				if rerr := json.Unmarshal(body.ZK, zInput); rerr != nil {
					ParamError.From("decode zookeeper instance error: " + rerr.Error()).AbortGin(ctx)
					return
				}
				if reflect.DeepEqual(zInput, zins) {
					zupdate = false
				} else {
					zins = zInput
				}
			}
			if zins.Name == "" {
				zins.Name = zkRlkName
			}
			kupdate := true
			rls := &cmp.ComponentInstance[types.ComponentKafka]{
				ComponentInstanceMeta: cmp.ComponentInstanceMeta{
					Name: kafkaRlsName,
					Type: "kafka",
				},
			}
			kins := &types.ComponentKafka{}
			if body.Instance == nil {
				err := cmp.Get(ctx, s.ccli, rls)
				if err != nil {
					UnknownError.From(err.Error()).AbortGin(ctx)
					return
				}
				kins = &rls.Instance
				kupdate = false
			} else {
				if rerr := json.Unmarshal(body.Instance, kins); rerr != nil {
					ParamError.From("decode kafka instance error: " + rerr.Error()).AbortGin(ctx)
					return
				}
			}
			if kins.Name == "" {
				kins.Name = kafkaRlsName
			}
			if kins.Dependencies == nil || kins.Dependencies.Zookeeper == "" {
				kins.Dependencies = &components.KafkaComponentDependencies{
					Zookeeper: zkRlkName,
				}
			}

			// update zk
			if zupdate {
				zRls.Instance = *zins
				UpdateInfoBindReleaseMarco(ctx, s, zRls)
				if ctx.IsAborted() {
					return
				}
			}

			// update kafka
			if kupdate {
				rls.Instance = *kins
				UpdateInfoBindReleaseMarco(ctx, s, rls)
				if ctx.IsAborted() {
					return
				}
			}

			err := cmp.Get(ctx, s.ccli, rls)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			if conf.ResourceConnectInfo.Mq == nil {
				conf.ResourceConnectInfo.Mq = &pcfg.MqInfo{}
			}
			conf.ResourceConnectInfo.Mq.FromKafkaInfo(*rls.Instance.Info)
			conf.Kafka = kins.Params
			conf.ZooKeeper = zins.Params
		} else if obj.MQType == "nsq" {
			if conf.Proton_mq_nsq == nil {
				NotFoundError.From("nsq not installed").AbortGin(ctx)
				return
			}
			obj.SourceType = "internal"
			obj.MQHosts = "proton-mq-nsq-nsqd.resource"
			obj.MQPort = 4151
			obj.MQLookupdHosts = "proton-mq-nsq-nsqlookupd.resource"
			obj.MQLookupdPort = 4161
			obj.Auth = nil
			conf.ResourceConnectInfo.Mq = &obj
		} else {
			ParamError.From(fmt.Sprintf("mq type '%s' don't support", obj.MQType)).AbortGin(ctx)
			return
		}
	} else {
		if obj.Auth != nil && obj.Auth.Password == "" && conf.ResourceConnectInfo.Mq != nil {
			obj.Auth.Password = conf.ResourceConnectInfo.Mq.Auth.Password
		}
		conf.ResourceConnectInfo.Mq = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}

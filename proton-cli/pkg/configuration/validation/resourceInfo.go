package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// 1、必要字段非空校验
// 2、枚举类型校验
// 3、部署了对应内置组件时，不允许使用外置组件；没有部署对应内置组件时，填内置组件信息非法
func ValidateResourceConnectInfo(c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {

	allErrs = append(allErrs, validateRds(c.ResourceConnectInfo.Rds, c, fldPath.Child("rds"))...)

	allErrs = append(allErrs, validateMongodb(c.ResourceConnectInfo.Mongodb, c, fldPath.Child("mongodb"))...)

	allErrs = append(allErrs, validateMq(c.ResourceConnectInfo.Mq, c, fldPath.Child("mq"))...)

	allErrs = append(allErrs, validateRedis(c.ResourceConnectInfo.Redis, c, fldPath.Child("redis"))...)

	allErrs = append(allErrs, validateSearchEngine(c.ResourceConnectInfo.OpenSearch, c, fldPath.Child("search_engine"))...)

	allErrs = append(allErrs, validatePolicyEngine(c.ResourceConnectInfo.PolicyEngine, c, fldPath.Child("policy_engine"))...)

	allErrs = append(allErrs, validateEtcd(c.ResourceConnectInfo.Etcd, c, fldPath.Child("etcd"))...)

	return
}

// validateRds
func validateRds(rds *configuration.RdsInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if rds == nil {
		return
	}

	// username/password非空
	allErrs = append(allErrs, ValidateRequiredString(rds.Username, fldPath.Child("username"))...)

	allErrs = append(allErrs, ValidateRequiredString(rds.Password, fldPath.Child("password"))...)

	// hosts非空
	allErrs = append(allErrs, ValidateRequiredString(rds.Hosts, fldPath.Child("hosts"))...)
	// port有效端口
	allErrs = append(allErrs, ValidatePort(rds.Port, fldPath.Child("port"))...)

	// rds_type只支持MariaDB MySQL GoldenDB DM8 TiDB
	if rds.RdsType == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("rds_type"), ""))
	} else if !isValidType(configuration.RdsTypeList, rds.RdsType) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("rds_type"), rds.RdsType, "valid rdsType: MariaDB MySQL GoldenDB DM8 TiDB KDB9"))
	}

	return
}

// validateMongodb
func validateMongodb(mongodb *configuration.MongodbInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if mongodb == nil {
		return
	}

	// username/password非空
	allErrs = append(allErrs, ValidateRequiredString(mongodb.Username, fldPath.Child("username"))...)

	allErrs = append(allErrs, ValidateRequiredString(mongodb.Password, fldPath.Child("password"))...)

	if mongodb.SourceType != configuration.Internal {
		// hosts非空
		allErrs = append(allErrs, ValidateRequiredString(mongodb.Hosts, fldPath.Child("hosts"))...)
		// port有效端口
		allErrs = append(allErrs, ValidatePort(mongodb.Port, fldPath.Child("port"))...)

		// replica_set 非空
		allErrs = append(allErrs, ValidateRequiredString(mongodb.ReplicaSet, fldPath.Child("replica_set"))...)

		// auth_source非空
		allErrs = append(allErrs, ValidateRequiredString(mongodb.AuthSource, fldPath.Child("auth_source"))...)
	}

	return
}

// validateRedis
func validateRedis(redis *configuration.RedisInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if redis == nil {
		return
	}
	// 该组件内置时连接信息在组件管理服务中生成
	if redis.SourceType == configuration.Internal {
		return
	}

	// connect_type只支持三种模式
	switch redis.ConnectType {
	case configuration.SentinelMode:
		allErrs = append(allErrs, ValidateRequiredString(redis.SentinelHosts, fldPath.Child("sentinel_hosts"))...)

		allErrs = append(allErrs, ValidatePort(redis.SentinelPort, fldPath.Child("sentinel_port"))...)

		allErrs = append(allErrs, ValidateRequiredString(redis.MasterGroupName, fldPath.Child("master_group_name"))...)

	case configuration.MasterSlaverMode:
		allErrs = append(allErrs, ValidateRequiredString(redis.MasterHosts, fldPath.Child("master_hosts"))...)
		allErrs = append(allErrs, ValidatePort(redis.MasterPort, fldPath.Child("master_port"))...)

		allErrs = append(allErrs, ValidateRequiredString(redis.SlaveHosts, fldPath.Child("slave_hosts"))...)
		allErrs = append(allErrs, ValidatePort(redis.SlavePort, fldPath.Child("slave_port"))...)

	case configuration.StandAlonelMode, configuration.ClusterMode:
		allErrs = append(allErrs, ValidateRequiredString(redis.Hosts, fldPath.Child("hosts"))...)
		allErrs = append(allErrs, ValidatePort(redis.Port, fldPath.Child("port"))...)

	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("connect_type"), ""))

	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("connect_type"), redis.ConnectType, "valid connectType: sentinel master-slave standalone"))
	}

	return
}

// validateMq
func validateMq(mq *configuration.MqInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if mq == nil {
		return
	}
	// sourceType必填
	switch mq.SourceType {
	case configuration.Internal:
		if (mq.MqType == configuration.Nsq && c.Proton_mq_nsq == nil) || (mq.MqType == configuration.KafkaType && c.Kafka == nil) {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("source_type"), " sourceType cannot be internal without installed internal module"))
		}
		if mq.MqType != configuration.Nsq && mq.MqType != configuration.KafkaType {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("mq_type"), " mqType only supported for nsq kafka with internal source"))
		}
	case configuration.External:
	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("source_type"), ""))
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("source_type"), mq.SourceType, "valid source_type: internal external"))
	}

	// mq_hosts/mq_port非空
	allErrs = append(allErrs, ValidateRequiredString(mq.MqHosts, fldPath.Child("mq_hosts"))...)

	allErrs = append(allErrs, ValidatePort(mq.MqPort, fldPath.Child("mq_port"))...)

	// rds_type只支持nsq kafka Tonglink htp20 bmq
	if mq.MqType == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("mq_type"), ""))
	} else if !isValidType(configuration.MqTypeList[:], mq.MqType) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mq_type"), mq.MqType, "valid mqType: nsq kafka Tonglink htp20 bmq htp202"))
	}
	// mqType=nsq时，mq_lookupd_hosts/mq_lookupd_port非空
	if mq.MqType == configuration.Nsq {
		allErrs = append(allErrs, ValidateRequiredString(mq.MqLookupdHosts, fldPath.Child("mq_lookupd_hosts"))...)

		allErrs = append(allErrs, ValidatePort(mq.MqLookupdPort, fldPath.Child("mq_lookupd_port"))...)
	}
	// 只有kafka支持auth
	if mq.MqType != configuration.KafkaType && isAssigned(mq.Auth) {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("auth"), fmt.Sprintf("auth only used by kafka;value: %v", mq.Auth)))
	}
	// 当前只有kafka支持开启tls
	if mq.MqType == configuration.KafkaType && isAssigned(mq.Auth) {
		// username/password非空
		allErrs = append(allErrs, ValidateRequiredString(mq.Auth.Username, fldPath.Child("auth.username"))...)

		allErrs = append(allErrs, ValidateRequiredString(mq.Auth.Password, fldPath.Child("auth.password"))...)

		if mq.Auth.Mechanism == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("auth.mechanism"), ""))

		} else if !isValidType(configuration.MechanTypeList[:], mq.Auth.Mechanism) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("auth.mechanism"), mq.Auth.Mechanism, "valid mq MechanType: PLAIN SCRAM-SHA-512 SCRAM-SHA-256"))
		}
	}

	return
}

// validateSearchEngine
func validateSearchEngine(openSearch *configuration.OpensearchInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if openSearch == nil {
		return
	}
	// 该组件内置时连接信息在组件管理服务中生成
	if openSearch.SourceType == configuration.Internal {
		return
	}

	// hosts非空
	allErrs = append(allErrs, ValidateRequiredString(openSearch.Hosts, fldPath.Child("hosts"))...)
	// port有效端口
	allErrs = append(allErrs, ValidatePort(openSearch.Port, fldPath.Child("port"))...)

	// username/password非空
	allErrs = append(allErrs, ValidateRequiredString(openSearch.Username, fldPath.Child("username"))...)

	allErrs = append(allErrs, ValidateRequiredString(openSearch.Password, fldPath.Child("password"))...)

	// 外置opensearch版本只支持5.6.4 7.10.0
	if openSearch.SourceType != configuration.Internal {
		switch openSearch.Version {
		case configuration.Version564, configuration.Version710:

		case "":
			allErrs = append(allErrs, field.Required(fldPath.Child("version"), ""))

		default:
			allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), openSearch.Version, "supported version: 5.6.4 7.10.0"))
		}
	}

	return
}

// validatePolicyEngine
func validatePolicyEngine(policyEngine *configuration.PolicyEngineInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if policyEngine == nil {
		return
	}
	// 该组件内置时连接信息在组件管理服务中生成
	if policyEngine.SourceType == configuration.Internal {
		return
	}

	// hosts非空
	allErrs = append(allErrs, ValidateRequiredString(policyEngine.Hosts, fldPath.Child("hosts"))...)
	// port有效端口
	allErrs = append(allErrs, ValidatePort(policyEngine.Port, fldPath.Child("port"))...)

	return
}

// validateEtcd
func validateEtcd(etcd *configuration.EtcdInfo, c *configuration.ClusterConfig, fldPath *field.Path) (allErrs field.ErrorList) {
	// 为空，直接返回
	if etcd == nil {
		return
	}
	// 该组件内置时连接信息在组件管理服务中生成
	if etcd.SourceType == configuration.Internal {
		return
	}

	// hosts非空
	allErrs = append(allErrs, ValidateRequiredString(etcd.Hosts, fldPath.Child("hosts"))...)
	// port有效端口
	allErrs = append(allErrs, ValidatePort(etcd.Port, fldPath.Child("port"))...)

	allErrs = append(allErrs, ValidateRequiredString(etcd.Secret, fldPath.Child("secret"))...)

	return
}

// isValidType
func isValidType[T configuration.RDSType | configuration.MqType | configuration.MechanType | configuration.SourceType](list []T, value T) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// 判断auth是否赋值
func isAssigned(auth *configuration.Auth) bool {
	if auth == nil || (auth.Username == "" && auth.Password == "" && auth.Mechanism == "") {
		return false
	}
	return true
}

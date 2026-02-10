package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/kafka"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/mariadb"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/mongodb"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/mq"
)

// 基础组件名称与 secret 中的常量保持一致，不然查不到
var serviceNameList = []string{"rds", "mongodb", "redis", "es", "mq", "proton-policy-engine", "proton-etcd"}

const (
	RedisInfoKey = "connectInfo"
	RedisTypeKey = "connectType"

	ResourceConnectInfoKey = "resourceConnectInfo"

	// 存放组件连接信息secret所在命名空间
	SecretNamespace = "anyshare"

	// secret名称都以 "cms-release-config-" + serverName 风格命名
	SecretPrefix = "cms-release-config"
)

// CompleteOldClusterConfFromSecret 旧版本基础组件连接信息存放在名为 cms-release-cofnig-<serviceName> 的 secret 中，
//升级时，从组件对应 secret 中获取连接信息，迁移到 clusterConfig 配置文件 resource_connet_info 下

func CompleteOldClusterConfFromSecret(c *configuration.ClusterConfig, kube *kubernetes.Clientset) error {
	var resourceNamespace = configuration.GetProtonResourceNSFromFile()
	// infoMap 为转成configuration.ResourceConnectInfo 准备的map
	infoMap := map[string]interface{}{}

	for _, server := range serviceNameList {

		data, err := getInfoFromSecret(context.Background(), server, kube)
		// 未找到对应 secret 或在 解析失败
		if err != nil {
			return err
		}
		// redis、proton-policy-engine、proton-etcd响应提需要适配
		switch server {

		case "redis":
			infoMap["redis"] = transformRedis(data)

		case "proton-policy-engine":
			infoMap["policyEngine"] = transformProtonPolicyEngine(server, data)
		case "proton-etcd":
			infoMap["etcd"] = transformProtonPolicyEngine(server, data)
		case "es":
			infoMap["opensearch"] = data
		default:
			infoMap[server] = data
		}
	}

	outMap := map[string]interface{}{}
	outMap[ResourceConnectInfoKey] = infoMap

	if err := mapstructure.Decode(outMap, c); err != nil {
		return err
	}

	// 为空不用补全source_type,提前退出
	if c.ResourceConnectInfo == nil {
		return nil
	}

	// secret 中没有保存区分内外置的字段，这里根据 secret 中保存的连接信息判断组件是内置还是外置，并补全
	// rds 内置时hosts一定是 mariadb 主 service 名称+"."+ 命名空间
	if c.ResourceConnectInfo.Rds != nil {
		if c.ResourceConnectInfo.Rds.Hosts == fmt.Sprintf("%s.%s", mariadb.MasterServiceName, mariadb.ClusterNamespace) {
			c.ResourceConnectInfo.Rds.SourceType = configuration.Internal
		} else {
			c.ResourceConnectInfo.Rds.SourceType = configuration.External
		}
	}
	// mogodb 内置时hosts一定是 mariadb 集群 statefulSet 名称 +"-" + service 名称+"."+ 命名空间格式安副本数拼接
	if c.ResourceConnectInfo.Mongodb != nil {
		if strings.Contains(c.ResourceConnectInfo.Mongodb.Hosts, fmt.Sprintf("%s-%d.%s.%s", mongodb.ClusterStatefulSetName, 0, mongodb.ClusterServiceName, mongodb.ClusterNamespace)) {
			c.ResourceConnectInfo.Mongodb.SourceType = configuration.Internal
		} else {
			c.ResourceConnectInfo.Mongodb.SourceType = configuration.External
		}
	}
	// mq 内置时hosts一定是nsqd service 名称+"."+ 命名空间
	if c.ResourceConnectInfo.Mq != nil {
		if c.ResourceConnectInfo.Mq.MqHosts == fmt.Sprintf("%s.%s", mq.NsqdServiceName, resourceNamespace) || c.ResourceConnectInfo.Mq.MqHosts == fmt.Sprintf("%s.%s.%s", mq.NsqdServiceName, resourceNamespace, mq.NsqdServiceNameSuffix) {
			c.ResourceConnectInfo.Mq.SourceType = configuration.Internal
		} else {
			c.ResourceConnectInfo.Mq.SourceType = configuration.External
		}
	}

	return nil
}

// CompleteInternalInfo 补全内置基础组件连接信息
// 只有当对应内置组件安装了，且传入Source_type=internal或者空，才会补全
func CompleteInternalInfo(c *configuration.ClusterConfig) {
	// 如果什么都没填，初始化
	var resourceNamespace = configuration.GetProtonResourceNSFromFile()
	if c.ResourceConnectInfo == nil {
		c.ResourceConnectInfo = new(configuration.ResourceConnectInfo)
	}
	if c.ResourceConnectInfo.Mongodb != nil {
		// 防止options为空字符串
		if s, ok := c.ResourceConnectInfo.Mongodb.Options.(string); ok && s == "" {
			c.ResourceConnectInfo.Mongodb.Options = nil
		}
	}

	if c.ResourceConnectInfo.Rds != nil && c.Proton_mariadb != nil && (c.ResourceConnectInfo.Rds.SourceType == configuration.Internal || c.ResourceConnectInfo.Rds.SourceType == "") {
		c.ResourceConnectInfo.Rds = &configuration.RdsInfo{
			SourceType: configuration.Internal,
			RdsType:    configuration.MariaDB,

			Hosts: fmt.Sprintf("%s.%s", mariadb.MasterServiceName, mariadb.ClusterNamespace),
			Port:  mariadb.ClusterMariaDBServicePort,

			Username: c.ResourceConnectInfo.Rds.Username,
			Password: c.ResourceConnectInfo.Rds.Password,

			HostsRead: fmt.Sprintf("%s.%s", mariadb.ClusterServiceName, mariadb.ClusterNamespace),
			PortRead:  mariadb.ClusterMariaDBServicePort,
		}
	}

	if c.Proton_mongodb != nil && c.ResourceConnectInfo.Mongodb != nil && (c.ResourceConnectInfo.Mongodb.SourceType == configuration.Internal || c.ResourceConnectInfo.Mongodb.SourceType == "") {
		// mongodb hosts根据节点数确定
		hosts := ""
		if c.Cs.Provisioner == configuration.KubernetesProvisionerExternal {
			hosts = getMongodbHosts(c.Proton_mongodb.ReplicaCount)
		} else {
			hosts = getMongodbHosts(len(c.Proton_mongodb.Hosts))
		}
		// mongodb 没填
		c.ResourceConnectInfo.Mongodb = &configuration.MongodbInfo{
			SourceType: configuration.Internal,
			Hosts:      hosts,
			Port:       mongodb.ClusterServicePort,

			Username: c.ResourceConnectInfo.Mongodb.Username,
			Password: c.ResourceConnectInfo.Mongodb.Password,

			ReplicaSet: mongodb.ReplicaSetName,
			SSL:        false,

			AuthSource: mongodb.AuthSourceName,
		}

	}

	// 不存在时，为MQ尽量选择内置 kafka 或者 nsq, kafka优先
	if c.ResourceConnectInfo.Mq == nil || c.ResourceConnectInfo.Mq.SourceType == configuration.Internal || c.ResourceConnectInfo.Mq.SourceType == "" {
		// 需要配置内置rds
		mqKafkaInfo := &configuration.MqInfo{
			SourceType: configuration.Internal,
			MqType:     configuration.KafkaType,

			MqHosts: fmt.Sprintf("%s.%s", kafka.KafkaServiceName, resourceNamespace),
			MqPort:  kafka.KafkaServicePort,
			Auth: &configuration.Auth{
				Username:  kafka.KafkaDefaultSSLUser,
				Password:  kafka.KafkaDefaultSSLPassword,
				Mechanism: configuration.Plain,
			},
		}
		mqNsqInfo := &configuration.MqInfo{
			SourceType: configuration.Internal,
			MqType:     configuration.Nsq,

			MqHosts: fmt.Sprintf("%s.%s", mq.NsqdServiceName, resourceNamespace),
			MqPort:  mq.NsqdServicePort,

			MqLookupdHosts: fmt.Sprintf("%s.%s", mq.NsqlookuodServiceName, resourceNamespace),
			MqLookupdPort:  mq.NsqlookuodServicePort,
		}

		if c.ResourceConnectInfo.Mq == nil {
			c.ResourceConnectInfo.Mq = &configuration.MqInfo{}
		}

		switch c.ResourceConnectInfo.Mq.MqType {
		case configuration.KafkaType:
			if c.Kafka != nil {
				c.ResourceConnectInfo.Mq = mqKafkaInfo
			} else {
				c.ResourceConnectInfo.Mq = nil
			}
		case configuration.Nsq:
			if c.Proton_mq_nsq != nil {
				c.ResourceConnectInfo.Mq = mqNsqInfo
			} else {
				c.ResourceConnectInfo.Mq = nil
			}
		default:
			if c.Kafka != nil {
				c.ResourceConnectInfo.Mq = mqKafkaInfo
			} else if c.Proton_mq_nsq != nil {
				c.ResourceConnectInfo.Mq = mqNsqInfo
			} else {
				c.ResourceConnectInfo.Mq = nil
			}
		}

	}

}

// getInfoFromSecret 从 secret 中取组件连接信息；组件 secret 必须存在，未找到抛错；解析失败抛错
func getInfoFromSecret(ctx context.Context, name string, k *kubernetes.Clientset) (data map[string]interface{}, err error) {

	log := logger.NewLogger()

	name = fmt.Sprintf("%s-%s", SecretPrefix, name)

	data = map[string]interface{}{}

	// 通过 k8s api 获取到secret
	secret, err := k.CoreV1().Secrets(SecretNamespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("get secret %s from namespace %s failed: %v", name, SecretNamespace, "not found")

			return nil, nil
		} else {
			log.Infof("get secret %s from namespace %s failed: %v", name, SecretNamespace, err)
			return nil, err
		}
	}

	err = yaml.Unmarshal(secret.Data["default.yaml"], &data)
	if err != nil {
		log.Errorf("unmarshal %s from default.yaml failed: %v", name, err)
	}

	return
}

// 根据内置mongodb节点上组装连接信息hosts
func getMongodbHosts(nodes int) (hosts string) {

	var list []string
	for i := 0; i < nodes; i++ {
		list = append(list, fmt.Sprintf("%s-%d.%s.%s", mongodb.ClusterStatefulSetName, i, mongodb.ClusterServiceName, mongodb.ClusterNamespace))
	}
	return strings.Join(list, ",")
}

// secret 中存放的redis结构和RedisInfo设计不一致，在这里转换
func transformRedis(data map[string]interface{}) (infoMap interface{}) {
	/*
		secret 接口响应的redis连接信息结构
		"default": {
			"connectInfo": {
				"masterGroupName": "mymaster",
				"password": "FAKE_PASSWORD",
				"sentinelHost": "proton-redis-proton-redis-sentinel.resource",
				"sentinelPassword": "FAKE_PASSWORD",
				"sentinelPort": 26379,
				"sentinelUsername": "FAKE_USERNAME",
				"username": "FAKE_USERNAME"
			},
			"connectType": "sentinel"
		}
	*/

	if info, ok := data[RedisInfoKey].(map[string]interface{}); ok {
		if connType, ok := data[RedisTypeKey].(string); ok {
			info[RedisTypeKey] = connType

			infoMap = info
		}
	}
	return
}

// secret 中存放的proton-policy-engine、proton-etcd结构和PolicyEngineInfo、EtcdInfo设计不一致，在这里转换
func transformProtonPolicyEngine(service string, data map[string]interface{}) (infoMap interface{}) {

	/*
		secret 接口响应的 proton-policy-engine 连接信息结构
		"default": {
			"proton-policy-engine": {
				"host": "proton-policy-engine-proton-policy-engine-cluster.resource",
				"port": 9800
			}
		}
		secret 接口响应的 proton-etcd 连接信息结构
		"default": {
			"proton-etcd": {
				"host": "proton-etcd.resource",
				"port": 2379,
				"secret": "etcdssl-secret"
			}
		}
	*/
	if info, ok := data[service].(map[string]interface{}); ok {
		infoMap = info
	}
	return
}

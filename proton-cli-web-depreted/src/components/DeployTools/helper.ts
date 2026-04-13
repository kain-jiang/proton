import { ConfigData, ProtonMonitor } from "./index.d";
import standard from "../../assets/img/standard.png";
import { isObject, isString } from "lodash";

function generatePassword(length = 10): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

// Chrony mode constants
export const CHRONY_MODE = {
  // 系统默认
  USER_MANAGED: "usermanaged",
  // 随机选择一个主节点作为时间服务器
  LOCAL_MASTER: "localmaster",
  // 外部时间服务器
  EXTERNAL_NTP: "externalntp",
};

export const FIREWALL_MODE = {
  // 用户自行管理
  USER_MANAGED: "usermanaged",
  // 使用firewalld（默认）
  FIREWALLD: "firewalld",
};

export const CHRONY_MODE_TEXT = {
  // 系统默认
  [CHRONY_MODE.USER_MANAGED]:
    "使用系统默认（不会变更时间源设置，适用于单节点或已经统一配置好操作系统NTP的场景）",
  // 随机选择一个主节点作为时间服务器
  [CHRONY_MODE.LOCAL_MASTER]:
    "使用内置NTP（设置其中一个Kubernetes Master节点作为集群的时间源）",
  // 外部时间服务器
  [CHRONY_MODE.EXTERNAL_NTP]:
    "使用外部NTP（设置输入的NTP服务器作为集群的时间源）",
};

export const FIREWALL_MODE_TEXT = {
  [FIREWALL_MODE.FIREWALLD]: "使用firewalld",
  [FIREWALL_MODE.USER_MANAGED]: "用户自行管理",
};

export const DEPLOY_MODE = {
  // 标准部署模式
  STANDARD: "standard",
  // 云部署模式
  CLOUD: "cloud",
};

// 产品型号
export const DEVICESPECS = {
  // Enterprise
  AS10000: "AS10000",
  AS9000: "AS9000",
  EnterpriseSoft: "Enterprise-Soft",
  EnterpriseSubscription: "Enterprise-Subscription",
  // Express
  AS2000: "AS2000",
  ExpressSoft: "Express-Soft",
  ExpressSubscription: "Express-Subscription",
  H3C: "h3c",
  // SMAS
  smas: "smas",
  smasStandard: "smas-standard",
  smasAdvanced: "smas-advanced",
  // ManagedCloud
  ManagedCloud: "Managed-Cloud",
  // Subscription
  asBasic: "asBasic",
  asStandard_1: "asStandard_1",
  asStandard_2: "asStandard_2",
  asProfessional_1: "asProfessional_1",
  asProfessional_2: "asProfessional_2",
  // SecureDiskBasic
  SecureDiskBasic: "secure-disk-basic",
  AS500: "AS500",
  AS550: "AS550",
  SecureDiskBasicSubscription: "secure-disk-basic-subscription",
  // SecureDiskStandard
  SecureDiskStandard: "secure-disk-standard",
  AS600: "AS600",
  AS650: "AS650",
  SecureDiskStandardSubscription: "secure-disk-standard-subscription",
  // InteliKnowledegeMgntStandard
  InteliKnowledegeMgntStandard: "inteli-knowledge-mgnt-standard",
  km900i: "km900i",
  InteliKnowledegeMgntSubscription: "inteli-knowledge-mgnt-subscription",
};

// kom颁发授权拼写错误
export const errorDeviceSpecsMap = {
  "inteli-knowledge-mgnt-standard": "inteli-knowledege-mgnt-standard",
  "inteli-knowledge-mgnt-subscription": "inteli-knowledege-mgnt-subscription",
};

// 产品型号类别
export const DEVICESPECSTYPE = {
  Enterprise: "Enterprise",
  Express: "Express",
  SMAS: "SMAS",
  ManagedCloud: "ManagedCloud",
  Subscription: "Subscription",
  SecureDiskBasic: "SecureDiskBasic",
  SecureDiskStandard: "SecureDiskStandard",
  InteliKnowledegeMgntStandard: "InteliKnowledgeMgntStandard",
};

export const DEVICESPECSMAP = {
  [DEVICESPECSTYPE.Enterprise]: [
    DEVICESPECS.AS10000,
    DEVICESPECS.AS9000,
    DEVICESPECS.EnterpriseSoft,
    DEVICESPECS.EnterpriseSubscription,
  ],
  [DEVICESPECSTYPE.Express]: [
    DEVICESPECS.AS2000,
    DEVICESPECS.ExpressSoft,
    DEVICESPECS.ExpressSubscription,
    DEVICESPECS.H3C,
  ],
  [DEVICESPECSTYPE.SMAS]: [
    DEVICESPECS.smasStandard,
    DEVICESPECS.smasAdvanced,
    DEVICESPECS.smas,
  ],
  [DEVICESPECSTYPE.ManagedCloud]: [DEVICESPECS.ManagedCloud],
  [DEVICESPECSTYPE.Subscription]: [
    DEVICESPECS.asBasic,
    DEVICESPECS.asStandard_1,
    DEVICESPECS.asStandard_2,
    DEVICESPECS.asProfessional_1,
    DEVICESPECS.asProfessional_2,
  ],
  [DEVICESPECSTYPE.SecureDiskBasic]: [
    DEVICESPECS.SecureDiskBasic,
    DEVICESPECS.AS500,
    DEVICESPECS.AS550,
    DEVICESPECS.SecureDiskBasicSubscription,
  ],
  [DEVICESPECSTYPE.SecureDiskStandard]: [
    DEVICESPECS.SecureDiskStandard,
    DEVICESPECS.AS600,
    DEVICESPECS.AS650,
    DEVICESPECS.SecureDiskStandardSubscription,
  ],
  [DEVICESPECSTYPE.InteliKnowledegeMgntStandard]: [
    DEVICESPECS.InteliKnowledegeMgntStandard,
    DEVICESPECS.km900i,
    DEVICESPECS.InteliKnowledegeMgntSubscription,
  ],
};

export const SERVICES = {
  ProtonMariadb: "proton_mariadb",
  ProtonMongodb: "proton_mongodb",
  ProtonRedis: "proton_redis",
  ProtonNSQ: "proton_mq_nsq",
  ProtonPolicyEngine: "proton_policy_engine",
  ProtonEtcd: "proton_etcd",
  Opensearch: "opensearch",
  Kafka: "kafka",
  Zookeeper: "zookeeper",
  Prometheus: "prometheus",
  Grafana: "grafana",
  Nebula: "nebula",
  ComponentManagement: "component_management",
  NvidiaDevicePlugin: "nvidia_device_plugin",
  PackageStore: "package-store",
  ECeph: "eceph",
  ProtonMonitor: "proton_monitor",
};

export const AddableServices = [
  SERVICES.ComponentManagement,
  SERVICES.NvidiaDevicePlugin,
];

// 配置状态
enum OptionSteps {
  //节点配置
  NodeConfig,

  //网络配置
  NetworkConfig,

  //仓库配置
  RepositoryConfig,

  //数据库配置
  DataBaseConfig,

  // 连接配置
  ConnectInfo,
}

// 数据库存储 类型
export const DataBaseStorageType = {
  // 标准版
  Standard: "standard",

  // 云主机
  Cloud: "cloud",

  // 托管Kubernetes
  DepositKubernetes: "deposit-kubernetes",
};

// 数据库存储 类型
export const DataBaseStorageTypeName = {
  // 标准版
  [DataBaseStorageType.Standard]: "标准模式部署",

  // 云主机
  [DataBaseStorageType.Cloud]: "云主机部署",

  // 托管Kubernetes部署
  [DataBaseStorageType.DepositKubernetes]: "托管Kubernetes部署",
};

// 组件类型
export const SOURCE_TYPE = {
  // 内置
  INTERNAL: "internal",
  // 外置
  EXTERNAL: "external",
};

// 外置关系型数据库类型
export const RDS_TYPE = {
  MARIADB: "MariaDB",
  DM8: "DM8",
  MYSQL: "MySQL",
  GOLDENDB: "GoldenDB",
  TIDB: "TiDB",
  TAURUSDB: "TaurusDB",
  KDB9: "KDB9",
};

// 消息队列
export const MQ_TYPE = {
  NSQ: "nsq",
  TONGLINK: "tonglink",
  HTP2: "htp20",
  HTP202: "htp202",
  KAFKA: "kafka",
  BMP: "bmq",
};

export const MQ_AUTH_MACHANISM = {
  PLAIN: "PLAIN",
  SCRAMSHA512: "SCRAM-SHA-512",
  SCRAMSHA256: "SCRAM-SHA-256",
};

export const OPENSEARCH_VERSION = {
  v7100: "7.10.0",
  v564: "5.6.4",
};

// 第三方搜索引擎类型
export const searchEngineType = {
  Opensearch: "opensearch",
  Elasticsearch: "elasticsearch",
};

export const REDIS_CONNECT_TYPE = {
  // 哨兵模式
  SENTINEL: "sentinel",
  // 主从模式
  MASTER_SLAVE: "master-slave",
  // 单机模式
  STANDALONE: "standalone",
  // 集群模式
  CLUSTER: "cluster",
};

export const NODES_LIMIT = {
  [SERVICES.Grafana]: 1,
  [SERVICES.Prometheus]: 2,
  [SERVICES.ProtonMonitor]: 2,
};

const IP_Family = {
  ipv6: "IPv6",
  ipv4: "IPv4",
  dualStack: "dualStack",
};

const IP_Family_LIST = [
  {
    label: IP_Family.ipv4,
    value: IP_Family.ipv4,
  },
  {
    label: IP_Family.ipv6,
    value: IP_Family.ipv6,
  },
  {
    label: "双栈",
    value: IP_Family.dualStack,
  },
];

export const RESOURCES = {
  ALL: "all",
  LIMITS: "limits",
  REQUESTS: "requests",
};

export const RESOURCES_TYPE = {
  CPU: "cpu",
  MEMORY: "memory",
};

export const NEBULA_COMPONENTS = {
  GRAPHD: "graphd",
  METAD: "metad",
  STORAGED: "storaged",
};

export const CONNECT_SERVICES = {
  RDS: "rds",
  MONGODB: "mongodb",
  REDIS: "redis",
  MQ: "mq",
  OPENSEARCH: "opensearch",
  POLICY_ENGINE: "policy_engine",
  ETCD: "etcd",
};

export const CONNECT_SERVICES_TEXT = {
  [CONNECT_SERVICES.RDS]: "RDS",
  [CONNECT_SERVICES.MONGODB]: "MongoDB",
  [CONNECT_SERVICES.REDIS]: "Redis",
  [CONNECT_SERVICES.MQ]: "MQ",
  [CONNECT_SERVICES.OPENSEARCH]: "OpenSearch",
  [CONNECT_SERVICES.POLICY_ENGINE]: "PolicyEngine",
  [CONNECT_SERVICES.ETCD]: "ETCD",
};

export const CSPlugin = {
  // https://github.com/prometheus/node_exporter
  NodeExporter: "node-exporter",
  // https://github.com/kubernetes/kube-state-metrics
  KubeStateMetrics: "kube-state-metrics",
};

export const CSPlugins = [
  {
    key: CSPlugin.NodeExporter,
    descritpion:
      "用于监控硬件和操作系统指标，可用于将指标收集到 Prometheus 或 AnyRobot 进行可观测性分析",
  },
  {
    key: CSPlugin.KubeStateMetrics,
    descritpion:
      "用于收集和报告Kubernetes集群中各种资源对象(如Pods、Deployments、Services等)的实时状态指标，可用于将指标收集到 Prometheus 或 AnyRobot 进行可观测性分析",
  },
];

/**
 * 页面状态
 */
export enum PageStatus {
  // 基础配置
  BasicInfo,

  // 集群配置
  ClusterConfig,

  // 容器仓库配置
  ContainerRegistryConfig,

  // 基础服务配置
  BasicServiceConfig,

  // 加载中
  Loading,
}

/**
 * IP校验状态
 */
enum nodeInfoCheckResult {
  // 正常, Normal不能移动位置，checkNodeIPByK8SFamily函数依赖这个
  Normal = 0,

  // 输入项为空
  ExistIpv6Ipv4,

  // IPv4 不合法。
  ValidatorIpv4,

  // IPV6 不合法
  validatorIpv6,

  // 该节点已经存在
  hasIpv6Ipv4,

  // 内部ip格式不正确
  InternalIPNotIPv4OrIPv6,

  // 内部ip冲突
  InternalIPConflict,

  // k8s协议栈是IPv4 节点IPv4为空
  K8SIPv4NodeIPv4Empty,

  // k8s协议栈是IPv6 节点IPv6为空
  K8SIPv6NodeIPv6Empty,

  // k8s协议栈是IPv4+IPv6 节点IPv46为空
  K8SIPv46NodeIP46Empty,

  // 节点名称重复
  hasRepeatNodeName,

  // 节点名称不合法
  validatorNodeName,
}

enum ConfigEditStatus {
  // 配置中
  Editing,

  // 初始化中
  initing,

  // 初始化完成
  Success,

  // 初始化出错
  Error,
}

const validatorMessge = {
  [nodeInfoCheckResult.ExistIpv6Ipv4]: {
    result: true,
    message: "ipv4 或者 ipv6至少存在一个。",
  },
  [nodeInfoCheckResult.ValidatorIpv4]: {
    result: true,
    message: "ipv4 格式不正确请重新输入。",
  },
  // ipv4合法性校验
  [nodeInfoCheckResult.validatorIpv6]: {
    // 校验结果
    result: true,

    // 提示信息
    message: "ipv6 格式不正确请重新输入。",
  },
  [nodeInfoCheckResult.hasIpv6Ipv4]: {
    result: true,
    message: "当前节点已经存在，请重新输入。",
  },
  [nodeInfoCheckResult.InternalIPNotIPv4OrIPv6]: {
    result: true,
    message: "内部ip格式不正确。",
  },
  [nodeInfoCheckResult.InternalIPConflict]: {
    result: true,
    message: "节点内部ip冲突。",
  },
  [nodeInfoCheckResult.K8SIPv4NodeIPv4Empty]: {
    result: true,
    message: "当K8S网络协议栈为IPV4时，节点IPV4地址必填。",
  },
  [nodeInfoCheckResult.K8SIPv6NodeIPv6Empty]: {
    result: true,
    message: "当K8S网络协议栈为IPV6时，节点IPV6地址必填。",
  },
  [nodeInfoCheckResult.K8SIPv46NodeIP46Empty]: {
    result: true,
    message: "当K8S网络协议栈为双栈时，节点IPV4和IPV6地址都必填。",
  },
  [nodeInfoCheckResult.hasRepeatNodeName]: {
    result: true,
    message: "当前节点名称已经存在，请重新输入。",
  },
  [nodeInfoCheckResult.validatorNodeName]: {
    result: true,
    message:
      "节点名称格式不正确，请重新输入。需满足正则：^[a-z]([-a-z0-9]*[a-z0-9])?$",
  },
};

const checkIPv6 = (value) => {
  return /^([\da-fA-F]{1,4}:){6}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^::([\da-fA-F]{1,4}:){0,4}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:):([\da-fA-F]{1,4}:){0,3}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){2}:([\da-fA-F]{1,4}:){0,2}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){3}:([\da-fA-F]{1,4}:){0,1}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){4}:((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}$|^:((:[\da-fA-F]{1,4}){1,6}|:)$|^[\da-fA-F]{1,4}:((:[\da-fA-F]{1,4}){1,5}|:)$|^([\da-fA-F]{1,4}:){2}((:[\da-fA-F]{1,4}){1,4}|:)$|^([\da-fA-F]{1,4}:){3}((:[\da-fA-F]{1,4}){1,3}|:)$|^([\da-fA-F]{1,4}:){4}((:[\da-fA-F]{1,4}){1,2}|:)$|^([\da-fA-F]{1,4}:){5}:([\da-fA-F]{1,4})?$|^([\da-fA-F]{1,4}:){6}:$/.test(
    value,
  );
};

const checkIPv4 = (value) => {
  return /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/.test(
    value,
  );
};

const checkNodeName = (value) => {
  return /^[a-z]([-a-z0-9]*[a-z0-9])?$/.test(value);
};

const DEFAULT_STANDARD_PACKAGE_STORE = {
  hosts: [],
  storage: {
    capacity: "10Gi",
    path: "/sysvol/package-store",
  },
  resources: undefined,
};

const DEFAULT_DEPOSIT_K8S_PACKAGE_STORE = {
  replicas: 3,
  storage: {
    capacity: "10Gi",
    storageClassName: undefined,
  },
  resources: undefined,
};

const DEGAULT_EXTERNAL_RDS = {
  source_type: SOURCE_TYPE.EXTERNAL,
  rds_type: RDS_TYPE.MYSQL,
  port: null,
  username: "",
  password: "",
  auto_create_database: true,
  admin_user: "",
  admin_passwd: "",
};

const DEFAULT_INTERNAL_RDS = {
  source_type: SOURCE_TYPE.INTERNAL,
  username: "anyshare",
  password: generatePassword(),
};

const DEFAULT_EXTERNAL_MONGODB = {
  source_type: SOURCE_TYPE.EXTERNAL,
  username: "",
  password: "",
  ssl: false,
  auth_source: "admin",
};

const DEFAULT_INTERNAL_MONGODB = {
  source_type: SOURCE_TYPE.INTERNAL,
  username: "anyshare",
  password: generatePassword(),
};

const DEFAULT_EXTERNAL_REDIS = {
  source_type: SOURCE_TYPE.EXTERNAL,
  connect_type: REDIS_CONNECT_TYPE.MASTER_SLAVE,
};

const DEFAULT_EXTERNAL_MQ = {
  source_type: SOURCE_TYPE.EXTERNAL,
  mq_type: MQ_TYPE.KAFKA,
};

const DEFAULT_INTERNAL_MQ = {
  source_type: SOURCE_TYPE.INTERNAL,
  mq_type: MQ_TYPE.KAFKA,
};

const DEFAULT_EXTERNAL_OPENSEARCH = {
  source_type: SOURCE_TYPE.EXTERNAL,
  version: OPENSEARCH_VERSION.v7100,
  distribution: searchEngineType.Elasticsearch,
};

const DEFAULT_INTERNAL_OPENSEARCH = {
  source_type: SOURCE_TYPE.INTERNAL,
  version: OPENSEARCH_VERSION.v7100,
  distribution: searchEngineType.Opensearch,
};

const DEFAULT_EXTERNAL_SERVICE = {
  source_type: SOURCE_TYPE.EXTERNAL,
};

const DEFAULT_INTERNAL_SERVICE = {
  source_type: SOURCE_TYPE.INTERNAL,
};

const DefaultConfigData = {
  chrony: { mode: CHRONY_MODE.EXTERNAL_NTP },
  firewall: { mode: FIREWALL_MODE.FIREWALLD },
  deploy: {
    mode: DEPLOY_MODE.STANDARD,
  },
  nodesInfo: [],
  // internal_cidr: "",
  // internal_nic: "",
  networkInfo: {
    provisioner: "local",
    master: [],
    addons: [CSPlugin.NodeExporter, CSPlugin.KubeStateMetrics],
    ipFamilies: [IP_Family.ipv4],
    hostNetwork: {
      bip: "172.33.0.1/16",
      podNetworkCidr: "192.169.0.0/16",
      serviceCidr: "10.96.0.0/12",
      ipv4Interface: "",
      ipv6Interface: "",
    },
    ha_port: "16643",
    etcdDataDir: "/sysvol/proton_data/cs_etcd_data",
    dockerDataDir: "/sysvol/proton_data/cs_docker_data",
    enableDualStack: false,
  },
  cr: {
    local: {
      master: [],
      ports: {
        chartMuseum: 5001,
        registry: 5000,
        rpm: 5003,
        crManager: 5002,
      },
      haPorts: {
        chartMuseum: 15001,
        registry: 15000,
        rpm: 15003,
        crManager: 15002,
      },
      storage: "/sysvol/proton_data/cr_data",
    },
  },
  [SERVICES.ProtonMariadb]: {
    hosts: [],
    config: {
      innodb_buffer_pool_size: "8G",
      resource_requests_memory: "24G",
      resource_limits_memory: "24G",
    },
    admin_user: "root",
    admin_passwd: generatePassword(),
    data_path: "/sysvol/mariadb",
    storage_capacity: "",
  },
  [SERVICES.ProtonMongodb]: {
    hosts: [],
    admin_user: "root",
    admin_passwd: generatePassword(),
    data_path: "/sysvol/mongodb/mongodb_data",
    storage_capacity: "",
    resources: {
      requests: {
        cpu: "100m",
        memory: "128Mi",
      },
    },
  },
  [SERVICES.ProtonRedis]: {
    hosts: [],
    admin_user: "root",
    admin_passwd: generatePassword(),
    data_path: "/sysvol/redis/redis_data",
    storage_capacity: "",
    resources: {
      requests: {
        cpu: "100m",
        memory: "30Mi",
      },
    },
  },
  [SERVICES.ProtonNSQ]: {
    hosts: [],
    data_path: "/sysvol/mq-nsq/mq-nsq_data",
    storage_capacity: "",
    resources: {
      requests: {
        cpu: "100m",
        memory: "15Mi",
      },
    },
  },
  [SERVICES.ProtonPolicyEngine]: {
    hosts: [],
    data_path: "/sysvol/policy-engine/policy-engine_data",
    storage_capacity: "",
    resources: {
      requests: {
        cpu: "100m",
        memory: "40Mi",
      },
    },
  },
  [SERVICES.ProtonEtcd]: {
    hosts: [],
    data_path: "/sysvol/proton-etcd/proton-etcd_data",
    storage_capacity: "",
  },
  [SERVICES.Opensearch]: {
    hosts: [],
    mode: "master",
    config: {
      jvmOptions: "-Xmx8g -Xms8g",
      hanlpRemoteextDict:
        "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_dict",
      hanlpRemoteextStopwords:
        "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_stopwords",
    },
    settings: {
      "cluster.routing.allocation.disk.watermark.low": "60%",
      "cluster.routing.allocation.disk.watermark.high": "65%",
      "cluster.routing.allocation.disk.watermark.flood_stage": "70%",
      "cluster.max_shards_per_node": "10000",
      "http.max_initial_line_length": "16kb",
      "bootstrap.memory_lock": true,
    },
    data_path: "/anyshare/opensearch",
    extraValues: {
      storage: {
        repo: {
          nfs: { enabled: false },
          hdfs: { enabled: false },
        },
      },
    },
    storage_capacity: "",
    resources: {
      limits: {
        cpu: "8",
        memory: "40Gi",
      },
      requests: {
        cpu: "1000m",
        memory: "2Gi",
      },
    },
    exporter_resources: {
      requests: {
        cpu: "100m",
        memory: "12Mi",
      },
    },
  },
  [SERVICES.NvidiaDevicePlugin]: {},
  [SERVICES.ComponentManagement]: {},
  [SERVICES.Kafka]: {
    hosts: [],
    data_path: "/sysvol/kafka/kafka_data",
    env: {
      KAFKA_HEAP_OPTS: "-Xmx1g -Xms1g",
    },
    storage_capacity: "",
    disable_external_service: false,
    external_service_list: [
      {
        name: "EXTERNAL",
        port: 9098,
        nodePortBase: 31000,
        enableSSL: false,
      },
      {
        name: "EXTERNAL-TLS",
        port: 9099,
        nodePortBase: 31100,
        enableSSL: true,
      },
    ],
    resources: {
      limits: {
        cpu: "4",
        memory: "2Gi",
      },
      requests: {
        cpu: "100m",
        memory: "128Mi",
      },
    },
    exporter_resources: {
      requests: {
        cpu: "100m",
        memory: "12Mi",
      },
    },
  },
  [SERVICES.Zookeeper]: {
    hosts: [],
    data_path: "/sysvol/zookeeper/zookeeper_data",
    env: {
      JVMFLAGS: "-Xmx500m -Xms500m",
    },
    storage_capacity: "",
    resources: {
      limits: {
        cpu: "1000m",
        memory: "2Gi",
      },
      requests: {
        cpu: "100m",
        memory: "270Mi",
      },
    },
  },
  [SERVICES.Prometheus]: {
    hosts: [],
    data_path: "/sysvol/prometheus",
    storage_capacity: "",
    resources: {
      limits: {
        cpu: "1000m",
        memory: "10Gi",
      },
      requests: {
        cpu: "100m",
        memory: "512Mi",
      },
    },
  },
  [SERVICES.Grafana]: {
    hosts: [],
    data_path: "/sysvol/grafana",
    storage_capacity: "",
    resources: {
      limits: {
        cpu: "1000m",
        memory: "3Gi",
      },
      requests: {
        cpu: "100m",
        memory: "128Mi",
      },
    },
  },
  [SERVICES.Nebula]: {
    hosts: [],
    data_path: "/sysvol/nebula",
    password: "",
    graphd: {
      config: {
        enable_authorize: "true",
        memory_tracker_limitratio: "0.99",
        system_memory_high_watermark_ratio: "0.99",
      },
      resources: null,
    },
    metad: {
      config: {
        enable_authorize: "true",
        memory_tracker_limitratio: "0.99",
        system_memory_high_watermark_ratio: "0.99",
      },
      resources: null,
    },
    storaged: {
      config: {
        enable_authorize: "true",
        memory_tracker_limitratio: "0.99",
        system_memory_high_watermark_ratio: "0.99",
      },
      resources: null,
    },
  },
  [SERVICES.PackageStore]: DEFAULT_STANDARD_PACKAGE_STORE,
  [SERVICES.ECeph]: {
    hosts: [],
    keepalived: {
      internal: "",
      external: "",
    },
    tls: {
      secret: "",
      ["certificate-data"]: "",
      ["key-data"]: "",
    },
  },
  [SERVICES.ProtonMonitor]: {
    hosts: [],
    data_path: "/sysvol/monitor",
    config: {
      grafana: { smtp: { enabled: false } },
    },
    resources: {
      fluentbit: null,
      dcgmExporter: null,
      nodeExporter: null,
      grafana: null,
      vmetrics: null,
      vlogs: null,
      vmagent: null,
    },
  },
  resource_connect_info: {
    [CONNECT_SERVICES.RDS]: DEFAULT_INTERNAL_RDS,
    [CONNECT_SERVICES.MONGODB]: DEFAULT_INTERNAL_MONGODB,
    [CONNECT_SERVICES.REDIS]: DEFAULT_INTERNAL_SERVICE,
    [CONNECT_SERVICES.MQ]: DEFAULT_INTERNAL_MQ,
    [CONNECT_SERVICES.OPENSEARCH]: DEFAULT_INTERNAL_OPENSEARCH,
    [CONNECT_SERVICES.POLICY_ENGINE]: DEFAULT_INTERNAL_SERVICE,
    [CONNECT_SERVICES.ETCD]: DEFAULT_INTERNAL_SERVICE,
  },
};

//其他服务
export const ConfigOtherService = [
  // {
  //   key: SERVICES.ComponentManagement,
  //   serviceKey: SERVICES.ComponentManagement,
  // },
  {
    // 避免客户疑问前端不展示，但是默认会安装
    key: SERVICES.NvidiaDevicePlugin,
    serviceKey: SERVICES.NvidiaDevicePlugin,
  },
];

// 可配置服务list
const ConfigServiceKeys = [
  {
    key: SERVICES.ProtonMariadb,
    serviceKey: SERVICES.ProtonMariadb,
  },
  {
    key: SERVICES.ProtonMongodb,
    serviceKey: SERVICES.ProtonMongodb,
  },
  {
    key: SERVICES.ProtonRedis,
    serviceKey: SERVICES.ProtonRedis,
  },
  {
    key: SERVICES.ProtonNSQ,
    serviceKey: SERVICES.ProtonNSQ,
  },
  {
    key: SERVICES.ProtonPolicyEngine,
    serviceKey: SERVICES.ProtonPolicyEngine,
  },
  {
    key: SERVICES.ProtonEtcd,
    serviceKey: SERVICES.ProtonEtcd,
  },
  {
    key: SERVICES.Opensearch,
    serviceKey: SERVICES.Opensearch,
  },
  {
    key: SERVICES.NvidiaDevicePlugin,
    serviceKey: SERVICES.NvidiaDevicePlugin,
  },
  {
    key: SERVICES.ComponentManagement,
    serviceKey: SERVICES.ComponentManagement,
  },
  {
    key: SERVICES.Kafka,
    serviceKey: SERVICES.Kafka,
  },
  {
    key: SERVICES.Zookeeper,
    serviceKey: SERVICES.Zookeeper,
  },
  {
    key: SERVICES.Prometheus,
    serviceKey: SERVICES.Prometheus,
  },
  {
    key: SERVICES.Grafana,
    serviceKey: SERVICES.Grafana,
  },
  {
    key: SERVICES.Nebula,
    serviceKey: SERVICES.Nebula,
  },
  {
    key: SERVICES.PackageStore,
    serviceKey: SERVICES.PackageStore,
  },
  {
    key: SERVICES.ECeph,
    serviceKey: SERVICES.ECeph,
  },
  {
    key: SERVICES.ProtonMonitor,
    serviceKey: SERVICES.ProtonMonitor,
  },
];

/**
 * 获取当前服务的部署节点
 * @param nodes 节点数
 */
const getDefaultNodes = (
  nodes,
  deployNodes,
  count = 3,
  tail = false,
  showAll = false,
) => {
  if (!deployNodes) {
    return nodes.slice(0, 1);
  }
  if (deployNodes.length === 0) {
    if (showAll) return nodes;
    if (tail) {
      switch (true) {
        // 当节点小于3的时候默认给1个节点
        case nodes.length < 3 && count === 3:
          return nodes.slice(0, 1);

        //面向容器仓库，特殊需要两节点的场景
        case nodes.length <= count && count < 3:
          return nodes;

        // 面向总结数大于要求节点数， 根据要求长度选择要求的节点数
        case nodes.length >= count:
          return nodes.slice(nodes.length - count, nodes.length);
      }
    } else {
      switch (true) {
        // 当节点小于3的时候默认给1个节点
        case nodes.length < 3 && count === 3:
          return nodes.slice(0, 1);

        //面向容器仓库，特殊需要两节点的场景
        case nodes.length <= count && count < 3:
          return nodes;

        // 面向总结数大于要求节点数， 根据要求长度选择要求的节点数
        case nodes.length >= count:
          return nodes.slice(0, count);
      }
    }
  } else {
    let data = [];
    deployNodes.forEach((value) => {
      let deployNode = nodes.find((node) => node.name === value);
      data = deployNode ? [...data, deployNode] : data;
    });

    return data;
  }
};

/**
 * 获取当前服务的部署节点 单个节点
 * @param node 节点数
 */
const getDefaultNode = (node, deployNodes, count = 3) => {
  if (deployNodes.length === 0) {
    return node.slice(0, 1);
  } else {
    let data = [];
    deployNodes.forEach((value) => {
      let deployNode = node.find((node) => node.name === value);
      data = deployNode ? [...data, deployNode] : data;
    });

    return data;
  }
};

/**
 * 将mongo的opotions转为object类型
 * @param data
 * @returns
 */
function transMongoOptions2Object(data) {
  if (!data?.resource_connect_info?.mongodb) {
    return {
      ...data?.resource_connect_info,
    };
  }
  let options = {};
  if (
    isString(data?.resource_connect_info?.mongodb?.options) &&
    data?.resource_connect_info?.mongodb?.options?.indexOf("=") !== -1
  ) {
    options = data?.resource_connect_info?.mongodb?.options
      .split("&")
      .reduce((pre, cur) => {
        const [k, v] = cur.split("=");
        return {
          ...pre,
          [k]: v,
        };
      }, {});
  }
  return {
    ...data?.resource_connect_info,
    mongodb: {
      ...data?.resource_connect_info?.mongodb,
      options,
    },
  };
}

// 将第三方rds的管理账户信息转为admin_key（base64格式）
function formatRDSAdminKey(connectInfo) {
  if (connectInfo?.rds?.source_type !== SOURCE_TYPE.EXTERNAL) {
    return {
      ...connectInfo,
    };
  }
  let data = {
    ...connectInfo,
    rds: {
      ...connectInfo?.rds,
    },
  };
  let admin_key;
  if (data?.rds?.auto_create_database) {
    admin_key = btoa(
      unescape(
        encodeURIComponent(
          `${data?.rds?.admin_user}:${data?.rds?.admin_passwd}`,
        ),
      ),
    );
  }
  delete data?.rds?.auto_create_database;
  delete data?.rds?.admin_user;
  delete data?.rds?.admin_passwd;
  return {
    ...data,
    rds: {
      ...data?.rds,
      admin_key,
    },
  };
}

function getDefaultServiceInfo(configData, dataBaseStorageType) {
  let packageStore;
  if (dataBaseStorageType === DataBaseStorageType.Standard) {
    packageStore = DEFAULT_STANDARD_PACKAGE_STORE;
  } else {
    packageStore = DEFAULT_DEPOSIT_K8S_PACKAGE_STORE;
  }
  return {
    ...configData,
    [SERVICES.PackageStore]: packageStore,
  };
}

function getDefaultServiceInfoByService(service, dataBaseStorageType) {
  if (service === SERVICES.PackageStore) {
    return dataBaseStorageType === DataBaseStorageType.Standard
      ? DEFAULT_STANDARD_PACKAGE_STORE
      : DEFAULT_DEPOSIT_K8S_PACKAGE_STORE;
  } else {
    return DefaultConfigData[service];
  }
}

/**
 * 删除托管Kubernetes部署不需要的项
 * @param initedData 原始数据
 * @param dataBaseStorageType 数据库类型
 * @returns
 */
function deleteDepositKubernetesConfig(initedData, dataBaseStorageType) {
  if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes) {
    let config = {
      ...initedData,
      cs: { provisioner: "external", addons: initedData?.cs?.addons },
    };
    delete config.nodes;
    delete config[SERVICES.ProtonMonitor];
    // delete config.internal_cidr;
    // delete config.internal_nic;
    return config;
  } else {
    return initedData;
  }
}

/**
 * 将托管Kubernetes部署服务的副本数转成number类型
 * @param initedData 原始数据
 * @param dataBaseStorageType 数据库类型
 * @returns
 */
function changeDepositKubernetesReplica(initedData, dataBaseStorageType) {
  if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes) {
    return DEFAULT_REPLICA_SERVICES.reduce(
      (config, service) => {
        if (!config[service]) {
          return config;
        } else {
          return {
            ...config,
            [service]: {
              ...config[service],
              replica_count: Number(config[service].replica_count),
            },
          };
        }
      },
      { ...initedData },
    );
  } else {
    return initedData;
  }
}

const getCS = (config) => {
  return config?.provisioner
    ? {
        provisioner: config?.provisioner,
        addons: config?.addons,
        master: config.master,
        ipFamilies: config.ipFamilies
          ? config.ipFamilies[0] === IP_Family.dualStack
            ? [IP_Family.ipv4, IP_Family.ipv6]
            : config.ipFamilies
          : [IP_Family.ipv4],
        ha_port: 16643,
        host_network: {
          bip: config.hostNetwork.bip,
          pod_network_cidr: config.hostNetwork.podNetworkCidr,
          service_cidr: config.hostNetwork.serviceCidr,
          ipv4_interface: config.hostNetwork.ipv4Interface,
          ipv6_interface: config.hostNetwork.ipv6Interface,
        },
        etcd_data_dir: config.etcdDataDir,
        docker_data_dir: config.dockerDataDir,
        enableDualStack: config.enableDualStack,
      }
    : {
        master: config.master,
        addons: config?.addons,
        ipFamilies: config.ipFamilies
          ? config.ipFamilies[0] === IP_Family.dualStack
            ? [IP_Family.ipv4, IP_Family.ipv6]
            : config.ipFamilies
          : [IP_Family.ipv4],
        ha_port: 16643,
        host_network: {
          bip: config.hostNetwork.bip,
          pod_network_cidr: config.hostNetwork.podNetworkCidr,
          service_cidr: config.hostNetwork.serviceCidr,
          ipv4_interface: config.hostNetwork.ipv4Interface,
          ipv6_interface: config.hostNetwork.ipv6Interface,
        },
        etcd_data_dir: config.etcdDataDir,
        docker_data_dir: config.dockerDataDir,
        enableDualStack: config.enableDualStack,
      };
};

const getCR = (config, crType) => {
  if (crType === CRType.LOCAL) {
    return config;
  } else {
    Object.values(RepositoryType).forEach((type) => {
      if (
        config.external.image_repository !== type &&
        config.external.chart_repository !== type
      ) {
        delete config.external[type];
      }
    });
    return config;
  }
};

/**
 * 清除非外部时间服务器的server参数
 * @param chrony
 * @returns
 */
const getChrony = (chrony, dataBaseStorageType) => {
  if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes)
    return { mode: CHRONY_MODE.USER_MANAGED };
  if (chrony.mode === CHRONY_MODE.EXTERNAL_NTP) {
    return chrony;
  } else {
    return {
      mode: chrony.mode,
    };
  }
};

/**
 * 获取防火墙配置
 * @param firewall
 * @returns
 */
const getFirewall = (firewall, dataBaseStorageType) => {
  if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes)
    return { mode: FIREWALL_MODE.USER_MANAGED };
  return firewall;
};

/**
 * 数据转换
 * @param data
 * @param crType
 * @returns
 */
const exChangeData = (
  data: ConfigData,
  crType: CRType,
  dataBaseStorageType,
) => {
  let connectInfo;
  if (data?.resource_connect_info) {
    connectInfo = transMongoOptions2Object(data);
    connectInfo = formatRDSAdminKey(connectInfo);
  }
  let initedData = {
    apiVersion: "v1",
    // internal_cidr: data.internal_cidr,
    // internal_nic: data.internal_nic,
    nodes: data.nodesInfo.map((value) => ({
      name: value.name,
      ip4: value.ipv4,
      ip6: value.ipv6,
      internal_ip: value.internal_ip,
    })),
    cs: getCS(data.networkInfo),
    cr: getCR(data.cr, crType),
    chrony: getChrony(data.chrony, dataBaseStorageType),
    firewall: getFirewall(data.firewall, dataBaseStorageType),
    deploy: {
      ...data.deploy,
      devicespec: [
        DEVICESPECS.InteliKnowledegeMgntStandard,
        DEVICESPECS.InteliKnowledegeMgntSubscription,
      ].includes(data.deploy.devicespec || DEVICESPECS.AS10000)
        ? errorDeviceSpecsMap[data.deploy.devicespec]
        : data.deploy.devicespec || DEVICESPECS.AS10000,
    },
    [SERVICES.ProtonMariadb]: data[SERVICES.ProtonMariadb],
    [SERVICES.ProtonMongodb]: data[SERVICES.ProtonMongodb],
    [SERVICES.ProtonRedis]: data[SERVICES.ProtonRedis],
    [SERVICES.ProtonNSQ]: data[SERVICES.ProtonNSQ],
    [SERVICES.ProtonPolicyEngine]: data[SERVICES.ProtonPolicyEngine],
    [SERVICES.ProtonEtcd]: data[SERVICES.ProtonEtcd],
    [SERVICES.Opensearch]: data[SERVICES.Opensearch],
    [SERVICES.NvidiaDevicePlugin]: data[SERVICES.NvidiaDevicePlugin]
      ? {}
      : null,
    [SERVICES.ComponentManagement]: data[SERVICES.ComponentManagement],
    [SERVICES.Kafka]: data[SERVICES.Kafka],
    [SERVICES.Nebula]: data[SERVICES.Nebula],
    [SERVICES.Prometheus]: data[SERVICES.Prometheus],
    [SERVICES.Grafana]: data[SERVICES.Grafana],
    [SERVICES.PackageStore]: data[SERVICES.PackageStore],
    [SERVICES.ECeph]: data[SERVICES.ECeph],
    [SERVICES.Zookeeper]: data[SERVICES.Zookeeper],
    [SERVICES.ProtonMonitor]: formatMonitorConfig(data[SERVICES.ProtonMonitor]),
    resource_connect_info: connectInfo,
  };
  initedData = filterEmptyObject(initedData);
  initedData = deleteDepositKubernetesConfig(initedData, dataBaseStorageType);
  initedData = changeDepositKubernetesReplica(initedData, dataBaseStorageType);
  if (crType === CRType.LOCAL) {
    return {
      ...initedData,
      cr: {
        local: {
          hosts: data.cr.local.master,
          ports: {
            chartmuseum: Number(data.cr.local.ports.chartMuseum),
            registry: Number(data.cr.local.ports.registry),
            rpm: Number(data.cr.local.ports.rpm),
            cr_manager: Number(data.cr.local.ports.crManager),
          },
          ha_ports: {
            chartmuseum: Number(data.cr.local.haPorts.chartMuseum),
            registry: Number(data.cr.local.haPorts.registry),
            rpm: Number(data.cr.local.haPorts.rpm),
            cr_manager: Number(data.cr.local.haPorts.crManager),
          },
          storage: data.cr.local.storage,
        },
      },
    };
  } else {
    return initedData;
  }
};

// 过滤空对象
const filterEmptyObject = (data) => {
  let newData = data;

  ConfigServiceKeys.forEach((value) => {
    if (!newData[value.serviceKey]) {
      delete newData[value.serviceKey];
    }
  });
  return newData;
};

// 切换数据库存储方式后，过滤数据库配置中不需要的参数
const filterEmptyKey = (data, dataType) => {
  let newData = data;
  if (dataType === DataBaseStorageType.Standard) {
    delete newData.storageClassName;
    delete newData.replica_count;
  } else {
    delete newData.hosts;
    delete newData.data_path;
  }
  return newData;
};

// 格式化proton monitor数据
const formatMonitorConfig = (monitorConfig: ProtonMonitor) => {
  if (monitorConfig?.config?.vmagent?.remoteWrite?.extraServers) {
    return {
      ...monitorConfig,
      config: {
        ...monitorConfig.config,
        vmagent: {
          ...monitorConfig.config.vmagent,
          remoteWrite: {
            ...monitorConfig.config.vmagent.remoteWrite,
            extraServers:
              monitorConfig?.config?.vmagent?.remoteWrite?.extraServers
                .split(",")
                .map((server) => ({ url: server })),
          },
        },
      },
    };
  } else {
    return monitorConfig;
  }
};

// CR 类型
enum CRType {
  // 本地
  LOCAL,

  //外置
  ExternalCRConfig,
}

/**
 * 数据存储类型
 */
export const DataBaseStorageList = [
  {
    type: DataBaseStorageType.Standard,
    title: DataBaseStorageTypeName[DataBaseStorageType.Standard],
    content: "使用一体机、物理机、虚拟机进行部署，请选择此模版",
    tips: "使用内置爱数 Kubernetes 部署",
    style: {
      backgroundImage: `url(${standard})`,
      backgroundPosition: "0 -5px",
      backgroundRepeat: "no-repeat",
      backgroundSize: "cover",
    },
  },
  {
    type: DataBaseStorageType.Cloud,
    title: DataBaseStorageTypeName[DataBaseStorageType.Cloud],
    tips: "使用云主机部署",
    content: "使用云主机（ECS） 进行部署，请选择此模版",
  },
  {
    type: DataBaseStorageType.DepositKubernetes,
    title: DataBaseStorageTypeName[DataBaseStorageType.DepositKubernetes],
    tips: "使用托管Kubernetes部署",
    content: "使用托管 Kubernetes 进行部署，请选择此模版",
  },
];

const notMultipleDateBase = [];

const initialAllNodesServices = [SERVICES.ECeph];

const DEGAULT_CONNECT_SERVICE = [
  CONNECT_SERVICES.RDS,
  CONNECT_SERVICES.MONGODB,
  CONNECT_SERVICES.REDIS,
  CONNECT_SERVICES.MQ,
  CONNECT_SERVICES.OPENSEARCH,
  CONNECT_SERVICES.POLICY_ENGINE,
  CONNECT_SERVICES.ETCD,
];

const DefaultSelectableServices = [
  {
    key: SERVICES.ProtonMariadb,
    name: "Proton MariaDB",
  },
  {
    key: SERVICES.ProtonMongodb,
    name: "Proton MongoDB",
  },
  {
    key: SERVICES.ProtonRedis,
    name: "Proton Redis",
  },
  {
    key: SERVICES.ProtonNSQ,
    name: "Proton MQ NSQ",
  },
  {
    key: SERVICES.ProtonPolicyEngine,
    name: "Proton Policy Engine",
  },
  {
    key: SERVICES.ProtonEtcd,
    name: "Proton ETCD",
  },
  {
    key: SERVICES.Opensearch,
    name: "Opensearch",
  },
  {
    key: SERVICES.NvidiaDevicePlugin,
    name: "NvidiaDevicePlugin",
  },
  {
    key: SERVICES.ComponentManagement,
    name: "ComponentManagement",
  },
  {
    key: SERVICES.Kafka,
    name: "Kafka",
  },
  {
    key: SERVICES.Zookeeper,
    name: "ZooKeeper",
  },
  {
    key: SERVICES.ProtonMonitor,
    name: "Proton Monitor",
  },
];

const DefalutAddableServices = [
  {
    key: SERVICES.Prometheus,
    name: "prometheus",
  },
  {
    key: SERVICES.Grafana,
    name: "grafana",
  },
  {
    key: SERVICES.PackageStore,
    name: "Package Store",
  },
  {
    key: SERVICES.ECeph,
    name: "ECeph",
  },
  {
    key: SERVICES.Nebula,
    name: "nebula",
  },
];

const DefaultDataBaseForm = {
  [SERVICES.ProtonMariadb]: null,
  [SERVICES.ProtonMongodb]: null,
  [SERVICES.ProtonRedis]: null,
  [SERVICES.ProtonNSQ]: null,
  [SERVICES.ProtonPolicyEngine]: null,
  [SERVICES.ProtonEtcd]: null,
  [SERVICES.Opensearch]: null,
  [SERVICES.Kafka]: null,
  [SERVICES.Zookeeper]: null,
  [SERVICES.Prometheus]: null,
  [SERVICES.Grafana]: null,
  [SERVICES.Nebula]: null,
  [SERVICES.PackageStore]: null,
  [SERVICES.ECeph]: null,
  [SERVICES.ProtonMonitor]: null,
};

const DefaultConnectInfoForm = {
  [CONNECT_SERVICES.RDS]: null,
  [CONNECT_SERVICES.MONGODB]: null,
  [CONNECT_SERVICES.REDIS]: null,
  [CONNECT_SERVICES.MQ]: null,
  [CONNECT_SERVICES.OPENSEARCH]: null,
  [CONNECT_SERVICES.POLICY_ENGINE]: null,
  [CONNECT_SERVICES.ETCD]: null,
};

// 端口数字验证规则
const portValidator = (rule, value: string) => {
  return new Promise<void>((resolve, reject) => {
    const regExp = /^[1-9][\d]{0,4}$/;
    if (value === undefined || value === null || value === "") {
      resolve();
    } else if (regExp.test(value) && Number(value) <= 65535) {
      resolve();
    } else {
      reject();
    }
  });
};

// 虚拟地址验证规则
const vipValidator = (rule, value: string) => {
  return new Promise<void>((resolve, reject) => {
    const regExp = /^[\w\.\:]+\/([1-9]|[1-5]\d|6[0-4])$/;
    if (value === undefined || value === null || value === "") {
      resolve();
    } else if (regExp.test(value)) {
      resolve();
    } else {
      reject();
    }
  });
};

// 副本数验证规则
const replicaValidator = (rule, value: string) => {
  return new Promise<void>((resolve, reject) => {
    const regExp = /^[1-9][\d]{0,1}$/;
    if (value === undefined || value === null || value === "") {
      resolve();
    } else if (regExp.test(value)) {
      resolve();
    } else {
      reject();
    }
  });
};

// 本地RDS和MongoDB用户名验证规则
const usernameValidator = (rule, value: string) => {
  return new Promise<void>((resolve, reject) => {
    if (value === "root") {
      reject();
    } else {
      resolve();
    }
  });
};

// 第三方RDS账户信息验证规则
const rdsUserInfoValidator = (value: string, comparisonValue: string) => {
  return new Promise<void>((resolve, reject) => {
    if (value !== comparisonValue && comparisonValue && value) {
      reject();
    } else {
      resolve();
    }
  });
};

// 布尔值验证规则
const booleanValidator = (rule, value: boolean) => {
  return new Promise<void>((resolve, reject) => {
    if ([true, false].includes(value)) {
      resolve();
    } else {
      reject();
    }
  });
};

// kafka日志配置验证规则
const logNumberValidator = (value: string, flag: boolean) => {
  return new Promise<void>((resolve, reject) => {
    if (flag && value === "-1") resolve();
    const regExp = /^[1-9][\d]*$/;
    if ([undefined, null, "", "0"].includes(value)) {
      resolve();
    } else if (regExp.test(value)) {
      resolve();
    } else {
      reject();
    }
  });
};

// 空值校验规则
const emptyValidatorRules = [
  {
    required: true,
    message: "此项不允许为空。",
  },
];

// 布尔空值校验规则
const booleanEmptyValidatorRules = [
  {
    validator: booleanValidator,
    message: "此项不允许为空。",
  },
];

// kafka日志配置验校验规则
const getLogNumberValidatorRules = (flag: boolean) => [
  {
    validator: (rule, value: string) => logNumberValidator(value, flag),
    message: "请输入非负整数。",
  },
];

// 端口校验规则
const portValidatorRules = [
  {
    required: true,
    message: "此项不允许为空。",
  },
  {
    validator: portValidator,
    message: "请输入1~65535范围内的整数。",
  },
];

// 虚拟地址校验规则
const vipValidatorRules = [
  {
    validator: vipValidator,
    message:
      "虚拟地址格式错误，请重新输入。虚拟地址需包含有效的IP地址和子网掩码。",
  },
];

// 副本数校验规则
const replicaValidatorRules = [
  {
    required: true,
    message: "此项不允许为空。",
  },
  {
    validator: replicaValidator,
    message: "请输入1~99范围内的整数。",
  },
];

// 本地RDS和MongoDB连接配置用户名校验规则
const getUsernameValidatorRules = (sourceType: string) => {
  return sourceType === SOURCE_TYPE.INTERNAL
    ? [
        {
          required: true,
          message: "此项不允许为空。",
        },
        {
          validator: usernameValidator,
          message: "此项不允许为root。",
        },
      ]
    : emptyValidatorRules;
};

// 第三方RDS连接配置账户信息校验规则
const getRDSUserInfoValidatorRules = (
  rdsType: string,
  comparisonValue: string,
) => {
  return rdsType === RDS_TYPE.DM8
    ? [
        {
          required: true,
          message: "此项不允许为空。",
        },
        {
          validator: (rule, value: string) =>
            rdsUserInfoValidator(value, comparisonValue),
          message: "达梦数据库的管理权限和普通账户信息必须相同。",
        },
      ]
    : emptyValidatorRules;
};

const NodeForm = {
  ServerForm: "serverForm",
  AccountInfoForm: "accountInfoForm",
};

// 校验状态
const enum ValidateState {
  // 正常
  Normal,
  // 输入为空
  Empty,
  // 节点数量错误
  NodesNumError,
}

const DefaultConnectInfoValidateState = {
  RDS_TYPE: ValidateState.Normal,
  MONGODB_SSL: ValidateState.Normal,
  REDIS_CONNECT_TYPE: ValidateState.Normal,
  MQ_RADIO: ValidateState.Normal,
  MQ_TYPE: ValidateState.Normal,
  MQ_AUTH_MACHANISM: ValidateState.Normal,
  OPENSEARCH_VERSION: ValidateState.Normal,
};

// 包含副本数配置的服务（packageStore副本数的实现逻辑保持不变）
const DEFAULT_REPLICA_SERVICES = [
  // SERVICES.PackageStore,
  SERVICES.ProtonMariadb,
  SERVICES.ProtonMongodb,
  SERVICES.ProtonNSQ,
  SERVICES.Opensearch,
  SERVICES.ProtonPolicyEngine,
  SERVICES.ProtonEtcd,
  SERVICES.ProtonRedis,
  SERVICES.Zookeeper,
  SERVICES.Kafka,
];

// JVM配置key
const JVMConfigKeys = {
  [SERVICES.Kafka]: "KAFKA_HEAP_OPTS",
  [SERVICES.Zookeeper]: "JVMFLAGS",
};

// 第三方容器仓库类型（镜像仓库和chart仓库）
const RepositoryType = {
  Registry: "registry",
  Chartmuseum: "chartmuseum",
  OCI: "oci",
};

const DefaultMonitorResourcesConfig = {
  fluentbit: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "10m",
      [RESOURCES_TYPE.MEMORY]: "30Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1",
      [RESOURCES_TYPE.MEMORY]: "500Mi",
    },
  },
  dcgmExporter: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "10m",
      [RESOURCES_TYPE.MEMORY]: "30Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1",
      [RESOURCES_TYPE.MEMORY]: "500Mi",
    },
  },
  nodeExporter: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "10m",
      [RESOURCES_TYPE.MEMORY]: "30Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1",
      [RESOURCES_TYPE.MEMORY]: "500Mi",
    },
  },
  grafana: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "100m",
      [RESOURCES_TYPE.MEMORY]: "100Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1",
      [RESOURCES_TYPE.MEMORY]: "2Gi",
    },
  },
  vmetrics: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "500m",
      [RESOURCES_TYPE.MEMORY]: "500Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1000m",
      [RESOURCES_TYPE.MEMORY]: "2Gi",
    },
  },
  vlogs: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "500m",
      [RESOURCES_TYPE.MEMORY]: "100Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1000m",
      [RESOURCES_TYPE.MEMORY]: "1Gi",
    },
  },
  vmagent: {
    [RESOURCES.REQUESTS]: {
      [RESOURCES_TYPE.CPU]: "100m",
      [RESOURCES_TYPE.MEMORY]: "100Mi",
    },
    [RESOURCES.LIMITS]: {
      [RESOURCES_TYPE.CPU]: "1000m",
      [RESOURCES_TYPE.MEMORY]: "1Gi",
    },
  },
};

export {
  checkIPv6,
  checkIPv4,
  validatorMessge,
  OptionSteps,
  DefaultConfigData,
  getDefaultNodes,
  getDefaultNode,
  exChangeData,
  filterEmptyObject,
  filterEmptyKey,
  getDefaultServiceInfo,
  getDefaultServiceInfoByService,
  nodeInfoCheckResult,
  CRType,
  ConfigEditStatus,
  ConfigServiceKeys,
  notMultipleDateBase,
  DefaultSelectableServices,
  DefalutAddableServices,
  IP_Family_LIST,
  IP_Family,
  DEGAULT_EXTERNAL_RDS,
  DEFAULT_INTERNAL_RDS,
  DEFAULT_EXTERNAL_MONGODB,
  DEFAULT_INTERNAL_MONGODB,
  DEFAULT_EXTERNAL_REDIS,
  DEFAULT_EXTERNAL_MQ,
  DEFAULT_INTERNAL_MQ,
  DEFAULT_EXTERNAL_OPENSEARCH,
  DEFAULT_INTERNAL_OPENSEARCH,
  DEFAULT_EXTERNAL_SERVICE,
  DEFAULT_INTERNAL_SERVICE,
  DEGAULT_CONNECT_SERVICE,
  emptyValidatorRules,
  portValidatorRules,
  replicaValidatorRules,
  NodeForm,
  ValidateState,
  DefaultDataBaseForm,
  DefaultConnectInfoForm,
  DefaultConnectInfoValidateState,
  initialAllNodesServices,
  vipValidatorRules,
  DEFAULT_REPLICA_SERVICES,
  JVMConfigKeys,
  getUsernameValidatorRules,
  booleanEmptyValidatorRules,
  RepositoryType,
  getRDSUserInfoValidatorRules,
  getLogNumberValidatorRules,
  DefaultMonitorResourcesConfig,
  checkNodeName,
};

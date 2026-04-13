import { FormInstance } from "@aishutech/ui";
import { OptionSteps, ConfigEditStatus, ValidateState } from "./helper";

declare namespace DeployToolsBaseType {
  interface Props {}
  interface State {
    // 数据库类型
    dataBaseStorageType: null | string;

    nextStepButtonDisable: boolean;

    // 控制配置编辑步骤状态
    stepStatus: OptionSteps;

    // 配置数据
    configData: ConfigData;

    // 节点账户
    sshAccount: string;

    // 节点密码
    sshPassword: string;

    // 配置编辑状态
    configEditStatus: ConfigEditStatus;

    // 可选服务集合
    selectableServices: Array<{
      key: string;
      name: string;
    }>;

    // 可添加服务集合
    addableServices: Array<{
      key: string;
      name: string;
    }>;

    // 节点配置form实例
    nodeForm: {
      serverForm: React.RefObject<FormInstance<any>>;
      accountInfoForm: React.RefObject<FormInstance<any>>;
    };

    // 网络配置form实例
    networkForm: {
      hostNetworkForm: React.RefObject<FormInstance<any>>;
      networkInfoForm: React.RefObject<FormInstance<any>>;
    };

    // 仓库配置form实例
    crForm: {
      localForm: {
        portsForm: React.RefObject<FormInstance<any>>;
        haPortsForm: React.RefObject<FormInstance<any>>;
        storageForm: React.RefObject<FormInstance<any>>;
      };
      externalForm: {
        chartmuseumForm: React.RefObject<FormInstance<any>>;
        registryForm: React.RefObject<FormInstance<any>>;
        ociForm: React.RefObject<FormInstance<any>>;
      };
    };

    // 基础服务配置form实例
    dataBaseForm: DefaultDataBaseFormRef;

    //连接配置form实例
    connectInfoForm: DefaultConnectInfoFormRef;

    // 节点校验状态
    nodesValidateState: ValidateState;

    // 网络配置部署节点校验状态
    networkNodesValidateState: ValidateState;

    // 仓库配置节点校验状态
    crNodesValidateState: ValidateState;

    // 基础服务配置grafana服务节点校验状态
    grafanaNodesValidateState: ValidateState;

    // 基础服务配置prometheus服务节点校验状态
    prometheusNodesValidateState: ValidateState;

    // 基础服务配置proton monitor服务节点校验状态
    monitorNodesValidateState: ValidateState;

    // 连接配置校验状态
    connectInfoValidateState: ConnectInfoValidateState;
  }
}

export interface Resource {
  limits: {
    cpu: string;
    memory: string;
  };
  requests: {
    cpu: string;
    memory: string;
  };
}

export interface Services {
  // proton-mariaDB
  proton_mariadb?: ProtonMariadb;

  // proton-MongoDB
  proton_mongodb: ProtonDataBase;

  // proton-Redis
  proton_redis: ProtonDataBase;

  // 消息队列
  proton_mq_nsq: ProtonServiceConfig;

  // 策略引擎
  proton_policy_engine: ProtonServiceConfig;

  //Etcd
  proton_etcd: ProtonServiceConfig;

  // opensearch
  opensearch: Opensearch;

  // kafka
  kafka: ProtonServiceConfig;

  //zookeeper
  zookeeper: ProtonServiceConfig;

  // prometheus
  prometheus: ProtonServiceConfig;

  // grafana
  grafana: ProtonServiceConfig;

  // nebula
  nebula: ProtonNebula;

  // package-store
  "package-store": PackageStoreConfig;

  // eceph
  eceph: ProtonECeph;

  // proton-Monitor
  proton_monitor: ProtonMonitor;

  // nvidia_device_plugin
  nvidia_device_plugin: {} | null;

  //component-management
  component_management: {} | null;

  // 外置连接信息
  resource_connect_info?: {
    rds: RDSConnectInfo;
    mongodb: MongoDBConnectInfo;
    redis: RedisConnectInfo;
    mq: MQConnectInfo;
    opensearch: OpensearchConnectInfo;
    policy_engine: PolicyEngineConnectInfo;
    etcd: ETCDConnectInfo;
  };
}

export interface ServicesNodes {
  // RDS 节点
  mariaDBNodes: Array<NodeInfo>;

  //MongoDB 节点
  mongoDBNodes: Array<NodeInfo>;

  // Redis 节点
  redisNodes: Array<NodeInfo>;

  // nsq 节点
  nsqNodes: Array<NodeInfo>;

  //kafka节点
  kafkaNodes: Array<NodeInfo>;

  // zookeeper节点
  zookeeperNodes: Array<NodeInfo>;

  //  prometheus 节点
  prometheusNodes: Array<NodeInfo>;

  //  nebula 节点
  nebulaNodes: Array<NodeInfo>;

  //  grafana 节点
  grafanaNodes: Array<NodeInfo>;

  // 策略引擎
  policyEngineNodes: Array<NodeInfo>;

  // etcd 节点
  etcdNodes: Array<NodeInfo>;

  // opensearch 节点
  opensearchNodes: Array<string>;

  // eceph 节点
  ecephNodes: Array<NodeInfo>;
}

export type NodeInfo = {
  //节点名称
  name: string;

  // IPv6地址
  ipv6: string;

  // IPv4 地址
  ipv4: string;

  // 内部ip
  internal_ip: string;
};

export type ChronyConfig = {
  // 模式
  mode: string;
  // 外部时间服务器时才会配置该字段
  server?: Array<string>;
};

export type FirewallConfig = {
  // 模式
  mode: string;
};

export type DeployConfig = {
  // 部署模式
  mode: string;
  // 产品型号
  devicespec?: string;
  // 命名空间
  namespace?: string;
  // 服务账号
  serviceaccount?: string;
};

// 部署配置
export interface ConfigData extends Services {
  // 时间同步服务器配置
  chrony: ChronyConfig;

  // 防火墙配置
  firewall: FirewallConfig;

  // 部署配置
  deploy: DeployConfig;

  // 集群节点
  nodesInfo: Array<NodeInfo>;

  // cs 集群网络配置
  networkInfo: NetworkInfo;

  // // 内部网段
  // internal_cidr: string;

  // // 网卡
  // internal_nic: string;

  // CR 配置
  cr: {
    // 本地CR
    local?: CRConfig;

    // 外置CR
    external?: ExternalCRConfig;
  };
}

//网路配置信息
export type NetworkInfo = {
  //网络节点
  master?: Array<string>;

  // 网段配置
  hostNetwork?: HostNetwork;

  // etcd 数据路径
  etcdDataDir?: string;

  //docker数据路径
  dockerDataDir?: string;

  // 网络类型
  provisioner?: string;

  // 插件列表
  addons?: ReadonlyArray<string>;

  // k8s IP协议栈
  ipFamilies: Array<string>;

  // 是否开启双栈能力
  enableDualStack?: boolean;
};

//网络配置
export type HostNetwork = {
  // Docker 网段
  bip: string;

  // Pod 网段
  podNetworkCidr: string;

  // Service 网段
  serviceCidr: string;

  // 网卡IPv4
  ipv4Interface: string;

  // 网卡IPv6
  ipv6Interface: string;
};

// CR 端口
export type CRPorts = {
  // chart 仓库端口
  chartMuseum: number;

  // 容器仓库端口
  registry: number;

  // rpm 仓库端口
  rpm: number;

  // 容器仓库管理端口
  crManager: number;
};

// CR 配置项
export type CRConfig = {
  // 部署节点
  master: Array<string>;

  // 端口
  ports: CRPorts;

  // 高可用端口
  haPorts: CRPorts;

  // 数据路径
  storage: string;
};

// MariaDB 部署配置
export type ProtonMariadb = {
  // 部署节点
  hosts?: Array<string>;

  // 部署配置
  config: {
    //缓存池配置
    innodb_buffer_pool_size: string;

    // 启动内存
    resource_requests_memory: string;

    // 资源内存上限
    resource_limits_memory: string;
  };

  // 管理员账户
  admin_user?: string;

  // 密码
  admin_passwd?: string;

  //数据路径
  data_path?: string;

  //存储类名
  storageClassName?: string;

  // 副本数
  replica_count?: number;

  // 存储卷容量
  storage_capacity?: string;
};

// MariaDB 部署配置
export type ProtonNebula = {
  // 部署节点
  hosts: Array<string>;

  // 密码
  password: string;

  //数据路径
  data_path: string;

  // graphd 的资源配额
  graphd: { config: ProtonNebulaConfig; resources: Resource };

  // metad 的资源配额
  metad: { config: ProtonNebulaConfig; resources: Resource };

  // storaged 的资源配额
  storaged: { config: ProtonNebulaConfig; resources: Resource };
};

export type ProtonNebulaConfig = {
  enable_authorize: string;
  memory_tracker_limitratio: string;
  system_memory_high_watermark_ratio: string;
};

export type ProtonECeph = {
  // 部署节点
  hosts: Array<string>;

  keepalived?: {
    // 内部虚拟ip，包含掩码
    internal?: string;

    // 外部虚拟ip，包含掩码
    external?: string;
  };

  // 数字证书信息
  tls?: {
    // 数字证书所在的secret名称
    secret?: string;

    // 数字证书，base64编码
    ["certificate-data"]?: string;

    // 数字证书的密钥，base64编码
    ["key-data"]?: string;
  };
};

export type RDSConnectInfo = {
  // 通用属性
  source_type: string;
  username: string;
  password: string;
  // 第三方RDS
  rds_type?: string;
  hosts?: string;
  port?: number;
  // replicaset?: string;
  // 是否自动化创建数据库
  auto_create_database?: boolean;
  // 用户名（管理权限）
  admin_user?: string;
  // 密码（管理权限）
  admin_passwd?: string;
};

export type MongoDBConnectInfo = {
  // 通用属性
  source_type: string;
  username: string;
  password: string;
  // 第三方 mongodb 才有
  hosts?: string;
  port?: number;
  ssl?: boolean;
  replica_set?: string;
  auth_source?: string;
  options?: {
    // 其他可选
    [key: string]: string;
  };
};

export type RedisConnectInfo = {
  // 通用属性
  source_type: string;
  connect_type?: string;
  username?: string;
  password?: string;
  // proton redis 哨兵 & 第三方哨兵多余的属性
  sentinel_username?: string;
  sentinel_password?: string;
  // 第三方哨兵多余的属性
  sentinel_hosts?: string;
  sentinel_port?: number;
  master_group_name?: string;
  // 第三方主从多余的属性
  master_hosts?: string;
  master_port?: number;
  slave_hosts?: string;
  slave_port?: number;
  // 第三方单机多余的属性
  hosts?: string;
  port?: string;
};

export type MQConnectInfo = {
  source_type: string; // 当前只有外置
  mq_type?: string;
  mq_hosts?: string;
  mq_port?: number;
  mq_lookupd_hosts?: string;
  mq_lookupd_port?: number;
  auth?: {
    username: string;
    password: string;
    mechanism: string;
  };
};

export type OpensearchConnectInfo = {
  source_type: string;
  hosts?: string;
  port?: number;
  username?: string;
  password?: string;
  version: string;
  distribution: string;
};

export type PolicyEngineConnectInfo = {
  source_type: string; // 当前只有外置
  hosts?: string;
  port?: number;
};

export type ETCDConnectInfo = {
  source_type: string; // 当前只有外置
  hosts?: string;
  port?: number;
  secret?: string;
};

// Proton-MongoDB 配置
export type ProtonDataBase = {
  // 部署节点
  hosts?: Array<string>;

  // 管理员账户
  admin_user: string;

  // 密码
  admin_passwd: string;

  //数据路径
  data_path?: string;

  //存储类名
  storageClassName?: string;

  // 副本数
  replica_count?: number;

  // 存储卷容量
  storage_capacity?: string;

  // 资源配置
  resources?: {
    requests: {
      cpu: string;
      memory: string;
    };
  };
};

export type ProtonServiceConfig = {
  // 部署节点
  hosts?: Array<string>;

  //数据路径
  data_path?: string;

  env?: {
    // kafka jvm配置
    KAFKA_HEAP_OPTS?: string;
    // kafka 日志保留字节数
    KAFKA_LOG_RETENTION_BYTES?: string;
    // kafka 日志保留小时数
    KAFKA_LOG_RETENTION_HOURS?: string;
    // kafka 日志段最大小时数
    KAFKA_LOG_ROLL_HOURS?: string;
    // zookeeper jvm配置
    JVMFLAGS?: string;
  };

  //存储类名
  storageClassName?: string;

  // kafka 禁用开放外部端口
  disable_external_service?: boolean;

  // 外部端口信息列表
  external_service_list?: ExternalServiceConfig[];

  // 资源配置
  resources?: {
    limits?: {
      cpu: string;
      memory: string;
    };
    requests?: {
      cpu: string;
      memory: string;
    };
  };

  // exporter资源配置
  exporter_resources?: {
    requests: {
      cpu: string;
      memory: string;
    };
  };

  // 副本数
  replica_count?: number;

  // 存储卷容量
  storage_capacity?: string;
};

export type ExternalServiceConfig = {
  // 访问名称
  name: string;
  // 地址
  ip?: string;
  // 外部访问端口
  port: number;
  // 外部访问nodePortBase端口
  nodePortBase: number;
  // 端口是否开启tls验证
  enableSSL: boolean;
};

export type Opensearch = {
  // 部署节点
  hosts?: Array<string>;

  // 部署模式
  mode: string;

  // 索引配置
  config: {
    // JVM
    jvmOptions: string;

    // 索引库
    hanlpRemoteextDict: string;

    // 去停词
    hanlpRemoteextStopwords: string;
  };

  //数据路径
  data_path?: string;

  extraValues: {
    // NFS快照仓库配置
    storage: {
      repo: {
        nfs: {
          enabled: boolean;
          server?: string;
          path?: string;
        };
        hdfs: {
          enabled: boolean;
        };
      };
    };
  };

  //存储类名
  storageClassName?: string;

  // 设置
  settings:
    | {
        // 低警戒水位线
        "cluster.routing.allocation.disk.watermark.low": string;
        // 高警戒水位线
        "cluster.routing.allocation.disk.watermark.high": string;
        // 洪泛警戒水位线
        "cluster.routing.allocation.disk.watermark.flood_stage": string;
        // 内存锁定
        "bootstrap.memory_lock": boolean;
      }
    | {};

  // 资源配置
  resources?: {
    limits: {
      cpu: string;
      memory: string;
    };
    requests: {
      cpu: string;
      memory: string;
    };
  };

  // exporter资源配置
  exporter_resources?: {
    requests: {
      cpu: string;
      memory: string;
    };
  };

  // 副本数
  replica_count?: number;

  // 存储卷容量
  storage_capacity?: string;
};

export type ResourceType = {
  limits: {
    cpu: string;
    memory: string;
  };
  requests: {
    cpu: string;
    memory: string;
  };
};

export type ProtonMonitor = {
  // 部署节点
  hosts?: Array<string>;

  // 配置
  config?: {
    // vmagent
    vmagent?: {
      remoteWrite?: { extraServers?: any };
    };

    // vmetrics
    vmetrics?: {
      retention?: string;
    };

    // vlogs
    vlogs?: {
      retention?: string;
    };

    // grafana
    grafana?: {
      smtp?: {
        host: string;
        user: string;
        password: string;
        skip_verify: boolean;
        from: string;
        from_name: string;
        startTLS_policy: string;
        enable_tracing: boolean;
      };
    };
  };

  //数据路径
  data_path?: string;

  resources?: {
    fluentbit: ResourceType | null;
    dcgmExporter: ResourceType | null;
    nodeExporter: ResourceType | null;
    grafana: ResourceType | null;
    vmetrics: ResourceType | null;
    vlogs: ResourceType | null;
    vmagent: ResourceType | null;
  };
};

export type ExternalCRConfig = {
  // 镜像仓库
  image_repository: string;
  // chart仓库
  chart_repository: string;
  // 仓库信息
  registry?: {
    // 地址
    host: string;

    // 账户名
    username: string;

    // 密码
    password: string;
  };

  // chart 信息
  chartmuseum?: {
    // 地址
    host: string;

    // 账户名
    username: string;

    // 密码
    password: string;
  };

  // oci 信息
  oci?: {
    // 地址
    registry: string;

    // 账户名
    username: string;

    // 密码
    password: string;

    // 是否使用http
    plain_http: boolean;
  };
};

export type PackageStoreConfig = {
  // 包管理服务运行在指定节点，持久化数据也在指定节点
  hosts: Array<string>;

  storage: {
    // 持久化存储的容量。
    capacity: string;

    // 数据目录的路径，应该为绝对路径
    path?: string;

    // 包管理服务使用指定的 storage class 作为持久化存储
    storageClassName?: string;
  };

  // 包管理服务的副本数
  replicas: number;

  // 包管理服务的资源配额
  resources: null | {
    limits: {
      cpu: string;
      memory: string;
    };
  };
};

type DataBaseFormRef =
  | React.RefObject<FormInstance<any>>
  | { [key: string]: React.RefObject<FormInstance<any>> };

export interface DefaultDataBaseFormRef {
  // proton-mariaDB
  proton_mariadb: DataBaseFormRef;

  // proton-MongoDB
  proton_mongodb: DataBaseFormRef;

  // proton-Redis
  proton_redis: DataBaseFormRef;

  // 消息队列
  proton_mq_nsq: DataBaseFormRef;

  // 策略引擎
  proton_policy_engine: DataBaseFormRef;

  //Etcd
  proton_etcd: DataBaseFormRef;

  // opensearch
  opensearch: DataBaseFormRef;

  // kafka
  kafka: DataBaseFormRef;

  //zookeeper
  zookeeper: DataBaseFormRef;

  // prometheus
  prometheus: DataBaseFormRef;

  // grafana
  grafana: DataBaseFormRef;

  // nebula
  nebula: DataBaseFormRef;

  // package-store

  "package-store": DataBaseFormRef;

  // eceph
  eceph: DataBaseFormRef;
}

export interface DefaultConnectInfoFormRef {
  // rds
  rds: React.RefObject<FormInstance<any>>;

  // mongoDB
  mongodb: React.RefObject<FormInstance<any>>;

  // redis
  redis: React.RefObject<FormInstance<any>>;

  // 消息队列
  mq: React.RefObject<FormInstance<any>>;

  // 策略引擎
  policy_engine: React.RefObject<FormInstance<any>>;

  // etcd
  etcd: React.RefObject<FormInstance<any>>;

  // opensearch
  opensearch: React.RefObject<FormInstance<any>>;
}

export interface ConnectInfoValidateState {
  // rds类型校验状态
  RDS_TYPE: ValidateState;

  // mongodb ssl校验状态
  MONGODB_SSL: ValidateState;

  // redis连接模式校验状态
  REDIS_CONNECT_TYPE: ValidateState;

  // mq资源类型校验状态
  MQ_RADIO: ValidateState;

  // mq类型校验状态
  MQ_TYPE: ValidateState;

  // mq认证机制类型状态
  MQ_AUTH_MACHANISM: ValidateState;

  // opensearch版本校验状态
  OPENSEARCH_VERSION: ValidateState;
}

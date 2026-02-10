export interface Services {
    // proton-mariaDB
    mariadb: ProtonMariadb;

    // proton-MongoDB
    mongodb: ProtonDataBase;

    // proton-Redis
    redis: ProtonDataBase;

    // 策略引擎
    policyengine: ProtonServiceConfig;

    //Etcd
    etcd: ProtonServiceConfig;

    // opensearch
    opensearch: Opensearch;

    // kafka
    kafka: ProtonServiceConfig;

    //zookeeper
    zookeeper: ProtonServiceConfig;

    // nebula
    nebula: ProtonNebula;
}

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

    // 连接账户
    username: string;

    // 连接密码
    password: string;
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
    storageClassName?: string;
};

export type ProtonNebulaConfig = {
    enable_authorize: string;
    memory_tracker_limitratio: string;
    system_memory_high_watermark_ratio: string;
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

    // 连接账户
    username?: string;

    // 连接密码
    password?: string;
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

    // kafka外部端口信息列表
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
                hdfs: { enabled: boolean };
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

type ComponentDataUnion = ProtonMariadb &
    ProtonDataBase &
    ProtonServiceConfig &
    Opensearch &
    ProtonNebula;

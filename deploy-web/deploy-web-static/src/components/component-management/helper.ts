// import { manageLog, ManagementOps, Level } from "mediator";
import {
    RDSConnectInfo,
    MongoDBConnectInfo,
    RedisConnectInfo,
    MQConnectInfo,
    OpensearchConnectInfo,
    PolicyEngineConnectInfo,
    ETCDConnectInfo,
    ConnectInfoUnion,
} from "../connect-info-management/index.d";
import { isObject } from "lodash";
import __ from "./locale";
import { ComponentDataUnion } from "./index.d";
// import {
//     ConnectInfoServices,
//     servicesText,
// } from "../connect-info-management/helper";

export const enum SERVICES {
    Kafka = "kafka",
    MariaDB = "mariadb",
    MongoDB = "mongodb",
    Redis = "redis",
    Zookeeper = "zookeeper",
    OpenSearch = "opensearch",
    PolicyEngine = "policyengine",
    ETCD = "etcd",
    Nebula = "nebula",
    ProtonNSQ = "nsq",
}

// 组件依赖映射
export const componentDependenciesMap = {
    [SERVICES.Kafka]: SERVICES.Zookeeper,
    [SERVICES.PolicyEngine]: SERVICES.ETCD,
};

export interface ComponentData {
    name?: string;
    type?: string;
    version?: string;
    params?: any;
    info?: object;
    // kafka组件依赖的zookeeper组件名称
    dependencies?: {
        zookeeper: string;
    };
}

export interface ConfigData {
    [SERVICES.MariaDB]?: ComponentData | null;
    [SERVICES.MongoDB]?: ComponentData | null;
    [SERVICES.Redis]?: ComponentData | null;
    [SERVICES.PolicyEngine]?: ComponentData | null;
    [SERVICES.ETCD]?: ComponentData | null;
    [SERVICES.OpenSearch]?: ComponentData | null;
    [SERVICES.Kafka]?: ComponentData | null;
    [SERVICES.Zookeeper]?: ComponentData | null;
    [SERVICES.Nebula]?: ComponentData | null;
}

export const buildInComponents = [
    SERVICES.MariaDB,
    SERVICES.MongoDB,
    SERVICES.Redis,
    SERVICES.PolicyEngine,
    SERVICES.ETCD,
    SERVICES.OpenSearch,
    SERVICES.Kafka,
    SERVICES.Zookeeper,
    SERVICES.Nebula,
];

export const buildInComponentsText = {
    [SERVICES.MariaDB]: "MariaDB",
    [SERVICES.MongoDB]: "MongoDB",
    [SERVICES.Redis]: "Redis",
    [SERVICES.PolicyEngine]: "PolicyEngine",
    [SERVICES.ETCD]: "ETCD",
    [SERVICES.OpenSearch]: "OpenSearch",
    [SERVICES.Kafka]: "Kafka",
    [SERVICES.Nebula]: "Nebula",
    [SERVICES.Zookeeper]: "Zookeeper",
};

export const ComponentName = {
    [SERVICES.MariaDB]: "mariadb",
    [SERVICES.MongoDB]: "mongodb",
    [SERVICES.Redis]: "proton-redis",
    [SERVICES.PolicyEngine]: "proton-policy-engine",
    [SERVICES.ETCD]: "proton-etcd",
    [SERVICES.OpenSearch]: "opensearch",
    [SERVICES.Kafka]: "kafka",
    [SERVICES.Nebula]: "nebula",
    [SERVICES.Zookeeper]: "zookeeper",
};

export const defaultAddableComponents = [
    SERVICES.MariaDB,
    SERVICES.MongoDB,
    SERVICES.Redis,
    SERVICES.PolicyEngine,
    SERVICES.ETCD,
    SERVICES.OpenSearch,
    SERVICES.Kafka,
    SERVICES.Nebula,
];

// 数据库存储 类型
export const DataBaseStorageType = {
    // 标准版
    Standard: "standard",

    // 托管Kubernetes
    DepositKubernetes: "deposit-kubernetes",
};

export const enum OperationType {
    Edit,
    Add,
}

// 组件类型
export const SOURCE_TYPE = {
    // 内置
    INTERNAL: "internal",
    // 外置
    EXTERNAL: "external",
    // 无
    NOOP: "noop",
};

export const CONNECT_SERVICES = {
    RDS: "rds",
    MONGODB: "mongodb",
    REDIS: "redis",
    MQ: "mq",
    OPENSEARCH: "opensearch",
    POLICY_ENGINE: "proton-policy-engine",
    ETCD: "proton-etcd",
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

/**
 * 连接类型 翻译
 */
export const RedisConnectText = {
    [REDIS_CONNECT_TYPE.SENTINEL]: __("哨兵模式"),
    [REDIS_CONNECT_TYPE.MASTER_SLAVE]: __("主从模式"),
    [REDIS_CONNECT_TYPE.STANDALONE]: __("单机模式"),
    [REDIS_CONNECT_TYPE.CLUSTER]: __("集群模式"),
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

// 第三方searchEngine的协议
export const searchEngineProtocolType = {
    Http: "http",
    Https: "https",
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
    username: "fake_username",
    password: "fake_password",
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
    username: "fake_username",
    password: "fake_password",
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

export const DefaultConfigData = {
    [SERVICES.MariaDB]: {
        hosts: [],
        config: {
            innodb_buffer_pool_size: "8G",
            resource_requests_memory: "24G",
            resource_limits_memory: "24G",
        },
        admin_user: "root",
        admin_passwd: "fake_password",
        data_path: "/sysvol/mariadb",
        storage_capacity: "",
        username: "fake_username",
        password: "fake_password",
    },
    [SERVICES.MongoDB]: {
        hosts: [],
        admin_user: "root",
        admin_passwd: "fake_password",
        data_path: "/sysvol/mongodb/mongodb_data",
        storage_capacity: "",
        resources: {
            requests: {
                cpu: "100m",
                memory: "128Mi",
            },
        },
        username: "fake_username",
        password: "fake_password",
    },
    [SERVICES.Redis]: {
        hosts: [],
        admin_user: "root",
        admin_passwd: "fake_password",
        data_path: "/sysvol/redis/redis_data",
        storage_capacity: "",
        resources: {
            requests: {
                cpu: "100m",
                memory: "30Mi",
            },
        },
    },
    [SERVICES.PolicyEngine]: {
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
    [SERVICES.ETCD]: {
        hosts: [],
        data_path: "/sysvol/proton-etcd/proton-etcd_data",
        storage_capacity: "",
    },
    [SERVICES.OpenSearch]: {
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

// 切换数据库存储方式后，过滤数据库配置中不需要的参数
export const filterEmptyKey = (data: any, sourceType: string) => {
    let newData = data;
    if (sourceType === SOURCE_TYPE.INTERNAL) {
        delete newData.storageClassName;
        delete newData.replica_count;
    } else {
        delete newData.hosts;
        delete newData.data_path;
    }
    return newData;
};

// 端口数字验证规则
const portValidator = (rule: any, value: string) => {
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

// 本地RDS和MongoDB用户名验证规则
const usernameValidator = (rule: any, value: string) => {
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
const booleanValidator = (rule: any, value: boolean) => {
    return new Promise<void>((resolve, reject) => {
        if ([true, false].includes(value)) {
            resolve();
        } else {
            reject();
        }
    });
};

// 空值校验规则
export const emptyValidatorRules = [
    {
        required: true,
        message: __("此项不允许为空。"),
    },
];

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

// kafka日志配置验校验规则
export const getLogNumberValidatorRules = (flag: boolean) => [
    {
        validator: (rule: any, value: string) =>
            logNumberValidator(value, flag),
        message: __("请输入非负整数。"),
    },
];

// 内置组件/连接信息名称校验规则
export const nameValidatorRules = [
    {
        required: true,
        message: __("此项不允许为空。"),
    },
    {
        pattern: /^[a-zA-Z0-9-]*$/,
        message: __("名称只能包含英文或数字或-字符。"),
    },
];

// 布尔空值校验规则
export const booleanEmptyValidatorRules = [
    {
        validator: booleanValidator,
        message: __("此项不允许为空。"),
    },
];

// 端口校验规则
export const portValidatorRules = [
    {
        required: true,
        message: __("此项不允许为空。"),
    },
    {
        validator: portValidator,
        message: __("请输入1~65535范围内的整数。"),
    },
];

// 副本数验证规则
const replicaValidator = (originVal: number) => {
    return (rule: any, value: string) => {
        return new Promise<void>((resolve, reject) => {
            const regExp = /^[1-9][\d]{0,1}$/;
            if (value === undefined || value === null || value === "") {
                resolve();
            } else if (regExp.test(value)) {
                if (Number(value) >= originVal) {
                    resolve();
                } else {
                    reject();
                }
            } else {
                reject();
            }
        });
    };
};

export const getReplicaValidatorRules = (originVal: number) => {
    return [
        {
            required: true,
            message: __("此项不允许为空。"),
        },
        {
            validator: replicaValidator(originVal),
            message: __("请输入${originVal}~99范围内的整数。", {
                originVal: originVal || 1,
            }),
        },
    ];
};

// 本地RDS和MongoDB连接配置用户名校验规则
export const getUsernameValidatorRules = (sourceType: string) => {
    return sourceType === SOURCE_TYPE.INTERNAL
        ? [
              {
                  required: true,
                  message: __("此项不允许为空。"),
              },
              {
                  validator: usernameValidator,
                  message: __("此项不允许为root。"),
              },
          ]
        : emptyValidatorRules;
};

// 第三方RDS连接配置账户信息校验规则
export const getRDSUserInfoValidatorRules = (
    rdsType: string,
    comparisonValue: string
) => {
    return rdsType === RDS_TYPE.DM8
        ? [
              {
                  required: true,
                  message: __("此项不允许为空。"),
              },
              {
                  validator: (rule: any, value: string) =>
                      rdsUserInfoValidator(value, comparisonValue),
                  message: __("达梦数据库的管理权限和普通账户信息必须相同。"),
              },
          ]
        : emptyValidatorRules;
};

export const RESOURCES = {
    ALL: "all",
    LIMITS: "limits",
    REQUESTS: "requests",
};

export const NEBULA_COMPONENTS = {
    GRAPHD: "graphd",
    METAD: "metad",
    STORAGED: "storaged",
};

export const RESOURCES_TYPE = {
    CPU: "cpu",
    MEMORY: "memory",
};

const pwdComponentfilter = (obj: ComponentDataUnion): string => {
    if (isObject(obj)) {
        let { password, admin_passwd, ...connectInfo } = obj;
        return JSON.stringify(connectInfo);
    } else {
        return String(obj);
    }
};

export const updateComponentLog = async (
    operationType: OperationType,
    params: any,
    componentType: SERVICES
) => {
    const exMsg = pwdComponentfilter(params);
    // await manageLog(
    //     operationType === OperationType.Add
    //         ? ManagementOps.ADD
    //         : ManagementOps.SET,
    //     operationType === OperationType.Add
    //         ? __("添加${component}成功", {
    //               component: buildInComponentsText[componentType],
    //           })
    //         : __("设置${component}配置成功", {
    //               component: buildInComponentsText[componentType],
    //           }),
    //     exMsg,
    //     Level.INFO
    // );
};

/**
 * 根据类型获取连接信息
 * @param curInfo 现信息
 * @returns
 */
export const getRDSInfoByResourceType = (curInfo: RDSConnectInfo) => {
    const {
        source_type: curSourceType,
        rds_type: curRdsType,
        username: curUsername,
        password: curPassword,
        hosts: curHosts,
        port: curPort,
        auto_create_database: curAutoCreateDatabase,
        admin_user: curAdminUser,
        admin_passwd: curAdminPasswd,
    } = curInfo || {};

    let info: RDSConnectInfo = {
        source_type: curSourceType,
        username: curUsername,
        password: curPassword,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
        info = {
            ...info,
            rds_type: curRdsType,
            hosts: curHosts,
            port: curPort,
            auto_create_database: curAutoCreateDatabase,
            admin_user: curAdminUser,
            admin_passwd: curAdminPasswd,
        };
    }
    return info;
};

/**
 * 根据类型获取连接信息
 * @param curInfo 现信息
 * @returns
 */
export const getMongoDBInfoByResourceType = (curInfo: MongoDBConnectInfo) => {
    const {
        source_type: curSourceType,
        username: curUsername,
        password: curPassword,
        replica_set: curReplicaSet,
        hosts: curHosts,
        port: curPort,
        ssl: curSSL,
        auth_source: curAuthSource,
        options: curOptions,
    } = curInfo || {};

    let info: MongoDBConnectInfo = {
        source_type: curSourceType,
        username: curUsername,
        password: curPassword,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
        info = {
            ...info,
            hosts: curHosts,
            port: curPort,
            ssl: curSSL,
            replica_set: curReplicaSet,
            auth_source: curAuthSource,
            options: curOptions,
        };
    }
    return info;
};

/**
 * 根据类型获取连接信息
 * @param preInfo 原信息
 * @param curInfo 现信息
 * @returns
 */
export const getRedisInfoByResourceType = (curInfo: RedisConnectInfo) => {
    const {
        source_type: curSourceType,
        connect_type: curConnectType,
        username: curUsername,
        password: curPassword,
        hosts: curHosts,
        port: curPort,
        sentinel_username: curSentinelUsername,
        sentinel_password: curSentinelPassword,
        sentinel_hosts: curSentinelHosts,
        sentinel_port: curSentinelPort,
        master_group_name: curMasterGroupName,
        master_hosts: curMasterHosts,
        master_port: curMasterPort,
        slave_hosts: curSlaveHosts,
        slave_port: curSlavePort,
    } = curInfo || {};

    let info: RedisConnectInfo = {
        source_type: curSourceType,
    };
    if (curSourceType === SOURCE_TYPE.INTERNAL) {
        return info;
    } else {
        info = {
            source_type: curSourceType,
            connect_type: curConnectType,
            username: curUsername,
            password: curPassword,
        };
        switch (curConnectType) {
            case REDIS_CONNECT_TYPE.SENTINEL:
                return {
                    ...info,
                    sentinel_username: curSentinelUsername,
                    sentinel_password: curSentinelPassword,
                    sentinel_hosts: curSentinelHosts,
                    sentinel_port: curSentinelPort,
                    master_group_name: curMasterGroupName,
                };
            case REDIS_CONNECT_TYPE.MASTER_SLAVE:
                return {
                    ...info,
                    master_hosts: curMasterHosts,
                    master_port: curMasterPort,
                    slave_hosts: curSlaveHosts,
                    slave_port: curSlavePort,
                };
            default:
                return {
                    ...info,
                    hosts: curHosts,
                    port: curPort,
                };
        }
    }
};

/**
 * 根据类型获取连接信息
 * @param preInfo 原信息
 * @param curInfo 现信息
 * @returns
 */
export const getMQInfoByResourceType = (curInfo: MQConnectInfo) => {
    const {
        source_type: curSourceType,
        mq_type: curMQType,
        mq_hosts: curHosts,
        mq_port: curPort,
        mq_lookupd_hosts: curLookupdHosts,
        mq_lookupd_port: curLookupdPort,
        auth: curAuth,
    } = curInfo || {};

    let info: MQConnectInfo = {
        source_type: curSourceType,
        mq_type: curMQType,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
        info = {
            ...info,
            mq_hosts: curHosts,
            mq_port: curPort,
        };
        if (curMQType === MQ_TYPE.KAFKA) {
            const {
                username: curUsername,
                password: curPassword,
                mechanism: curMechanism,
            } = curAuth || {};
            info = {
                ...info,
                auth: {
                    username: curUsername!,
                    password: curPassword!,
                    mechanism: curMechanism!,
                },
            };
        }
        if (curMQType !== MQ_TYPE.NSQ) {
            info = {
                ...info,
                mq_lookupd_hosts: curHosts,
                mq_lookupd_port: curPort,
            };
        } else {
            info = {
                ...info,
                mq_lookupd_hosts: curLookupdHosts,
                mq_lookupd_port: curLookupdPort,
            };
        }
    }
    return info;
};

/**
 * 根据类型获取连接信息
 * @param curInfo 现信息
 * @returns
 */
export const getOpenSearchInfoByResourceType = (
    curInfo: OpensearchConnectInfo
) => {
    const {
        source_type: curSourceType,
        username: curUsername,
        password: curPassword,
        hosts: curHosts,
        port: curPort,
        version: curVersion,
        distribution: curDistribution,
        protocol: curProtocol,
    } = curInfo || {};

    let info: OpensearchConnectInfo = {
        source_type: curSourceType,
        version: curVersion,
        distribution: curDistribution,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
        info = {
            ...info,
            username: curUsername,
            password: curPassword,
            hosts: curHosts,
            port: curPort,
            version: curVersion,
            protocol: curProtocol,
        };
    }
    return info;
};

/**
 * 根据类型获取连接信息
 * @param curInfo 现信息
 * @returns
 */
export const getPolicyEngineInfoByResourceType = (
    curInfo: PolicyEngineConnectInfo
) => {
    const {
        source_type: curSourceType,
        hosts: curHosts,
        port: curPort,
    } = curInfo || {};

    let info: PolicyEngineConnectInfo = {
        source_type: curSourceType,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
        info = {
            ...info,
            hosts: curHosts,
            port: curPort,
        };
    }
    return info;
};

/**
 * 根据类型获取连接信息
 * @param curInfo 现信息
 * @returns
 */
export const getETCDInfoByResourceType = (curInfo: ETCDConnectInfo) => {
    const {
        source_type: curSourceType,
        hosts: curHosts,
        port: curPort,
        secret: curSecret,
    } = curInfo || {};

    let info: ETCDConnectInfo = {
        source_type: curSourceType,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
        info = {
            ...info,
            hosts: curHosts,
            port: curPort,
            secret: curSecret,
        };
    }
    return info;
};

export const changeResourceConnectServiceTypeByService = (
    isDel: boolean,
    service: any,
    type?: string
) => {
    if (service === SERVICES.MariaDB) {
        let rds;
        if (isDel) {
            rds = getRDSInfoByResourceType(
                type
                    ? { ...DEGAULT_EXTERNAL_RDS, rds_type: type }
                    : (DEGAULT_EXTERNAL_RDS as any)
            );
        } else {
            rds = getRDSInfoByResourceType(DEFAULT_INTERNAL_RDS);
        }
        return rds;
    } else if (service === SERVICES.MongoDB) {
        let mongodb;
        if (isDel) {
            mongodb = getMongoDBInfoByResourceType(DEFAULT_EXTERNAL_MONGODB);
        } else {
            mongodb = getMongoDBInfoByResourceType(DEFAULT_INTERNAL_MONGODB);
        }
        return mongodb;
    } else if (service === SERVICES.Redis) {
        let redis;
        if (isDel) {
            redis = getRedisInfoByResourceType(
                type
                    ? {
                          ...DEFAULT_EXTERNAL_REDIS,
                          connect_type: type,
                      }
                    : DEFAULT_EXTERNAL_REDIS
            );
        } else {
            redis = getRedisInfoByResourceType(DEFAULT_INTERNAL_SERVICE);
        }
        return redis;
    } else if (service === SERVICES.ProtonNSQ) {
        if (isDel) {
            return getMQInfoByResourceType(DEFAULT_EXTERNAL_MQ);
        } else {
            return getMQInfoByResourceType({
                ...DEFAULT_INTERNAL_MQ,
                mq_type: MQ_TYPE.NSQ,
            });
        }
    } else if (service === SERVICES.Kafka) {
        if (isDel) {
            return getMQInfoByResourceType(
                type
                    ? { ...DEFAULT_EXTERNAL_MQ, mq_type: type }
                    : DEFAULT_EXTERNAL_MQ
            );
        } else {
            return getMQInfoByResourceType(DEFAULT_INTERNAL_MQ);
        }
    } else if (service === SERVICES.OpenSearch) {
        let opensearch;
        if (isDel) {
            opensearch = getOpenSearchInfoByResourceType(
                DEFAULT_EXTERNAL_OPENSEARCH
            );
        } else {
            opensearch = getOpenSearchInfoByResourceType(
                DEFAULT_INTERNAL_OPENSEARCH
            );
        }
        return opensearch;
    } else if (service === SERVICES.PolicyEngine) {
        let policyEngine;
        if (isDel) {
            policyEngine = getPolicyEngineInfoByResourceType(
                DEFAULT_EXTERNAL_SERVICE
            );
        } else {
            policyEngine = getPolicyEngineInfoByResourceType(
                DEFAULT_INTERNAL_SERVICE
            );
        }
        return policyEngine;
    } else if (service === SERVICES.ETCD) {
        let etcd;
        if (isDel) {
            etcd = getETCDInfoByResourceType(DEFAULT_EXTERNAL_SERVICE);
        } else {
            etcd = getETCDInfoByResourceType(DEFAULT_INTERNAL_SERVICE);
        }
        return etcd;
    }
};

// 校验状态
export const enum ValidateState {
    // 正常
    Normal,
    // 输入为空
    Empty,
    // 节点数量错误
    NodesNumError,
}

export const DefaultConnectInfoValidateState = {
    RDS_TYPE: ValidateState.Normal,
    MONGODB_SSL: ValidateState.Normal,
    REDIS_CONNECT_TYPE: ValidateState.Normal,
    MQ_RADIO: ValidateState.Normal,
    MQ_TYPE: ValidateState.Normal,
    MQ_AUTH_MACHANISM: ValidateState.Normal,
    OPENSEARCH_VERSION: ValidateState.Normal,
};

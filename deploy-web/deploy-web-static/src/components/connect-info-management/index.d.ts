import { ValidateState } from "../component-management/helper";

export type MQAuthConnectInfo = {
    username: string;
    password: string;
    mechanism: string;
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
    auto_create_database?: boolean;
    admin_user?: string;
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
    auth?: MQAuthConnectInfo;
};

export type OpensearchConnectInfo = {
    source_type: string;
    hosts?: string;
    port?: number;
    username?: string;
    password?: string;
    version: string;
    distribution: string;
    protocol?: string;
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

type ConnectInfoUnion = RDSConnectInfo &
    MongoDBConnectInfo &
    RedisConnectInfo &
    MQConnectInfo &
    OpensearchConnectInfo &
    PolicyEngineConnectInfo &
    ETCDConnectInfo;

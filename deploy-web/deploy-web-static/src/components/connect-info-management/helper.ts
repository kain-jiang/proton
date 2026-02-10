import { ConnectInfoUnion } from "./index.d";
import { OperationType, SERVICES } from "../component-management/helper";
import __ from "./locale";
// import { manageLog, ManagementOps, Level } from "mediator";
import { isObject, isString } from "lodash";

export enum ConnectInfoServices {
    ETCD = "proton-etcd",
    MongoDB = "mongodb",
    MQ = "mq",
    OpenSearch = "opensearch",
    PolicyEngine = "proton-policy-engine",
    RDS = "rds",
    Redis = "redis",
}

export const serviceMapComponent = {
    [ConnectInfoServices.RDS]: SERVICES.MariaDB,
    [ConnectInfoServices.MongoDB]: SERVICES.MongoDB,
    [ConnectInfoServices.Redis]: SERVICES.Redis,
    [ConnectInfoServices.MQ]: SERVICES.Kafka,
    [ConnectInfoServices.OpenSearch]: SERVICES.OpenSearch,
    [ConnectInfoServices.PolicyEngine]: SERVICES.PolicyEngine,
    [ConnectInfoServices.ETCD]: SERVICES.ETCD,
};

export interface ConnectInfoData {
    // 连接配置
    info?: any;
    // 内置组件配置
    instance?: any;
    // 连接信息名称
    name?: string;
    zookeeper?: any;
    // 连接信息类型
    type?: string;
    // 系统空间id
    sid?: number;
    // 系统空间名称
    systemName?: string;
    // 命名空间
    namespace?: string;
}

export const connectInfoServices = [
    ConnectInfoServices.RDS,
    ConnectInfoServices.MongoDB,
    ConnectInfoServices.Redis,
    ConnectInfoServices.MQ,
    ConnectInfoServices.OpenSearch,
    ConnectInfoServices.PolicyEngine,
    ConnectInfoServices.ETCD,
];

export interface ConfigData {
    [ConnectInfoServices.RDS]?: ConnectInfoData | null;
    [ConnectInfoServices.MongoDB]?: ConnectInfoData | null;
    [ConnectInfoServices.Redis]?: ConnectInfoData | null;
    [ConnectInfoServices.MQ]?: ConnectInfoData | null;
    [ConnectInfoServices.OpenSearch]?: ConnectInfoData | null;
    [ConnectInfoServices.PolicyEngine]?: ConnectInfoData | null;
    [ConnectInfoServices.ETCD]?: ConnectInfoData | null;
}

export const defaultAddableServices = [
    ConnectInfoServices.RDS,
    ConnectInfoServices.MongoDB,
    ConnectInfoServices.Redis,
    ConnectInfoServices.MQ,
    ConnectInfoServices.OpenSearch,
    ConnectInfoServices.PolicyEngine,
    ConnectInfoServices.ETCD,
];

export const servicesText = {
    [ConnectInfoServices.RDS]: __("关系型数据库（RDS）"),
    [ConnectInfoServices.MongoDB]: "MongoDB",
    [ConnectInfoServices.Redis]: "Redis",
    [ConnectInfoServices.MQ]: __("消息队列（MQ）"),
    [ConnectInfoServices.OpenSearch]: "SearchEngine",
    [ConnectInfoServices.PolicyEngine]: "PolicyEngine",
    [ConnectInfoServices.ETCD]: "ETCD",
};

export const servicesSourceTypeText = {
    [ConnectInfoServices.RDS]: "RDS",
    [ConnectInfoServices.MongoDB]: "MongoDB",
    [ConnectInfoServices.Redis]: "Redis",
    [ConnectInfoServices.MQ]: "MQ",
    [ConnectInfoServices.OpenSearch]: __("搜索与分析引擎"),
    [ConnectInfoServices.PolicyEngine]: "PolicyEngine",
    [ConnectInfoServices.ETCD]: "ETCD",
};

export const servicesInternalSourceTypeText = {
    [ConnectInfoServices.RDS]: "MariaDB",
    [ConnectInfoServices.MongoDB]: "MongoDB",
    [ConnectInfoServices.Redis]: "Redis",
    [ConnectInfoServices.MQ]: "MQ",
    [ConnectInfoServices.OpenSearch]: "OpenSearch",
    [ConnectInfoServices.PolicyEngine]: "PolicyEngine",
    [ConnectInfoServices.ETCD]: "ETCD",
};

const pwdServicefilter = (obj: ConnectInfoUnion): string => {
    if (isObject(obj)) {
        let { password, sentinel_password, ...connectInfo } = obj;
        if (isObject(connectInfo.auth)) {
            connectInfo = {
                ...connectInfo,
                auth: {
                    ...connectInfo.auth,
                    password: "",
                },
            };
        }
        return JSON.stringify(connectInfo);
    } else {
        return String(obj);
    }
};

export const updateServiceLog = async (
    operationType: OperationType,
    params: any,
    serviceType: ConnectInfoServices
) => {
    const exMsg = pwdServicefilter(params);
    // await manageLog(
    //     operationType === OperationType.Add
    //         ? ManagementOps.ADD
    //         : ManagementOps.SET,
    //     operationType === OperationType.Add
    //         ? __("添加${service}连接信息成功", {
    //               service: servicesText[serviceType],
    //           })
    //         : __("设置${service}连接信息成功", {
    //               service: servicesText[serviceType],
    //           }),
    //     exMsg,
    //     Level.INFO
    // );
};

/**
 * 将mongo的opotions转为sting类型
 * @param data
 * @returns
 */
export function transMongoOptions2String(data: any) {
    if (isObject(data?.options)) {
        let options = "",
            keys = Object.keys(data?.options);
        if (keys.length) {
            options = keys
                .map((k) => `${k}=${data?.options[k]}`)
                .join("&")
                .toString();
        }
        return {
            ...data,
            options,
        };
    } else {
        return data;
    }
}

/**
 * 将mongo的opotions转为object类型
 * @param data
 * @returns
 */
export function transMongoOptions2Object(data: any) {
    let options = {};
    if (isString(data?.options) && data?.options?.indexOf("=") !== -1) {
        options = data?.options
            .split("&")
            .reduce((pre: object, cur: string) => {
                const [k, v] = cur.split("=");
                return {
                    ...pre,
                    [k]: v,
                };
            }, {});
    }
    return {
        ...data,
        options,
    };
}

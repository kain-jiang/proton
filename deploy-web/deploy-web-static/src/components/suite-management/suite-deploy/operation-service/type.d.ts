import { ServiceMode } from "../../../../core/service-management/service-deploy";
import __ from "./locale";
/**
 * @enum OperationServiceStepsEnum
 * @description 安装或更新服务步骤枚举
 * @param CHOOSESERVICE 0 选择服务
 * @param CREATECONFIG 1 填写配置项
 * @param CONFIRMCONFIG 2 确认配置
 */
export enum OperationServiceStepsEnum {
  CHOOSESERVICE,
  CREATECONFIG,
  CONFIRMCONFIG,
}

export type OperationType = ServiceMode.Install | ServiceMode.Update;

export enum ConnectInfoServices {
  ETCD = "etcd",
  MongoDB = "mongodb",
  MQ = "mq",
  OpenSearch = "opensearch",
  PolicyEngine = "policyengine",
  RDS = "rds",
  Redis = "redis",
}

export const componentsText = {
  [ConnectInfoServices.RDS]: __("关系型数据库（RDS）"),
  [ConnectInfoServices.MongoDB]: "MongoDB",
  [ConnectInfoServices.Redis]: "Redis",
  [ConnectInfoServices.MQ]: __("消息队列（MQ）"),
  [ConnectInfoServices.OpenSearch]: "OpenSearch",
  [ConnectInfoServices.PolicyEngine]: "PolicyEngine",
  [ConnectInfoServices.ETCD]: "ETCD",
};

import { ServiceMode } from "../../../../core/service-management/service-deploy";

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

// 操作类型
export type OperationType = ServiceMode.Install | ServiceMode.Update;

import { Color } from "../../common/components";
import __ from "../locale";

/**
 * @enum ServiceConfigStatusEnum
 * @description 服务细分状态枚举
 * @param INITIALIZING 0 初始化
 * @param CONFIGCONFIRMED 1 配置已确认
 * @param WAITINGEXECUTE 2 任务等待执行
 * @param EXECUTING 3 任务正在执行
 * @param SUCCESS 4 任务执行成功
 * @param FAILURE 5 任务失败
 * @param STOPPED 6 任务已暂停
 * @param STOPPING 7 任务停止中
 * @param MISSINGDEPENDENCE 8 任务缺少依赖组件，失败
 * @param UNINSTALLLASTVERSIONFAILURE 9 任务卸载上一版本组件失败
 * @param SKIPPED 10 略过
 * @param DELETINGLASTVERSION 11 删除上一版本组件中
 * @param UPDATECOMPLETE 12 目标版本服务更新完毕
 * @param PARENTCOMPONENTUPDATING 13 父组件更新阶段
 * @param PARENTCOMPONENTUPDATEFAILURE 14 父组件更新失败
 */
export enum ServiceConfigStatusEnum {
  // 初始化
  INITIALIZING,
  // 配置已确认
  CONFIGCONFIRMED,
  // 任务等待执行
  WAITINGEXECUTE,
  // 任务正在执行
  EXECUTING,
  // 任务执行成功
  SUCCESS,
  // 任务失败
  FAILURE,
  // 任务已暂停
  STOPPED,
  // 任务停止中
  STOPPING,
  // 任务缺少依赖组件，失败
  MISSINGDEPENDENCE,
  // 任务卸载上一版本组件失败
  UNINSTALLLASTVERSIONFAILURE,
  // 略过
  SKIPPED,
  // 删除上一版本组件中
  DELETINGLASTVERSION,
  // 目标版本服务更新完毕
  UPDATECOMPLETE,
  // 父组件更新阶段
  PARENTCOMPONENTUPDATING,
  // 父组件更新失败
  PARENTCOMPONENTUPDATEFAILURE,
}

// 任务状态
export const serviceConfigStatus = {
  [ServiceConfigStatusEnum.INITIALIZING]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.INITIALIZING,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.CONFIGCONFIRMED]: {
    text: __("运行中"),
    value: ServiceConfigStatusEnum.CONFIGCONFIRMED,
    color: Color.SERVICE_GREEN,
  },
  [ServiceConfigStatusEnum.WAITINGEXECUTE]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.WAITINGEXECUTE,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.EXECUTING]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.EXECUTING,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.SUCCESS]: {
    text: __("运行中"),
    value: ServiceConfigStatusEnum.SUCCESS,
    color: Color.SERVICE_GREEN,
  },
  [ServiceConfigStatusEnum.FAILURE]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.FAILURE,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.STOPPED]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.STOPPED,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.STOPPING]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.STOPPING,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.MISSINGDEPENDENCE]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.MISSINGDEPENDENCE,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.UNINSTALLLASTVERSIONFAILURE]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.UNINSTALLLASTVERSIONFAILURE,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.SKIPPED]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.SKIPPED,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.DELETINGLASTVERSION]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.DELETINGLASTVERSION,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.UPDATECOMPLETE]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.UPDATECOMPLETE,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.PARENTCOMPONENTUPDATING]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.PARENTCOMPONENTUPDATING,
    color: Color.SERVICE_RED,
  },
  [ServiceConfigStatusEnum.PARENTCOMPONENTUPDATEFAILURE]: {
    text: __("运行失败"),
    value: ServiceConfigStatusEnum.PARENTCOMPONENTUPDATEFAILURE,
    color: Color.SERVICE_RED,
  },
};

/**
 * @enum ServiceCategoryStatusEnum
 * @description 服务状态枚举
 * @param RUNNING RUNNING 运行中
 * @param FAILED FAILED 运行失败
 */
export enum ServiceCategoryStatusEnum {
  RUNNING = "RUNNING",
  FAILED = "FAILED",
}

// 服务状态
export const serviceCategoryStatus = {
  [ServiceCategoryStatusEnum.RUNNING]: {
    text: __("运行中"),
    value: ServiceCategoryStatusEnum.RUNNING,
    color: Color.SERVICE_GREEN,
  },
  [ServiceCategoryStatusEnum.FAILED]: {
    text: __("运行失败"),
    value: ServiceCategoryStatusEnum.FAILED,
    color: Color.SERVICE_RED,
  },
};

// 服务状态内容
export const serviceCategoryStatusItems = {
  [ServiceCategoryStatusEnum.RUNNING]: [
    ServiceConfigStatusEnum.CONFIGCONFIRMED,
    ServiceConfigStatusEnum.SUCCESS,
  ],
  [ServiceCategoryStatusEnum.FAILED]: [
    ServiceConfigStatusEnum.INITIALIZING,
    ServiceConfigStatusEnum.WAITINGEXECUTE,
    ServiceConfigStatusEnum.EXECUTING,
    ServiceConfigStatusEnum.STOPPED,
    ServiceConfigStatusEnum.STOPPING,
    ServiceConfigStatusEnum.DELETINGLASTVERSION,
    ServiceConfigStatusEnum.SKIPPED,
    ServiceConfigStatusEnum.UPDATECOMPLETE,
    ServiceConfigStatusEnum.PARENTCOMPONENTUPDATING,
    ServiceConfigStatusEnum.FAILURE,
    ServiceConfigStatusEnum.MISSINGDEPENDENCE,
    ServiceConfigStatusEnum.UNINSTALLLASTVERSIONFAILURE,
    ServiceConfigStatusEnum.PARENTCOMPONENTUPDATEFAILURE,
  ],
};

/**
 * @enum SelectStatusEnum
 * @description 服务确认状态枚举
 * @param Selected true 是
 * @param notSelected false 否
 */
export const SelectStatusEnum = {
  Selected: true,
  notSelected: false,
};

// 服务确认状态
export const selectStatus = {
  [SelectStatusEnum.Selected]: {
    text: __("是"),
    value: SelectStatusEnum.Selected,
    color: Color.SERVICE_GREEN,
  },
  [SelectStatusEnum.notSelected]: {
    text: __("否"),
    value: SelectStatusEnum.notSelected,
    color: Color.SERVICE_RED,
  },
};

// 组件类型枚举
export enum ComponentDefineTypeEnum {
  Service = "helm/service",
  Resource = "proton/resource",
  Addtional = "helm/addtional",
  Hole = "hole",
}

// 组件类型过滤
export const componentDefineTypeFilter = [
  { text: __("服务"), value: ComponentDefineTypeEnum.Service },
  { text: __("资源"), value: ComponentDefineTypeEnum.Resource },
  { text: __("附加"), value: ComponentDefineTypeEnum.Addtional },
  { text: __("补充"), value: ComponentDefineTypeEnum.Hole },
];

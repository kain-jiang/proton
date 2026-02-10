import { Color } from "../../common/components";
import __ from "../locale";

/**
 * @enum TaskConfigStatusEnum
 * @description 任务细分状态枚举
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
export enum TaskConfigStatusEnum {
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
export const taskConfigStatus = {
  [TaskConfigStatusEnum.INITIALIZING]: {
    categoryText: __("运行中"),
    text: __("初始化"),
    value: TaskConfigStatusEnum.INITIALIZING,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.CONFIGCONFIRMED]: {
    categoryText: __("配置已确认"),
    text: __("配置已确认"),
    value: TaskConfigStatusEnum.CONFIGCONFIRMED,
    color: Color.SERVICE_GREEN,
  },
  [TaskConfigStatusEnum.WAITINGEXECUTE]: {
    categoryText: __("运行中"),
    text: __("任务等待执行"),
    value: TaskConfigStatusEnum.WAITINGEXECUTE,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.EXECUTING]: {
    categoryText: __("运行中"),
    text: __("任务正在执行"),
    value: TaskConfigStatusEnum.EXECUTING,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.SUCCESS]: {
    categoryText: __("成功"),
    text: __("任务执行成功"),
    value: TaskConfigStatusEnum.SUCCESS,
    color: Color.SERVICE_GREEN,
  },
  [TaskConfigStatusEnum.FAILURE]: {
    categoryText: __("失败"),
    text: __("任务失败"),
    value: TaskConfigStatusEnum.FAILURE,
    color: Color.SERVICE_RED,
  },
  [TaskConfigStatusEnum.STOPPED]: {
    categoryText: __("已暂停"),
    text: __("任务已暂停"),
    value: TaskConfigStatusEnum.STOPPED,
    color: Color.SERVICE_GREY,
  },
  [TaskConfigStatusEnum.STOPPING]: {
    categoryText: __("运行中"),
    text: __("任务停止中"),
    value: TaskConfigStatusEnum.STOPPING,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.MISSINGDEPENDENCE]: {
    categoryText: __("失败"),
    text: __("任务缺少依赖组件，失败"),
    value: TaskConfigStatusEnum.MISSINGDEPENDENCE,
    color: Color.SERVICE_RED,
  },
  [TaskConfigStatusEnum.UNINSTALLLASTVERSIONFAILURE]: {
    categoryText: __("失败"),
    text: __("任务卸载上一版本组件失败"),
    value: TaskConfigStatusEnum.UNINSTALLLASTVERSIONFAILURE,
    color: Color.SERVICE_RED,
  },
  [TaskConfigStatusEnum.SKIPPED]: {
    categoryText: __("运行中"),
    text: __("略过"),
    value: TaskConfigStatusEnum.SKIPPED,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.DELETINGLASTVERSION]: {
    categoryText: __("运行中"),
    text: __("删除上一版本组件中"),
    value: TaskConfigStatusEnum.DELETINGLASTVERSION,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.UPDATECOMPLETE]: {
    categoryText: __("运行中"),
    text: __("目标版本服务更新完毕"),
    value: TaskConfigStatusEnum.UPDATECOMPLETE,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.PARENTCOMPONENTUPDATING]: {
    categoryText: __("运行中"),
    text: __("父组件更新阶段"),
    value: TaskConfigStatusEnum.PARENTCOMPONENTUPDATING,
    color: Color.SERVICE_BLUE,
  },
  [TaskConfigStatusEnum.PARENTCOMPONENTUPDATEFAILURE]: {
    categoryText: __("失败"),
    text: __("父组件更新失败"),
    value: TaskConfigStatusEnum.PARENTCOMPONENTUPDATEFAILURE,
    color: Color.SERVICE_RED,
  },
};

/**
 * @enum TaskCategoryStatusEnum
 * @description 任务状态枚举
 * @param RUNNING RUNNING 运行中
 * @param STOPPED STOPPED 已暂停
 * @param SUCCEEDED SUCCEEDED 成功
 * @param FAILED FAILED 失败
 * @param CONFIGCONFIRMED CONFIGCONFIRMED 配置已确认
 */
export enum TaskCategoryStatusEnum {
  RUNNING = "RUNNING",
  STOPPED = "STOPPED",
  SUCCEEDED = "SUCCEEDED",
  FAILED = "FAILED",
  CONFIGCONFIRMED = "CONFIGCONFIRMED",
}

// 任务状态
export const taskCategoryStatus = {
  [TaskCategoryStatusEnum.RUNNING]: {
    text: __("运行中"),
    value: TaskCategoryStatusEnum.RUNNING,
    color: Color.SERVICE_BLUE,
  },
  [TaskCategoryStatusEnum.STOPPED]: {
    text: __("已暂停"),
    value: TaskCategoryStatusEnum.STOPPED,
    color: Color.SERVICE_GREY,
  },
  [TaskCategoryStatusEnum.SUCCEEDED]: {
    text: __("成功"),
    value: TaskCategoryStatusEnum.SUCCEEDED,
    color: Color.SERVICE_GREEN,
  },
  [TaskCategoryStatusEnum.FAILED]: {
    text: __("失败"),
    value: TaskCategoryStatusEnum.FAILED,
    color: Color.SERVICE_RED,
  },
  [TaskCategoryStatusEnum.CONFIGCONFIRMED]: {
    text: __("配置已确认"),
    value: TaskCategoryStatusEnum.CONFIGCONFIRMED,
    color: Color.SERVICE_GREEN,
  },
};

// 任务状态内容
export const taskCategoryStatusItems = {
  [TaskCategoryStatusEnum.RUNNING]: [
    TaskConfigStatusEnum.INITIALIZING,
    TaskConfigStatusEnum.WAITINGEXECUTE,
    TaskConfigStatusEnum.EXECUTING,
    TaskConfigStatusEnum.STOPPING,
    TaskConfigStatusEnum.DELETINGLASTVERSION,
    TaskConfigStatusEnum.SKIPPED,
    TaskConfigStatusEnum.UPDATECOMPLETE,
    TaskConfigStatusEnum.PARENTCOMPONENTUPDATING,
  ],
  [TaskCategoryStatusEnum.STOPPED]: [TaskConfigStatusEnum.STOPPED],
  [TaskCategoryStatusEnum.SUCCEEDED]: [TaskConfigStatusEnum.SUCCESS],
  [TaskCategoryStatusEnum.FAILED]: [
    TaskConfigStatusEnum.FAILURE,
    TaskConfigStatusEnum.MISSINGDEPENDENCE,
    TaskConfigStatusEnum.UNINSTALLLASTVERSIONFAILURE,
    TaskConfigStatusEnum.PARENTCOMPONENTUPDATEFAILURE,
  ],
  [TaskCategoryStatusEnum.CONFIGCONFIRMED]: [
    TaskConfigStatusEnum.CONFIGCONFIRMED,
  ],
};

// 任务操作类型
export enum JobOperateType {
  // 更新
  Update,
  // 安装
  Install,
  // 回滚
  Revert,
  // 卸载
  Uninstall,
}

export const jobOperateTypeStatus = {
  [JobOperateType.Update]: {
    text: __("更新"),
    value: JobOperateType.Update,
  },
  [JobOperateType.Install]: {
    text: __("安装"),
    value: JobOperateType.Install,
  },
  [JobOperateType.Revert]: {
    text: __("回滚"),
    value: JobOperateType.Revert,
  },
  [JobOperateType.Uninstall]: {
    text: __("卸载"),
    value: JobOperateType.Uninstall,
  },
};

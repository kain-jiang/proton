import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ServiceConfigStatusEnum } from "../../../components/service-management/service-deploy/type";
import {
  JobOperateType,
  TaskConfigStatusEnum,
} from "../../../components/service-management/task-monitor/type";

export interface ServiceItem {
  // 服务id
  id: number;
  // 服务标识
  name: string;
  // 服务名称
  title: string;
  // 服务版本
  version: string;
  // 服务状态（1：运行失败；2：运行中）
  status: number;
  // 服务备注
  comment: string;
  // 结束时间
  endTime: number;
  // 应用包id
  aid: number;
  // 应用定义类型与版本
  appDefineType?: string;
  // config
  config?: object;
  // 创建时间
  createTime: number;
  // 命名空间
  namespace?: string;
  // 执行器ID
  owner?: number;
  // 系统id
  sid?: number;
  // 任务开始时间（更新时间）
  startTime: number;
  // 系统空间名称
  systemName?: string;
  // 任务操作类型
  operateType?: JobOperateType;
}

/**
 * @interface IGetServiceTableParams
 * @param name 搜索字段-服务名称
 * @param title 搜索字段-服务标识
 * @param status 筛选字段-状态类型
 * @param sid 筛选字段-系统空间ID
 */
export interface IGetServiceTableParams {
  title?: string;
  name?: string;
  status?: ServiceConfigStatusEnum[];
  sid?: number;
}

/**
 * @interface IGetServiceParams
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetServiceParams extends IGetServiceTableParams {
  offset: number;
  limit: number;
}

/**
 * @description 服务详细信息和配置
 * @param {number} jid 任务id
 * @param {string} formData 配置项数据
 * @param {string} schema 配置项类型定义
 */
export interface ServiceJSONSchemaItem extends ServiceItem {
  jid?: number;
  formData?: RJSFSchema;
  schema?: RJSFSchema;
  uiSchema?: UiSchema;
}

export interface ApplicationItem {
  // 应用包id
  aid: number;
  // 应用定义类型与版本
  appDefineType?: string;
  // 应用标识
  name: string;
  // 应用名称
  title: string;
  // 应用版本
  version: string;
}

/**
 * @interface IGetApplicationParams
 * @param offset 索引
 * @param limit 页数
 * @param name 名称筛选
 * @param nowork 是否过滤已安装应用
 * @param sid 系统空间id
 */
export interface IGetApplicationParams {
  offset: number;
  limit: number;
  name?: string;
  nowork?: boolean;
  sid?: number;
}

/**
 * @interface IConfigJobParams
 * @param jid 任务id
 * @param comment 备注
 * @param config 配置项
 */
export interface IConfigJobParams {
  jid: number;
  comment?: string;
  formData: RJSFSchema;
}

/**
 * @interface ICreateAndExecuteJobParams
 * @param aid 应用id
 * @param comment 备注
 * @param config 配置项
 * @param operateType 任务类型 （仅在回滚时需要设置，其他类型由后端自动设置）
 * @param sid 系统空间id
 */
export interface ICreateAndExecuteJobParams {
  aid: number;
  comment?: string;
  formData: RJSFSchema;
  operateType?: JobOperateType;
  sid: number;
}

export interface ComponentInstanceType {
  // 组件定义方式
  componentDefineType?: string;
  // 组件名称
  name: string;
  // 特殊可复用组件定义下的组件类型
  type?: string;
  // 组件版本
  version: string;
}
export interface ComponentType {
  // 结束时间
  endTime: number;
  // 应用定义内组件定义的ID
  acid?: number;
  // 组件实例所属的应用实例ID
  aiid?: number;
  // 组件所属应用名称
  appName?: string;
  // 组件属性配置
  attribute?: object;
  // 组件id
  cid: number;
  // 组件基础信息
  component: componentInstanceType;
  // 组件配置
  config?: object;
  // 创建时间
  createTime: number;
  // 开始时间（更新时间）
  startTime: number;
  // 系统信息
  system?: object;
  // 组件状态等
  trait: {
    status: number;
    timeout?: number;
  };
}

/**
 * @interface IGetJobTableParams
 * @param name 搜索字段-任务标识
 * @param title 搜索字段-任务名称
 * @param status 搜索字段-任务状态
 * @param jtype 搜索字段-任务操作类型
 * @param sid 搜索字段-系统空间id
 */
export interface IGetJobTableParams {
  name?: string;
  title?: string;
  status?: TaskConfigStatusEnum[];
  jtype?: JobOperateType[];
  sid?: number;
}

/**
 * @interface IGetJobParams
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetJobParams extends IGetJobTableParams {
  offset: number;
  limit: number;
}

export interface ApplicationType {
  // 应用包id
  aid: number;
  appDefineType?: string;
  components?: Array<object>;
  configSchema?: object;
  graph?: Array<object>;
  // 应用包标识
  name: string;
  // 应用包名称
  title: string;
  // 应用包版本
  version: string;
}

export interface JobRecordType {
  // 结束时间
  endTime: number;
  // 应用级配置
  appConfig?: object;
  // 应用基础信息
  application: ApplicationType;
  // 备注
  comment: string;
  // 组件实例
  components?: ComponentType[];
  // 配置
  config?: object;
  // 创建时间
  createTime: number;
  // 应用实例id
  id: number;
  // 命名空间
  namespace?: string;
  // 执行器id
  owner?: number;
  // 系统id
  sid?: number;
  // 开始时间（更新时间）
  startTime: number;
  // 状态
  status: number;
  // 系统名称
  systemName?: string;
  // 任务操作类型
  operateType: JobOperateType;
}
export interface JobItem {
  current: JobRecordType | null;
  jid: number;
  target: JobRecordType;
}

export interface ComponentItem {
  formData: ComponentType;
  schema: RJSFSchema;
  uiSchema?: UiSchema;
}

export interface IGetLogTableParams {
  // 任务id
  jid?: number;
  // 组件实例id
  cid?: number;
  // 秒级时间戳过滤（负数为时间戳以前，正数为）
  timestamp?: number;
  // 排序方式
  sort?: string;
  // 是否返回对应过滤条件下的数据数量
  count?: string;
}

/**
 * @interface IGetLogParams
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetLogParams extends IGetLogTableParams {
  offset: number;
  limit: number;
}

export interface JobLogItem {
  // 任务日志id
  jlid?: number;
  // 任务id
  jid?: number;
  // 组件实例id
  cid?: number;
  // 应用实例id
  aiid?: number;
  // 应用标识
  aname: string;
  // 应用名称
  title: string;
  // 组件名称
  cname: string;
  // 错误码
  code: number;
  // 日志信息
  msg: string;
  // 日志记录日期时间戳
  time: number;
  // 错误码描述
  description: string;
}

/**
 * @interface IGetConfigTemplateParams
 * @param offset 索引
 * @param limit 页数
 * @param l 标签过滤器中标签
 * @param lc 标签过滤器类型(0或1或不设置时为'或'关系，2为'与关系')
 * @param v 版本过滤器
 * @param vt 版本过滤器类型(0或1或不设置时为精准匹配,2为patch版本匹配,3为大于等于当前版本匹配)
 * @param aname 应用名
 * @param count 是否返回对应过滤条件下的数据数量
 */
export interface IGetConfigTemplateParams {
  offset: number;
  limit: number;
  l?: string[];
  lc?: number;
  v: string;
  vt: number;
  aname: string;
  count?: string;
}

/**
 * 配置模板类型
 */
export interface ConfigTemplateItem {
  // 应用名称
  aname: string;
  // 应用版本
  aversion: string;
  // 配置内容
  config?: object;
  // 标签
  labels: string[];
  // 模板描述
  tdescription: string;
  // 模板id
  tid?: number;
  // 模板名称
  tname: string;
  // 模板版本
  tversion: string;
}

export interface ISortServiceParams {
  name: string;
  version: string;
  title?: string;
}

export interface IGetDependenciesListParams {
  // 服务名称
  name: string;
  // 服务版本
  version: string;
  // 是否已确认
  select: boolean;
}

export interface DependenciesListItem {
  aid: number;
  name: string;
  title: string;
  version: string;
  versions: string[];
  installed: boolean | undefined;
  select: boolean;
  dependencies: DependenciesServiceItem[];
}

export interface DependenciesServiceItem {
  name: string;
  title: string;
}

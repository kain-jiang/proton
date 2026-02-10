import { ConfigEditStatusEnum } from "../../../components/suite-management/suite-config/helper";
import { ConnectInfoServices } from "../../../components/suite-management/suite-deploy/operation-service/type";
import { ServiceConfigStatusEnum } from "../../../components/suite-management/suite-deploy/type";
import { TaskConfigStatusEnum } from "../../../components/suite-management/task-monitor/type";
import { JobType } from "../../../core/suite-management/suite-deploy";
import { ServiceItem } from "../../service-management/service-deploy/declare";
import { RJSFSchema, UiSchema } from "@rjsf/utils";

/**
 * @interface IGetSuiteTableParams
 * @param name 搜索字段-服务名称
 * @param title 搜索字段-服务标识
 * @param status 筛选字段-状态类型
 */
export interface IGetSuiteTableParams {
  title?: string;
  name?: string;
  status?: ServiceConfigStatusEnum[];
}

/**
 * @interface IGetSuiteParams
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetSuiteParams extends IGetSuiteTableParams {
  offset: number;
  limit: number;
}

export interface ServiceSchemaItem extends ServiceItem {
  formData?: RJSFSchema;
  editStatus?: ConfigEditStatusEnum;
  schema?: RJSFSchema;
  uiSchema?: UiSchema;
}

export interface SuiteConfig {
  apps: ServiceSchemaItem[];
  pcomponents: Array<{ type: ConnectInfoServices }>;
}

export interface SuiteItem {
  config: SuiteConfig;
  createTime?: number;
  description?: string;
  endTime?: number;
  jid: number;
  jname: string;
  mversion: string;
  namespace?: string;
  processed?: number;
  sid?: number;
  startTime?: number;
  status: number;
  systemName?: string;
  total?: number;
  title?: string;
}

export interface ApplicationItem {
  // 应用包描述
  description?: string;
  // 应用标识
  mname: string;
  // 应用名称
  title: string;
  // 应用版本
  mversion: string;
}

/**
 * @interface IGetApplicationParams
 * @param offset 索引
 * @param limit 页数
 * @param name 名称筛选
 * @param nowork 是否过滤已安装应用
 */
export interface IGetApplicationParams {
  offset: number;
  limit: number;
  name?: string;
  nowork?: boolean;
}

export interface SuiteManifestsItem {
  config: SuiteConfig;
  description: string;
  mname: string;
  mversion: string;
}

export interface ICreateComposeJobParams {
  jname: string;
  mversion?: string;
  description: string;
  config: {
    apps: Partial<ServiceSchemaItem>[];
    pcomponents: null;
  };
  sid?: number;
}

/**
 * @interface IGetJobTableParams
 * @param name 搜索字段-任务标识
 * @param title 搜索字段-任务名称
 * @param status 搜索字段-任务状态
 * @param type 搜索字段-任务类型
 * @param sid 搜索字段-系统空间id
 */
export interface IGetJobTableParams {
  name?: string;
  title?: string;
  status?: TaskConfigStatusEnum[];
  type?: JobType;
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

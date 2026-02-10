import { VerifyResultEnum } from "../../../components/service-management/task-monitor/task-detail/verify-record/type";

export interface VerifyInfoType {
  // 验证结束时间
  verifyEndTime: string;
  // 验证结果
  verifyResult: VerifyResultEnum;
}

export interface DataSchemaVerifyItem extends VerifyInfoType {
  // 数据验证列表id
  did: number;
}

export interface FuncVerifyItem extends VerifyInfoType {
  // 功能验证列表id
  fid: number;
}

export interface VerifyItemType {
  // 数据验证列表
  dataSchemaVerifyList: DataSchemaVerifyItem[];
  // 功能验证列表
  funcVerifyList: FuncVerifyItem[];
}

/**
 * @interface IGetFuncDetailParams
 * @param fid 功能验证记录id
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetFuncDetailParams {
  fid: number;
  offset: number;
  limit: number;
}

export interface FuncDetailItem {
  // 测试用例id
  tid: number;
  // 测试描述
  testDescription: string;
  // 测试功能名称
  testFunctionName: string;
  // 测试结果
  testResult: string;
  // 时间
  time: string;
}

/**
 * @interface IGetDataDetailParams
 * @param did 数据验证记录id
 * @param offset 索引
 * @param limit 页数
 */
export interface IGetDataDetailParams {
  did: number;
  offset: number;
  limit: number;
}

export interface DataDetailItem {
  // 测试用例id
  tid: number;
  // 测试的服务名称
  serviceName: string;
  // 测试结果
  testResult: string;
  // 测试结果详情
  testResultDetail: string;
}

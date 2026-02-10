import { request } from "../../../tools/request";
import {
  VerifyItemType,
  IGetFuncDetailParams,
  FuncDetailItem,
  IGetDataDetailParams,
  DataDetailItem,
} from "./declare";
import { paramsSerializer } from "../../../components/common/utils/request";

class Verify {
  url = "/api/deploy-installer/v1/verification";

  /**
   * 获取指定job所有的验证记录与结果
   */
  get(jid: number): Promise<VerifyItemType> {
    return request.get(`${this.url}/${jid}`);
  }

  /**
   * 获取某次功能验证详情
   */
  getFuncDetail(params: IGetFuncDetailParams): Promise<FuncDetailItem[]> {
    return request.get(`${this.url}/function?${paramsSerializer(params)}`);
  }

  /**
   * 获取某次数据验证详情
   */
  getDataDetail(params: IGetDataDetailParams): Promise<DataDetailItem[]> {
    return request.get(`${this.url}/database?${paramsSerializer(params)}`);
  }
}

export const verify = new Verify();

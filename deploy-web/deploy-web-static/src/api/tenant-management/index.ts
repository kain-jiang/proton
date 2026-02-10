import { paramsSerializer } from "../../components/common/utils/request";
import { request } from "../../tools/request";
import { IGetSystemParams, SystemConfig } from "./declare";

// 系统空间（命名空间）
class System {
  url = "/api/deploy-installer/v1/system";

  // 获取所有系统空间
  get(params: IGetSystemParams): Promise<SystemConfig[]> {
    return request.get(`${this.url}?${paramsSerializer(params)}`);
  }

  // 获取系统空间信息
  getSystemInfo(sid: number): Promise<SystemConfig> {
    return request.get(`${this.url}/${sid}`);
  }
}

export const system = new System();

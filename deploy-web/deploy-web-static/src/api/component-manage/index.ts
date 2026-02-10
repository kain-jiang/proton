import {
  ComponentData,
  ComponentName,
  SOURCE_TYPE,
} from "../../components/component-management/helper";
import {
  ConnectInfoData,
  ConnectInfoServices,
} from "../../components/connect-info-management/helper";
import { paramsSerializer } from "../../components/common/utils/request";
import { request } from "../../tools/request";
import {
  ComponentListItem,
  ConnectInfoListItem,
  IGetComponentListParams,
  IGetConnectInfoListParams,
} from "./declare";

class ComponentManage {
  url = "/api/deploy-installer/v1/components";

  /**
   * 获取连接信息
   * @description 特殊处理rds的连接信息，根据后端返回的admin_key，分离出admin_user和admin_passwd
   */
  async getConnectInfo(
    service: string,
    name: string,
    sid?: number
  ): Promise<ConnectInfoData> {
    const type =
      service === ConnectInfoServices.ETCD
        ? "etcd"
        : service === ConnectInfoServices.PolicyEngine
        ? "policyengine"
        : service;
    if (type !== ConnectInfoServices.RDS) {
      return request.get(
        `${this.url}/info/${type}${sid ? "/" + sid : ""}/${name}`
      );
    }
    const result: ConnectInfoData = await request.get(
      `${this.url}/info/${type}${sid ? "/" + sid : ""}/${name}`
    );
    if (result?.info?.source_type === SOURCE_TYPE.EXTERNAL) {
      if (result?.info?.admin_key) {
        const [admin_user, admin_passwd] = decodeURIComponent(
          atob(result.info.admin_key)
        ).split(":");
        return {
          ...result,
          info: {
            ...result.info,
            auto_create_database: true,
            admin_user,
            admin_passwd,
          },
        };
      } else {
        return {
          ...result,
          info: {
            ...result.info,
            auto_create_database: false,
            admin_user: "",
            admin_passwd: "",
          },
        };
      }
    }
    return result;
  }

  /**
   * 配置连接信息
   */
  putConnectInfo(service: string, params: ConnectInfoData): Promise<null> {
    const serviceType =
      service === ConnectInfoServices.ETCD
        ? "etcd"
        : service === ConnectInfoServices.PolicyEngine
        ? "policyengine"
        : service;
    return request.put(`${this.url}/info/${serviceType}`, params);
  }

  /**
   * 获取内置组件信息
   */
  getComponentInfo(type: string, name: string): Promise<ComponentData> {
    return request.get(`${this.url}/release/${type}/${name}`);
  }

  /**
   * 配置内置组件信息
   */
  putComponentInfo(component: string, params: ComponentData): Promise<null> {
    return request.put(`${this.url}/release/${component}`, params);
  }

  /**
   * 获取内置组件列表信息
   */
  getComponentList(
    params: IGetComponentListParams
  ): Promise<ComponentListItem[]> {
    return request.get(`${this.url}/release?${paramsSerializer(params)}`);
  }

  /**
   * 获取连接信息列表
   */
  getConnectInfoList(
    params: IGetConnectInfoListParams
  ): Promise<ConnectInfoListItem[]> {
    return request.get(`${this.url}/info?${paramsSerializer(params)}`);
  }
}

export const componentManage = new ComponentManage();

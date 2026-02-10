import { ConfigData, DeployConfig } from "..";

export interface Props {
  /**
   * 改变模板类型触发
   */
  changeDataBaseStorageType: (templateType: string) => any;

  /**
   * 配置数据
   */
  configData: ConfigData;

  /**
   * 改变部署配置信息
   */
  updateDeploy: (value: Partial<DeployConfig>) => void;
}

import { PolicyEngineConnectInfo, ConfigData } from "../../index";

export interface Props {
  // 模板类型
  dataBaseStorageType: string;

  // 所有数据
  configData: ConfigData;

  // 更新MariaDB的数据
  updateConnectInfo: (o: PolicyEngineConnectInfo) => void;

  // 删除组件
  onDeleteResource: (services) => void;

  // 更新连接配置form实例
  updateConnectInfoForm: (form) => void;
}

export interface State {
  // 编辑中的数据
  policy_engine: PolicyEngineConnectInfo;
}

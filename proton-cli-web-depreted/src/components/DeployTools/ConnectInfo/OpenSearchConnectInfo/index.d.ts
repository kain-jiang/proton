import {
  OpensearchConnectInfo,
  ConfigData,
  ConnectInfoValidateState,
} from "../../index";

export interface Props {
  // 模板类型
  dataBaseStorageType: string;

  // 所有数据
  configData: ConfigData;

  // 连接配置校验状态
  connectInfoValidateState: ConnectInfoValidateState;

  // 更新MariaDB的数据
  updateConnectInfo: (o: OpensearchConnectInfo) => void;

  // 删除组件
  onDeleteResource: (services) => void;

  // 更新连接配置form实例
  updateConnectInfoForm: (form) => void;

  // 更新连接配置校验状态
  updateConnectInfoValidateState: (
    value: Partial<ConnectInfoValidateState>,
  ) => void;
}

export interface State {
  // 编辑中的数据
  opensearch: OpensearchConnectInfo;
}

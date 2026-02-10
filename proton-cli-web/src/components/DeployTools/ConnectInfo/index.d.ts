import {
  Services,
  ConfigData,
  ServicesNodes,
  ConnectInfoValidateState,
} from "../index";

declare namespace DataBaseConfigType {
  interface Props {
    // 模板类型
    dataBaseStorageType: string;

    // 所有数据
    configData: ConfigData;

    // 连接配置部分配置项校验状态
    connectInfoValidateState: ConnectInfoValidateState;

    // 更新源数据
    onUpdateConnectInfo: (key, o) => void;

    // 删除组件
    onDeleteResource: (services) => void;

    // 添加组件
    onAddResource: (services) => void;

    // 更新连接配置form实例
    updateConnectInfoForm: (value) => void;

    // 更新连接配置校验状态
    updateConnectInfoValidateState: (
      value: Partial<ConnectInfoValidateState>,
    ) => void;
  }

  interface State extends Services, ServicesNodes {}
}

import { ValidateState } from "../../helper";
import { ProtonDataBase, NodeInfo, ConfigData } from "../../index";

declare namespace MonitorType {
  interface Props {
    // 模板类型
    dataBaseStorageType: string;

    // 所有数据
    configData: ConfigData;

    // title
    service: {
      key: string;
      name: string;
    };

    // prometheus节点校验状态
    monitorNodesValidateState: ValidateState;

    // 更新MariaDB的数据
    onUpdateMonitorData: (config: ProtonDataBase, dataBaseStorageType) => void;

    // 删除Monitor
    onDeleteMonitorConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;

    // 更新proton monitor节点校验状态
    updateMonitorNodesValidateState: () => void;
  }

  interface State {
    // MariaDB编辑中的数据
    monitorConfig: ProtonDataBase;

    // monitorNodes 节点
    monitorNodes: Array<NodeInfo>;
  }
}

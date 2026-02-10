import { ValidateState } from "../../helper";
import { ProtonServiceConfig, NodeInfo, ConfigData } from "../../index";

declare namespace ServiceType {
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
    prometheusNodesValidateState: ValidateState;

    // grafana节点校验状态
    grafanaNodesValidateState: ValidateState;

    // 更新MariaDB的数据
    onUpdateServiceData: (
      config: ProtonServiceConfig,
      dataType: string,
    ) => void;

    // 删除服务
    onDeleteServiceConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;

    // 更新grafana节点校验状态
    updateGrafanaNodesValidateState: () => void;

    // 更新prometheus节点校验状态
    updatePrometheusNodesValidateState: () => void;
  }

  interface State {
    // 编辑中的数据
    serviceConfig: ProtonServiceConfig;

    // 节点
    serviceNodes: Array<NodeInfo>;
  }
}

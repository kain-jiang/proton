import { ValidateState } from "../helper";
import { Services, ConfigData, ServicesNodes } from "../index";

declare namespace DataBaseConfigType {
  interface Props {
    // 模板类型
    dataBaseStorageType: string;

    // 可添加服务集合
    addableServices: Array<{
      key: string;
      name: string;
    }>;

    // 可选服务集合
    selectableServices: Array<{
      key: string;
      name: string;
    }>;

    // 所有数据
    configData: ConfigData;

    // grafana节点校验状态
    grafanaNodesValidateState: ValidateState;

    // prometheus节点校验状态
    prometheusNodesValidateState: ValidateState;

    // proton monitor节点校验状态
    monitorNodesValidateState: ValidateState;

    // 更新源数据
    onUpdateDataBaseConfig: (value) => void;

    // 删除服务
    onDeleteService: (value: string) => void;

    // 新增服务
    onAddService: (value: string) => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;

    // 更新grafana节点校验状态
    updateGrafanaNodesValidateState: () => void;

    // 更新prometheus节点校验状态
    updatePrometheusNodesValidateState: () => void;

    // 更新proton monitor节点校验状态
    updateMonitorNodesValidateState: () => void;
  }

  interface State extends Services, ServicesNodes {}
}

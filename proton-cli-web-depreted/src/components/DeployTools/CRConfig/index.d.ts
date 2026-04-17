import { NodeInfo, ConfigData, CRConfig, ExternalCRConfig } from "../index";
import { CRType, ValidateState } from "../helper";

declare namespace CRConfigType {
  interface Props {
    // 当前的CR的类型
    cRType: CRType;

    // 数据库类型
    dataBaseStorageType: string;

    // 部署配置
    configData: ConfigData;

    // 节点校验状态
    crNodesValidateState: ValidateState;

    // 更新数据源
    onUpdateCRConfig: (value) => void;

    // 更新crType
    onUpDateCRTypeConfig: (value) => void;

    // 更新仓库配置form实例
    updateCRForm: (value) => void;

    // 更新节点校验状态
    updateCRNodesValidateState: () => void;
  }

  interface State {
    // 本地CR 配置
    crConfig: CRConfig;

    // 外置CR配置
    externalCRConfig: ExternalCRConfig;

    //部署节点
    nodes: Array<NodeInfo>;

    // 当前应用的CR
    selectCRType: CRType;
  }
}

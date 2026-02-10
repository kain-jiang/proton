import { ValidateState } from "../helper";
import { NodeInfo, NetworkInfo, ConfigData, DeployConfig } from "../index.d";
declare namespace NetWorkConfigType {
  interface Props {
    // 节点信息
    configData: ConfigData;

    // 数据库类型
    dataBaseStorageType: string;

    // 部署节点校验状态
    networkNodesValidateState: ValidateState;

    // 更新数据源
    onUpdateNetworkConfig: (value) => void;

    // 更新网络配置form实例
    updateNetworkForm: (value) => void;

    // 更新网络配置节点校验状态
    updateNetworkNodesValidateState: () => void;

    updateDeploy: (deploy: DeployConfig) => void;
  }
  interface State {
    // 节点信息
    nodes: Array<NodeInfo>;

    // 网络配置信息
    networkInfo: NetworkInfo;

    // 部署配置
    deploy: DeployConfig;
  }
}

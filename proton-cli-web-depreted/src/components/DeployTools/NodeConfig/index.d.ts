import { ConfigData, NodeInfo, ChronyConfig, FirewallConfig } from "../index.d";
import { ValidateState, nodeInfoCheckResult } from "../helper";

declare namespace NodeConfigType {
  interface Props {
    // 节点信息
    configData: ConfigData;

    setNextStepButtonDisable: (value) => void;

    //账户信息默认值
    accountInfo: {
      // 节点账户
      sshAccount: string;

      // 节点密码
      sshPassword: string;
    };

    // // IP配置
    // ipConfig: {
    //   // 内部网段
    //   internal_cidr: string;

    //   // 网卡
    //   internal_nic: string;
    // };
    // 数据库类型
    dataBaseStorageType: string;

    // 节点配置校验状态
    nodesValidateState: ValidateState;

    //数据更新事件
    updateNodesInfo: (nodesInfo: Array<NodeInfo>) => void;

    // 更新账户信息
    updateSSHInfo: (value) => void;

    // 更新内部网段和网卡
    updateNicCidr: (allvalues) => void;

    // 更新网络配置信息
    updateNetworkConfig: (value) => void;

    // 更新时间服务器配置
    updateChrony: (value: ChronyConfig) => void;

    // 更新防火墙配置
    updateFirewall: (value: FirewallConfig) => void;

    // 更新节点配置form实例
    updateNodeForm: (value) => void;

    // 更新节点校验状态
    updateNodesValidateState: () => void;
  }

  interface State {
    /**
     * 时间服务器配置
     */
    chrony: ChronyConfig;

    /**
     * 防火墙配置
     */
    firewall: FirewallConfig;

    /**
     * 节点信息
     */
    nodesInfo: Array<NodeInfo>;

    /**
     * 抽屉开关
     */
    addNodeStatus: boolean;

    /**
     * 正在编辑的节点索引
     */
    isEditingNodeIndex: number;

    /**
     * 正在编辑的是新增节点
     */
    isCreatingNode: boolean;

    /**
     * 验证表单非空
     */
    validator: nodeInfoCheckResult;

    // // IP配置
    // ipConfig: {
    //   // 内部网段
    //   internal_cidr: string;

    //   // 网卡
    //   internal_nic: string;
    // };

    /**
     * IP配置
     */
    CSIPFamily: Array<string>;

    // 是否开启双栈能力
    enableDualStack: boolean;
  }
}

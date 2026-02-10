import * as React from "react";
import {
  nodeInfoCheckResult,
  checkIPv4,
  checkIPv6,
  IP_Family,
  DefaultConfigData,
  checkNodeName,
} from "../helper";
import { NodeConfigType } from "./index";
import { FormInstance, message } from "@aishutech/ui";
import { isEqual } from "lodash";

export default class NodeConfigBase extends React.Component<
  NodeConfigType.Props,
  NodeConfigType.State
> {
  state = {
    nodesInfo: [],
    addNodeStatus: false,
    isEditingNodeIndex: -1,
    isCreatingNode: false,
    validator: nodeInfoCheckResult.Normal,
    // ipConfig: {
    //   internal_cidr: "",
    //   internal_nic: "",
    // },
    CSIPFamily: [IP_Family.ipv4],
    chrony: DefaultConfigData.chrony,
    firewall: DefaultConfigData.firewall,
    enableDualStack: false,
  };

  nodeInfo = {
    name: "",
    ipv4: "",
    ipv6: "",
  };

  nodeForm = {
    serverForm: React.createRef<FormInstance>(),
    acountInfoForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const { nodesInfo, networkInfo, chrony, firewall } = this.props.configData;
    this.setState({
      nodesInfo,
      chrony,
      firewall,
      CSIPFamily: this.getIPFamily(networkInfo.ipFamilies),
      enableDualStack: !!networkInfo.enableDualStack,
    });
    this.props.updateNodeForm(this.nodeForm);
  }

  componentDidUpdate(
    prevProps: Readonly<NodeConfigType.Props>,
    prevState: Readonly<NodeConfigType.State>,
    snapshot?: any,
  ): void {
    const { nodesInfo, networkInfo, chrony, firewall } = this.props.configData;
    if (
      prevProps.configData.nodesInfo != nodesInfo ||
      prevProps.configData.chrony !== chrony ||
      prevProps.configData.firewall !== firewall
    ) {
      this.setState({
        chrony,
        firewall,
        nodesInfo,
        CSIPFamily: this.getIPFamily(networkInfo.ipFamilies),
        enableDualStack: !!networkInfo.enableDualStack,
      });
    }
    // if (prevProps.ipConfig != this.props.ipConfig) {
    //   this.setState({
    //     ipConfig: this.props.ipConfig,
    //   });
    // }
  }

  /**
   * 获取k8s IP协议栈
   * @param nodesInfo 节点信息
   * @returns
   */
  private getIPFamily(ipFamilies) {
    return ipFamilies ? ipFamilies : [IP_Family.ipv4];
  }

  /**
   * 切换IP类型
   * @param value
   */
  public onChangeCSIPFamily(value) {
    const enableDualStack = value[0] === IP_Family.dualStack ? true : false;
    this.setState(
      {
        isEditingNodeIndex: -1,
        CSIPFamily: value,
        enableDualStack,
      },
      () => {
        this.props.updateNetworkConfig({ enableDualStack, ipFamilies: value });
      },
    );
  }

  /**
   * 切换是否开启双栈能力
   * @param value
   */
  public onChangeEnableDualStack(value) {
    this.setState(
      {
        enableDualStack: value,
      },
      () => {
        this.props.updateNetworkConfig({ enableDualStack: value });
      },
    );
  }

  /**
   * 打开添加抽屉
   */
  public onAddNode() {
    this.setState({
      addNodeStatus: true,
    });
  }

  /**
   * 编辑ipV4
   */
  public onChangeRowData(value, allValues) {
    this.nodeInfo = { ...this.nodeInfo, ...allValues };
  }

  /**
   * 添加当前节点
   */
  public onSaveEdit(record) {
    this.props.setNextStepButtonDisable(false);
    if (!this.checkValidator(this.nodeInfo)) {
      return;
    } else {
      this.setState(
        {
          nodesInfo: this.state.nodesInfo.map((nodeInfo, index) => {
            return index === this.state.isEditingNodeIndex
              ? {
                  ...this.nodeInfo,
                  name: this.nodeInfo.name
                    ? this.nodeInfo.name
                    : this.getNodeName(this.nodeInfo),
                }
              : nodeInfo;
          }),
        },
        () => {
          this.props.updateNodesInfo(this.state.nodesInfo);
          this.props.updateNodesValidateState();
        },
      );

      this.setState({
        isEditingNodeIndex: -1,
        isCreatingNode: false,
        validator: nodeInfoCheckResult.Normal,
      });
      this.nodeInfo = {
        name: "",
        ipv4: "",
        ipv6: "",
      };
    }
  }

  /**
   * 取消本次添加
   */
  public onCancelEditing() {
    this.props.setNextStepButtonDisable(false);
    if (this.state.isCreatingNode) {
      this.setState({
        nodesInfo: this.state.nodesInfo.filter((nodeInfo, index) => {
          return index < this.state.nodesInfo.length - 1;
        }),
        isCreatingNode: false,
      });
    }
    this.setState({
      isEditingNodeIndex: -1,
      validator: nodeInfoCheckResult.Normal,
    });
    this.nodeInfo = {
      name: "",
      ipv4: "",
      ipv6: "",
    };
  }

  /**
   * 删除节点
   */
  public onDeleteNode(deleteNodeIndex) {
    this.setState(
      {
        nodesInfo: this.state.nodesInfo.filter((value, index) => {
          return index != deleteNodeIndex;
        }),
      },
      () => {
        this.props.updateNodesInfo(this.state.nodesInfo);
      },
    );
  }

  /**
   * 触发编辑
   * @param node 被编辑的节点
   */
  public onClickEdit(node, index) {
    this.props.setNextStepButtonDisable(true);
    this.setState({
      isEditingNodeIndex: index,
    });
    this.nodeInfo = node;
  }

  /**
   * 当前节点的被编辑状态
   * @param node 当前节点
   * @returns 被编辑的状态
   */
  public getNodeEditingStatus(index) {
    return Boolean(index === this.state.isEditingNodeIndex);
  }

  /**
   * 添加节点
   */
  public addNodeInfoData() {
    this.props.setNextStepButtonDisable(true);
    const index = this.state.nodesInfo.length;
    this.setState({
      nodesInfo: [
        ...this.state.nodesInfo,
        {
          name: "",
          ipv4: "",
          ipv6: "",
        },
      ],

      isEditingNodeIndex: index,
      isCreatingNode: true,
    });
    this.nodeInfo = {
      name: "",
      ipv4: "",
      ipv6: "",
    };
  }

  /**
   * 检查内部ip是否冲突
   * @param value 当前节点信息
   * @param nodesInfo 所有节点的信息
   * @returns
   */
  checkInternalIPIsConflict(value, nodesInfo) {
    const nodes = nodesInfo.filter(
      (node, index) => index !== this.state.isEditingNodeIndex,
    );
    const internalIPs = nodes.map((node) => node.internal_ip);

    return internalIPs.includes(value.internal_ip);
  }

  /**
   * 检查输入合法性
   * @param nodeInfo
   * @returns
   */
  private checkValidator(nodeInfo) {
    const { CSIPFamily, nodesInfo } = this.state;
    if (nodeInfo.internal_ip && !/^\s+$/g.test(nodeInfo.internal_ip)) {
      if (
        !checkIPv4(nodeInfo.internal_ip) &&
        !checkIPv6(nodeInfo.internal_ip)
      ) {
        this.setState({
          validator: nodeInfoCheckResult.InternalIPNotIPv4OrIPv6,
        });
        return false;
      } else if (this.checkInternalIPIsConflict(nodeInfo, nodesInfo)) {
        this.setState({
          validator: nodeInfoCheckResult.InternalIPConflict,
        });
        return false;
      }
    }
    switch (true) {
      case !nodeInfo.ipv4 && !nodeInfo.ipv6:
        this.setState({
          validator: nodeInfoCheckResult.ExistIpv6Ipv4,
        });
        return false;
      case nodeInfo.ipv4 && !checkIPv4(nodeInfo.ipv4):
        this.setState({
          validator: nodeInfoCheckResult.ValidatorIpv4,
        });
        return false;
      case nodeInfo.ipv6 && !checkIPv6(nodeInfo.ipv6):
        this.setState({
          validator: nodeInfoCheckResult.validatorIpv6,
        });
        return false;
      case this.checkRepeatNode(nodeInfo):
        this.setState({
          validator: nodeInfoCheckResult.hasIpv6Ipv4,
        });
        return false;
      case nodeInfo.name && !checkNodeName(nodeInfo.name):
        this.setState({
          validator: nodeInfoCheckResult.validatorNodeName,
        });
        return false;
      case this.checkRepeatNodeName(nodeInfo):
        this.setState({
          validator: nodeInfoCheckResult.hasRepeatNodeName,
        });
        return false;
      case !!this.checkNodeIPByK8SFamily(nodeInfo, [...CSIPFamily]):
        this.setState({
          validator: this.checkNodeIPByK8SFamily(nodeInfo, [...CSIPFamily]),
        });
        return false;
      default:
        return true;
    }
  }

  /**
   * 检查是否重复
   * @param value 当前节点
   * @returns
   */
  private checkRepeatNode(value) {
    let data = this.state.nodesInfo.filter(
      (nodeInfo, index) => index !== this.state.isEditingNodeIndex,
    );
    if (data.length === 0) {
      return false;
    } else {
      return !!data.find(
        (node) =>
          (node.ipv4 && node.ipv4 === value.ipv4) ||
          (node.ipv6 && node.ipv6 === value.ipv6),
      );
    }
  }

  /**
   * 检查是否重复节点名称
   * @param value 当前节点
   * @returns
   */
  private checkRepeatNodeName(value) {
    let data = this.state.nodesInfo.filter(
      (nodeInfo, index) => index !== this.state.isEditingNodeIndex,
    );
    if (data.length === 0 || !value.name) {
      return false;
    } else {
      return !!data.find((node) => node.name === value.name);
    }
  }

  /**
   * 检查协议栈
   * @param nodeInfo 节点信息
   * @param CSIPFamily k8s协议栈
   */
  private checkNodeIPByK8SFamily(nodeInfo, CSIPFamily) {
    if (isEqual(CSIPFamily, [IP_Family.ipv4]) && !nodeInfo.ipv4) {
      return nodeInfoCheckResult.K8SIPv4NodeIPv4Empty;
    } else if (isEqual(CSIPFamily, [IP_Family.ipv6]) && !nodeInfo.ipv6) {
      return nodeInfoCheckResult.K8SIPv6NodeIPv6Empty;
    } else if (
      isEqual(CSIPFamily, [IP_Family.dualStack]) &&
      (!nodeInfo.ipv4 || !nodeInfo.ipv6)
    ) {
      return nodeInfoCheckResult.K8SIPv46NodeIP46Empty;
    } else {
      return nodeInfoCheckResult.Normal;
    }
  }

  /**
   * 获取名称
   */
  private getNodeName(nodeInfo) {
    // 节点名称需要符合dns1035规则
    // [a-z]([-a-z0-9]*[a-z0-9])?$
    let v4Name = "",
      v6Name = "";
    if (nodeInfo.ipv4) {
      const tmp = nodeInfo.ipv4.split(".");
      v4Name = `-${tmp[2]}-${tmp[3]}`;
    }
    if (nodeInfo.ipv6) {
      let tmp = nodeInfo.ipv6.split(":").pop();
      if (tmp.indexOf(".") !== -1) {
        // 兼容 ipv6 内嵌ipv4版本
        // tmp = tmp.split(".").pop()
        tmp = `${tmp.split(".")[2]}-${tmp.split(".")[3]}`;
      }
      v6Name = `-${tmp}`;
    }
    return `node${v4Name}${v6Name}`;
  }
}

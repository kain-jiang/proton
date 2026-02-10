import * as React from "react";
import {
  getDefaultNodes,
  DefaultConfigData,
  DataBaseStorageType,
  IP_Family,
} from "../helper";
import { NetWorkConfigType } from "./index";
import { FormInstance } from "@aishutech/ui";
import { DeployConfig } from "..";

export default class NetworkConfigBase extends React.Component<
  NetWorkConfigType.Props,
  NetWorkConfigType.State
> {
  state = {
    networkInfo: DefaultConfigData.networkInfo,
    deploy: DefaultConfigData.deploy as DeployConfig,
    nodes: [],
  };

  networkForm = {
    hostNetworkForm: React.createRef<FormInstance>(),
    networkInfoForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const {
      configData: { nodesInfo, networkInfo, deploy },
      dataBaseStorageType,
    } = this.props;
    const masterNode = getDefaultNodes(nodesInfo, networkInfo.master);
    this.props.updateNetworkForm(this.networkForm);
    this.setState(
      {
        networkInfo: {
          ...networkInfo,
          master: masterNode.map((node) => node.name),
          hostNetwork:
            dataBaseStorageType === DataBaseStorageType.DepositKubernetes
              ? {
                  ...networkInfo,
                }
              : {
                  ...networkInfo.hostNetwork,
                  podNetworkCidr: this.getDefaultPodNetworkCidr(networkInfo),
                  serviceCidr: this.getDefaultServiceCidr(networkInfo),
                  ipv4Interface:
                    networkInfo.ipFamilies &&
                    networkInfo.ipFamilies[0] === IP_Family.dualStack
                      ? networkInfo.hostNetwork.ipv4Interface
                      : "",
                  ipv6Interface:
                    networkInfo.ipFamilies &&
                    networkInfo.ipFamilies[0] === IP_Family.dualStack
                      ? networkInfo.hostNetwork.ipv6Interface
                      : "",
                },
        },
        nodes: masterNode,
        deploy: { ...deploy },
      },
      () => {
        this.props.onUpdateNetworkConfig({
          networkInfo: this.state.networkInfo,
        });
        if (dataBaseStorageType === DataBaseStorageType.Standard) {
          this.networkForm.hostNetworkForm.current.setFieldsValue({
            ...this.state.networkInfo.hostNetwork,
          });
          this.networkForm.networkInfoForm.current.setFieldsValue({
            ...this.state.networkInfo,
          });
        }
      },
    );
  }

  /**
   * 更新网络配置
   * @param config 更新网络配置
   */
  public onChangeNetworkConfig(config) {
    this.setState(
      {
        networkInfo: {
          ...this.state.networkInfo,
          hostNetwork: {
            ...this.state.networkInfo.hostNetwork,
            ...config,
          },
        },
      },
      () => {
        this.props.onUpdateNetworkConfig({
          networkInfo: this.state.networkInfo,
        });
      },
    );
  }

  /**
   * 更新数据路径
   * @param config 数据路径配置
   */
  public onChangeDataDir(config) {
    this.setState(
      {
        networkInfo: {
          ...this.state.networkInfo,
          ...config,
        },
      },
      () => {
        this.props.onUpdateNetworkConfig({
          networkInfo: this.state.networkInfo,
        });
      },
    );
  }

  /**
   * 部署节点
   * @param nodes 部署节点
   */
  public onChangeMasterNode(nodes) {
    this.setState(
      {
        nodes,
        networkInfo: {
          ...this.state.networkInfo,
          master: nodes.map((value) => value.name),
        },
      },
      () => {
        this.props.onUpdateNetworkConfig({
          networkInfo: this.state.networkInfo,
        });
        this.props.updateNetworkNodesValidateState();
      },
    );
  }

  public onChangeCSPlugins(e, val) {
    const { networkInfo } = this.state;
    const addons = e.target.checked
      ? [...networkInfo.addons, val]
      : networkInfo?.addons?.filter((item) => item !== val);
    this.setState(
      {
        networkInfo: {
          ...networkInfo,
          addons,
        },
      },
      () => {
        this.props.onUpdateNetworkConfig({
          networkInfo: this.state.networkInfo,
        });
      },
    );
  }

  private getDefaultPodNetworkCidr(networkInfo) {
    switch (networkInfo.ipFamilies && networkInfo.ipFamilies[0]) {
      case IP_Family.ipv6:
        return "fc00:b36f:c1c3:2000::/64";
      case IP_Family.dualStack:
        return "192.169.0.0/16,fc00:b36f:c1c3:2000::/64";
      default:
        return "192.169.0.0/16";
    }
  }

  private getDefaultServiceCidr(networkInfo) {
    switch (networkInfo.ipFamilies && networkInfo.ipFamilies[0]) {
      case IP_Family.ipv6:
        return "fc01:b36f:c1c3:1000::/108";
      case IP_Family.dualStack:
        return "10.96.0.0/12,fc01:b36f:c1c3:1000::/108";
      default:
        return "10.96.0.0/12";
    }
  }

  /**
   * 更新网络配置
   * @param config 更新网络配置
   */
  public onChangeDeployConfig(config) {
    this.setState(
      {
        deploy: {
          ...this.state.deploy,
          ...config,
        },
      },
      () => {
        this.props.updateDeploy(this.state.deploy);
      },
    );
  }
}

import * as React from "react";
import { FormInstance } from "@aishutech/ui";
import { eq, isEqual, isArray } from "lodash";
import { getDefaultNodes } from "../../helper";
import { ECephType } from "./index";

export default class ECephConfigBase extends React.Component<
  ECephType.Props,
  ECephType.State
> {
  state = {
    ecephConfig: null,
    ecephNodes: [],
  };

  ecephForm = {
    keepalivedForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.ecephForm);
    this.initConfig(
      configData.eceph,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(
    prevProps: Readonly<ECephType.Props>,
    prevState: Readonly<ECephType.State>,
    snapshot?: any,
  ): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.eceph, configData.eceph)) {
      this.initConfig(
        configData.eceph,
        configData.nodesInfo,
        dataBaseStorageType,
      );
    }
  }

  /**
   * 初始化配置
   * @param oldConfig 原始配置
   * @param nodesInfo 原始节点信息配置
   * @param dataBaseStorageType 配置类型
   */
  private initConfig(oldConfig, nodesInfo, dataBaseStorageType) {
    const nodes = oldConfig
      ? getDefaultNodes(
          nodesInfo ? nodesInfo : [],
          isArray(oldConfig?.hosts) ? oldConfig?.hosts : [],
          3,
          false,
          true,
        )
      : [];
    const hosts = nodes.map((node) => node.name);
    let newConfig = oldConfig;
    if (
      isArray(oldConfig?.hosts) &&
      !isEqual(oldConfig?.hosts.sort(), hosts.sort())
    ) {
      newConfig = {
        ...oldConfig,
        hosts,
      };
    }
    this.setState(
      {
        ecephNodes: nodes,
        ecephConfig: newConfig,
      },
      () => {
        this.ecephForm.keepalivedForm.current.setFieldsValue({
          ...this.state.ecephConfig?.keepalived,
        });
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateECephData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新keepalived
   * @param config
   */
  public onChangeECephKeepalived(config) {
    const ecephConfig = {
      ...this.state.ecephConfig,
      keepalived: {
        ...this.state.ecephConfig.keepalived,
        ...config,
      },
    };
    this.setState(
      {
        ecephConfig,
      },
      () => {
        this.props.onUpdateECephData(
          ecephConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新tls
   * @param config
   */
  public onChangeECephTLS(config) {
    const ecephConfig = {
      ...this.state.ecephConfig,
      tls: {
        ...this.state.ecephConfig.tls,
        ...config,
      },
    };
    this.setState(
      {
        ecephConfig,
      },
      () => {
        this.props.onUpdateECephData(
          ecephConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 修改节点
   * @param nodes 节点信息
   * @param config 配置信息
   */
  public onChangeECephNode(nodes, config) {
    this.props.onUpdateECephData(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }
}

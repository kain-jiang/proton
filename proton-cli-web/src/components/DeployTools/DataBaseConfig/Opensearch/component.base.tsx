import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import {
  DataBaseStorageType,
  DefaultConfigData,
  getDefaultNodes,
} from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { OpensearchType } from "./index";

export default class OpensearchConfigBase extends React.Component<
  OpensearchType.Props,
  OpensearchType.State
> {
  state = {
    opensearchConfig: null,
    opensearchNodes: [],
  };

  opensearchForm = {
    modeForm: React.createRef<FormInstance>(),
    dataForm: React.createRef<FormInstance>(),
    replicaForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.opensearchForm);
    this.initConfig(
      configData.opensearch,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(
    prevProps: Readonly<OpensearchType.Props>,
    prevState: Readonly<OpensearchType.State>,
  ): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.opensearch, configData.opensearch)) {
      this.initConfig(
        configData.opensearch,
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
        opensearchNodes: nodes,
        opensearchConfig: newConfig,
      },
      () => {
        this.opensearchForm.modeForm.current.setFieldsValue({
          ...this.state.opensearchConfig,
        });
        this.opensearchForm.dataForm.current.setFieldsValue({
          ...this.state.opensearchConfig,
        });
        if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes) {
          this.opensearchForm.replicaForm.current.setFieldsValue({
            ...this.state.opensearchConfig,
          });
        }
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateOpensearchData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新opensearch的配置
   * @param settings 设置
   */
  public onChangeSettings(settings) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      settings: {
        ...this.state.opensearchConfig?.settings,
        ...settings,
      },
    };
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新opensearch的nfs配置
   */
  public onChangeNFSConfig(type, config) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      extraValues: {
        ...this.state.opensearchConfig?.extraValues,
        storage: {
          ...this.state.opensearchConfig?.extraValues?.storage,
          repo: {
            ...this.state.opensearchConfig?.extraValues?.storage?.repo,
            [type]:
              type === "nfs" && Object.keys(config).includes("enabled")
                ? Object.values(config).includes(true)
                  ? { enabled: true, server: "", path: "" }
                  : { enabled: false }
                : {
                    ...this.state.opensearchConfig?.extraValues?.storage
                      ?.repo?.[type],
                    ...config,
                  },
          },
        },
      },
    };

    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新opensearch的http配置
   * @param http 配置
   */
  public onChangeHttp(http) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      http: {
        ...this.state.opensearchConfig?.http,
        ...http,
      },
    };
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新opensearch的 cluster 配置
   * @param cluster 配置
   */
  public onChangeCluster(cluster) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      cluster: {
        ...this.state.opensearchConfig?.cluster,
        ...cluster,
      },
    };
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 修改资源配置触发
   * @param resources limits | requests
   * @param key
   * @param val
   */
  public onChangeResource(resources, key, val) {
    let opensearchConfig = { ...this.state.opensearchConfig };
    if (resources === "limits") {
      opensearchConfig = {
        ...opensearchConfig,
        resources: {
          ...opensearchConfig?.resources,
          limits: {
            ...opensearchConfig?.resources?.limits,
            [key]: val,
          },
        },
      };
    } else {
      opensearchConfig = {
        ...opensearchConfig,
        resources: {
          ...opensearchConfig?.resources,
          requests: {
            ...opensearchConfig?.resources?.requests,
            [key]: val,
          },
        },
      };
    }
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新OpenSearch的配置
   * @param config 更新的配置
   */
  public onChangeOpensearchConfig(config) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      config: {
        ...this.state.opensearchConfig?.config,
        ...config,
      },
    };
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangeOpensearch(config) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      ...config,
    };
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
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
  public onChangeOpensearchNode(nodes, config) {
    this.props.onUpdateOpensearchData(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }

  /**
   * 是否配置资源限制
   */
  public onChangeOpensearchConfigResources(val) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      resources: val
        ? DefaultConfigData[this.props.service.key].resources
        : undefined,
    };
    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 修改子组件资源配置触发
   * @param key
   * @param val
   */
  public onChangeComponentResource(key, val) {
    let opensearchConfig = { ...this.state.opensearchConfig };

    opensearchConfig = {
      ...opensearchConfig,
      exporter_resources: {
        ...opensearchConfig?.exporter_resources,
        requests: {
          ...opensearchConfig?.exporter_resources?.requests,
          [key]: val,
        },
      },
    };

    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 是否配置子组件资源限制
   */
  public onChangeOpensearchConfigComponentResources(val) {
    const opensearchConfig = {
      ...this.state.opensearchConfig,
      exporter_resources: val
        ? DefaultConfigData[this.props.service.key]?.exporter_resources
        : undefined,
    };

    this.setState(
      {
        opensearchConfig,
      },
      () => {
        this.props.onUpdateOpensearchData(
          opensearchConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }
}

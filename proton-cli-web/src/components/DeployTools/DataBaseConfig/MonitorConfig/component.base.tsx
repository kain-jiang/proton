import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import {
  DefaultMonitorResourcesConfig,
  NODES_LIMIT,
  RESOURCES,
  SERVICES,
  getDefaultNodes,
} from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { MonitorType } from "./index";

export default class MonitorConfigBase extends React.Component<
  MonitorType.Props,
  MonitorType.State
> {
  state = {
    monitorConfig: null,
    monitorNodes: [],
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.form);
    this.initConfig(
      configData.proton_monitor,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(
    prevProps: Readonly<MonitorType.Props>,
    prevState: Readonly<MonitorType.State>,
  ): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.proton_monitor, configData.proton_monitor)) {
      this.initConfig(
        configData.proton_monitor,
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
          NODES_LIMIT[SERVICES.ProtonMonitor],
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
        monitorNodes: nodes,
        monitorConfig: newConfig,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.monitorConfig,
        });
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateMonitorData(newConfig, dataBaseStorageType);
    }
  }

  public onChangeMonitor(config) {
    const monitorConfig = {
      ...this.state.monitorConfig,
      ...config,
    };
    this.setState(
      {
        monitorConfig,
      },
      () => {
        this.props.onUpdateMonitorData(
          monitorConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  public onChangeMonitorConfig(config) {
    const monitorConfig = {
      ...this.state.monitorConfig,
      config: {
        ...this.state.monitorConfig?.config,
        ...config,
      },
    };
    this.setState(
      {
        monitorConfig,
      },
      () => {
        this.props.onUpdateMonitorData(
          monitorConfig,
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
  public onChangeMonitorNode(nodes, config) {
    if (nodes && nodes.length > NODES_LIMIT[SERVICES.ProtonMonitor]) {
      message.info(
        `Proton Monitor 部署节点数量应当小于等于 ${
          NODES_LIMIT[SERVICES.ProtonMonitor]
        }`,
      );
    } else {
      this.props.updateMonitorNodesValidateState();
    }
    this.props.onUpdateMonitorData(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }

  public onChangeMonitorComponentResources(
    monitorComponent,
    resourcesType,
    config,
  ) {
    let monitorConfig;
    if (resourcesType === RESOURCES.ALL) {
      let resources = DefaultMonitorResourcesConfig[monitorComponent];

      monitorConfig = {
        ...this.state.monitorConfig,
        resources: {
          ...this.state.monitorConfig?.resources,
          [monitorComponent]: config ? resources : null,
        },
      };
    } else if (resourcesType === RESOURCES.LIMITS) {
      monitorConfig = {
        ...this.state.monitorConfig,
        resources: {
          ...this.state.monitorConfig?.resources,
          [monitorComponent]: {
            ...this.state.monitorConfig?.resources?.[monitorComponent],
            [RESOURCES.LIMITS]: {
              ...this.state.monitorConfig?.resources?.[monitorComponent]?.[
                RESOURCES.LIMITS
              ],
              ...config,
            },
          },
        },
      };
    } else {
      monitorConfig = {
        ...this.state.monitorConfig,
        resources: {
          ...this.state.monitorConfig?.resources,
          [monitorComponent]: {
            ...this.state.monitorConfig?.resources?.[monitorComponent],
            [RESOURCES.REQUESTS]: {
              ...this.state.monitorConfig?.resources?.[monitorComponent]?.[
                RESOURCES.REQUESTS
              ],
              ...config,
            },
          },
        },
      };
    }
    this.setState(
      {
        monitorConfig,
      },
      () => {
        this.props.onUpdateMonitorData(
          monitorConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  public onChangeMonitorConfigGrafana(val) {
    const monitorConfig = {
      ...this.state.monitorConfig,
      config: {
        ...this.state.monitorConfig?.config,
        grafana: val
          ? {
              smtp: {
                enabled: true,
                skip_verify: false,
                enable_tracing: false,
              },
            }
          : {
              smtp: {
                enabled: false,
              },
            },
      },
    };
    this.setState(
      {
        monitorConfig,
      },
      () => {
        this.props.onUpdateMonitorData(
          monitorConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  public onChangeMonitorConfigVmagent(val) {
    const monitorConfig = {
      ...this.state.monitorConfig,
      config: {
        ...this.state.monitorConfig?.config,
        vmagent: val ? { remoteWrite: { extraServers: "" } } : undefined,
      },
    };
    this.setState(
      {
        monitorConfig,
      },
      () => {
        this.props.onUpdateMonitorData(
          monitorConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  public onChangeGrafanaSMTP(key, val) {
    let monitorConfig = { ...this.state.monitorConfig };

    monitorConfig = {
      ...monitorConfig,
      config: {
        ...monitorConfig.config,
        grafana: {
          ...monitorConfig.config?.grafana,
          smtp: {
            ...monitorConfig.config?.grafana?.smtp,
            [key]: val,
          },
        },
      },
    };

    this.setState(
      {
        monitorConfig,
      },
      () => {
        this.props.onUpdateMonitorData(
          monitorConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }
}

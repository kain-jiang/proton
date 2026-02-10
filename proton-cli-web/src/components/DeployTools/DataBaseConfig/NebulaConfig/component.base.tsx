import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import {
  getDefaultNodes,
  RESOURCES,
  RESOURCES_TYPE,
  NEBULA_COMPONENTS,
} from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { NebulaType } from "./index";

export default class NebulaConfigBase extends React.Component<
  NebulaType.Props,
  NebulaType.State
> {
  state = {
    nebulaNodes: [],
    nebulaConfig: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.form);
    this.initConfig(
      configData.nebula,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(prevProps: Readonly<NebulaType.Props>): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.nebula, configData.nebula)) {
      this.initConfig(
        configData.nebula,
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
          isArray(oldConfig?.hosts) ? oldConfig.hosts : [],
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
        nebulaNodes: nodes,
        nebulaConfig: newConfig,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.nebulaConfig,
        });
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateNebulaData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新 Graphd 资源配置
   * @param config 更新的配置
   */
  public onChangeNebulaComponentResources(
    nebulaComponent,
    resourcesType,
    config,
  ) {
    let nebulaConfig;
    if (resourcesType === RESOURCES.ALL) {
      let resources = {
        [RESOURCES.LIMITS]: {
          [RESOURCES_TYPE.CPU]: "",
          [RESOURCES_TYPE.MEMORY]: "",
        },
        [RESOURCES.REQUESTS]: {
          [RESOURCES_TYPE.CPU]: "100m",
          [RESOURCES_TYPE.MEMORY]: "128Mi",
        },
      };

      nebulaConfig = {
        ...this.state.nebulaConfig,
        [nebulaComponent]: {
          ...this.state.nebulaConfig[nebulaComponent],
          resources: config ? resources : null,
        },
      };
    } else if (resourcesType === RESOURCES.LIMITS) {
      nebulaConfig = {
        ...this.state.nebulaConfig,
        [nebulaComponent]: {
          ...this.state.nebulaConfig[nebulaComponent],
          resources: {
            ...this.state.nebulaConfig[nebulaComponent].resources,
            [RESOURCES.LIMITS]: {
              ...this.state.nebulaConfig[nebulaComponent].resources[
                RESOURCES.LIMITS
              ],
              ...config,
            },
          },
        },
      };
    } else {
      nebulaConfig = {
        ...this.state.nebulaConfig,
        [nebulaComponent]: {
          ...this.state.nebulaConfig[nebulaComponent],
          resources: {
            ...this.state.nebulaConfig[nebulaComponent].resources,
            [RESOURCES.REQUESTS]: {
              ...this.state.nebulaConfig[nebulaComponent].resources[
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
        nebulaConfig,
      },
      () => {
        this.props.onUpdateNebulaData(
          nebulaConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangeNebula(config) {
    const nebulaConfig = {
      ...this.state.nebulaConfig,
      ...config,
    };
    this.setState(
      {
        nebulaConfig,
      },
      () => {
        this.props.onUpdateNebulaData(
          nebulaConfig,
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
  public onChangeNebulaNode(nodes, config) {
    this.props.onUpdateNebulaData(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }

  /**
   * 更新graphd,metad,storaged的config配置
   */
  public onChangeNebulaComponentConfig(nebulaComponent, config) {
    const nebulaConfig = {
      ...this.state.nebulaConfig,
      [nebulaComponent]: {
        ...this.state.nebulaConfig[nebulaComponent],
        config: {
          ...this.state.nebulaConfig[nebulaComponent]?.config,
          ...config,
        },
      },
    };
    this.setState(
      {
        nebulaConfig,
      },
      () => {
        this.props.onUpdateNebulaData(
          nebulaConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }
}

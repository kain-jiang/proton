import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import { getDefaultNodes } from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { Props, State } from "./index";

export default class PackageStoreBase extends React.Component<Props, State> {
  state = {
    packageStoreNodes: [],
    packageStoreConfig: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.form);
    this.initConfig(
      configData["package-store"],
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData, dataBaseStorageType } = this.props;
    if (
      !eq(prevProps.configData["package-store"], configData["package-store"])
    ) {
      this.initConfig(
        configData["package-store"],
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
        packageStoreNodes: nodes,
        packageStoreConfig: newConfig,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.packageStoreConfig,
        });
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdatePackageStore(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新mariaDB的配置
   * @param config 更新的配置
   */
  public onChangePackageStoreStorage(config) {
    const packageStoreConfig = {
      ...this.state.packageStoreConfig,
      storage: {
        ...this.state.packageStoreConfig.storage,
        ...config,
      },
    };
    this.setState(
      {
        packageStoreConfig,
      },
      () => {
        this.props.onUpdatePackageStore(
          packageStoreConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 是否配置资源限制
   */
  public onChangePackageStoreResources(val) {
    const packageStoreConfig = {
      ...this.state.packageStoreConfig,
      resources: val
        ? {
            limits: {
              cpu: "",
              memory: "",
            },
          }
        : undefined,
    };
    this.setState(
      {
        packageStoreConfig,
      },
      () => {
        this.props.onUpdatePackageStore(
          packageStoreConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangePackageStoreResourcesLimits(config) {
    const packageStoreConfig = {
      ...this.state.packageStoreConfig,
      resources: {
        limits: {
          ...this.state.packageStoreConfig.resources.limits,
          ...config,
        },
      },
    };
    this.setState(
      {
        packageStoreConfig,
      },
      () => {
        this.props.onUpdatePackageStore(
          packageStoreConfig,
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
  public onChangePackageStoreNode(nodes, config) {
    this.props.onUpdatePackageStore(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }

  /**
   * 修改节点
   * @param nodes 节点信息
   * @param config 配置信息
   */
  public onChangePackageStoreReplicas(replicas) {
    const packageStoreConfig = {
      ...this.state.packageStoreConfig,
      ...replicas,
    };
    this.setState(
      {
        packageStoreConfig,
      },
      () => {
        this.props.onUpdatePackageStore(
          packageStoreConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }
}

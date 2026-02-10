import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import { DataBaseStorageType, getDefaultNodes } from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { MariaDBType } from "./index";

export default class MariaDBConfigBase extends React.Component<
  MariaDBType.Props,
  MariaDBType.State
> {
  state = {
    mariaDBNodes: [],
    mariaDBConfig: null,
  };

  mariaDBForm = {
    configForm: React.createRef<FormInstance>(),
    accountForm: React.createRef<FormInstance>(),
    pathForm: React.createRef<FormInstance>(),
    storageForm: React.createRef<FormInstance>(),
    replicaForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.mariaDBForm);
    this.initConfig(
      configData.proton_mariadb,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(prevProps: Readonly<MariaDBType.Props>): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.proton_mariadb, configData.proton_mariadb)) {
      this.initConfig(
        configData.proton_mariadb,
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
        mariaDBNodes: nodes,
        mariaDBConfig: newConfig,
      },
      () => {
        this.mariaDBForm.configForm.current.setFieldsValue({
          ...this.state.mariaDBConfig?.config,
        });
        this.mariaDBForm.accountForm.current.setFieldsValue({
          ...this.state.mariaDBConfig,
        });
        if (dataBaseStorageType === DataBaseStorageType.Standard) {
          this.mariaDBForm.pathForm.current.setFieldsValue({
            ...this.state.mariaDBConfig,
          });
        } else {
          this.mariaDBForm.storageForm.current.setFieldsValue({
            ...this.state.mariaDBConfig,
          });
          this.mariaDBForm.replicaForm.current.setFieldsValue({
            ...this.state.mariaDBConfig,
          });
        }
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateMariDBData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新mariaDB的配置
   * @param config 更新的配置
   */
  public onChangeMariaDBConfig(config) {
    const mariaDBConfig = {
      ...this.state.mariaDBConfig,
      config: {
        ...this.state.mariaDBConfig.config,
        ...config,
      },
    };
    this.setState(
      {
        mariaDBConfig,
      },
      () => {
        this.props.onUpdateMariDBData(
          mariaDBConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangeMariDB(config) {
    const mariaDBConfig = {
      ...this.state.mariaDBConfig,
      ...config,
    };
    this.setState(
      {
        mariaDBConfig,
      },
      () => {
        this.props.onUpdateMariDBData(
          mariaDBConfig,
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
  public onChangeMariaDBNode(nodes, config) {
    this.props.onUpdateMariDBData(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }
}

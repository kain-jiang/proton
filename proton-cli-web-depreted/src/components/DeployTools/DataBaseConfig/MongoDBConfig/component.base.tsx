import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import {
  DataBaseStorageType,
  DefaultConfigData,
  getDefaultNodes,
} from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { MongoDBType } from "./index";

export default class MongoDBConfigBase extends React.Component<
  MongoDBType.Props,
  MongoDBType.State
> {
  state = {
    mongoDBConfig: null,
    mongoDBNodes: [],
  };

  mongoDBForm = {
    accountForm: React.createRef<FormInstance>(),
    pathForm: React.createRef<FormInstance>(),
    storageForm: React.createRef<FormInstance>(),
    replicaForm: React.createRef<FormInstance>(),
    resourcesForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.mongoDBForm);
    this.initConfig(
      configData.proton_mongodb,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(
    prevProps: Readonly<MongoDBType.Props>,
    prevState: Readonly<MongoDBType.State>,
  ): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.proton_mongodb, configData.proton_mongodb)) {
      this.initConfig(
        configData.proton_mongodb,
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
        mongoDBNodes: nodes,
        mongoDBConfig: newConfig,
      },
      () => {
        this.mongoDBForm.accountForm.current.setFieldsValue({
          ...this.state.mongoDBConfig,
        });
        this.mongoDBForm.resourcesForm.current.setFieldsValue({
          ...this.state.mongoDBConfig,
        });
        if (dataBaseStorageType === DataBaseStorageType.Standard) {
          this.mongoDBForm.pathForm.current.setFieldsValue({
            ...this.state.mongoDBConfig,
          });
        } else {
          this.mongoDBForm.storageForm.current.setFieldsValue({
            ...this.state.mongoDBConfig,
          });
          this.mongoDBForm.replicaForm.current.setFieldsValue({
            ...this.state.mongoDBConfig,
          });
        }
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateMongoDBData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangeMongoDB(config) {
    const mongoDBConfig = {
      ...this.state.mongoDBConfig,
      ...config,
    };
    this.setState(
      {
        mongoDBConfig,
      },
      () => {
        this.props.onUpdateMongoDBData(
          mongoDBConfig,
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
  public onChangeMongoDBNode(nodes, config) {
    this.props.onUpdateMongoDBData(
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
  public onChangeMongoDBConfigResources(val) {
    const mongoDBConfig = {
      ...this.state.mongoDBConfig,
      resources: val
        ? DefaultConfigData[this.props.service.key].resources
        : undefined,
    };
    this.setState(
      {
        mongoDBConfig,
      },
      () => {
        this.props.onUpdateMongoDBData(
          mongoDBConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 修改资源配置触发
   * @param key
   * @param val
   */
  public onChangeResource(key, val) {
    let mongoDBConfig = { ...this.state.mongoDBConfig };

    mongoDBConfig = {
      ...mongoDBConfig,
      resources: {
        ...mongoDBConfig?.resources,
        requests: {
          ...mongoDBConfig?.resources?.requests,
          [key]: val,
        },
      },
    };

    this.setState(
      {
        mongoDBConfig,
      },
      () => {
        this.props.onUpdateMongoDBData(
          mongoDBConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }
}

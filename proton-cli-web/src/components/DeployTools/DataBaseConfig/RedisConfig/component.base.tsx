import * as React from "react";
import { FormInstance, message } from "@aishutech/ui";
import { eq, isEqual, isArray } from "lodash";
import {
  DataBaseStorageType,
  DefaultConfigData,
  getDefaultNodes,
} from "../../helper";
import { RedisType } from "./index";

export default class RedisConfigBase extends React.Component<
  RedisType.Props,
  RedisType.State
> {
  state = {
    redisConfig: null,
    redisNodes: [],
  };

  redisForm = {
    accountForm: React.createRef<FormInstance>(),
    pathForm: React.createRef<FormInstance>(),
    replicaForm: React.createRef<FormInstance>(),
    resourcesForm: React.createRef<FormInstance>(),
  };

  componentDidMount(): void {
    const { configData, dataBaseStorageType, updateDataBaseForm } = this.props;
    updateDataBaseForm(this.redisForm);
    this.initConfig(
      configData.proton_redis,
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(
    prevProps: Readonly<RedisType.Props>,
    prevState: Readonly<RedisType.State>,
    snapshot?: any,
  ): void {
    const { configData, dataBaseStorageType } = this.props;
    if (!eq(prevProps.configData.proton_redis, configData.proton_redis)) {
      this.initConfig(
        configData.proton_redis,
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
        redisNodes: nodes,
        redisConfig: newConfig,
      },
      () => {
        this.redisForm.accountForm.current.setFieldsValue({
          ...this.state.redisConfig,
        });
        this.redisForm.pathForm.current.setFieldsValue({
          ...this.state.redisConfig,
        });
        this.redisForm.resourcesForm.current.setFieldsValue({
          ...this.state.redisConfig,
        });
        if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes) {
          this.redisForm.replicaForm.current.setFieldsValue({
            ...this.state.redisConfig,
          });
        }
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateRedisData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangeRedis(config) {
    const redisConfig = {
      ...this.state.redisConfig,
      ...config,
    };
    this.setState(
      {
        redisConfig,
      },
      () => {
        this.props.onUpdateRedisData(
          redisConfig,
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
  public onChangeRedisNode(nodes, config) {
    this.props.onUpdateRedisData(
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
  public onChangeRedisConfigResources(val) {
    const redisConfig = {
      ...this.state.redisConfig,
      resources: val
        ? DefaultConfigData[this.props.service.key].resources
        : undefined,
    };
    this.setState(
      {
        redisConfig,
      },
      () => {
        this.props.onUpdateRedisData(
          redisConfig,
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
    let redisConfig = { ...this.state.redisConfig };

    redisConfig = {
      ...redisConfig,
      resources: {
        ...redisConfig?.resources,
        requests: {
          ...redisConfig?.resources?.requests,
          [key]: val,
        },
      },
    };

    this.setState(
      {
        redisConfig,
      },
      () => {
        this.props.onUpdateRedisData(
          redisConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }
}

import * as React from "react";
import axios from "axios";
import { notification, message } from "@aishutech/ui";
import { isEmpty, isObject } from "lodash";
import {
  OptionSteps,
  DefaultConfigData,
  exChangeData,
  CRType,
  CSPlugin,
  ConfigEditStatus,
  ConfigServiceKeys,
  getDefaultNode,
  getDefaultNodes,
  getDefaultServiceInfoByService,
  notMultipleDateBase,
  DataBaseStorageType,
  DefaultSelectableServices,
  DefalutAddableServices,
  NODES_LIMIT,
  SOURCE_TYPE,
  REDIS_CONNECT_TYPE,
  SERVICES,
  CONNECT_SERVICES,
  DEGAULT_EXTERNAL_RDS,
  DEFAULT_INTERNAL_RDS,
  DEFAULT_EXTERNAL_MONGODB,
  DEFAULT_INTERNAL_MONGODB,
  DEFAULT_EXTERNAL_REDIS,
  DEFAULT_EXTERNAL_MQ,
  DEFAULT_INTERNAL_MQ,
  DEFAULT_EXTERNAL_OPENSEARCH,
  DEFAULT_INTERNAL_OPENSEARCH,
  DEFAULT_EXTERNAL_SERVICE,
  DEFAULT_INTERNAL_SERVICE,
  MQ_TYPE,
  CHRONY_MODE,
  NodeForm,
  ValidateState,
  DefaultDataBaseForm,
  DefaultConnectInfoForm,
  DefaultConnectInfoValidateState,
  RDS_TYPE,
  MQ_AUTH_MACHANISM,
  OPENSEARCH_VERSION,
  initialAllNodesServices,
  DEPLOY_MODE,
  RepositoryType,
} from "./helper";
import {
  ChronyConfig,
  ConnectInfoValidateState,
  DeployConfig,
  DeployToolsBaseType,
  ETCDConnectInfo,
  FirewallConfig,
  MQConnectInfo,
  MongoDBConnectInfo,
  NodeInfo,
  OpensearchConnectInfo,
  PolicyEngineConnectInfo,
  RDSConnectInfo,
  RedisConnectInfo,
} from "./index";

export default class DeployToolsBase extends React.Component<
  DeployToolsBaseType.Props,
  DeployToolsBaseType.State
> {
  crType = CRType.LOCAL;

  configData = null;

  state = {
    dataBaseStorageType: null,
    nextStepButtonDisable: false,
    stepStatus: OptionSteps.NodeConfig,
    configData: DefaultConfigData,
    sshAccount: "",
    sshPassword: "",
    configEditStatus: ConfigEditStatus.Editing,
    selectableServices: [...DefaultSelectableServices],
    addableServices: [...DefalutAddableServices],
    nodeForm: {
      serverForm: null,
      accountInfoForm: null,
    },
    networkForm: {
      hostNetworkForm: null,
      networkInfoForm: null,
    },
    crForm: {
      localForm: {
        portsForm: null,
        haPortsForm: null,
        storageForm: null,
      },
      externalForm: {
        chartmuseumForm: null,
        registryForm: null,
        ociForm: null,
      },
    },
    dataBaseForm: DefaultDataBaseForm,
    connectInfoForm: DefaultConnectInfoForm,
    nodesValidateState: ValidateState.Normal,
    networkNodesValidateState: ValidateState.Normal,
    crNodesValidateState: ValidateState.Normal,
    grafanaNodesValidateState: ValidateState.Normal,
    prometheusNodesValidateState: ValidateState.Normal,
    monitorNodesValidateState: ValidateState.Normal,
    connectInfoValidateState: DefaultConnectInfoValidateState,
  };

  componentDidMount(): void {
    this.setState({
      configData: {
        ...this.state.configData,
        ...this.getDefaultServiceConfig(),
      },
    });
  }

  /**
   * 改变模板类型触发
   * @param dataBaseStorageType 模板类型
   */
  changeDataBaseStorageType(dataBaseStorageType) {
    const { configData, selectableServices, addableServices } = this.state;
    const newSelectableServices = selectableServices.filter((service) => {
      return service.key !== SERVICES.Nebula && service.key !== SERVICES.ECeph;
    });
    const newAddableServices = addableServices.filter((service) => {
      return service.key !== SERVICES.Nebula && service.key !== SERVICES.ECeph;
    });
    let newConfigData;
    if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes) {
      newConfigData = {
        ...configData,
        deploy: {
          ...configData.deploy,
          mode: DEPLOY_MODE.CLOUD,
        },
        networkInfo: {
          // ...configData.networkInfo,
          provisioner: "external",
          addons: configData.networkInfo.addons || [
            CSPlugin.KubeStateMetrics,
            CSPlugin.NodeExporter,
          ],
        },
        cr: {
          external: {
            image_repository: RepositoryType.Registry,
            chart_repository: RepositoryType.Chartmuseum,
            chartmuseum: { host: "", password: "", username: "" },
            registry: { host: "", password: "", username: "" },
            oci: {
              registry: "",
              password: "",
              username: "",
              plain_http: false,
            },
          },
        },
      };
      // 删除nebula参数。当前托管k8s不支持nebula
      delete newConfigData[SERVICES.Nebula];
      // 删除eceph参数。托管k8s不支持eceph
      delete newConfigData[SERVICES.ECeph];
    } else {
      newConfigData = {
        ...configData,
        deploy: {
          ...configData.deploy,
          mode:
            dataBaseStorageType === DataBaseStorageType.Standard
              ? DEPLOY_MODE.STANDARD
              : DEPLOY_MODE.CLOUD,
        },
        networkInfo: {
          ...configData.networkInfo,
          provisioner: "local",
          addons: configData.networkInfo.addons || [
            CSPlugin.KubeStateMetrics,
            CSPlugin.NodeExporter,
          ],
        },
      };
    }

    this.setState({
      dataBaseStorageType:
        dataBaseStorageType === DataBaseStorageType.Cloud
          ? DataBaseStorageType.Standard
          : dataBaseStorageType,
      selectableServices:
        dataBaseStorageType === DataBaseStorageType.DepositKubernetes
          ? newSelectableServices
          : selectableServices,
      addableServices:
        dataBaseStorageType === DataBaseStorageType.DepositKubernetes
          ? newAddableServices
          : addableServices,
      stepStatus:
        dataBaseStorageType === DataBaseStorageType.DepositKubernetes
          ? OptionSteps.NetworkConfig
          : OptionSteps.NodeConfig,
      configData: newConfigData,
    });
    this.crType =
      dataBaseStorageType === DataBaseStorageType.DepositKubernetes
        ? CRType.ExternalCRConfig
        : this.crType;
  }

  /**
   * 切换配置事件
   * @param nextStatus 变更的事件
   */
  public onChangeStepStatus(nextStatus: OptionSteps): void {
    this.setState({
      stepStatus: nextStatus,
    });
  }

  /**
   * 初始化平台
   */
  public async onInitPlatform() {
    const { sshAccount, sshPassword, configData, dataBaseStorageType } =
      this.state;
    try {
      this.setState({
        configEditStatus: ConfigEditStatus.initing,
      });
      await axios({
        method: "post",
        url: `/init?accout=${encodeURIComponent(
          sshAccount,
        )}&password=${encodeURIComponent(sshPassword)}&_t=${Date.now()}`,
        data: exChangeData(configData, this.crType, dataBaseStorageType),
        timeout: 0,
      });
      this.setState({
        configEditStatus: ConfigEditStatus.Success,
        stepStatus: OptionSteps.NodeConfig,
      });
    } catch (err) {
      const { response } = err;
      if (!response) {
        console.log(err);
        return;
      }
      const { status, data } = response;
      if (status === 409 && data === "/init cannot be called concurrently") {
        setTimeout(() => {
          this.checkInitStatus();
        }, 5 * 1000);
      } else {
        if (err.message === "Network Error") {
          setTimeout(() => {
            this.checkInitStatus();
          }, 5 * 1000);
        } else {
          this.setState({
            configEditStatus: ConfigEditStatus.Editing,
          });
          notification["error"]({
            message: "初始化失败!",
            description: <div>{JSON.stringify(data)}</div>,
            placement: "top",
            duration: null,
          });
        }
      }
    }
  }

  /**
   * 当触发二次请求（arm+chrome触发tcp重发bug）时检查安装状态
   */
  private async checkInitStatus() {
    try {
      const { data } = await axios({
        method: "get",
        url: `/alpha/result`,
        timeout: 0,
      });
      if (data && data.indexOf("fail") !== -1) {
        this.setState({
          configEditStatus: ConfigEditStatus.Editing,
        });
        notification["error"]({
          message: "初始化失败!",
          description: <div>{JSON.stringify(data)}</div>,
          placement: "top",
          duration: null,
        });
      } else {
        this.setState({
          configEditStatus: ConfigEditStatus.Success,
          stepStatus: OptionSteps.NodeConfig,
        });
      }
    } catch ({ response }) {
      const { status, data } = response;
      if (status === 404 && data === "The initialization is running.") {
        setTimeout(() => {
          this.checkInitStatus();
        }, 10 * 1000);
      } else {
        this.setState({
          configEditStatus: ConfigEditStatus.Editing,
        });
        notification["error"]({
          message: "初始化失败!",
          description: <div>{JSON.stringify(data)}</div>,
          placement: "top",
          duration: null,
        });
      }
    }
  }

  /**
   * 更新节点数据
   * @param nodesInfo 节点信息
   */
  public updateNodesInfo(nodesInfo: Array<NodeInfo>): void {
    this.setState({
      configData: { ...this.state.configData, nodesInfo },
    });
  }

  /**
   * updateNetworkConfig
   */
  public updateNetworkConfig(config) {
    this.setState({
      configData: {
        ...this.state.configData,
        networkInfo: {
          ...this.state.configData.networkInfo,
          ...config,
        },
      },
    });
  }

  /**
   * update 时间服务器配置
   * chrony 时间服务器配置
   */
  public updateChrony(chrony: ChronyConfig) {
    this.setState({
      configData: {
        ...this.state.configData,
        chrony: {
          ...this.state.configData.chrony,
          ...chrony,
        },
      },
    });
  }

  /**
   * update 防火墙配置
   * firewall 防火墙配置
   */
  public updateFirewall(firewall: FirewallConfig) {
    this.setState({
      configData: {
        ...this.state.configData,
        firewall: {
          ...this.state.configData.firewall,
          ...firewall,
        },
      },
    });
  }

  /**
   * 更新部署配置
   */
  public updateDeploy(deploy: Partial<DeployConfig>) {
    this.setState({
      configData: {
        ...this.state.configData,
        deploy: {
          ...this.state.configData.deploy,
          ...deploy,
        },
      },
    });
  }

  /**
   * 更新配置信息
   */
  public onUpdateConfigInfo(value) {
    this.setState((prevState) => {
      return {
        configData: {
          ...prevState.configData,
          ...value,
        },
      };
    });
  }

  /**
   * 更新连接信息
   * @param key 连接信息键
   * @param curInfo 连接信息新值
   */
  public onUpdateConnectInfo(key, curInfo) {
    this.setState((prevState) => {
      const { resource_connect_info: preInfo } = prevState.configData;
      let info;
      if (key === CONNECT_SERVICES.RDS) {
        info = {
          ...preInfo,
          [key]: this.getRDSInfoByResourceType(curInfo),
        };
      } else if (key === CONNECT_SERVICES.MONGODB) {
        info = {
          ...preInfo,
          [key]: this.getMongoDBInfoByResourceType(curInfo),
        };
      } else if (key === CONNECT_SERVICES.REDIS) {
        info = {
          ...preInfo,
          [key]: this.getRedisInfoByResourceType(curInfo),
        };
      } else if (key === CONNECT_SERVICES.MQ) {
        info = {
          ...preInfo,
          [key]: this.getMQInfoByResourceType(curInfo),
        };
      } else if (key === CONNECT_SERVICES.OPENSEARCH) {
        info = {
          ...preInfo,
          [key]: this.getOpenSearchInfoByResourceType(curInfo),
        };
      } else if (key === CONNECT_SERVICES.POLICY_ENGINE) {
        info = {
          ...preInfo,
          [key]: this.getPolicyEngineInfoByResourceType(curInfo),
        };
      } else if (key === CONNECT_SERVICES.ETCD) {
        info = {
          ...preInfo,
          [key]: this.getETCDInfoByResourceType(curInfo),
        };
      }
      return {
        configData: {
          ...prevState.configData,
          resource_connect_info: info,
        },
      };
    });
  }

  /**
   * 根据类型获取连接信息
   * @param curInfo 现信息
   * @returns
   */
  private getRDSInfoByResourceType(curInfo: RDSConnectInfo) {
    const {
      source_type: curSourceType,
      rds_type: curRdsType,
      username: curUsername,
      password: curPassword,
      hosts: curHosts,
      port: curPort,
      auto_create_database: curAutoCreateDatabase,
      admin_user: curAdminUser,
      admin_passwd: curAdminPasswd,
    } = curInfo || {};

    let info: RDSConnectInfo = {
      source_type: curSourceType,
      username: curUsername,
      password: curPassword,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
      info = {
        ...info,
        rds_type: curRdsType,
        hosts: curHosts,
        port: curPort,
        auto_create_database: curAutoCreateDatabase,
        admin_user: curAdminUser,
        admin_passwd: curAdminPasswd,
      };
    }
    return info;
  }

  /**
   * 根据类型获取连接信息
   * @param curInfo 现信息
   * @returns
   */
  private getMongoDBInfoByResourceType(curInfo: MongoDBConnectInfo) {
    const {
      source_type: curSourceType,
      username: curUsername,
      password: curPassword,
      replica_set: curReplicaSet,
      hosts: curHosts,
      port: curPort,
      ssl: curSSL,
      auth_source: curAuthSource,
      options: curOptions,
    } = curInfo || {};

    let info: MongoDBConnectInfo = {
      source_type: curSourceType,
      username: curUsername,
      password: curPassword,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
      info = {
        ...info,
        hosts: curHosts,
        port: curPort,
        ssl: curSSL,
        replica_set: curReplicaSet,
        auth_source: curAuthSource,
        options: curOptions,
      };
    }
    return info;
  }

  /**
   * 根据类型获取连接信息
   * @param preInfo 原信息
   * @param curInfo 现信息
   * @returns
   */
  private getRedisInfoByResourceType(curInfo: RedisConnectInfo) {
    const {
      source_type: curSourceType,
      connect_type: curConnectType,
      username: curUsername,
      password: curPassword,
      hosts: curHosts,
      port: curPort,
      sentinel_username: curSentinelUsername,
      sentinel_password: curSentinelPassword,
      sentinel_hosts: curSentinelHosts,
      sentinel_port: curSentinelPort,
      master_group_name: curMasterGroupName,
      master_hosts: curMasterHosts,
      master_port: curMasterPort,
      slave_hosts: curSlaveHosts,
      slave_port: curSlavePort,
    } = curInfo || {};

    let info: RedisConnectInfo = {
      source_type: curSourceType,
    };
    if (curSourceType === SOURCE_TYPE.INTERNAL) {
      return info;
    } else {
      info = {
        source_type: curSourceType,
        connect_type: curConnectType,
        username: curUsername,
        password: curPassword,
      };
      switch (curConnectType) {
        case REDIS_CONNECT_TYPE.SENTINEL:
          return {
            ...info,
            sentinel_username: curSentinelUsername,
            sentinel_password: curSentinelPassword,
            sentinel_hosts: curSentinelHosts,
            sentinel_port: curSentinelPort,
            master_group_name: curMasterGroupName,
          };
        case REDIS_CONNECT_TYPE.MASTER_SLAVE:
          return {
            ...info,
            master_hosts: curMasterHosts,
            master_port: curMasterPort,
            slave_hosts: curSlaveHosts,
            slave_port: curSlavePort,
          };
        default:
          return {
            ...info,
            hosts: curHosts,
            port: curPort,
          };
      }
    }
  }

  /**
   * 根据类型获取连接信息
   * @param preInfo 原信息
   * @param curInfo 现信息
   * @returns
   */
  private getMQInfoByResourceType(curInfo: MQConnectInfo) {
    const {
      source_type: curSourceType,
      mq_type: curMQType,
      mq_hosts: curHosts,
      mq_port: curPort,
      mq_lookupd_hosts: curLookupdHosts,
      mq_lookupd_port: curLookupdPort,
      auth: curAuth,
    } = curInfo || {};

    let info: MQConnectInfo = {
      source_type: curSourceType,
      mq_type: curMQType,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
      info = {
        ...info,
        mq_hosts: curHosts,
        mq_port: curPort,
      };
      if (curMQType === MQ_TYPE.KAFKA) {
        const {
          username: curUsername,
          password: curPassword,
          mechanism: curMechanism,
        } = curAuth || {};
        info = {
          ...info,
          auth: {
            username: curUsername,
            password: curPassword,
            mechanism: curMechanism,
          },
        };
      }
      if (curMQType !== MQ_TYPE.NSQ) {
        info = {
          ...info,
          mq_lookupd_hosts: curHosts,
          mq_lookupd_port: curPort,
        };
      } else {
        info = {
          ...info,
          mq_lookupd_hosts: curLookupdHosts,
          mq_lookupd_port: curLookupdPort,
        };
      }
    }
    return info;
  }

  /**
   * 根据类型获取连接信息
   * @param curInfo 现信息
   * @returns
   */
  private getOpenSearchInfoByResourceType(curInfo: OpensearchConnectInfo) {
    const {
      source_type: curSourceType,
      username: curUsername,
      password: curPassword,
      hosts: curHosts,
      port: curPort,
      version: curVersion,
      distribution: curDistribution,
    } = curInfo || {};

    let info: OpensearchConnectInfo = {
      source_type: curSourceType,
      version: curVersion,
      distribution: curDistribution,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
      info = {
        ...info,
        username: curUsername,
        password: curPassword,
        hosts: curHosts,
        port: curPort,
        version: curVersion,
      };
    }
    return info;
  }

  /**
   * 根据类型获取连接信息
   * @param curInfo 现信息
   * @returns
   */
  private getPolicyEngineInfoByResourceType(curInfo: PolicyEngineConnectInfo) {
    const {
      source_type: curSourceType,
      hosts: curHosts,
      port: curPort,
    } = curInfo || {};

    let info: PolicyEngineConnectInfo = {
      source_type: curSourceType,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
      info = {
        ...info,
        hosts: curHosts,
        port: curPort,
      };
    }
    return info;
  }

  /**
   * 根据类型获取连接信息
   * @param curInfo 现信息
   * @returns
   */
  private getETCDInfoByResourceType(curInfo: ETCDConnectInfo) {
    const {
      source_type: curSourceType,
      hosts: curHosts,
      port: curPort,
      secret: curSecret,
    } = curInfo || {};

    let info: ETCDConnectInfo = {
      source_type: curSourceType,
    };
    if (curSourceType === SOURCE_TYPE.EXTERNAL) {
      info = {
        ...info,
        hosts: curHosts,
        port: curPort,
        secret: curSecret,
      };
    }
    return info;
  }

  /**
   * 更新配置信息
   */
  public onUpdateCRConfigInfo(value) {
    this.setState({
      configData: {
        ...this.state.configData,
        cr: value,
      },
    });
  }
  /**
   * 更新CRType
   */
  public onUpDateCRTypeConfig(value) {
    this.crType = value;
  }

  /**
   * 更新账户信息
   */
  public onChangeSSHInfo(value) {
    this.setState({
      ...this.state,
      ...value,
    });
  }

  /**
   * 更新资源信息
   * @param type 类型
   * @param service 服务
   * @param resource_connect_info 连接信息
   * @returns
   */
  private changeResourceConnectServiceTypeByService(
    isDel,
    service,
    configData,
  ) {
    const { resource_connect_info } = configData;
    if (service === SERVICES.ProtonMariadb) {
      let rds;
      if (isDel) {
        rds = this.getRDSInfoByResourceType(DEGAULT_EXTERNAL_RDS);
      } else {
        rds = this.getRDSInfoByResourceType(DEFAULT_INTERNAL_RDS);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.RDS]: rds,
      };
    } else if (service === SERVICES.ProtonMongodb) {
      let mongodb;
      if (isDel) {
        mongodb = this.getMongoDBInfoByResourceType(DEFAULT_EXTERNAL_MONGODB);
      } else {
        mongodb = this.getMongoDBInfoByResourceType(DEFAULT_INTERNAL_MONGODB);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.MONGODB]: mongodb,
      };
    } else if (service === SERVICES.ProtonRedis) {
      let redis;
      if (isDel) {
        redis = this.getRedisInfoByResourceType(DEFAULT_EXTERNAL_REDIS);
      } else {
        redis = this.getRedisInfoByResourceType(DEFAULT_INTERNAL_SERVICE);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.REDIS]: redis,
      };
    } else if (service === SERVICES.ProtonNSQ) {
      if (isDel) {
        if (configData[SERVICES.Kafka]) {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]:
              this.getMQInfoByResourceType(DEFAULT_INTERNAL_MQ),
          };
        } else {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]:
              this.getMQInfoByResourceType(DEFAULT_EXTERNAL_MQ),
          };
        }
      } else {
        if (configData[SERVICES.Kafka]) {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]:
              this.getMQInfoByResourceType(DEFAULT_INTERNAL_MQ),
          };
        } else {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]: this.getMQInfoByResourceType({
              ...DEFAULT_INTERNAL_MQ,
              mq_type: MQ_TYPE.NSQ,
            }),
          };
        }
      }
    } else if (service === SERVICES.Kafka) {
      if (isDel) {
        if (configData[SERVICES.ProtonNSQ]) {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]: this.getMQInfoByResourceType({
              ...DEFAULT_INTERNAL_MQ,
              mq_type: MQ_TYPE.NSQ,
            }),
          };
        } else {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]:
              this.getMQInfoByResourceType(DEFAULT_EXTERNAL_MQ),
          };
        }
      } else {
        if (configData[SERVICES.ProtonNSQ]) {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]:
              this.getMQInfoByResourceType(DEFAULT_INTERNAL_MQ),
          };
        } else {
          return {
            ...resource_connect_info,
            [CONNECT_SERVICES.MQ]:
              this.getMQInfoByResourceType(DEFAULT_INTERNAL_MQ),
          };
        }
      }
    } else if (service === SERVICES.Opensearch) {
      let opensearch;
      if (isDel) {
        opensearch = this.getOpenSearchInfoByResourceType(
          DEFAULT_EXTERNAL_OPENSEARCH,
        );
      } else {
        opensearch = this.getOpenSearchInfoByResourceType(
          DEFAULT_INTERNAL_OPENSEARCH,
        );
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.OPENSEARCH]: opensearch,
      };
    } else if (service === SERVICES.ProtonPolicyEngine) {
      let policyEngine;
      if (isDel) {
        policyEngine = this.getPolicyEngineInfoByResourceType(
          DEFAULT_EXTERNAL_SERVICE,
        );
      } else {
        policyEngine = this.getPolicyEngineInfoByResourceType(
          DEFAULT_INTERNAL_SERVICE,
        );
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.POLICY_ENGINE]: policyEngine,
      };
    } else if (service === SERVICES.ProtonEtcd) {
      let etcd;
      if (isDel) {
        return {
          ...resource_connect_info,
          [CONNECT_SERVICES.ETCD]: undefined,
        };
      } else {
        etcd = this.getETCDInfoByResourceType(DEFAULT_INTERNAL_SERVICE);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.ETCD]: etcd,
      };
    } else {
      return resource_connect_info;
    }
  }

  /**
   * 更新资源信息
   * @param type 类型
   * @param service 服务
   * @param resource_connect_info 连接信息
   * @returns
   */
  private addResourceConnectServiceType(service, configData) {
    const { resource_connect_info } = configData;
    if (service === CONNECT_SERVICES.RDS) {
      let rds;
      if (!configData[SERVICES.ProtonMariadb]) {
        rds = this.getRDSInfoByResourceType(DEGAULT_EXTERNAL_RDS);
      } else {
        rds = this.getRDSInfoByResourceType(DEFAULT_INTERNAL_RDS);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.RDS]: rds,
      };
    } else if (service === CONNECT_SERVICES.MONGODB) {
      let mongodb;
      if (!configData[SERVICES.ProtonMongodb]) {
        mongodb = this.getMongoDBInfoByResourceType(DEFAULT_EXTERNAL_MONGODB);
      } else {
        mongodb = this.getMongoDBInfoByResourceType(DEFAULT_INTERNAL_MONGODB);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.MONGODB]: mongodb,
      };
    } else if (service === CONNECT_SERVICES.REDIS) {
      let redis;
      if (!configData[SERVICES.ProtonRedis]) {
        redis = this.getRedisInfoByResourceType(DEFAULT_EXTERNAL_REDIS);
      } else {
        redis = this.getRedisInfoByResourceType(DEFAULT_INTERNAL_SERVICE);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.REDIS]: redis,
      };
    } else if (service === CONNECT_SERVICES.MQ) {
      if (!configData[SERVICES.ProtonNSQ] && !configData[SERVICES.Kafka]) {
        return {
          ...resource_connect_info,
          [CONNECT_SERVICES.MQ]:
            this.getMQInfoByResourceType(DEFAULT_EXTERNAL_MQ),
        };
      } else if (configData[SERVICES.Kafka]) {
        return {
          ...resource_connect_info,
          [CONNECT_SERVICES.MQ]:
            this.getMQInfoByResourceType(DEFAULT_INTERNAL_MQ),
        };
      } else {
        return {
          ...resource_connect_info,
          [CONNECT_SERVICES.MQ]: this.getMQInfoByResourceType({
            ...DEFAULT_INTERNAL_MQ,
            mq_type: MQ_TYPE.NSQ,
          }),
        };
      }
    } else if (service === CONNECT_SERVICES.OPENSEARCH) {
      let opensearch;
      if (!configData[SERVICES.Opensearch]) {
        opensearch = this.getOpenSearchInfoByResourceType(
          DEFAULT_EXTERNAL_OPENSEARCH,
        );
      } else {
        opensearch = this.getOpenSearchInfoByResourceType(
          DEFAULT_INTERNAL_OPENSEARCH,
        );
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.OPENSEARCH]: opensearch,
      };
    } else if (service === CONNECT_SERVICES.POLICY_ENGINE) {
      let policyEngine;
      if (!configData[SERVICES.ProtonPolicyEngine]) {
        policyEngine = this.getPolicyEngineInfoByResourceType(
          DEFAULT_EXTERNAL_SERVICE,
        );
      } else {
        policyEngine = this.getPolicyEngineInfoByResourceType(
          DEFAULT_INTERNAL_SERVICE,
        );
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.POLICY_ENGINE]: policyEngine,
      };
    } else if (service === CONNECT_SERVICES.ETCD) {
      let etcd;
      if (!configData[SERVICES.ProtonEtcd]) {
        etcd = this.getETCDInfoByResourceType(DEFAULT_EXTERNAL_SERVICE);
      } else {
        etcd = this.getETCDInfoByResourceType(DEFAULT_INTERNAL_SERVICE);
      }
      return {
        ...resource_connect_info,
        [CONNECT_SERVICES.ETCD]: etcd,
      };
    } else {
      return resource_connect_info;
    }
  }

  /**
   * 删除服务
   * @param value
   * @param key
   */
  public onDeleteService(value) {
    switch (value) {
      case SERVICES.Grafana:
        this.setState({
          grafanaNodesValidateState: ValidateState.Normal,
        });
        break;
      case SERVICES.Prometheus:
        this.setState({
          prometheusNodesValidateState: ValidateState.Normal,
        });
        break;
      case SERVICES.ProtonMonitor:
        this.setState({
          monitorNodesValidateState: ValidateState.Normal,
        });
        break;
    }
    this.setState({
      selectableServices: this.state.selectableServices.filter(
        (service) => service.key != value,
      ),
      addableServices: [
        ...this.state.addableServices,
        this.state.selectableServices.find((service) => service.key === value),
      ],
      configData: {
        ...this.state.configData,
        [value]: null,
        resource_connect_info: this.changeResourceConnectServiceTypeByService(
          true,
          value,
          this.state.configData,
        ),
      },
    });
  }

  /**
   * 删除连接信息
   * @param connectService 组件
   */
  public onDeleteResource(connectService) {
    this.clearResourceValidateState(connectService);
    const resourceConnectInfo = {
      ...this.state.configData.resource_connect_info,
    };
    delete resourceConnectInfo[connectService];
    this.setState({
      configData: {
        ...this.state.configData,
        resource_connect_info: resourceConnectInfo,
      },
    });
  }

  /**
   * 删除连接信息校验状态
   * @param connectService 组件
   */
  public clearResourceValidateState(connectService) {
    switch (connectService) {
      case CONNECT_SERVICES.RDS:
        this.updateConnectInfoValidateState({
          RDS_TYPE: ValidateState.Normal,
        });
        break;
      case CONNECT_SERVICES.MONGODB:
        this.updateConnectInfoValidateState({
          MONGODB_SSL: ValidateState.Normal,
        });
        break;
      case CONNECT_SERVICES.REDIS:
        this.updateConnectInfoValidateState({
          REDIS_CONNECT_TYPE: ValidateState.Normal,
        });
        break;
      case CONNECT_SERVICES.MQ:
        this.updateConnectInfoValidateState({
          MQ_AUTH_MACHANISM: ValidateState.Normal,
          MQ_RADIO: ValidateState.Normal,
          MQ_TYPE: ValidateState.Normal,
        });
        break;
      case CONNECT_SERVICES.OPENSEARCH:
        this.updateConnectInfoValidateState({
          OPENSEARCH_VERSION: ValidateState.Normal,
        });
        break;
    }
  }

  /**
   * 添加连接信息
   */
  public onAddResource(service) {
    const { configData } = this.state;
    this.setState({
      configData: {
        ...configData,
        resource_connect_info: this.addResourceConnectServiceType(
          service,
          configData,
        ),
      },
    });
  }

  /**
   * 增加服务
   * @param value
   */
  public onAddService(value) {
    const configData = { ...this.state.configData };
    const { nodesInfo } = configData;
    const serviceLimits = Object.keys(NODES_LIMIT);

    if (notMultipleDateBase.includes(value)) {
      DefaultConfigData[value].hosts = getDefaultNode(
        nodesInfo,
        DefaultConfigData[value].hosts,
      ).map((node) => node.name);
    } else {
      if (serviceLimits.includes(value)) {
        DefaultConfigData[value].hosts = getDefaultNodes(
          nodesInfo,
          DefaultConfigData[value].hosts,
          NODES_LIMIT[value],
          true,
        );
      } else if (initialAllNodesServices.includes(value)) {
        DefaultConfigData[value].hosts = getDefaultNodes(
          nodesInfo,
          DefaultConfigData[value].hosts,
          3,
          false,
          true,
        ).map((node) => node.name);
      } else {
        DefaultConfigData[value].hosts = getDefaultNodes(
          nodesInfo,
          DefaultConfigData[value].hosts,
        ).map((node) => node.name);
      }
    }

    let newconfigData = Object.assign({}, configData, {
      [value]: getDefaultServiceInfoByService(
        value,
        this.state.dataBaseStorageType,
      ),
      resource_connect_info: this.changeResourceConnectServiceTypeByService(
        false,
        value,
        configData,
      ),
    });
    this.setState({
      selectableServices: [
        this.state.addableServices.find((service) => service.key === value),
        ...this.state.selectableServices,
      ],
      addableServices: this.state.addableServices.filter(
        (service) => service.key != value,
      ),
      configData: newconfigData,
    });
  }

  /**
   * 获取可配置服务的模块的默认值
   */
  private getDefaultServiceConfig() {
    return ConfigServiceKeys.reduce((preValue, value) => {
      return {
        ...preValue,
        [value.key]: !!this.state.addableServices.find(
          (services) => services.key === value.key,
        )
          ? null
          : DefaultConfigData[value.key],
      };
    }, {});
  }

  /**
   * 设置下一步按钮是否灰化
   * @param value boole
   */
  protected setNextStepButtonDisable(value) {
    this.setState({
      nextStepButtonDisable: value,
    });
  }

  /**
   * 更新节点配置form实例
   */
  public updateNodeForm(value) {
    this.setState({
      nodeForm: {
        ...value,
      },
    });
  }

  /**
   * 更新网络配置form实例
   */
  public updateNetworkForm(value) {
    this.setState({
      networkForm: {
        ...value,
      },
    });
  }

  /**
   * 更新仓库配置form实例
   */
  public updateCRForm(value) {
    this.setState({
      crForm: {
        ...value,
      },
    });
  }

  /**
   * 更新基础服务配置form实例
   */
  public updateDataBaseForm(value) {
    this.setState((preState) => {
      return {
        dataBaseForm: {
          ...preState.dataBaseForm,
          ...value,
        },
      };
    });
  }

  /**
   * 更新连接配置form实例
   */
  public updateConnectInfoForm(value) {
    this.setState((preState) => {
      return {
        connectInfoForm: {
          ...preState.connectInfoForm,
          ...value,
        },
      };
    });
  }

  /**
   * 更新节点校验状态
   */
  public updateNodesValidateState() {
    this.setState({
      nodesValidateState: ValidateState.Normal,
    });
  }

  /**
   * 更新网络配置节点校验状态
   */
  public updateNetworkNodesValidateState() {
    this.setState({
      networkNodesValidateState: ValidateState.Normal,
    });
  }

  /**
   * 更新仓库配置节点校验状态
   */
  public updateCRNodesValidateState() {
    this.setState({
      crNodesValidateState: ValidateState.Normal,
    });
  }

  /**
   * 更新基础服务配置grafana节点校验状态
   */
  public updateGrafanaNodesValidateState() {
    this.setState({
      grafanaNodesValidateState: ValidateState.Normal,
    });
  }

  /**
   * 更新基础服务配置prometheus节点校验状态
   */
  public updatePrometheusNodesValidateState() {
    this.setState({
      prometheusNodesValidateState: ValidateState.Normal,
    });
  }

  /**
   * 更新基础服务配置proton monitor节点校验状态
   */
  public updateMonitorNodesValidateState() {
    this.setState({
      monitorNodesValidateState: ValidateState.Normal,
    });
  }

  /**
   * 更新连接配置校验状态
   */
  public updateConnectInfoValidateState(
    value: Partial<ConnectInfoValidateState>,
  ) {
    this.setState((preState) => {
      return {
        connectInfoValidateState: {
          ...preState.connectInfoValidateState,
          ...value,
        },
      };
    });
  }

  /**
   * 校验节点配置
   */
  public checkNodeConfig() {
    const nodeCheck = new Promise<void>((resolve, reject) => {
      if (!this.state.configData.nodesInfo.length) {
        this.setState({
          nodesValidateState: ValidateState.Empty,
        });
        reject();
      } else {
        this.setState({
          nodesValidateState: ValidateState.Normal,
        });
        resolve();
      }
    });
    const nodeFormCheck = Object.entries(this.state.nodeForm).map(
      async (form) => {
        if (
          this.state.configData.chrony.mode !== CHRONY_MODE.EXTERNAL_NTP &&
          form[0] === NodeForm.ServerForm
        ) {
          return true;
        } else {
          return await form[1].current.validateFields();
        }
      },
    );
    Promise.all([...nodeFormCheck, nodeCheck])
      .then(() => {
        this.onChangeStepStatus(OptionSteps.NetworkConfig);
      })
      .catch(() => {});
  }

  /**
   * 校验网络配置
   */
  public checkNetworkConfig() {
    if (
      this.state.dataBaseStorageType === DataBaseStorageType.DepositKubernetes
    ) {
      this.onChangeStepStatus(OptionSteps.RepositoryConfig);
    } else {
      const nodeCheck = new Promise<void>((resolve, reject) => {
        if (!this.state.configData.networkInfo.master.length) {
          this.setState({
            networkNodesValidateState: ValidateState.Empty,
          });
          reject();
        } else {
          this.setState({
            networkNodesValidateState: ValidateState.Normal,
          });
          resolve();
        }
      });
      const networkFormCheck = Object.values(this.state.networkForm).map(
        async (form) => {
          return await form.current.validateFields();
        },
      );
      Promise.all([...networkFormCheck, nodeCheck])
        .then(() => {
          this.onChangeStepStatus(OptionSteps.RepositoryConfig);
        })
        .catch(() => {});
    }
  }

  /**
   * 校验仓库配置
   */
  public checkRepositoryConfig() {
    const nodeCheck = new Promise<void>((resolve, reject) => {
      if (this.crType === CRType.ExternalCRConfig) {
        return resolve();
      }
      if (!this.state.configData.cr.local.master.length) {
        this.setState({
          crNodesValidateState: ValidateState.Empty,
        });
        reject();
      } else {
        this.setState({
          crNodesValidateState: ValidateState.Normal,
        });
        resolve();
      }
    });
    const crFormCheck = Object.values(
      this.crType === CRType.LOCAL
        ? this.state.crForm.localForm
        : this.state.crForm.externalForm,
    ).map(async (form) => {
      return await form.current?.validateFields();
    });
    Promise.all([...crFormCheck, nodeCheck])
      .then(() => {
        this.onChangeStepStatus(OptionSteps.DataBaseConfig);
      })
      .catch(() => {});
  }

  /**
   * 校验基础服务配置
   */
  public checkDataBaseConfig() {
    const nodesCheck = new Promise<void>((resolve, reject) => {
      const grafana = this.state.configData[SERVICES.Grafana];
      const prometheus = this.state.configData[SERVICES.Prometheus];
      const monitor = this.state.configData[SERVICES.ProtonMonitor];
      if (grafana?.hosts?.length > NODES_LIMIT[SERVICES.Grafana]) {
        this.setState({
          grafanaNodesValidateState: ValidateState.NodesNumError,
        });
        reject();
      }
      if (prometheus?.hosts?.length > NODES_LIMIT[SERVICES.Prometheus]) {
        this.setState({
          prometheusNodesValidateState: ValidateState.NodesNumError,
        });
        reject();
      }
      if (monitor?.hosts?.length > NODES_LIMIT[SERVICES.ProtonMonitor]) {
        this.setState({
          monitorNodesValidateState: ValidateState.NodesNumError,
        });
        reject();
      }
      resolve();
    });
    const dataBaseFormCheck = Object.values(this.state.dataBaseForm)
      .filter((serviceForm) => serviceForm)
      .map(async (serviceForm) => {
        if (serviceForm.current) {
          return await serviceForm.current.validateFields();
        } else {
          return await Promise.all(
            Object.values(serviceForm).map(async (form: any) => {
              return await form?.current?.validateFields();
            }),
          );
        }
      });
    Promise.all([...dataBaseFormCheck, nodesCheck])
      .then(() => {
        this.onChangeStepStatus(OptionSteps.ConnectInfo);
      })
      .catch(() => {});
  }

  /**
   * 校验连接配置
   */
  public checkConnectInfoConfig() {
    const configCheck = new Promise<void>((resolve, reject) => {
      const resource_connect_info =
        this.state.configData?.resource_connect_info;
      // rds类型
      if (
        resource_connect_info?.rds &&
        resource_connect_info?.rds.source_type === SOURCE_TYPE.EXTERNAL &&
        !Object.values(RDS_TYPE).includes(resource_connect_info?.rds?.rds_type)
      ) {
        this.updateConnectInfoValidateState({ RDS_TYPE: ValidateState.Empty });
        reject();
      }
      // mongodb ssl
      if (
        resource_connect_info?.mongodb &&
        resource_connect_info?.mongodb.source_type === SOURCE_TYPE.EXTERNAL &&
        ![true, false].includes(resource_connect_info?.mongodb?.ssl)
      ) {
        this.updateConnectInfoValidateState({
          MONGODB_SSL: ValidateState.Empty,
        });
        reject();
      }
      // redis连接模式
      if (
        resource_connect_info?.redis &&
        resource_connect_info?.redis.source_type === SOURCE_TYPE.EXTERNAL &&
        !Object.values(REDIS_CONNECT_TYPE).includes(
          resource_connect_info?.redis?.connect_type,
        )
      ) {
        this.updateConnectInfoValidateState({
          REDIS_CONNECT_TYPE: ValidateState.Empty,
        });
        reject();
      }
      // mq资源类型
      if (
        resource_connect_info?.mq &&
        !Object.values(SOURCE_TYPE).includes(
          resource_connect_info?.mq?.source_type,
        )
      ) {
        this.updateConnectInfoValidateState({
          MQ_RADIO: ValidateState.Empty,
        });
        reject();
      }
      // mq类型
      if (
        resource_connect_info?.mq &&
        !Object.values(MQ_TYPE).includes(resource_connect_info?.mq?.mq_type)
      ) {
        this.updateConnectInfoValidateState({
          MQ_TYPE: ValidateState.Empty,
        });
        reject();
      }
      // mq认证机制
      if (
        resource_connect_info?.mq?.auth &&
        !Object.values(MQ_AUTH_MACHANISM).includes(
          resource_connect_info?.mq?.auth?.mechanism,
        )
      ) {
        this.updateConnectInfoValidateState({
          MQ_AUTH_MACHANISM: ValidateState.Empty,
        });
        reject();
      }
      // opensearch版本
      if (
        resource_connect_info?.opensearch &&
        !Object.values(OPENSEARCH_VERSION).includes(
          resource_connect_info?.opensearch?.version,
        )
      ) {
        this.updateConnectInfoValidateState({
          OPENSEARCH_VERSION: ValidateState.Empty,
        });
        reject();
      }
      resolve();
    });
    const connectInfoFormCheck = Object.values(this.state.connectInfoForm).map(
      async (form) => {
        return await form?.current?.validateFields();
      },
    );
    Promise.all([...connectInfoFormCheck, configCheck])
      .then(() => {
        this.onInitPlatform();
      })
      .catch(() => {});
  }
}

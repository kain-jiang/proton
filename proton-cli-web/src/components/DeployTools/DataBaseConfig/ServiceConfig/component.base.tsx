import * as React from "react";
import {
  DefaultConfigData,
  NODES_LIMIT,
  SERVICES,
  getDefaultNodes,
} from "../../helper";
import { eq, isEqual, isArray } from "lodash";
import { FormInstance, message } from "@aishutech/ui";
import { ServiceType } from "./index";

export default class ServiceConfigBase extends React.Component<
  ServiceType.Props,
  ServiceType.State
> {
  state = {
    serviceConfig: null,
    serviceNodes: [],
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, dataBaseStorageType, service, updateDataBaseForm } =
      this.props;
    updateDataBaseForm(this.form);
    this.initConfig(
      service,
      configData[service.key],
      configData.nodesInfo,
      dataBaseStorageType,
    );
  }

  componentDidUpdate(
    prevProps: Readonly<ServiceType.Props>,
    prevState: Readonly<ServiceType.State>,
    snapshot?: any,
  ): void {
    const { configData, dataBaseStorageType, service } = this.props;
    if (!eq(prevProps.configData[service.key], configData[service.key])) {
      this.initConfig(
        service,
        configData[service.key],
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
  private initConfig(service, oldConfig, nodesInfo, dataBaseStorageType) {
    let nodes = [];
    if (service.key === SERVICES.Prometheus) {
      nodes = getDefaultNodes(
        nodesInfo ? nodesInfo : [],
        isArray(oldConfig?.hosts) ? oldConfig?.hosts : [],
        NODES_LIMIT.prometheus,
        true,
      );
    } else if (service.key === SERVICES.Grafana) {
      nodes = getDefaultNodes(
        nodesInfo ? nodesInfo : [],
        isArray(oldConfig?.hosts) ? oldConfig?.hosts : [],
        NODES_LIMIT.grafana,
        true,
      );
    } else {
      nodes = getDefaultNodes(
        nodesInfo ? nodesInfo : [],
        isArray(oldConfig?.hosts) ? oldConfig?.hosts : [],
      );
    }
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
        serviceNodes: nodes,
        serviceConfig: newConfig,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.serviceConfig,
        });
      },
    );
    if (!eq(newConfig, oldConfig)) {
      this.props.onUpdateServiceData(newConfig, dataBaseStorageType);
    }
  }

  /**
   * 更新adminUser, admin_passwd，pathData
   * @param config
   */
  public onChangeService(config) {
    const serviceConfig = {
      ...this.state.serviceConfig,
      ...config,
    };
    this.setState({ serviceConfig }, () => {
      this.props.onUpdateServiceData(
        serviceConfig,
        this.props.dataBaseStorageType,
      );
    });
  }

  /**
   * 更新env
   * @param config
   */
  public onChangeEnv(config) {
    const serviceConfig = {
      ...this.state.serviceConfig,
      env: {
        ...this.state.serviceConfig?.env,
        ...config,
      },
    };
    this.setState({ serviceConfig }, () => {
      this.props.onUpdateServiceData(
        serviceConfig,
        this.props.dataBaseStorageType,
      );
    });
  }

  /**
   * 修改资源配置触发
   * @param resources limits | requests
   * @param key
   * @param val
   */
  public onChangeResource(resources, key, val) {
    let serviceConfig = { ...this.state.serviceConfig };
    if (resources === "limits") {
      serviceConfig = {
        ...serviceConfig,
        resources: {
          ...serviceConfig?.resources,
          limits: {
            ...serviceConfig?.resources?.limits,
            [key]: val,
          },
        },
      };
    } else {
      serviceConfig = {
        ...serviceConfig,
        resources: {
          ...serviceConfig?.resources,
          requests: {
            ...serviceConfig?.resources?.requests,
            [key]: val,
          },
        },
      };
    }
    this.setState(
      {
        serviceConfig,
      },
      () => {
        this.props.onUpdateServiceData(
          serviceConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 是否配置资源限制
   */
  public onChangeServiceConfigResources(val) {
    const serviceConfig = {
      ...this.state.serviceConfig,
      resources: val
        ? DefaultConfigData[this.props.service.key].resources
        : undefined,
    };
    this.setState(
      {
        serviceConfig,
      },
      () => {
        this.props.onUpdateServiceData(
          serviceConfig,
          this.props.dataBaseStorageType,
        );
      },
    );
  }

  /**
   * 是否配置子组件资源限制
   */
  public onChangeServiceConfigComponentResources(val) {
    const serviceConfig = {
      ...this.state.serviceConfig,
      exporter_resources: val
        ? DefaultConfigData[this.props.service.key]?.exporter_resources
        : undefined,
    };

    this.setState(
      {
        serviceConfig,
      },
      () => {
        this.props.onUpdateServiceData(
          serviceConfig,
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
    let serviceConfig = { ...this.state.serviceConfig };

    serviceConfig = {
      ...serviceConfig,
      exporter_resources: {
        ...serviceConfig.exporter_resources,
        requests: {
          ...serviceConfig?.exporter_resources?.requests,
          [key]: val,
        },
      },
    };

    this.setState(
      {
        serviceConfig,
      },
      () => {
        this.props.onUpdateServiceData(
          serviceConfig,
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
  public onChangeServiceNode(nodes, config) {
    const { service } = this.props;
    if (
      service.key === SERVICES.Prometheus &&
      nodes &&
      nodes.length > NODES_LIMIT.prometheus
    ) {
      message.info(
        `Prometheus 部署节点数量应当小于等于 ${NODES_LIMIT.prometheus}`,
      );
    }
    if (
      service.key === SERVICES.Grafana &&
      nodes &&
      nodes.length > NODES_LIMIT.grafana
    ) {
      message.info(`Grafana 部署节点只允许 ${NODES_LIMIT.grafana} 节点`);
    }
    if (service.key === SERVICES.Grafana) {
      this.props.updateGrafanaNodesValidateState();
    }
    if (service.key === SERVICES.Prometheus) {
      this.props.updatePrometheusNodesValidateState();
    }
    this.props.onUpdateServiceData(
      {
        ...config,
        hosts: nodes.map((node) => node.name),
      },
      this.props.dataBaseStorageType,
    );
  }

  /**
   * 更新外部端口列表
   * @param index 数组索引
   * @param updateData 要更新的数据
   */
  public onChangeExternalServiceList(index, updateData) {
    let serviceConfig = { ...this.state.serviceConfig };
    const externalServiceList = [
      ...(serviceConfig.external_service_list || []),
    ];

    externalServiceList[index] = {
      ...externalServiceList[index],
      ...updateData,
    };

    serviceConfig = {
      ...serviceConfig,
      external_service_list: externalServiceList,
    };

    this.setState({ serviceConfig }, () => {
      this.props.onUpdateServiceData(
        serviceConfig,
        this.props.dataBaseStorageType,
      );
    });
  }

  /**
   * 删除外部端口列表
   * @param index 要删除的索引
   */
  public deleteExternalServiceList(index) {
    const serviceConfig = { ...this.state.serviceConfig };
    const externalServiceList = [
      ...(serviceConfig.external_service_list || []),
    ];

    externalServiceList.splice(index, 1);
    serviceConfig.external_service_list = externalServiceList;

    this.setState({ serviceConfig }, () => {
      this.props.onUpdateServiceData(
        serviceConfig,
        this.props.dataBaseStorageType,
      );
    });
  }

  /**
   * 添加新的外部端口信息
   */
  public addExternalServiceList() {
    let serviceConfig = { ...this.state.serviceConfig };
    serviceConfig = {
      ...serviceConfig,
      external_service_list: [
        ...serviceConfig.external_service_list,
        { name: "", port: null, nodePortBase: null, enableSSL: false },
      ],
    };

    this.setState({ serviceConfig }, () => {
      this.props.onUpdateServiceData(
        serviceConfig,
        this.props.dataBaseStorageType,
      );
    });
  }
}

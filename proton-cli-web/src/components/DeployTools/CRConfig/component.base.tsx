import * as React from "react";
import {
  DefaultConfigData,
  getDefaultNodes,
  CRType,
  RepositoryType,
} from "../helper";
import { CRConfigType } from "./index";
import { FormInstance } from "@aishutech/ui";
import { ExternalCRConfig } from "..";

export default class CRConfigBase extends React.Component<
  CRConfigType.Props,
  CRConfigType.State
> {
  state = {
    crConfig: DefaultConfigData.cr.local,
    externalCRConfig: {
      image_repository: RepositoryType.Registry,
      chart_repository: RepositoryType.Chartmuseum,
      registry: {
        host: "",
        username: "",
        password: "",
      },
      chartmuseum: {
        host: "",
        username: "",
        password: "",
      },
      oci: {
        registry: "",
        username: "",
        password: "",
        plain_http: false,
      },
    } as ExternalCRConfig,
    nodes: [],
    selectCRType: CRType.LOCAL,
  };

  crForm = {
    localForm: {
      portsForm: React.createRef<FormInstance>(),
      haPortsForm: React.createRef<FormInstance>(),
      storageForm: React.createRef<FormInstance>(),
    },
    externalForm: {
      chartmuseumForm: React.createRef<FormInstance>(),
      registryForm: React.createRef<FormInstance>(),
      ociForm: React.createRef<FormInstance>(),
    },
  };
  componentDidMount(): void {
    this.props.updateCRForm(this.crForm);
    if (this.props.cRType === CRType.LOCAL) {
      const masterNode = getDefaultNodes(
        this.props.configData.nodesInfo,
        this.props.configData.cr.local.master,
        2,
      );
      this.setState(
        {
          crConfig: {
            ...this.props.configData.cr.local,
            master: masterNode.map((node) => node.name),
          },
          nodes: masterNode,
          selectCRType: CRType.LOCAL,
        },
        () => {
          this.props.onUpdateCRConfig({ local: this.state.crConfig });
          this.crForm.localForm.portsForm.current.setFieldsValue({
            ...this.state.crConfig.ports,
          });
          this.crForm.localForm.haPortsForm.current.setFieldsValue({
            ...this.state.crConfig.haPorts,
          });
          this.crForm.localForm.storageForm.current.setFieldsValue({
            ...this.state.crConfig,
          });
        },
      );
    } else {
      const masterNode = getDefaultNodes(
        this.props.configData.nodesInfo,
        this.state.crConfig.master,
        2,
      );
      const { externalCRConfig } = this.state;
      this.setState(
        {
          externalCRConfig: this.props.configData.cr.external
            ? this.props.configData.cr.external
            : externalCRConfig,
          selectCRType: CRType.ExternalCRConfig,
          nodes: masterNode,
        },
        () => {
          this.crForm.externalForm.chartmuseumForm.current?.setFieldsValue({
            ...this.state.externalCRConfig.chartmuseum,
          });
          this.crForm.externalForm.registryForm.current?.setFieldsValue({
            ...this.state.externalCRConfig.registry,
          });
          this.crForm.externalForm.ociForm.current?.setFieldsValue({
            ...this.state.externalCRConfig.oci,
          });
        },
      );
    }
  }

  /**
   * 部署节点
   * @param nodes 部署节点
   */
  public onChangeMasterNode(nodes) {
    this.setState(
      {
        nodes,
        crConfig: {
          ...this.state.crConfig,
          master: nodes.map((value) => value.name),
        },
      },
      () => {
        this.props.onUpdateCRConfig({ local: this.state.crConfig });
        this.props.updateCRNodesValidateState();
      },
    );
  }

  /**
   * 修改端口配置
   * @param value 被修改的配置
   */
  public onChangePorts(config) {
    this.setState(
      {
        crConfig: {
          ...this.state.crConfig,
          ports: {
            ...this.state.crConfig.ports,
            ...config,
          },
        },
      },
      () => {
        this.props.onUpdateCRConfig({ local: this.state.crConfig });
      },
    );
  }
  /**
   * 修改端口配置
   * @param config 被修改的配置
   */

  public onChangeHaPorts(config) {
    this.setState(
      {
        crConfig: {
          ...this.state.crConfig,
          haPorts: {
            ...this.state.crConfig.haPorts,
            ...config,
          },
        },
      },
      () => {
        this.props.onUpdateCRConfig({ local: this.state.crConfig });
      },
    );
  }

  /**
   * 修改数据路径
   * @param value
   */
  public onChangeStorage(storage) {
    this.setState(
      {
        crConfig: {
          ...this.state.crConfig,
          storage,
        },
      },
      () => {
        this.props.onUpdateCRConfig({ local: this.state.crConfig });
      },
    );
  }

  /**
   * 切换表单
   */
  public onChangeCRType(value) {
    this.setState({
      selectCRType: value,
    });
    this.props.onUpDateCRTypeConfig(value);
    value
      ? this.props.onUpdateCRConfig({ external: this.state.externalCRConfig })
      : this.props.onUpdateCRConfig({
          local: {
            ...this.state.crConfig,
            master: this.state.nodes.map((node) => node.name),
          },
        });
  }

  /**
   *  配置容器仓库
   */
  public onChangeregistryConfig(value) {
    this.setState(
      {
        externalCRConfig: {
          ...this.state.externalCRConfig,
          registry: {
            ...this.state.externalCRConfig.registry,
            ...value,
          },
        },
      },
      () => {
        this.props.onUpdateCRConfig({ external: this.state.externalCRConfig });
      },
    );
  }

  /**
   *  配置chart仓库
   */
  public onChangeChartmuseumConfig(value) {
    this.setState(
      {
        externalCRConfig: {
          ...this.state.externalCRConfig,
          chartmuseum: {
            ...this.state.externalCRConfig.chartmuseum,
            ...value,
          },
        },
      },
      () => {
        this.props.onUpdateCRConfig({ external: this.state.externalCRConfig });
      },
    );
  }

  /**
   *  配置oci仓库
   */
  public onChangeOCIConfig(value) {
    this.setState(
      {
        externalCRConfig: {
          ...this.state.externalCRConfig,
          oci: {
            ...this.state.externalCRConfig.oci,
            ...value,
          },
        },
      },
      () => {
        this.props.onUpdateCRConfig({ external: this.state.externalCRConfig });
      },
    );
  }

  /**
   *  配置仓库类型
   */
  public onChangeRepository(value) {
    this.setState(
      {
        externalCRConfig: {
          ...this.state.externalCRConfig,
          ...value,
        },
      },
      () => {
        this.props.onUpdateCRConfig({ external: this.state.externalCRConfig });
      },
    );
  }
}

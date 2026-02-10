import * as React from "react";
import {
  Form,
  Input,
  Row,
  Col,
  Radio,
  Divider,
  Alert,
  Select,
  Switch,
} from "@aishutech/ui";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import {
  CRType,
  DataBaseStorageType,
  RepositoryType,
  emptyValidatorRules,
  portValidatorRules,
} from "../helper";
import SelectNode from "../SelectNode/component.view";
import CRConfigBase from "./component.base";
import "./styles.view.scss";

export default class CRConfig extends CRConfigBase {
  render(): React.ReactNode {
    return (
      <div className="cr-config-contain">
        {this.props.dataBaseStorageType === DataBaseStorageType.Standard ? (
          <Radio.Group
            style={{
              margin: "10px 0",
            }}
            value={this.state.selectCRType}
            buttonStyle="solid"
            onChange={(e) => {
              this.onChangeCRType(e.target.value);
            }}
          >
            <Radio.Button value={CRType.LOCAL}>本地容器仓库</Radio.Button>
            <Radio.Button value={CRType.ExternalCRConfig}>
              第三方容器仓库
            </Radio.Button>
          </Radio.Group>
        ) : null}
        {this.getCRForm()}
      </div>
    );
  }

  /**
   * 获取本地CR的配置表单
   * @returns
   */
  getLocalCRConfig() {
    return (
      <div
        style={{
          margin: "10px 0",
        }}
      >
        <Form>
          <Form.Item label="部署节点" required>
            <SelectNode
              mode={false}
              nodes={this.props.configData.nodesInfo}
              selectedNodes={this.state.nodes}
              onSelectedChange={this.onChangeMasterNode.bind(this)}
            />
            {this.props.crNodesValidateState ? (
              <div style={{ color: "#FF4D4F" }}>请至少添加一个节点</div>
            ) : null}
          </Form.Item>
        </Form>
        <Divider orientation="left" orientationMargin="0">
          端口配置
        </Divider>
        <Form
          layout="inline"
          validateTrigger="onBlur"
          initialValues={this.state.crConfig.ports}
          ref={this.crForm.localForm.portsForm}
        >
          <Form.Item
            label="Chart仓库"
            name="chartMuseum"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.ports.chartMuseum}
              onChange={(e) => {
                this.onChangePorts({ chartMuseum: e.target.value });
              }}
            />
          </Form.Item>
          <Form.Item
            label="容器仓库"
            name="crManager"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.ports.crManager}
              onChange={(e) => {
                this.onChangePorts({ crManager: e.target.value });
              }}
            />
          </Form.Item>
          <Form.Item
            label="RPM仓库"
            name="rpm"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.ports.rpm}
              onChange={(e) => {
                this.onChangePorts({ rpm: e.target.value });
              }}
            />
          </Form.Item>
          <Form.Item
            label="Registry"
            name="registry"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.ports.registry}
              onChange={(e) => {
                this.onChangePorts({ registry: e.target.value });
              }}
            />
          </Form.Item>
        </Form>
        <Divider orientation="left" orientationMargin="0">
          高可用端口配置
        </Divider>
        <Form
          layout="inline"
          validateTrigger="onBlur"
          initialValues={this.state.crConfig.haPorts}
          ref={this.crForm.localForm.haPortsForm}
        >
          <Form.Item
            label="Chart仓库"
            name="chartMuseum"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.haPorts.chartMuseum}
              onChange={(e) => {
                this.onChangeHaPorts({ chartMuseum: e.target.value });
              }}
            />
          </Form.Item>
          <Form.Item
            label="容器仓库"
            name="crManager"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.haPorts.crManager}
              onChange={(e) => {
                this.onChangeHaPorts({ crManager: e.target.value });
              }}
            />
          </Form.Item>
          <Form.Item
            label="RPM仓库"
            name="rpm"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.haPorts.rpm}
              onChange={(e) => {
                this.onChangeHaPorts({ rpm: e.target.value });
              }}
            />
          </Form.Item>
          <Form.Item
            label="Registry"
            name="registry"
            required
            rules={portValidatorRules}
          >
            <Input
              style={{ width: "100px" }}
              value={this.state.crConfig.haPorts.registry}
              onChange={(e) => {
                this.onChangeHaPorts({ registry: e.target.value });
              }}
            />
          </Form.Item>
        </Form>
        <div
          style={{
            margin: "20px 0",
          }}
        >
          <Form
            validateTrigger="onBlur"
            initialValues={this.state.crConfig}
            ref={this.crForm.localForm.storageForm}
          >
            <Form.Item
              label="chart与image的存储路径"
              name="storage"
              required
              rules={emptyValidatorRules}
            >
              <Input
                value={this.state.crConfig.storage}
                onChange={(e) => {
                  this.onChangeStorage(e.target.value);
                }}
              />
            </Form.Item>
          </Form>
        </div>
      </div>
    );
  }

  /**
   * 获取第三方CR的表单
   */
  getExternalCRConfig() {
    return (
      <>
        <Alert
          message='如果使用 Harbor 2.x， 可同时作为registry、oci、chartmuseum仓库提供，确保账号具有推送权限，且包含以下项目：["proton", "public", "ict", "as", "dip", 以及作为Chart存储的项目]'
          type="info"
        />
        <Form
          layout="horizontal"
          style={{ width: "420px", paddingTop: "24px" }}
        >
          <Form.Item label="镜像仓库" labelAlign="left" labelCol={{ span: 4 }}>
            <Select
              value={this.state.externalCRConfig.image_repository}
              style={{ width: "330px" }}
              onChange={(value) =>
                this.onChangeRepository({ image_repository: value })
              }
              options={[
                {
                  value: RepositoryType.Registry,
                  label: RepositoryType.Registry,
                },
                { value: RepositoryType.OCI, label: RepositoryType.OCI },
              ]}
            />
          </Form.Item>
          <Form.Item label="Chart仓库" labelAlign="left" labelCol={{ span: 4 }}>
            <Select
              value={this.state.externalCRConfig.chart_repository}
              style={{ width: "330px" }}
              onChange={(value) =>
                this.onChangeRepository({ chart_repository: value })
              }
              options={[
                {
                  value: RepositoryType.Chartmuseum,
                  label: RepositoryType.Chartmuseum,
                },
                { value: RepositoryType.OCI, label: RepositoryType.OCI },
              ]}
            />
          </Form.Item>
        </Form>
        <Row>
          {this.state.externalCRConfig.image_repository ===
          RepositoryType.Registry ? (
            <Col span={12}>
              <Divider
                orientation="left"
                orientationMargin="0"
                style={{ width: "420px" }}
              >
                registry
              </Divider>
              <Form
                layout="horizontal"
                style={{ width: "420px" }}
                validateTrigger="onBlur"
                initialValues={this.state.externalCRConfig.registry}
                ref={this.crForm.externalForm.registryForm}
              >
                <Form.Item
                  label="地址"
                  name="host"
                  required
                  rules={emptyValidatorRules}
                  labelAlign="left"
                  labelCol={{ span: 4 }}
                >
                  <div>
                    <Input
                      value={this.state.externalCRConfig.registry.host}
                      onChange={(e) => {
                        this.onChangeregistryConfig({ host: e.target.value });
                      }}
                      style={{
                        width: "330px",
                      }}
                      placeholder="请输入有效的registry仓库地址"
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={
                        '格式为 IP:PORT或者域名:PORT（https默认为443时PORT可以省略）。\n如果 registry 仓库由 Harbor 2.x 提供，请确保仓库包含以下项目：["proton", "public", "ict", "as"]'
                      }
                    />
                  </div>
                </Form.Item>
                <Form.Item
                  label="账号"
                  labelAlign="left"
                  labelCol={{ span: 4 }}
                >
                  <Input
                    value={this.state.externalCRConfig.registry.username}
                    onChange={(e) => {
                      this.onChangeregistryConfig({ username: e.target.value });
                    }}
                    style={{
                      width: "330px",
                    }}
                  />
                </Form.Item>
                <Form.Item
                  label="密码"
                  labelAlign="left"
                  labelCol={{ span: 4 }}
                >
                  <Input.Password
                    value={this.state.externalCRConfig.registry.password}
                    onChange={(e) => {
                      this.onChangeregistryConfig({ password: e.target.value });
                    }}
                    style={{
                      width: "330px",
                    }}
                  />
                </Form.Item>
              </Form>
            </Col>
          ) : null}
          {this.state.externalCRConfig.chart_repository ===
          RepositoryType.Chartmuseum ? (
            <Col span={12}>
              <Divider
                orientation="left"
                orientationMargin="0"
                style={{ width: "420px" }}
              >
                chartmuseum
              </Divider>
              <Form
                layout="horizontal"
                style={{ width: "420px" }}
                labelAlign="left"
                validateTrigger="onBlur"
                initialValues={this.state.externalCRConfig.chartmuseum}
                ref={this.crForm.externalForm.chartmuseumForm}
              >
                <Form.Item
                  label="地址"
                  name="host"
                  required
                  rules={emptyValidatorRules}
                  labelCol={{ span: 4 }}
                >
                  <div>
                    <Input
                      value={this.state.externalCRConfig.chartmuseum.host}
                      onChange={(e) => {
                        this.onChangeChartmuseumConfig({
                          host: e.target.value,
                        });
                      }}
                      style={{
                        width: "330px",
                      }}
                      placeholder="请输入有效的chartmuseum服务地址(包含协议)"
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={
                        "如果 chartmuseum 仓库由 Harbor 2.x 提供，仓库地址请具体到某一项目，例如：https://acr.domain.cn/chartrepo/all-charts"
                      }
                    />
                  </div>
                </Form.Item>
                <Form.Item label="账号" labelCol={{ span: 4 }}>
                  <Input
                    value={this.state.externalCRConfig.chartmuseum.username}
                    onChange={(e) => {
                      this.onChangeChartmuseumConfig({
                        username: e.target.value,
                      });
                    }}
                    style={{
                      width: "330px",
                    }}
                  />
                </Form.Item>
                <Form.Item label="密码" labelCol={{ span: 4 }}>
                  <Input.Password
                    value={this.state.externalCRConfig.chartmuseum.password}
                    onChange={(e) => {
                      this.onChangeChartmuseumConfig({
                        password: e.target.value,
                      });
                    }}
                    style={{
                      width: "330px",
                    }}
                  />
                </Form.Item>
              </Form>
            </Col>
          ) : null}
          {this.state.externalCRConfig.image_repository ===
            RepositoryType.OCI ||
          this.state.externalCRConfig.chart_repository ===
            RepositoryType.OCI ? (
            <Col span={12}>
              <Divider
                orientation="left"
                orientationMargin="0"
                style={{ width: "420px" }}
              >
                oci
              </Divider>
              <Form
                layout="horizontal"
                style={{ width: "420px" }}
                validateTrigger="onBlur"
                labelAlign="left"
                initialValues={this.state.externalCRConfig.oci}
                ref={this.crForm.externalForm.ociForm}
              >
                <Form.Item
                  label="地址"
                  name="registry"
                  required
                  rules={emptyValidatorRules}
                  labelCol={{ span: 4 }}
                >
                  <div>
                    <Input
                      value={this.state.externalCRConfig.oci.registry}
                      onChange={(e) => {
                        this.onChangeOCIConfig({ registry: e.target.value });
                      }}
                      style={{
                        width: "330px",
                      }}
                      placeholder="请输入有效的oci仓库地址"
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={
                        "如果 oci 仓库由 Harbor 2.x 提供，仓库地址请具体到某一项目，例如：acr.domain.cn/all-charts"
                      }
                    />
                  </div>
                </Form.Item>
                <Form.Item label="账号" labelCol={{ span: 4 }}>
                  <Input
                    value={this.state.externalCRConfig.oci.username}
                    onChange={(e) => {
                      this.onChangeOCIConfig({
                        username: e.target.value,
                      });
                    }}
                    style={{
                      width: "330px",
                    }}
                  />
                </Form.Item>
                <Form.Item label="密码" labelCol={{ span: 4 }}>
                  <Input.Password
                    value={this.state.externalCRConfig.oci.password}
                    onChange={(e) => {
                      this.onChangeOCIConfig({
                        password: e.target.value,
                      });
                    }}
                    style={{
                      width: "330px",
                    }}
                  />
                </Form.Item>
                <Form.Item label="是否使用http" required>
                  <Switch
                    checked={this.state.externalCRConfig.oci.plain_http}
                    onChange={(value) => {
                      this.onChangeOCIConfig({
                        plain_http: value,
                      });
                    }}
                  />
                </Form.Item>
              </Form>
            </Col>
          ) : null}
        </Row>
      </>
    );
  }

  /**
   * 获取表单
   */
  getCRForm() {
    switch (this.state.selectCRType) {
      case CRType.LOCAL:
        return this.getLocalCRConfig();
      case CRType.ExternalCRConfig:
        return this.getExternalCRConfig();
    }
  }
}

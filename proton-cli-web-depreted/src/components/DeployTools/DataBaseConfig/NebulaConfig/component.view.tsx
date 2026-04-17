import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { Form, Input, Row, Col, Radio } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import NebulaConfigBase from "./component.base";
import {
  DataBaseStorageType,
  NEBULA_COMPONENTS,
  RESOURCES,
  RESOURCES_TYPE,
  emptyValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export default class NebulaConfig extends NebulaConfigBase {
  render(): React.ReactNode {
    const { configData, dataBaseStorageType, service } = this.props;
    const { nebulaConfig, nebulaNodes } = this.state;

    return (
      <div className="wrapper">
        <Row>
          <Col
            span={23}
            style={{
              color: "#000000",
              height: "30px",
              lineHeight: "30px",
              fontSize: "14px",
              fontWeight: "bold",
            }}
          >
            <span className="split"></span>
            <span className="title">{service.name}</span>
          </Col>
          <Col className="delete" span={1}>
            <DeleteOutlined
              onClick={this.props.onDeleteNebulaConfig.bind(this)}
            />
          </Col>
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
        <div>
          <Form
            style={{
              margin: "10px 0",
              width: "100%",
            }}
            name={service.key}
            validateTrigger="onBlur"
            initialValues={nebulaConfig}
            ref={this.form}
          >
            {dataBaseStorageType === DataBaseStorageType.Standard ? (
              <Form.Item
                labelAlign="left"
                labelCol={{ span: 4 }}
                label="部署节点"
                required
              >
                <SelectNode
                  mode={false}
                  nodes={configData.nodesInfo}
                  selectedNodes={nebulaNodes}
                  onSelectedChange={(nodes) =>
                    this.onChangeNebulaNode(nodes, nebulaConfig)
                  }
                />
              </Form.Item>
            ) : null}
            <Form.Item labelAlign="left" labelCol={{ span: 4 }} label="密码:">
              <Input.Password
                style={{ width: "200px" }}
                value={nebulaConfig?.password}
                onChange={(e) => {
                  this.onChangeNebula({ password: e.target.value });
                }}
              />
              <QuestionCircleOutlined
                style={{
                  marginLeft: "6px",
                }}
                title="Nebula Graph 的 root 帐户的密码，长度不超过 24。如果为空则使用生成的随机密码。"
              />
            </Form.Item>
            {dataBaseStorageType === DataBaseStorageType.Standard ? (
              <Form.Item
                labelAlign="left"
                labelCol={{ span: 4 }}
                label="数据路径:"
                name="data_path"
                required
                rules={emptyValidatorRules}
              >
                <Input
                  value={nebulaConfig?.data_path}
                  onChange={(e) => {
                    this.onChangeNebula({ data_path: e.target.value });
                  }}
                />
              </Form.Item>
            ) : null}
            <Form.Item label="Metad自定义配置:"></Form.Item>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="memory_tracker_limitratio:"
                  name={["metad", "config", "memory_tracker_limitratio"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      nebulaConfig?.metad?.config?.memory_tracker_limitratio
                    }
                    onChange={(e) => {
                      this.onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.METAD,
                        { memory_tracker_limitratio: e.target.value },
                      );
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="system_memory_high_watermark_ratio:"
                  name={[
                    "metad",
                    "config",
                    "system_memory_high_watermark_ratio",
                  ]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      nebulaConfig?.metad?.config
                        ?.system_memory_high_watermark_ratio
                    }
                    onChange={(e) => {
                      this.onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.METAD,
                        { system_memory_high_watermark_ratio: e.target.value },
                      );
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              labelAlign="left"
              labelCol={{ span: 4 }}
              label="enable_authorize:"
              name={["metad", "config", "enable_authorize"]}
              required
              rules={emptyValidatorRules}
            >
              <Radio.Group
                value={nebulaConfig?.metad?.config?.enable_authorize}
                onChange={(e) => {
                  this.onChangeNebulaComponentConfig(NEBULA_COMPONENTS.METAD, {
                    enable_authorize: e.target.value,
                  });
                }}
              >
                <Radio value={"true"}>是</Radio>
                <Radio value={"false"}>否</Radio>
              </Radio.Group>
            </Form.Item>

            <Form.Item label="graphd自定义配置:"></Form.Item>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="memory_tracker_limitratio:"
                  name={["graphd", "config", "memory_tracker_limitratio"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      nebulaConfig?.graphd?.config?.memory_tracker_limitratio
                    }
                    onChange={(e) => {
                      this.onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.GRAPHD,
                        { memory_tracker_limitratio: e.target.value },
                      );
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="system_memory_high_watermark_ratio:"
                  name={[
                    "graphd",
                    "config",
                    "system_memory_high_watermark_ratio",
                  ]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      nebulaConfig?.graphd?.config
                        ?.system_memory_high_watermark_ratio
                    }
                    onChange={(e) => {
                      this.onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.GRAPHD,
                        { system_memory_high_watermark_ratio: e.target.value },
                      );
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              labelAlign="left"
              labelCol={{ span: 4 }}
              label="enable_authorize:"
              name={["graphd", "config", "enable_authorize"]}
              required
              rules={emptyValidatorRules}
            >
              <Radio.Group
                value={nebulaConfig?.graphd?.config?.enable_authorize}
                onChange={(e) => {
                  this.onChangeNebulaComponentConfig(NEBULA_COMPONENTS.GRAPHD, {
                    enable_authorize: e.target.value,
                  });
                }}
              >
                <Radio value={"true"}>是</Radio>
                <Radio value={"false"}>否</Radio>
              </Radio.Group>
            </Form.Item>

            <Form.Item label="Storaged自定义配置:"></Form.Item>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="memory_tracker_limitratio:"
                  name={["storaged", "config", "memory_tracker_limitratio"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      nebulaConfig?.storaged?.config?.memory_tracker_limitratio
                    }
                    onChange={(e) => {
                      this.onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.STORAGED,
                        { memory_tracker_limitratio: e.target.value },
                      );
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelAlign="left"
                  label="system_memory_high_watermark_ratio:"
                  name={[
                    "storaged",
                    "config",
                    "system_memory_high_watermark_ratio",
                  ]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      nebulaConfig?.storaged?.config
                        ?.system_memory_high_watermark_ratio
                    }
                    onChange={(e) => {
                      this.onChangeNebulaComponentConfig(
                        NEBULA_COMPONENTS.STORAGED,
                        { system_memory_high_watermark_ratio: e.target.value },
                      );
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              labelAlign="left"
              labelCol={{ span: 4 }}
              label="enable_authorize:"
              name={["storaged", "config", "enable_authorize"]}
              required
              rules={emptyValidatorRules}
            >
              <Radio.Group
                value={nebulaConfig?.storaged?.config?.enable_authorize}
                onChange={(e) => {
                  this.onChangeNebulaComponentConfig(
                    NEBULA_COMPONENTS.STORAGED,
                    {
                      enable_authorize: e.target.value,
                    },
                  );
                }}
              >
                <Radio value={"true"}>是</Radio>
                <Radio value={"false"}>否</Radio>
              </Radio.Group>
            </Form.Item>
            <Form.Item label="Metad自定义配置资源限制:">
              <Radio.Group
                value={!!nebulaConfig?.metad?.resources}
                onChange={(e) => {
                  this.onChangeNebulaComponentResources(
                    NEBULA_COMPONENTS.METAD,
                    RESOURCES.ALL,
                    e.target.value,
                  );
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            {nebulaConfig?.metad?.resources ? (
              <>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.CPU:"
                      name={["metad", "resources", "limits", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={nebulaConfig?.metad?.resources?.limits?.cpu}
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.Memory:"
                      name={["metad", "resources", "limits", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={nebulaConfig?.metad?.resources?.limits?.memory}
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.CPU:"
                      name={["metad", "resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={nebulaConfig?.metad?.resources?.requests?.cpu}
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.Memory:"
                      name={["metad", "resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={
                            nebulaConfig?.metad?.resources?.requests?.memory
                          }
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.METAD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
            <Form.Item label="graphd自定义配置资源限制:">
              <Radio.Group
                value={!!nebulaConfig?.graphd?.resources}
                onChange={(e) => {
                  this.onChangeNebulaComponentResources(
                    NEBULA_COMPONENTS.GRAPHD,
                    RESOURCES.ALL,
                    e.target.value,
                  );
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            {nebulaConfig?.graphd?.resources ? (
              <>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.CPU:"
                      name={["graphd", "resources", "limits", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={nebulaConfig?.graphd?.resources?.limits?.cpu}
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.Memory:"
                      name={["graphd", "resources", "limits", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={
                            nebulaConfig?.graphd?.resources?.limits?.memory
                          }
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.CPU:"
                      name={["graphd", "resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={nebulaConfig?.graphd?.resources?.requests?.cpu}
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.Memory:"
                      name={["graphd", "resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={
                            nebulaConfig?.graphd?.resources?.requests?.memory
                          }
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.GRAPHD,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
            <Form.Item label="Storaged自定义配置资源限制:">
              <Radio.Group
                value={!!nebulaConfig?.storaged?.resources}
                onChange={(e) => {
                  this.onChangeNebulaComponentResources(
                    NEBULA_COMPONENTS.STORAGED,
                    RESOURCES.ALL,
                    e.target.value,
                  );
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            {nebulaConfig?.storaged?.resources ? (
              <>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.CPU:"
                      name={["storaged", "resources", "limits", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={nebulaConfig?.storaged?.resources?.limits?.cpu}
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Limits.Memory:"
                      name={["storaged", "resources", "limits", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={
                            nebulaConfig?.storaged?.resources?.limits?.memory
                          }
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.LIMITS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.CPU:"
                      name={["storaged", "resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={
                            nebulaConfig?.storaged?.resources?.requests?.cpu
                          }
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.CPU]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      label="Requests.Memory:"
                      name={["storaged", "resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={
                            nebulaConfig?.storaged?.resources?.requests?.memory
                          }
                          onChange={(e) => {
                            this.onChangeNebulaComponentResources(
                              NEBULA_COMPONENTS.STORAGED,
                              RESOURCES.REQUESTS,
                              {
                                [RESOURCES_TYPE.MEMORY]: e.target.value,
                              },
                            );
                          }}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            ) : null}
          </Form>
        </div>
      </div>
    );
  }
}

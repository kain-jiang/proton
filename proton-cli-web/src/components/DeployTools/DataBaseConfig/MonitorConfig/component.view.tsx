import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { Form, Input, Radio, Row, Col } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import MonitorConfigBase from "./component.base";
import {
  DefaultConfigData,
  RESOURCES,
  RESOURCES_TYPE,
  SERVICES,
  ValidateState,
  booleanEmptyValidatorRules,
  emptyValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export default class MonitorConfig extends MonitorConfigBase {
  render(): React.ReactNode {
    const { configData, service, monitorNodesValidateState } = this.props;
    const { monitorConfig, monitorNodes } = this.state;

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
              onClick={this.props.onDeleteMonitorConfig.bind(this)}
            />
          </Col>
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
        <Form
          layout="horizontal"
          name={service.key}
          validateTrigger="onBlur"
          initialValues={monitorConfig}
          ref={this.form}
        >
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label="部署节点"
            required
          >
            <SelectNode
              mode={false}
              nodes={configData.nodesInfo}
              selectedNodes={monitorNodes}
              onSelectedChange={(nodes) =>
                this.onChangeMonitorNode(nodes, monitorConfig)
              }
            />
            {monitorNodesValidateState === ValidateState.NodesNumError ? (
              <div
                style={{ color: "#FF4D4F" }}
              >{`Proton Monitor 部署节点数量应当小于等于2`}</div>
            ) : null}
          </Form.Item>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label="数据路径:"
            name="data_path"
            required
            rules={emptyValidatorRules}
          >
            <Input
              value={monitorConfig?.data_path}
              onChange={(e) => {
                this.onChangeMonitor({ data_path: e.target.value });
              }}
            />
          </Form.Item>
          <Row>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 8 }}
                labelAlign="left"
                label="指标数据保留时间:"
                name={["config", "vmetrics", "retention"]}
              >
                <Input
                  style={{ width: "200px" }}
                  value={monitorConfig?.config?.vmetrics?.retention}
                  placeholder="默认值：10d"
                  onChange={(e) => {
                    this.onChangeMonitorConfig({
                      vmetrics: {
                        retention: e.target.value,
                      },
                    });
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 8 }}
                labelAlign="left"
                label="日志数据保留时间:"
                name={["config", "vlogs", "retention"]}
              >
                <Input
                  style={{ width: "200px" }}
                  value={monitorConfig?.config?.vlogs?.retention}
                  placeholder="默认值：10d"
                  onChange={(e) => {
                    this.onChangeMonitorConfig({
                      vlogs: {
                        retention: e.target.value,
                      },
                    });
                  }}
                />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            label="转发指标到第三方:"
            labelCol={{ span: 4 }}
            labelAlign="left"
          >
            <Radio.Group
              value={!!monitorConfig?.config?.vmagent?.remoteWrite}
              onChange={(e) => {
                this.onChangeMonitorConfigVmagent(e.target.value);
              }}
            >
              <Radio value={true}>是</Radio>
              <Radio value={false}>否</Radio>
            </Radio.Group>
          </Form.Item>
          {monitorConfig?.config?.vmagent?.remoteWrite ? (
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="第三方服务器地址:"
              name={["config", "vmagent", "remoteWrite", "extraServers"]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{ width: "95%" }}
                  value={
                    monitorConfig?.config?.vmagent?.remoteWrite?.extraServers
                  }
                  onChange={(e) => {
                    this.onChangeMonitorConfig({
                      vmagent: {
                        remoteWrite: { extraServers: e.target.value },
                      },
                    });
                  }}
                />
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title={`多个第三方服务器地址使用,连接`}
                />
              </div>
            </Form.Item>
          ) : null}
          <Form.Item
            label="邮件告警配置:"
            labelCol={{ span: 4 }}
            labelAlign="left"
          >
            <Radio.Group
              value={monitorConfig?.config?.grafana?.smtp?.enabled}
              onChange={(e) => {
                this.onChangeMonitorConfigGrafana(e.target.value);
              }}
            >
              <Radio value={true}>是</Radio>
              <Radio value={false}>否</Radio>
            </Radio.Group>
          </Form.Item>
          {monitorConfig?.config?.grafana?.smtp?.enabled ? (
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="host:"
                  name={["config", "grafana", "smtp", "host"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={monitorConfig?.config?.grafana?.smtp?.host}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP("host", e.target.value);
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="user:"
                  name={["config", "grafana", "smtp", "user"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={monitorConfig?.config?.grafana?.smtp?.user}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP("user", e.target.value);
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="password:"
                  name={["config", "grafana", "smtp", "password"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={monitorConfig?.config?.grafana?.smtp?.password}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP("password", e.target.value);
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="skip_verify:"
                  name={["config", "grafana", "smtp", "skip_verify"]}
                  required
                  rules={booleanEmptyValidatorRules}
                >
                  <Radio.Group
                    value={monitorConfig?.config?.grafana?.smtp?.skip_verify}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP("skip_verify", e.target.value);
                    }}
                  >
                    <Radio value={true}>是</Radio>
                    <Radio value={false}>否</Radio>
                  </Radio.Group>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="from:"
                  name={["config", "grafana", "smtp", "from"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={monitorConfig?.config?.grafana?.smtp?.from}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP("from", e.target.value);
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="from_name:"
                  name={["config", "grafana", "smtp", "from_name"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={monitorConfig?.config?.grafana?.smtp?.from_name}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP("from_name", e.target.value);
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="startTLS_policy:"
                  name={["config", "grafana", "smtp", "startTLS_policy"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      monitorConfig?.config?.grafana?.smtp?.startTLS_policy
                    }
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP(
                        "startTLS_policy",
                        e.target.value,
                      );
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="enable_tracing:"
                  name={["config", "grafana", "smtp", "enable_tracing"]}
                  required
                  rules={booleanEmptyValidatorRules}
                >
                  <Radio.Group
                    value={monitorConfig?.config?.grafana?.smtp?.enable_tracing}
                    onChange={(e) => {
                      this.onChangeGrafanaSMTP(
                        "enable_tracing",
                        e.target.value,
                      );
                    }}
                  >
                    <Radio value={true}>是</Radio>
                    <Radio value={false}>否</Radio>
                  </Radio.Group>
                </Form.Item>
              </Col>
            </Row>
          ) : null}
          {Object.keys(DefaultConfigData[SERVICES.ProtonMonitor].resources).map(
            (key) => {
              return (
                <>
                  <Form.Item label={`${key}自定义配置资源限制:`}>
                    <Radio.Group
                      value={!!monitorConfig?.resources?.[key]}
                      onChange={(e) => {
                        this.onChangeMonitorComponentResources(
                          key,
                          RESOURCES.ALL,
                          e.target.value,
                        );
                      }}
                    >
                      <Radio value={true}>是</Radio>
                      <Radio value={false}>否</Radio>
                    </Radio.Group>
                  </Form.Item>
                  {monitorConfig?.resources?.[key] ? (
                    <Row>
                      <Col span={12}>
                        <Form.Item
                          labelAlign="left"
                          labelCol={{ span: 8 }}
                          label="Limits.CPU:"
                          name={["resources", key, "limits", "cpu"]}
                          required
                          rules={emptyValidatorRules}
                        >
                          <div>
                            <Input
                              style={{ width: "200px" }}
                              value={
                                monitorConfig?.resources?.[key]?.limits?.cpu
                              }
                              onChange={(e) => {
                                this.onChangeMonitorComponentResources(
                                  key,
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
                          labelAlign="left"
                          labelCol={{ span: 8 }}
                          label="Limits.Memory:"
                          name={["resources", key, "limits", "memory"]}
                          required
                          rules={emptyValidatorRules}
                        >
                          <div>
                            <Input
                              style={{ width: "200px" }}
                              value={
                                monitorConfig?.resources?.[key]?.limits?.memory
                              }
                              onChange={(e) => {
                                this.onChangeMonitorComponentResources(
                                  key,
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
                      <Col span={12}>
                        <Form.Item
                          labelAlign="left"
                          labelCol={{ span: 8 }}
                          label="Requests.CPU:"
                          name={["resources", key, "requests", "cpu"]}
                          required
                          rules={emptyValidatorRules}
                        >
                          <div>
                            <Input
                              style={{ width: "200px" }}
                              value={
                                monitorConfig?.resources?.[key]?.requests?.cpu
                              }
                              onChange={(e) => {
                                this.onChangeMonitorComponentResources(
                                  key,
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
                          labelAlign="left"
                          labelCol={{ span: 8 }}
                          label="Requests.Memory:"
                          name={["resources", key, "requests", "memory"]}
                          required
                          rules={emptyValidatorRules}
                        >
                          <div>
                            <Input
                              style={{ width: "200px" }}
                              value={
                                monitorConfig?.resources?.[key]?.requests
                                  ?.memory
                              }
                              onChange={(e) => {
                                this.onChangeMonitorComponentResources(
                                  key,
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
                  ) : null}
                </>
              );
            },
          )}
        </Form>
      </div>
    );
  }
}

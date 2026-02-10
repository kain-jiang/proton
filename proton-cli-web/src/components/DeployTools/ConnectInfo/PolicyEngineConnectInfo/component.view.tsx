import * as React from "react";
import {
  Form,
  Input,
  InputNumber,
  Radio,
  Divider,
  Row,
  Col,
} from "@aishutech/ui";
import { Title } from "../../Title/component.view";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import { PolicyEngineConnectInfoBase } from "./component.base";
import {
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export class PolicyEngineConnectInfo extends PolicyEngineConnectInfoBase {
  render(): React.ReactNode {
    const { policy_engine } = this.state;

    return (
      <div className="service-box">
        <Title
          title={
            CONNECT_SERVICES_TEXT[CONNECT_SERVICES.POLICY_ENGINE] + "连接信息"
          }
          deleteCallback={
            policy_engine?.source_type === SOURCE_TYPE.EXTERNAL &&
            this.props.onDeleteResource
          }
        />
        <Divider orientation="left" orientationMargin="0">
          资源类型
        </Divider>
        <Radio.Group
          style={{
            margin: "10px 0",
          }}
          disabled
          value={policy_engine?.source_type}
          onChange={(e) => {
            this.changePolicyEngineConnectInfo(
              "source_type",
              e.target.value,
              policy_engine,
            );
          }}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.POLICY_ENGINE]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>
            第三方 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.POLICY_ENGINE]}
          </Radio>
        </Radio.Group>
        {policy_engine?.source_type === SOURCE_TYPE.EXTERNAL ? (
          <Form
            layout="horizontal"
            name="policy_engine"
            validateTrigger="onBlur"
            initialValues={policy_engine}
            ref={this.form}
          >
            <Divider orientation="left" orientationMargin="0">
              连接信息
            </Divider>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 4 }}
                  labelAlign="left"
                  label="地址:"
                  name="hosts"
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={policy_engine?.hosts}
                      onChange={(e) => {
                        this.changePolicyEngineConnectInfo(
                          "hosts",
                          e.target.value,
                          policy_engine,
                        );
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title="多个ip或者域名请以英文逗号分割。"
                    />
                  </div>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 4 }}
                  labelAlign="left"
                  label="端口:"
                  name="port"
                  required
                  rules={portValidatorRules}
                >
                  <InputNumber
                    style={{ width: "200px" }}
                    value={policy_engine?.port}
                    onChange={(val) => {
                      this.changePolicyEngineConnectInfo(
                        "port",
                        val,
                        policy_engine,
                      );
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        ) : null}
      </div>
    );
  }
}

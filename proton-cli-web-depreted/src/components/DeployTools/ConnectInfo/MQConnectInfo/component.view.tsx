import * as React from "react";
import {
  Form,
  Input,
  Divider,
  Select,
  Radio,
  Row,
  Col,
  InputNumber,
} from "@aishutech/ui";
import { Title } from "../../Title/component.view";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import { MQConnectInfoBase } from "./component.base";
import {
  MQ_TYPE,
  MQ_AUTH_MACHANISM,
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
} from "../../helper";
import "./styles.view.scss";

export class MQConnectInfo extends MQConnectInfoBase {
  render(): React.ReactNode {
    const { mq, mqTypeList } = this.state;

    return (
      <div className="service-box">
        <Title
          title={CONNECT_SERVICES_TEXT[CONNECT_SERVICES.MQ] + "连接信息"}
          deleteCallback={this.props.onDeleteResource}
        />
        <Divider orientation="left" orientationMargin="0">
          资源类型
        </Divider>
        <Radio.Group
          style={{
            margin: "10px 0",
          }}
          disabled
          value={mq?.source_type}
          onChange={(e) => {
            this.changeMQConnectInfo("source_type", e.target.value, mq);
            this.props.updateConnectInfoValidateState({
              MQ_RADIO: ValidateState.Normal,
              MQ_TYPE: ValidateState.Normal,
              MQ_AUTH_MACHANISM: ValidateState.Normal,
            });
          }}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.MQ]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>
            第三方 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.MQ]}
          </Radio>
        </Radio.Group>
        {this.props.connectInfoValidateState.MQ_RADIO ? (
          <div style={{ color: "#FF4D4F" }}>此项不允许为空。</div>
        ) : null}
        <Form
          layout="horizontal"
          name="mq"
          validateTrigger="onBlur"
          initialValues={mq}
          ref={this.form}
        >
          <Divider orientation="left" orientationMargin="0">
            MQ 类型
          </Divider>
          <Form.Item
            labelCol={{ span: 2 }}
            labelAlign="left"
            label="MQ类型:"
            required
          >
            <Select
              showArrow
              style={{
                width: "200px",
              }}
              optionLabelProp="label"
              placeholder="请选择 MQ 类型"
              value={mq?.mq_type}
              onSelect={(val) => {
                this.changeMQConnectInfo("mq_type", val, mq);
                this.props.updateConnectInfoValidateState({
                  MQ_TYPE: ValidateState.Normal,
                  MQ_AUTH_MACHANISM: ValidateState.Normal,
                });
              }}
            >
              {mqTypeList.map((val) => (
                <Select.Option key={val} label={val}>
                  {val}
                </Select.Option>
              ))}
            </Select>
            {this.props.connectInfoValidateState.MQ_TYPE ? (
              <div style={{ color: "#FF4D4F" }}>此项不允许为空。</div>
            ) : null}
          </Form.Item>
          {mq?.source_type === SOURCE_TYPE.EXTERNAL ? (
            <>
              <Divider orientation="left" orientationMargin="0">
                连接信息
              </Divider>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label="地址:"
                    name="mq_hosts"
                    required
                    rules={emptyValidatorRules}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={mq?.mq_hosts}
                        onChange={(e) => {
                          this.changeMQConnectInfo(
                            "mq_hosts",
                            e.target.value,
                            mq,
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
                    name="mq_port"
                    required
                    rules={portValidatorRules}
                  >
                    <InputNumber
                      style={{ width: "200px" }}
                      value={mq?.mq_port}
                      onChange={(value) => {
                        this.changeMQConnectInfo("mq_port", value, mq);
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
              {mq?.mq_type === MQ_TYPE.NSQ ? (
                <>
                  <Divider orientation="left" orientationMargin="0">
                    lookupd 连接信息
                  </Divider>
                  <Row>
                    <Col span={12}>
                      <Form.Item
                        labelCol={{ span: 4 }}
                        labelAlign="left"
                        label="地址:"
                        name="mq_lookupd_hosts"
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{ width: "200px" }}
                            value={mq?.mq_lookupd_hosts}
                            onChange={(e) => {
                              this.changeMQConnectInfo(
                                "mq_lookupd_hosts",
                                e.target.value,
                                mq,
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
                        name="mq_lookupd_port"
                        required
                        rules={portValidatorRules}
                      >
                        <InputNumber
                          style={{ width: "200px" }}
                          value={mq?.mq_lookupd_port}
                          onChange={(value) => {
                            this.changeMQConnectInfo(
                              "mq_lookupd_port",
                              value,
                              mq,
                            );
                          }}
                        />
                      </Form.Item>
                    </Col>
                  </Row>
                </>
              ) : null}
              {mq?.mq_type === MQ_TYPE.KAFKA ? (
                <>
                  <Divider orientation="left" orientationMargin="0">
                    认证账户信息
                  </Divider>
                  <Row>
                    <Col span={8}>
                      <Form.Item
                        labelCol={{ span: 5 }}
                        labelAlign="left"
                        label="用户名:"
                        name={["auth", "username"]}
                        required
                        rules={emptyValidatorRules}
                      >
                        <Input
                          style={{ width: "150px" }}
                          value={mq?.auth?.username}
                          onChange={(e) => {
                            this.changeMQAuthConnectInfo(
                              "username",
                              e.target.value,
                              mq,
                            );
                          }}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        labelCol={{ span: 5 }}
                        labelAlign="left"
                        label="密码:"
                        name={["auth", "password"]}
                        required
                        rules={emptyValidatorRules}
                      >
                        <Input.Password
                          style={{ width: "150px" }}
                          value={mq?.auth?.password}
                          onChange={(e) => {
                            this.changeMQAuthConnectInfo(
                              "password",
                              e.target.value,
                              mq,
                            );
                          }}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        labelCol={{ span: 6 }}
                        labelAlign="left"
                        label="认证机制:"
                        required
                      >
                        <Select
                          showArrow
                          style={{ width: "150px" }}
                          optionLabelProp="label"
                          placeholder="请选择认证机制"
                          value={mq?.auth?.mechanism}
                          onSelect={(val) => {
                            this.changeMQAuthConnectInfo("mechanism", val, mq);
                            this.props.updateConnectInfoValidateState({
                              MQ_AUTH_MACHANISM: ValidateState.Normal,
                            });
                          }}
                        >
                          {Object.keys(MQ_AUTH_MACHANISM).map((key) => (
                            <Select.Option
                              key={MQ_AUTH_MACHANISM[key]}
                              label={MQ_AUTH_MACHANISM[key]}
                            >
                              {MQ_AUTH_MACHANISM[key]}
                            </Select.Option>
                          ))}
                        </Select>
                        {this.props.connectInfoValidateState
                          .MQ_AUTH_MACHANISM ? (
                          <div style={{ color: "#FF4D4F" }}>
                            此项不允许为空。
                          </div>
                        ) : null}
                      </Form.Item>
                    </Col>
                  </Row>
                </>
              ) : null}
            </>
          ) : null}
        </Form>
      </div>
    );
  }
}

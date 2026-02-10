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
} from "@kweaver-ai/ui";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/icons";
import { MQConnectInfoBase } from "./component.base";
import {
  MQ_TYPE,
  MQ_AUTH_MACHANISM,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
} from "../../../component-management/helper";
import styles from "./styles.module.less";
import __ from "../locale";
import { noop } from "lodash";

export class MQConnectInfo extends MQConnectInfoBase {
  render(): React.ReactNode {
    const { mq, mqTypeList } = this.state;

    return (
      <div>
        <Form
          layout="horizontal"
          name="mq"
          validateTrigger="onBlur"
          initialValues={mq}
          ref={this.form}
        >
          {mq?.source_type === SOURCE_TYPE.EXTERNAL ? (
            <>
              <Form.Item
                labelCol={{ span: 2 }}
                labelAlign="left"
                label={__("MQ类型")}
                required
              >
                <Select
                  showArrow
                  style={{
                    width: "200px",
                  }}
                  optionLabelProp="label"
                  placeholder={__("请选择 MQ 类型")}
                  getPopupContainer={(node) =>
                    node.parentElement || document.body
                  }
                  value={mq?.mq_type}
                  onSelect={(val) => {
                    this.props.updateConnectInfo(mq, val);
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
                  <div style={{ color: "#FF4D4F" }}>
                    {__("此项不允许为空。")}
                  </div>
                ) : null}
              </Form.Item>
              <Divider orientation="left" orientationMargin="0">
                {__("连接配置")}
              </Divider>
              <div className={styles["component-title"]}>{__("连接信息")}</div>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("地址")}
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
                            mq
                          );
                        }}
                      />
                      <QuestionCircleOutlined
                        onPointerEnterCapture={noop}
                        onPointerLeaveCapture={noop}
                        style={{
                          marginLeft: "6px",
                        }}
                        title={__("多个ip或者域名请以英文逗号分割。")}
                      />
                    </div>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("端口")}
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
                  <div className={styles["component-title"]}>
                    {__("lookupd 连接信息")}
                  </div>
                  <Row>
                    <Col span={12}>
                      <Form.Item
                        labelCol={{ span: 4 }}
                        labelAlign="left"
                        label={__("地址")}
                        name="mq_lookupd_hosts"
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{
                              width: "200px",
                            }}
                            value={mq?.mq_lookupd_hosts}
                            onChange={(e) => {
                              this.changeMQConnectInfo(
                                "mq_lookupd_hosts",
                                e.target.value,
                                mq
                              );
                            }}
                          />
                          <QuestionCircleOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                            style={{
                              marginLeft: "6px",
                            }}
                            title={__("多个ip或者域名请以英文逗号分割。")}
                          />
                        </div>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        labelCol={{ span: 4 }}
                        labelAlign="left"
                        label={__("端口")}
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
                              mq
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
                  <div className={styles["component-title"]}>
                    {__("认证账户信息")}
                  </div>
                  <Row>
                    <Col span={8}>
                      <Form.Item
                        labelCol={{ span: 5 }}
                        labelAlign="left"
                        label={__("用户名")}
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
                              mq
                            );
                          }}
                          disabled={this.getIsDisabled(mq)}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        labelCol={{ span: 5 }}
                        labelAlign="left"
                        label={__("密码")}
                        name={["auth", "password"]}
                        required
                        rules={
                          this.getIsDisabled(mq)
                            ? undefined
                            : emptyValidatorRules
                        }
                      >
                        <div>
                          <Input.Password
                            style={{
                              width: "150px",
                            }}
                            value={
                              this.getIsDisabled(mq)
                                ? "******"
                                : mq?.auth?.password
                            }
                            onChange={(e) => {
                              this.changeMQAuthConnectInfo(
                                "password",
                                e.target.value,
                                mq
                              );
                            }}
                            disabled={this.getIsDisabled(mq)}
                          />
                        </div>
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        labelCol={{ span: 6 }}
                        labelAlign="left"
                        label={__("认证机制")}
                        required
                      >
                        <Select
                          showArrow
                          style={{ width: "150px" }}
                          optionLabelProp="label"
                          placeholder={__("请选择认证机制")}
                          value={mq?.auth?.mechanism}
                          getPopupContainer={(node) =>
                            node.parentElement || document.body
                          }
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
                          <div
                            style={{
                              color: "#FF4D4F",
                            }}
                          >
                            {__("此项不允许为空。")}
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

import * as React from "react";
import {
  Form,
  Input,
  InputNumber,
  Radio,
  Divider,
  Row,
  Col,
} from "@kweaver-ai/ui";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/icons";
import { MongoDBConnectInfoBase } from "./component.base";
import {
  SOURCE_TYPE,
  ValidateState,
  emptyValidatorRules,
  getUsernameValidatorRules,
  portValidatorRules,
} from "../../../component-management/helper";
import styles from "./styles.module.less";
import __ from "../locale";
import { noop } from "lodash";

export class MongoDBConnectInfo extends MongoDBConnectInfoBase {
  render(): React.ReactNode {
    const { mongodb } = this.state;

    return (
      <>
        {mongodb?.source_type === SOURCE_TYPE.EXTERNAL ? (
          <div className={styles["service-box"]}>
            <Divider orientation="left" orientationMargin="0">
              {__("连接配置")}
            </Divider>
            <Form
              layout="horizontal"
              name="mongodb"
              validateTrigger="onBlur"
              initialValues={mongodb}
              ref={this.form}
            >
              <div className={styles["component-title"]}>{__("账户信息")}</div>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("用户名")}
                    name="username"
                    required
                    rules={getUsernameValidatorRules(mongodb?.source_type)}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={mongodb?.username}
                      onChange={(e) => {
                        this.changeMongoDBConnectInfo(
                          "username",
                          e.target.value,
                          mongodb
                        );
                      }}
                      disabled={
                        mongodb?.source_type === this.props.originSourceType
                      }
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("密码")}
                    name="password"
                    required
                    rules={
                      mongodb?.source_type === this.props.originSourceType
                        ? undefined
                        : emptyValidatorRules
                    }
                  >
                    <div>
                      <Input.Password
                        style={{ width: "200px" }}
                        value={
                          mongodb?.source_type === this.props.originSourceType
                            ? "******"
                            : mongodb?.password
                        }
                        onChange={(e) => {
                          this.changeMongoDBConnectInfo(
                            "password",
                            e.target.value,
                            mongodb
                          );
                        }}
                        disabled={
                          mongodb?.source_type === this.props.originSourceType
                        }
                      />
                      {mongodb?.source_type === SOURCE_TYPE.INTERNAL ? (
                        <QuestionCircleOutlined
                          onPointerEnterCapture={noop}
                          onPointerLeaveCapture={noop}
                          style={{
                            marginLeft: "6px",
                          }}
                          title={__(
                            "密码要求3种字符，支持大写、小写、数字、特殊字符（!@#$%^&*()_+-.=）。"
                          )}
                        />
                      ) : null}
                    </div>
                  </Form.Item>
                </Col>
              </Row>
              <div className={styles["component-title"]}>{__("连接信息")}</div>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("地址")}
                    name="hosts"
                    required
                    rules={emptyValidatorRules}
                  >
                    <div>
                      <Input
                        style={{
                          width: "200px",
                        }}
                        value={mongodb?.hosts}
                        onChange={(e) => {
                          this.changeMongoDBConnectInfo(
                            "hosts",
                            e.target.value,
                            mongodb
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
                    labelCol={{ span: 5 }}
                    labelAlign="left"
                    label={__("端口")}
                    name="port"
                    required
                    rules={portValidatorRules}
                  >
                    <InputNumber
                      style={{ width: "200px" }}
                      value={mongodb?.port}
                      onChange={(val) => {
                        this.changeMongoDBConnectInfo("port", val, mongodb);
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <div className={styles["component-title"]}>{__("鉴权信息")}</div>
              <Row>
                <Col span={6}>
                  <Form.Item
                    labelCol={{ span: 5 }}
                    labelAlign="left"
                    label="ssl:"
                    required
                  >
                    <div className={styles["ssl-radio"]}>
                      <Radio.Group
                        value={mongodb?.ssl}
                        onChange={(e) => {
                          this.changeMongoDBConnectInfo(
                            "ssl",
                            e.target.value,
                            mongodb
                          );
                          this.props.updateConnectInfoValidateState({
                            MONGODB_SSL: ValidateState.Normal,
                          });
                        }}
                      >
                        <Radio value={true}>{__("开启")}</Radio>
                        <Radio value={false}>{__("关闭")}</Radio>
                      </Radio.Group>
                    </div>
                    {this.props.connectInfoValidateState.MONGODB_SSL ? (
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
                <Col span={8}>
                  <Form.Item
                    labelCol={{ span: 5 }}
                    labelAlign="left"
                    label={__("副本集")}
                    name="replica_set"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={mongodb?.replica_set}
                      onChange={(e) => {
                        this.changeMongoDBConnectInfo(
                          "replica_set",
                          e.target.value,
                          mongodb
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item
                    labelCol={{ span: 6 }}
                    labelAlign="left"
                    label="authSource:"
                    name="auth_source"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={mongodb?.auth_source}
                      onChange={(e) => {
                        this.changeMongoDBConnectInfo(
                          "auth_source",
                          e.target.value,
                          mongodb
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <div className={styles["component-title"]}>{__("可选参数")}</div>
              <Form.Item labelAlign="left" label={__("可选参数")}>
                <Input
                  style={{ width: "200px" }}
                  value={mongodb?.options}
                  onChange={(e) => {
                    this.changeMongoDBConnectInfo(
                      "options",
                      e.target.value,
                      mongodb
                    );
                  }}
                />
                <QuestionCircleOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                  style={{
                    marginLeft: "6px",
                  }}
                  title={__("形如k1=v1&k2=v2,多个参数使用&连接。")}
                />
              </Form.Item>
            </Form>
          </div>
        ) : null}
      </>
    );
  }
}

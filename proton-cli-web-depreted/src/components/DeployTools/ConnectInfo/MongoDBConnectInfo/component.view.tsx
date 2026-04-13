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
import { MongoDBConnectInfoBase } from "./component.base";
import {
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
  ValidateState,
  emptyValidatorRules,
  getUsernameValidatorRules,
  portValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export class MongoDBConnectInfo extends MongoDBConnectInfoBase {
  render(): React.ReactNode {
    const { mongodb } = this.state;

    return (
      <div className="service-box">
        <Title
          title={CONNECT_SERVICES_TEXT[CONNECT_SERVICES.MONGODB] + "连接信息"}
          deleteCallback={
            mongodb?.source_type === SOURCE_TYPE.EXTERNAL &&
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
          value={mongodb?.source_type}
          onChange={(e) => {
            this.changeMongoDBConnectInfo(
              "source_type",
              e.target.value,
              mongodb,
            );
          }}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.MONGODB]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>
            第三方 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.MONGODB]}
          </Radio>
        </Radio.Group>
        <Form
          layout="horizontal"
          name="mongodb"
          validateTrigger="onBlur"
          initialValues={mongodb}
          ref={this.form}
        >
          <Divider orientation="left" orientationMargin="0">
            账户信息
          </Divider>
          <Row>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 4 }}
                labelAlign="left"
                label="用户名:"
                name="username"
                required
                rules={getUsernameValidatorRules(mongodb?.source_type)}
              >
                <div>
                  <Input
                    style={{ width: "200px" }}
                    value={mongodb?.username}
                    onChange={(e) => {
                      this.changeMongoDBConnectInfo(
                        "username",
                        e.target.value,
                        mongodb,
                      );
                    }}
                  />
                  <QuestionCircleOutlined
                    style={{
                      marginLeft: "6px",
                    }}
                    title="该账号用于各产品服务使用数据库，如对数据增删改查。"
                  />
                </div>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                labelCol={{ span: 4 }}
                labelAlign="left"
                label="密码:"
                name="password"
                required
                rules={emptyValidatorRules}
              >
                <div>
                  <Input.Password
                    style={{ width: "200px" }}
                    value={mongodb?.password}
                    onChange={(e) => {
                      this.changeMongoDBConnectInfo(
                        "password",
                        e.target.value,
                        mongodb,
                      );
                    }}
                  />
                  {mongodb?.source_type === SOURCE_TYPE.INTERNAL ? (
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title="密码要求3种字符，支持大写、小写、数字、特殊字符（!@#$%^&*()_+-.=）。"
                    />
                  ) : null}
                </div>
              </Form.Item>
            </Col>
          </Row>
          {mongodb?.source_type === SOURCE_TYPE.EXTERNAL ? (
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
                    name="hosts"
                    required
                    rules={emptyValidatorRules}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={mongodb?.hosts}
                        onChange={(e) => {
                          this.changeMongoDBConnectInfo(
                            "hosts",
                            e.target.value,
                            mongodb,
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
                    labelCol={{ span: 5 }}
                    labelAlign="left"
                    label="端口:"
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
              <Divider orientation="left" orientationMargin="0">
                鉴权信息
              </Divider>
              <Row>
                <Col span={6}>
                  <Form.Item
                    labelCol={{ span: 5 }}
                    labelAlign="left"
                    label="ssl:"
                    required
                  >
                    <div className="ssl-radio">
                      <Radio.Group
                        value={mongodb?.ssl}
                        onChange={(e) => {
                          this.changeMongoDBConnectInfo(
                            "ssl",
                            e.target.value,
                            mongodb,
                          );
                          this.props.updateConnectInfoValidateState({
                            MONGODB_SSL: ValidateState.Normal,
                          });
                        }}
                      >
                        <Radio value={true}>开启</Radio>
                        <Radio value={false}>关闭</Radio>
                      </Radio.Group>
                    </div>
                    {this.props.connectInfoValidateState.MONGODB_SSL ? (
                      <div style={{ color: "#FF4D4F" }}>此项不允许为空。</div>
                    ) : null}
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    labelCol={{ span: 5 }}
                    labelAlign="left"
                    label="副本集:"
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
                          mongodb,
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
                          mongodb,
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Divider orientation="left" orientationMargin="0">
                可选参数
              </Divider>
              <Form.Item labelAlign="left" label="可选参数:">
                <Input
                  style={{ width: "200px" }}
                  value={mongodb?.options}
                  onChange={(e) => {
                    this.changeMongoDBConnectInfo(
                      "options",
                      e.target.value,
                      mongodb,
                    );
                  }}
                />
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title="形如k1=v1&k2=v2,多个参数使用&连接。"
                />
              </Form.Item>
            </>
          ) : null}
        </Form>
      </div>
    );
  }
}

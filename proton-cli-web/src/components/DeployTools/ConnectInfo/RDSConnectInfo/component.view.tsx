import * as React from "react";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Radio,
  Divider,
  Row,
  Col,
  Switch,
} from "@aishutech/ui";
import { Title } from "../../Title/component.view";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import { RDSConnectInfoBase } from "./component.base";
import {
  RDS_TYPE,
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
  getUsernameValidatorRules,
  getRDSUserInfoValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export class RDSConnectInfo extends RDSConnectInfoBase {
  render(): React.ReactNode {
    const { rds } = this.state;

    return (
      <div className="service-box">
        <Title
          title={CONNECT_SERVICES_TEXT[CONNECT_SERVICES.RDS] + "连接信息"}
          deleteCallback={
            rds?.source_type === SOURCE_TYPE.EXTERNAL &&
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
          value={rds?.source_type}
          disabled
          onChange={(e) => {
            this.changeRDSConnectInfo("source_type", e.target.value, rds);
          }}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.RDS]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>
            第三方 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.RDS]}
          </Radio>
        </Radio.Group>
        <Form
          layout="horizontal"
          name="rds"
          validateTrigger="onBlur"
          initialValues={rds}
          ref={this.form}
        >
          {rds?.source_type === SOURCE_TYPE.EXTERNAL ? (
            <>
              <Divider orientation="left" orientationMargin="0">
                RDS 类型
              </Divider>
              <Form.Item
                labelCol={{ span: 4 }}
                labelAlign="left"
                label="RDS 类型:"
                required
              >
                <Select
                  showArrow
                  style={{
                    width: "200px",
                  }}
                  optionLabelProp="label"
                  placeholder="请选择 RDS 类型"
                  value={rds?.rds_type}
                  onSelect={(val) => {
                    this.changeRDSConnectInfo("rds_type", val, rds);
                    this.handleClearValidation();
                    this.props.updateConnectInfoValidateState({
                      RDS_TYPE: ValidateState.Normal,
                    });
                  }}
                >
                  {Object.keys(RDS_TYPE).map((key) => (
                    <Select.Option key={RDS_TYPE[key]} label={RDS_TYPE[key]}>
                      {RDS_TYPE[key]}
                    </Select.Option>
                  ))}
                </Select>
                {this.props.connectInfoValidateState.RDS_TYPE ? (
                  <div style={{ color: "#FF4D4F" }}>此项不允许为空。</div>
                ) : null}
              </Form.Item>
              <Divider orientation="left" orientationMargin="0">
                账户信息
              </Divider>
              <Form.Item
                labelCol={{ span: 4 }}
                labelAlign="left"
                label="自动化创建数据库:"
                name="auto_create_database"
                required
              >
                <div>
                  <Switch
                    checked={rds?.auto_create_database}
                    onChange={(value) => {
                      this.changeRDSConnectInfo(
                        "auto_create_database",
                        value,
                        rds,
                      );
                      if (!value) {
                        this.form.current.validateFields(
                          ["username", "password"].filter((key) => rds[key]),
                        );
                      }
                    }}
                  />
                  <QuestionCircleOutlined
                    style={{
                      marginLeft: "6px",
                      verticalAlign: "middle",
                    }}
                    title={`针对人大金仓，代表自动创建schema，仍需要手工建人大金仓的数据库。`}
                  />
                </div>
              </Form.Item>
              {rds?.auto_create_database ? (
                <Row>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      labelAlign="left"
                      label="用户名（管理权限）:"
                      name="admin_user"
                      required
                      rules={getRDSUserInfoValidatorRules(
                        rds?.rds_type,
                        rds?.username,
                      )}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={rds?.admin_user}
                          onChange={(e) => {
                            this.changeRDSConnectInfo(
                              "admin_user",
                              e.target.value,
                              rds,
                            );
                          }}
                          onBlur={() => this.checkLinkItem("username", rds)}
                        />
                        <QuestionCircleOutlined
                          style={{
                            marginLeft: "6px",
                          }}
                          title={`该账号用于${
                            rds?.rds_type || RDS_TYPE.MYSQL
                          }实例的管理，如创建数据库。`}
                        />
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      labelCol={{ span: 8 }}
                      labelAlign="left"
                      label="密码:"
                      name="admin_passwd"
                      required
                      rules={getRDSUserInfoValidatorRules(
                        rds?.rds_type,
                        rds?.password,
                      )}
                    >
                      <Input.Password
                        style={{ width: "200px" }}
                        value={rds?.admin_passwd}
                        onChange={(e) => {
                          this.changeRDSConnectInfo(
                            "admin_passwd",
                            e.target.value,
                            rds,
                          );
                        }}
                        onBlur={() => this.checkLinkItem("password", rds)}
                      />
                    </Form.Item>
                  </Col>
                </Row>
              ) : null}
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="用户名:"
                    name="username"
                    required
                    rules={getRDSUserInfoValidatorRules(
                      rds?.rds_type,
                      rds?.admin_user,
                    )}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={rds?.username}
                        onChange={(e) => {
                          this.changeRDSConnectInfo(
                            "username",
                            e.target.value,
                            rds,
                          );
                        }}
                        onBlur={() => this.checkLinkItem("admin_user", rds)}
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
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="密码:"
                    name="password"
                    required
                    rules={getRDSUserInfoValidatorRules(
                      rds?.rds_type,
                      rds?.admin_passwd,
                    )}
                  >
                    <Input.Password
                      style={{ width: "200px" }}
                      value={rds?.password}
                      onChange={(e) => {
                        this.changeRDSConnectInfo(
                          "password",
                          e.target.value,
                          rds,
                        );
                      }}
                      onBlur={() => this.checkLinkItem("admin_passwd", rds)}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Divider orientation="left" orientationMargin="0">
                连接信息
              </Divider>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="地址:"
                    name="hosts"
                    required
                    rules={emptyValidatorRules}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={rds?.hosts}
                        onChange={(e) => {
                          this.changeRDSConnectInfo(
                            "hosts",
                            e.target.value,
                            rds,
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
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="端口:"
                    name="port"
                    required
                    rules={portValidatorRules}
                  >
                    <InputNumber
                      style={{ width: "200px" }}
                      value={rds?.port}
                      onChange={(val) => {
                        this.changeRDSConnectInfo("port", val, rds);
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </>
          ) : (
            <>
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
                    rules={getUsernameValidatorRules(rds?.source_type)}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={rds?.username}
                        onChange={(e) => {
                          this.changeRDSConnectInfo(
                            "username",
                            e.target.value,
                            rds,
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
                        value={rds?.password}
                        onChange={(e) => {
                          this.changeRDSConnectInfo(
                            "password",
                            e.target.value,
                            rds,
                          );
                        }}
                      />
                      <QuestionCircleOutlined
                        style={{
                          marginLeft: "6px",
                        }}
                        title="密码要求3种字符，支持大写、小写、数字、特殊字符（!@#$%^&*()_+-.=）。"
                      />
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}
        </Form>
      </div>
    );
  }
}

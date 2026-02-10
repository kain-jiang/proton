import * as React from "react";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Divider,
  Row,
  Col,
  Switch,
} from "@kweaver-ai/ui";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/icons";
import { RDSConnectInfoBase } from "./component.base";
import {
  RDS_TYPE,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
  getUsernameValidatorRules,
  getRDSUserInfoValidatorRules,
} from "../../../component-management/helper";
import styles from "./styles.module.less";
import __ from "../locale";
import { noop } from "lodash";

export class RDSConnectInfo extends RDSConnectInfoBase {
  render(): React.ReactNode {
    const { rds } = this.state;

    return (
      <div>
        {rds?.source_type === SOURCE_TYPE.EXTERNAL ? (
          <Form
            layout="horizontal"
            name="rds"
            validateTrigger="onBlur"
            initialValues={rds}
            ref={this.form}
          >
            <Form.Item
              labelCol={{ span: 2 }}
              labelAlign="left"
              label={__("RDS 类型")}
              required
            >
              <Select
                showArrow
                style={{
                  width: "200px",
                }}
                optionLabelProp="label"
                placeholder={__("请选择 RDS 类型")}
                getPopupContainer={(node) =>
                  node.parentElement || document.body
                }
                value={rds?.rds_type}
                onSelect={(val) => {
                  this.props.updateConnectInfo(rds, val);
                  this.props.updateConnectInfoValidateState({
                    RDS_TYPE: ValidateState.Normal,
                  });
                  this.handleClearValidation();
                }}
              >
                {Object.keys(RDS_TYPE).map((key) => (
                  <Select.Option key={RDS_TYPE[key]} label={RDS_TYPE[key]}>
                    {RDS_TYPE[key]}
                  </Select.Option>
                ))}
              </Select>
              {this.props.connectInfoValidateState.RDS_TYPE ? (
                <div style={{ color: "#FF4D4F" }}>{__("此项不允许为空。")}</div>
              ) : null}
            </Form.Item>
            <Divider orientation="left" orientationMargin="0">
              {__("连接配置")}
            </Divider>
            <div className={styles["component-title"]}>{__("账户信息")}</div>
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label={
                <span
                  className={styles["label-overflow"]}
                  title={__("自动化创建数据库")}
                >
                  {__("自动化创建数据库")}
                </span>
              }
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
                      rds
                    );
                    if (!value) {
                      this.form?.current?.validateFields(
                        ["username", "password"].filter((key) => rds[key])
                      );
                    }
                  }}
                />
                <QuestionCircleOutlined
                  onPointerEnterCapture={noop}
                  onPointerLeaveCapture={noop}
                  style={{
                    marginLeft: "6px",
                    verticalAlign: "middle",
                  }}
                  title={__(
                    "针对人大金仓，代表自动创建schema，仍需要手工建人大金仓的数据库。"
                  )}
                />
              </div>
            </Form.Item>
            {rds?.auto_create_database ? (
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label={
                      <span
                        className={styles["label-overflow"]}
                        title={__("用户名（管理权限）")}
                      >
                        {__("用户名（管理权限）")}
                      </span>
                    }
                    name="admin_user"
                    required
                    rules={getRDSUserInfoValidatorRules(
                      rds?.rds_type,
                      rds?.username
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
                            rds
                          );
                        }}
                        onBlur={() => this.checkLinkItem("username", rds)}
                      />
                      <QuestionCircleOutlined
                        onPointerEnterCapture={noop}
                        onPointerLeaveCapture={noop}
                        style={{
                          marginLeft: "6px",
                        }}
                        title={__(
                          "该账号用于${rds_type}实例的管理，如创建数据库。",
                          {
                            rds_type: rds?.rds_type || RDS_TYPE.MYSQL,
                          }
                        )}
                      />
                    </div>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label={__("密码")}
                    name="admin_passwd"
                    required
                    rules={
                      this.getIsDisabled(rds)
                        ? emptyValidatorRules
                        : getRDSUserInfoValidatorRules(
                            rds?.rds_type,
                            rds?.password
                          )
                    }
                  >
                    <div>
                      <Input.Password
                        style={{ width: "200px" }}
                        value={rds?.admin_passwd}
                        onChange={(e) => {
                          this.changeRDSConnectInfo(
                            "admin_passwd",
                            e.target.value,
                            rds
                          );
                        }}
                        onBlur={() => this.checkLinkItem("password", rds)}
                      />
                    </div>
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label={__("用户名")}
                  name="username"
                  required
                  rules={
                    this.getIsDisabled(rds)
                      ? emptyValidatorRules
                      : getRDSUserInfoValidatorRules(
                          rds?.rds_type,
                          rds?.admin_user
                        )
                  }
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={rds?.username}
                      onChange={(e) => {
                        this.changeRDSConnectInfo(
                          "username",
                          e.target.value,
                          rds
                        );
                      }}
                      onBlur={() => this.checkLinkItem("admin_user", rds)}
                      disabled={this.getIsDisabled(rds)}
                    />
                    <QuestionCircleOutlined
                      onPointerEnterCapture={noop}
                      onPointerLeaveCapture={noop}
                      style={{
                        marginLeft: "6px",
                      }}
                      title={__(
                        "该账号用于各产品服务使用数据库，如对数据增删改查。"
                      )}
                    />
                  </div>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label={__("密码")}
                  name="password"
                  required
                  rules={
                    this.getIsDisabled(rds)
                      ? undefined
                      : getRDSUserInfoValidatorRules(
                          rds?.rds_type,
                          rds?.admin_passwd
                        )
                  }
                >
                  <div>
                    <Input.Password
                      style={{ width: "200px" }}
                      value={this.getIsDisabled(rds) ? "******" : rds?.password}
                      onChange={(e) => {
                        this.changeRDSConnectInfo(
                          "password",
                          e.target.value,
                          rds
                        );
                      }}
                      onBlur={() => this.checkLinkItem("admin_passwd", rds)}
                      disabled={this.getIsDisabled(rds)}
                    />
                  </div>
                </Form.Item>
              </Col>
            </Row>
            <div className={styles["component-title"]}>{__("连接信息")}</div>
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label={__("地址")}
                  name="hosts"
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={rds?.hosts}
                      onChange={(e) => {
                        this.changeRDSConnectInfo("hosts", e.target.value, rds);
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
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label={__("端口")}
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
          </Form>
        ) : null}
      </div>
    );
  }
}

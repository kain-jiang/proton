import * as React from "react";
import {
  Form,
  Input,
  Row,
  Col,
  Select,
  Radio,
  Divider,
  InputNumber,
} from "@kweaver-ai/ui";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/icons";
import {
  REDIS_CONNECT_TYPE,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
} from "../../../component-management/helper";
import { RedisConnectInfoBase } from "./component.base";
import styles from "./styles.module.less";
import __ from "../locale";
import { noop } from "lodash";

export class RedisConnectInfo extends RedisConnectInfoBase {
  render(): React.ReactNode {
    const { redis } = this.state;

    return (
      <div>
        <Form
          layout="horizontal"
          name="redis"
          validateTrigger="onBlur"
          initialValues={redis}
          ref={this.form}
        >
          {redis?.source_type === SOURCE_TYPE.EXTERNAL ? (
            <>
              <Divider orientation="left" orientationMargin="0">
                {__("连接配置")}
              </Divider>
              <div className={styles["component-title"]}>
                {__("Redis 连接模式")}
              </div>

              <Form.Item
                labelCol={{ span: 2 }}
                labelAlign="left"
                label={__("模式")}
                required
              >
                <Select
                  showArrow
                  style={{
                    width: "200px",
                  }}
                  optionLabelProp="label"
                  placeholder={__("请选择 Redis 连接模式")}
                  getPopupContainer={(node) =>
                    node.parentElement || document.body
                  }
                  value={redis?.connect_type}
                  onSelect={(val) => {
                    this.props.updateConnectInfo(redis, val);
                    this.props.updateConnectInfoValidateState({
                      REDIS_CONNECT_TYPE: ValidateState.Normal,
                    });
                  }}
                >
                  {Object.keys(REDIS_CONNECT_TYPE).map((key) => (
                    <Select.Option
                      key={REDIS_CONNECT_TYPE[key]}
                      label={REDIS_CONNECT_TYPE[key]}
                    >
                      {REDIS_CONNECT_TYPE[key]}
                    </Select.Option>
                  ))}
                </Select>
                {this.props.connectInfoValidateState.REDIS_CONNECT_TYPE ? (
                  <div style={{ color: "#FF4D4F" }}>
                    {__("此项不允许为空。")}
                  </div>
                ) : null}
              </Form.Item>
              {this.getTemplateByConnectType(redis)}
            </>
          ) : null}
        </Form>
      </div>
    );
  }

  getTemplateByConnectType(redis: any) {
    const commonAccount = [
      <div className={styles["component-title"]}>{__("账户配置")}</div>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("用户名")}
          >
            <Input
              style={{ width: "200px" }}
              value={redis?.username}
              onChange={(e) => {
                this.changeRedisConnectInfo("username", e.target.value, redis);
              }}
              disabled={this.getIsDisabled(redis)}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("密码")}
          >
            <Input.Password
              style={{ width: "200px" }}
              value={
                this.getIsDisabled(redis)
                  ? redis?.username
                    ? "******"
                    : ""
                  : redis?.password
              }
              onChange={(e) => {
                this.changeRedisConnectInfo("password", e.target.value, redis);
              }}
              disabled={this.getIsDisabled(redis)}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const commonLink = [
      <div className={styles["component-title"]}>{__("连接信息")}</div>,
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
                style={{ width: "200px" }}
                value={redis?.hosts}
                onChange={(e) => {
                  this.changeRedisConnectInfo("hosts", e.target.value, redis);
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
            name="port"
            required
            rules={portValidatorRules}
          >
            <InputNumber
              style={{ width: "200px" }}
              value={redis?.port}
              onChange={(val) => {
                this.changeRedisConnectInfo("port", val, redis);
              }}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const masterLink = [
      <div className={styles["component-title"]}>{__("master连接信息")}</div>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("地址")}
            name="master_hosts"
            required
            rules={emptyValidatorRules}
          >
            <div>
              <Input
                style={{ width: "200px" }}
                value={redis?.master_hosts}
                onChange={(e) => {
                  this.changeRedisConnectInfo(
                    "master_hosts",
                    e.target.value,
                    redis
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
            name="master_port"
            required
            rules={portValidatorRules}
          >
            <InputNumber
              style={{ width: "200px" }}
              value={redis?.master_port}
              onChange={(val) => {
                this.changeRedisConnectInfo("master_port", val, redis);
              }}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const slaveLink = [
      <div className={styles["component-title"]}>{__("slave连接信息")}</div>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("地址")}
            name="slave_hosts"
            required
            rules={emptyValidatorRules}
          >
            <div>
              <Input
                style={{ width: "200px" }}
                value={redis?.slave_hosts}
                onChange={(e) => {
                  this.changeRedisConnectInfo(
                    "slave_hosts",
                    e.target.value,
                    redis
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
            name="slave_port"
            required
            rules={portValidatorRules}
          >
            <InputNumber
              style={{ width: "200px" }}
              value={redis?.slave_port}
              onChange={(val) => {
                this.changeRedisConnectInfo("slave_port", val, redis);
              }}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const sentinelLink = [
      <div className={styles["component-title"]}>{__("哨兵连接信息")}</div>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("地址")}
            name="sentinel_hosts"
            required
            rules={emptyValidatorRules}
          >
            <div>
              <Input
                style={{ width: "200px" }}
                value={redis?.sentinel_hosts}
                onChange={(e) => {
                  this.changeRedisConnectInfo(
                    "sentinel_hosts",
                    e.target.value,
                    redis
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
            name="sentinel_port"
            required
            rules={portValidatorRules}
          >
            <InputNumber
              style={{ width: "200px" }}
              value={redis?.sentinel_port}
              onChange={(val) => {
                this.changeRedisConnectInfo("sentinel_port", val, redis);
              }}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const sentinelAccount = [
      <div className={styles["component-title"]}>{__("哨兵账户信息")}</div>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("用户名")}
          >
            <Input
              style={{ width: "200px" }}
              value={redis?.sentinel_username}
              onChange={(e) => {
                this.changeRedisConnectInfo(
                  "sentinel_username",
                  e.target.value,
                  redis
                );
              }}
              disabled={this.getIsDisabled(redis)}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label={__("密码")}
          >
            <Input.Password
              style={{ width: "200px" }}
              value={
                this.getIsDisabled(redis)
                  ? redis?.sentinel_username
                    ? "******"
                    : ""
                  : redis?.sentinel_password
              }
              onChange={(e) => {
                this.changeRedisConnectInfo(
                  "sentinel_password",
                  e.target.value,
                  redis
                );
              }}
              disabled={this.getIsDisabled(redis)}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const sentinelMasterGroupName = [
      <div className={styles["component-title"]}>{__("复制组名")}</div>,
      <Form.Item
        labelCol={{ span: 2 }}
        labelAlign="left"
        label={__("复制组名")}
        name="master_group_name"
        required
        rules={emptyValidatorRules}
      >
        <Input
          style={{ width: "200px" }}
          value={redis?.master_group_name}
          onChange={(e) => {
            this.changeRedisConnectInfo(
              "master_group_name",
              e.target.value,
              redis
            );
          }}
        />
      </Form.Item>,
    ];
    switch (redis?.connect_type) {
      case REDIS_CONNECT_TYPE.STANDALONE:
      case REDIS_CONNECT_TYPE.CLUSTER:
        return [commonLink, commonAccount];
      case REDIS_CONNECT_TYPE.MASTER_SLAVE:
        return [masterLink, slaveLink, commonAccount];
      default:
        return [
          sentinelLink,
          sentinelAccount,
          commonAccount,
          sentinelMasterGroupName,
        ];
    }
  }
}

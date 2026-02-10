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
} from "@aishutech/ui";
import { Title } from "../../Title/component.view";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import {
  REDIS_CONNECT_TYPE,
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
} from "../../helper";
import { RedisConnectInfoBase } from "./component.base";
import "./styles.view.scss";

export class RedisConnectInfo extends RedisConnectInfoBase {
  render(): React.ReactNode {
    const { redis } = this.state;

    return (
      <div className="service-box">
        <Title
          title={CONNECT_SERVICES_TEXT[CONNECT_SERVICES.REDIS] + "连接信息"}
          deleteCallback={
            redis?.source_type === SOURCE_TYPE.EXTERNAL &&
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
          value={redis?.source_type}
          onChange={(e) => {
            this.changeRedisConnectInfo("source_type", e.target.value, redis);
          }}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.REDIS]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>
            第三方 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.REDIS]}
          </Radio>
        </Radio.Group>
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
                Redis 连接模式
              </Divider>
              <Form.Item
                labelCol={{ span: 2 }}
                labelAlign="left"
                label="模式:"
                required
              >
                <Select
                  showArrow
                  style={{
                    width: "200px",
                  }}
                  optionLabelProp="label"
                  placeholder="请选择 Redis 连接模式"
                  value={redis?.connect_type}
                  onSelect={(val) => {
                    this.changeRedisConnectInfo("connect_type", val, redis);
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
                  <div style={{ color: "#FF4D4F" }}>此项不允许为空。</div>
                ) : null}
              </Form.Item>
              {this.getTemplateByConnectType(redis)}
            </>
          ) : null}
        </Form>
      </div>
    );
  }

  getTemplateByConnectType(redis) {
    const commonAccount = [
      <Divider orientation="left" orientationMargin="0">
        账户配置
      </Divider>,
      <Row>
        <Col span={12}>
          <Form.Item labelCol={{ span: 4 }} labelAlign="left" label="用户名:">
            <div>
              <Input
                style={{ width: "200px" }}
                value={redis?.username}
                onChange={(e) => {
                  this.changeRedisConnectInfo(
                    "username",
                    e.target.value,
                    redis,
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
          <Form.Item labelCol={{ span: 4 }} labelAlign="left" label="密码:">
            <Input.Password
              style={{ width: "200px" }}
              value={redis?.password}
              onChange={(e) => {
                this.changeRedisConnectInfo("password", e.target.value, redis);
              }}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const commonLink = [
      <Divider orientation="left" orientationMargin="0">
        连接信息
      </Divider>,
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
                value={redis?.hosts}
                onChange={(e) => {
                  this.changeRedisConnectInfo("hosts", e.target.value, redis);
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
      <Divider orientation="left" orientationMargin="0">
        master连接信息
      </Divider>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label="地址:"
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
                    redis,
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
      <Divider orientation="left" orientationMargin="0">
        slave连接信息
      </Divider>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label="地址:"
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
                    redis,
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
      <Divider orientation="left" orientationMargin="0">
        哨兵连接信息
      </Divider>,
      <Row>
        <Col span={12}>
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label="地址:"
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
                    redis,
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
      <Divider orientation="left" orientationMargin="0">
        哨兵账户信息
      </Divider>,
      <Row>
        <Col span={12}>
          <Form.Item labelCol={{ span: 4 }} labelAlign="left" label="用户名:">
            <Input
              style={{ width: "200px" }}
              value={redis?.sentinel_username}
              onChange={(e) => {
                this.changeRedisConnectInfo(
                  "sentinel_username",
                  e.target.value,
                  redis,
                );
              }}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item labelCol={{ span: 4 }} labelAlign="left" label="密码:">
            <Input.Password
              style={{ width: "200px" }}
              value={redis?.sentinel_password}
              onChange={(e) => {
                this.changeRedisConnectInfo(
                  "sentinel_password",
                  e.target.value,
                  redis,
                );
              }}
            />
          </Form.Item>
        </Col>
      </Row>,
    ];
    const sentinelMasterGroupName = [
      <Divider orientation="left" orientationMargin="0">
        复制组名
      </Divider>,
      <Form.Item
        labelCol={{ span: 2 }}
        labelAlign="left"
        label="复制组名:"
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
              redis,
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

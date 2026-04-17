import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { Form, Input, Row, Col, Radio, InputNumber } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import {
  DataBaseStorageType,
  emptyValidatorRules,
  replicaValidatorRules,
} from "../../helper";
import RedisConfigBase from "./component.base";
import "./styles.view.scss";

export default class RedisConfig extends RedisConfigBase {
  render(): React.ReactNode {
    const { configData, dataBaseStorageType, service } = this.props;
    const { redisConfig, redisNodes } = this.state;

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
              onClick={this.props.onDeleteRedisConfig.bind(this)}
            />
          </Col>
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
        {dataBaseStorageType === DataBaseStorageType.Standard ? (
          <Form.Item
            labelCol={{ span: 4 }}
            labelAlign="left"
            label="部署节点"
            required
          >
            <SelectNode
              mode={false}
              nodes={configData.nodesInfo}
              selectedNodes={redisNodes}
              onSelectedChange={(nodes) =>
                this.onChangeRedisNode(nodes, redisConfig)
              }
            />
          </Form.Item>
        ) : null}
        {dataBaseStorageType === DataBaseStorageType.DepositKubernetes ? (
          <Form
            labelCol={{ span: 4 }}
            labelAlign="left"
            name="redisForm_replicaForm"
            initialValues={redisConfig}
            validateTrigger="onBlur"
            ref={this.redisForm.replicaForm}
          >
            <Form.Item
              label="副本数:"
              name="replica_count"
              required
              rules={replicaValidatorRules}
            >
              <InputNumber
                style={{
                  width: "100%",
                }}
                value={redisConfig?.replica_count}
                onChange={(val) => {
                  this.onChangeRedis({
                    replica_count: val,
                  });
                }}
              />
            </Form.Item>
          </Form>
        ) : null}
        <div
          style={{
            marginTop: "10px",
          }}
        >
          <Form
            name="redisForm_accountForm"
            layout="horizontal"
            validateTrigger="onBlur"
            initialValues={redisConfig}
            ref={this.redisForm.accountForm}
          >
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="用户名（管理权限）:"
                  name="admin_user"
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={redisConfig?.admin_user}
                      onChange={(e) => {
                        this.onChangeRedis({ admin_user: e.target.value });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title="该账号用于Redis实例的管理，如创建数据库。"
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
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input.Password
                      style={{ width: "200px" }}
                      type="password"
                      value={redisConfig?.admin_passwd}
                      onChange={(e) => {
                        this.onChangeRedis({ admin_passwd: e.target.value });
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
          </Form>
          <Form
            name="redisForm_pathForm"
            layout="horizontal"
            validateTrigger="onBlur"
            initialValues={redisConfig}
            ref={this.redisForm.pathForm}
          >
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="存储卷容量:"
                >
                  <Input
                    style={{ width: "200px" }}
                    value={redisConfig?.storage_capacity}
                    onChange={(e) => {
                      this.onChangeRedis({ storage_capacity: e.target.value });
                    }}
                  />
                  <QuestionCircleOutlined
                    style={{
                      marginLeft: "6px",
                    }}
                    title="填写规则为整数或浮点数+单位，如(Mi,Gi,Ti)。"
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                {dataBaseStorageType === DataBaseStorageType.Standard ? (
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="数据路径:"
                    name="data_path"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={redisConfig?.data_path}
                      onChange={(e) => {
                        this.onChangeRedis({ data_path: e.target.value });
                      }}
                    />
                  </Form.Item>
                ) : (
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="storageClassName:"
                    name="storageClassName"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={redisConfig?.storageClassName}
                      onChange={(e) => {
                        this.onChangeRedis({
                          storageClassName: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                )}
              </Col>
            </Row>
          </Form>
          <Form
            layout="horizontal"
            name="redisForm_resourcesForm"
            validateTrigger="onBlur"
            initialValues={redisConfig}
            ref={this.redisForm.resourcesForm}
          >
            <Row>
              <Col span={12}>
                <Form.Item
                  label="自定义配置资源限制:"
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                >
                  <Radio.Group
                    value={!!redisConfig?.resources}
                    onChange={(e) => {
                      this.onChangeRedisConfigResources(e.target.value);
                    }}
                  >
                    <Radio value={true}>是</Radio>
                    <Radio value={false}>否</Radio>
                  </Radio.Group>
                </Form.Item>
              </Col>
            </Row>
            {redisConfig?.resources ? (
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="Requests.CPU:"
                    name={["resources", "requests", "cpu"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={redisConfig?.resources?.requests?.cpu}
                      onChange={(e) => {
                        this.onChangeResource("cpu", e.target.value);
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="Requests.Memory:"
                    name={["resources", "requests", "memory"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={redisConfig?.resources?.requests?.memory}
                      onChange={(e) => {
                        this.onChangeResource("memory", e.target.value);
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
          </Form>
        </div>
      </div>
    );
  }
}

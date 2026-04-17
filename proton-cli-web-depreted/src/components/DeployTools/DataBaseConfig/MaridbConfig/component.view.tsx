import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { Form, Input, Row, Col, InputNumber } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import MariaDBConfigBase from "./component.base";
import {
  DataBaseStorageType,
  emptyValidatorRules,
  replicaValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export default class MariaDBConfig extends MariaDBConfigBase {
  render(): React.ReactNode {
    const { configData, dataBaseStorageType, service } = this.props;
    const { mariaDBConfig, mariaDBNodes } = this.state;

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
              onClick={this.props.onDeleteMariaDBConfig.bind(this)}
            />
          </Col>
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
        <div>
          {dataBaseStorageType === DataBaseStorageType.Standard ? (
            <Form labelCol={{ span: 4 }} labelAlign="left">
              <Form.Item label="部署节点" required>
                <SelectNode
                  mode={false}
                  nodes={configData.nodesInfo}
                  selectedNodes={mariaDBNodes}
                  onSelectedChange={(nodes) =>
                    this.onChangeMariaDBNode(nodes, mariaDBConfig)
                  }
                />
              </Form.Item>
            </Form>
          ) : null}
          {dataBaseStorageType === DataBaseStorageType.DepositKubernetes ? (
            <Form
              labelCol={{ span: 4 }}
              labelAlign="left"
              name="mariaDBForm_replicaForm"
              initialValues={mariaDBConfig}
              validateTrigger="onBlur"
              ref={this.mariaDBForm.replicaForm}
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
                  value={mariaDBConfig?.replica_count}
                  onChange={(val) => {
                    this.onChangeMariDB({
                      replica_count: val,
                    });
                  }}
                />
              </Form.Item>
            </Form>
          ) : null}
          <Form
            style={{
              margin: "10px 0",
              width: "100%",
            }}
            name="mariaDBForm_configForm"
            validateTrigger="onBlur"
            initialValues={mariaDBConfig?.config}
            ref={this.mariaDBForm.configForm}
          >
            <Row gutter={24}>
              <Col span={8}>
                <Form.Item
                  label="Innodb_buffer_size:"
                  name="innodb_buffer_pool_size"
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "100px" }}
                    value={mariaDBConfig?.config?.innodb_buffer_pool_size}
                    onChange={(e) => {
                      this.onChangeMariaDBConfig({
                        innodb_buffer_pool_size: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  label="Requests.Memory:"
                  name="resource_requests_memory"
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "100px" }}
                      value={mariaDBConfig?.config?.resource_requests_memory}
                      onChange={(e) => {
                        this.onChangeMariaDBConfig({
                          resource_requests_memory: e.target.value,
                        });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                    />
                  </div>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  label="Limits.Memory:"
                  name="resource_limits_memory"
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "100px" }}
                      value={mariaDBConfig?.config?.resource_limits_memory}
                      onChange={(e) => {
                        this.onChangeMariaDBConfig({
                          resource_limits_memory: e.target.value,
                        });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                    />
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </Form>
          <Form
            style={{
              marginTop: "10px",
            }}
            name="mariaDBForm_accountForm"
            validateTrigger="onBlur"
            initialValues={mariaDBConfig}
            ref={this.mariaDBForm.accountForm}
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
                      value={mariaDBConfig?.admin_user}
                      onChange={(e) => {
                        this.onChangeMariDB({ admin_user: e.target.value });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title="该账号用于MariaDB实例的管理，如创建数据库。"
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
                      value={mariaDBConfig?.admin_passwd}
                      onChange={(e) => {
                        this.onChangeMariDB({ admin_passwd: e.target.value });
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
          {dataBaseStorageType === DataBaseStorageType.Standard ? (
            <Form
              layout="horizontal"
              name="mariaDBForm_pathForm"
              validateTrigger="onBlur"
              initialValues={mariaDBConfig}
              ref={this.mariaDBForm.pathForm}
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
                      placeholder="默认值：10Gi"
                      value={mariaDBConfig?.storage_capacity}
                      onChange={(e) => {
                        this.onChangeMariDB({
                          storage_capacity: e.target.value,
                        });
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
                  <Form.Item
                    labelCol={{ span: 8 }}
                    labelAlign="left"
                    label="数据路径:"
                    name="data_path"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={mariaDBConfig?.data_path}
                      onChange={(e) => {
                        this.onChangeMariDB({ data_path: e.target.value });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Form>
          ) : (
            <Form
              layout="horizontal"
              name="mariaDBForm_storageForm"
              validateTrigger="onBlur"
              initialValues={mariaDBConfig}
              ref={this.mariaDBForm.storageForm}
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
                      placeholder="默认值：10Gi"
                      value={mariaDBConfig?.storage_capacity}
                      onChange={(e) => {
                        this.onChangeMariDB({
                          storage_capacity: e.target.value,
                        });
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
                      value={mariaDBConfig?.storageClassName}
                      onChange={(e) => {
                        this.onChangeMariDB({
                          storageClassName: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Form>
          )}
        </div>
      </div>
    );
  }
}

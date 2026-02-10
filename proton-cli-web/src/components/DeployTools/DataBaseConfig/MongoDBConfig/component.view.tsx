import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { Form, Input, Radio, Row, Col, InputNumber } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import MongoDBConfigBase from "./component.base";
import {
  DataBaseStorageType,
  emptyValidatorRules,
  replicaValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export default class MongoDBConfig extends MongoDBConfigBase {
  render(): React.ReactNode {
    const { configData, dataBaseStorageType, service } = this.props;
    const { mongoDBConfig, mongoDBNodes } = this.state;

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
              onClick={this.props.onDeleteMongoDBConfig.bind(this)}
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
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="部署节点"
              required
            >
              <SelectNode
                mode={false}
                nodes={configData.nodesInfo}
                selectedNodes={mongoDBNodes}
                onSelectedChange={(nodes) =>
                  this.onChangeMongoDBNode(nodes, mongoDBConfig)
                }
              />
            </Form.Item>
          ) : null}
          {dataBaseStorageType === DataBaseStorageType.DepositKubernetes ? (
            <Form
              labelCol={{ span: 4 }}
              labelAlign="left"
              name="mongoDBForm_replicaForm"
              initialValues={mongoDBConfig}
              validateTrigger="onBlur"
              ref={this.mongoDBForm.replicaForm}
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
                  value={mongoDBConfig?.replica_count}
                  onChange={(val) => {
                    this.onChangeMongoDB({
                      replica_count: val,
                    });
                  }}
                />
              </Form.Item>
            </Form>
          ) : null}
          <Form
            layout="horizontal"
            style={{
              marginTop: "10px",
            }}
            name="mongoDBForm_accountForm"
            validateTrigger="onBlur"
            initialValues={mongoDBConfig}
            ref={this.mongoDBForm.accountForm}
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
                      value={mongoDBConfig?.admin_user}
                      onChange={(e) => {
                        this.onChangeMongoDB({ admin_user: e.target.value });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title="该账号用于MongoDB实例的管理，如创建数据库。"
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
                      value={mongoDBConfig?.admin_passwd}
                      onChange={(e) => {
                        this.onChangeMongoDB({ admin_passwd: e.target.value });
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
              name="mongoDBForm_pathForm"
              validateTrigger="onBlur"
              initialValues={mongoDBConfig}
              ref={this.mongoDBForm.pathForm}
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
                      value={mongoDBConfig?.storage_capacity}
                      onChange={(e) => {
                        this.onChangeMongoDB({
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
                      value={mongoDBConfig?.data_path}
                      onChange={(e) => {
                        this.onChangeMongoDB({ data_path: e.target.value });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Form>
          ) : (
            <Form
              layout="horizontal"
              name="mongoDBForm_storageForm"
              validateTrigger="onBlur"
              initialValues={mongoDBConfig}
              ref={this.mongoDBForm.storageForm}
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
                      value={mongoDBConfig?.storage_capacity}
                      onChange={(e) => {
                        this.onChangeMongoDB({
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
                      value={mongoDBConfig?.storageClassName}
                      onChange={(e) => {
                        this.onChangeMongoDB({
                          storageClassName: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Form>
          )}
          <Form
            layout="horizontal"
            name="mongoDBForm_resourcesForm"
            validateTrigger="onBlur"
            initialValues={mongoDBConfig}
            ref={this.mongoDBForm.resourcesForm}
          >
            <Row>
              <Col span={12}>
                <Form.Item
                  label="自定义配置资源限制:"
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                >
                  <Radio.Group
                    value={!!mongoDBConfig?.resources}
                    onChange={(e) => {
                      this.onChangeMongoDBConfigResources(e.target.value);
                    }}
                  >
                    <Radio value={true}>是</Radio>
                    <Radio value={false}>否</Radio>
                  </Radio.Group>
                </Form.Item>
              </Col>
            </Row>
            {mongoDBConfig?.resources ? (
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
                      value={mongoDBConfig?.resources?.requests?.cpu}
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
                      value={mongoDBConfig?.resources?.requests?.memory}
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

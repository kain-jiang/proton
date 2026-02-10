import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import { Form, Input, Row, Col, Radio, InputNumber } from "@aishutech/ui";
import {
  DataBaseStorageType,
  emptyValidatorRules,
  replicaValidatorRules,
} from "../../helper";
import PackageStoreBase from "./component.base";
import "./styles.view.scss";

export default class PackageStore extends PackageStoreBase {
  render(): React.ReactNode {
    const { configData, dataBaseStorageType, service } = this.props;
    const { packageStoreConfig, packageStoreNodes } = this.state;

    return (
      <div className="package-store">
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
              onClick={this.props.onDeletePackageStoreConfig.bind(this)}
            />
          </Col>
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
        <Form
          name={service.key}
          layout="horizontal"
          validateTrigger="onBlur"
          initialValues={packageStoreConfig}
          ref={this.form}
        >
          {dataBaseStorageType === DataBaseStorageType.Standard ? (
            <>
              <Form.Item label="部署节点" required>
                <SelectNode
                  mode={false}
                  nodes={configData.nodesInfo}
                  selectedNodes={packageStoreNodes}
                  onSelectedChange={(nodes) =>
                    this.onChangePackageStoreNode(nodes, packageStoreConfig)
                  }
                />
              </Form.Item>
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="存储卷容量:"
                    name={["storage", "capacity"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={packageStoreConfig?.storage?.capacity}
                        onChange={(e) => {
                          this.onChangePackageStoreStorage({
                            capacity: e.target.value,
                          });
                        }}
                      />
                      <QuestionCircleOutlined
                        style={{
                          marginLeft: "6px",
                        }}
                        title="填写规则为整数或浮点数+单位，如(Mi,Gi,Ti)。"
                      />
                    </div>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="数据目录:"
                    name={["storage", "path"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={packageStoreConfig?.storage?.path}
                      onChange={(e) => {
                        this.onChangePackageStoreStorage({
                          path: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </>
          ) : (
            <>
              <Form.Item
                label="副本数:"
                name="replicas"
                required
                rules={replicaValidatorRules}
              >
                <InputNumber
                  style={{
                    width: "100%",
                  }}
                  value={packageStoreConfig?.replicas}
                  onChange={(val) => {
                    this.onChangePackageStoreReplicas({
                      replicas: val,
                    });
                  }}
                />
              </Form.Item>
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="存储卷容量:"
                    name={["storage", "capacity"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <div>
                      <Input
                        style={{ width: "200px" }}
                        value={packageStoreConfig?.storage?.capacity}
                        onChange={(e) => {
                          this.onChangePackageStoreStorage({
                            capacity: e.target.value,
                          });
                        }}
                      />
                      <QuestionCircleOutlined
                        style={{
                          marginLeft: "6px",
                        }}
                        title="填写规则为整数或浮点数+单位，如(Mi,Gi,Ti)。"
                      />
                    </div>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="storageClassName:"
                    name={["storage", "storageClassName"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={packageStoreConfig?.storage?.storageClassName}
                      onChange={(e) => {
                        this.onChangePackageStoreStorage({
                          storageClassName: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}
          <Form.Item label="自定义配置资源限制:">
            <Radio.Group
              value={!!packageStoreConfig?.resources}
              onChange={(e) => {
                this.onChangePackageStoreResources(e.target.value);
              }}
            >
              <Radio value={true}>是</Radio>
              <Radio value={false}>否</Radio>
            </Radio.Group>
            <QuestionCircleOutlined
              style={{
                marginLeft: "6px",
              }}
              title="当选择否时使用的是推荐值：CPU为1C，Memory为200m。"
            />
          </Form.Item>
          {packageStoreConfig?.resources ? (
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item
                  label="Limits.CPU:"
                  name={["resources", "limits", "cpu"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={packageStoreConfig?.resources?.limits?.cpu}
                      onChange={(e) => {
                        this.onChangePackageStoreResourcesLimits({
                          cpu: e.target.value,
                        });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={`填写规则为整数或浮点数+单位，如(C,m)。`}
                    />
                  </div>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label="Limits.Memory:"
                  name={["resources", "limits", "memory"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={packageStoreConfig?.resources?.limits?.memory}
                      onChange={(e) => {
                        this.onChangePackageStoreResourcesLimits({
                          memory: e.target.value,
                        });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。`}
                    />
                  </div>
                </Form.Item>
              </Col>
            </Row>
          ) : null}
        </Form>
      </div>
    );
  }
}

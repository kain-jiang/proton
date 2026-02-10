import * as React from "react";
import {
  DeleteOutlined,
  PlusOutlined,
  QuestionCircleOutlined,
} from "@aishutech/ui/icons";
import SelectNode from "../../SelectNode/component.view";
import {
  Form,
  Input,
  Row,
  Col,
  Radio,
  InputNumber,
  Button,
  Switch,
} from "@aishutech/ui";
import ServiceConfigBase from "./component.base";
import {
  DEFAULT_REPLICA_SERVICES,
  DataBaseStorageType,
  JVMConfigKeys,
  NODES_LIMIT,
  SERVICES,
  ValidateState,
  booleanEmptyValidatorRules,
  emptyValidatorRules,
  getLogNumberValidatorRules,
  notMultipleDateBase,
  portValidatorRules,
  replicaValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export default class ServiceConfig extends ServiceConfigBase {
  render(): React.ReactNode {
    const {
      service,
      dataBaseStorageType,
      configData,
      grafanaNodesValidateState,
      prometheusNodesValidateState,
    } = this.props;
    const { serviceNodes, serviceConfig } = this.state;

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
              onClick={this.props.onDeleteServiceConfig.bind(this)}
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
          labelAlign="left"
          validateTrigger="onBlur"
          initialValues={serviceConfig}
          ref={this.form}
        >
          {dataBaseStorageType === DataBaseStorageType.DepositKubernetes &&
          DEFAULT_REPLICA_SERVICES.includes(service.key) ? (
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
                value={serviceConfig?.replica_count}
                onChange={(val) => {
                  this.onChangeService({
                    replica_count: val,
                  });
                }}
              />
            </Form.Item>
          ) : null}
          {dataBaseStorageType === DataBaseStorageType.Standard ? (
            <Form.Item className="label" label={"部署节点"} required>
              <SelectNode
                mode={notMultipleDateBase.includes(service.key)}
                nodes={configData.nodesInfo}
                selectedNodes={serviceNodes}
                onSelectedChange={(nodes) =>
                  this.onChangeServiceNode(nodes, serviceConfig)
                }
              />
              {prometheusNodesValidateState === ValidateState.NodesNumError &&
              service.key === SERVICES.Prometheus ? (
                <div
                  style={{ color: "#FF4D4F" }}
                >{`Prometheus 部署节点数量应当小于等于 ${NODES_LIMIT.prometheus}`}</div>
              ) : null}
              {grafanaNodesValidateState === ValidateState.NodesNumError &&
              service.key === SERVICES.Grafana ? (
                <div
                  style={{ color: "#FF4D4F" }}
                >{`Grafana 部署节点只允许 ${NODES_LIMIT.grafana} 节点`}</div>
              ) : null}
            </Form.Item>
          ) : null}
          {[SERVICES.Zookeeper].includes(this.props.service.key) ? (
            <Form.Item
              label="JVM配置"
              name={["env", JVMConfigKeys[this.props.service.key]]}
              required
              rules={emptyValidatorRules}
            >
              <Input
                value={
                  serviceConfig?.env?.[JVMConfigKeys[this.props.service.key]]
                }
                onChange={(e) => {
                  this.onChangeEnv({
                    [JVMConfigKeys[this.props.service.key]]: e.target.value,
                  });
                }}
              />
            </Form.Item>
          ) : null}
          {[SERVICES.Kafka].includes(this.props.service.key) ? (
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item
                  label="JVM配置"
                  name={["env", JVMConfigKeys[this.props.service.key]]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={
                      serviceConfig?.env?.[
                        JVMConfigKeys[this.props.service.key]
                      ]
                    }
                    onChange={(e) => {
                      this.onChangeEnv({
                        [JVMConfigKeys[this.props.service.key]]: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label="日志保留字节数"
                  name={["env", "KAFKA_LOG_RETENTION_BYTES"]}
                  rules={getLogNumberValidatorRules(true)}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder="默认值：-1"
                    value={serviceConfig?.env?.KAFKA_LOG_RETENTION_BYTES}
                    onChange={(e) => {
                      this.onChangeEnv({
                        KAFKA_LOG_RETENTION_BYTES: e.target.value || undefined,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label="日志保留小时数"
                  name={["env", "KAFKA_LOG_RETENTION_HOURS"]}
                  rules={getLogNumberValidatorRules(false)}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder="默认值：168"
                    value={serviceConfig?.env?.KAFKA_LOG_RETENTION_HOURS}
                    onChange={(e) => {
                      this.onChangeEnv({
                        KAFKA_LOG_RETENTION_HOURS: e.target.value || undefined,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label="日志段最大小时数"
                  name={["env", "KAFKA_LOG_ROLL_HOURS"]}
                  rules={getLogNumberValidatorRules(false)}
                >
                  <Input
                    style={{ width: "200px" }}
                    placeholder="默认值：24"
                    value={serviceConfig?.env?.KAFKA_LOG_ROLL_HOURS}
                    onChange={(e) => {
                      this.onChangeEnv({
                        KAFKA_LOG_ROLL_HOURS: e.target.value || undefined,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          ) : null}
          {dataBaseStorageType === DataBaseStorageType.Standard ? (
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label="存储卷容量">
                  <Input
                    style={{ width: "200px" }}
                    value={serviceConfig?.storage_capacity}
                    onChange={(e) => {
                      this.onChangeService({
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
                  label="数据路径"
                  name="data_path"
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    value={serviceConfig?.data_path}
                    onChange={(e) => {
                      this.onChangeService({ data_path: e.target.value });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          ) : (
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label="存储卷容量">
                  <Input
                    style={{ width: "200px" }}
                    value={serviceConfig?.storage_capacity}
                    onChange={(e) => {
                      this.onChangeService({
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
                  label="storageClassName"
                  name="storageClassName"
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={serviceConfig?.storageClassName}
                    onChange={(e) => {
                      this.onChangeService({
                        storageClassName: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          )}
          {[SERVICES.Kafka].includes(this.props.service.key) ? (
            <Form.Item
              label="禁用开放外部端口:"
              name="disable_external_service"
              required
              rules={booleanEmptyValidatorRules}
            >
              <div>
                <Radio.Group
                  value={serviceConfig?.disable_external_service}
                  onChange={(e) => {
                    this.onChangeService({
                      disable_external_service: e.target.value,
                    });
                  }}
                >
                  <Radio value={true}>是</Radio>
                  <Radio value={false}>否</Radio>
                </Radio.Group>
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title={`未禁用时将根据下列端口信息进行开放。`}
                />
              </div>
            </Form.Item>
          ) : null}
          {[SERVICES.Kafka].includes(this.props.service.key) ? (
            <Form.Item
              label="外部端口信息:"
              name="external_service_list"
              required
              style={{ marginBottom: "0" }}
            >
              <Button
                type="default"
                icon={<PlusOutlined />}
                style={{ marginBottom: "24px" }}
                onClick={this.addExternalServiceList.bind(this)}
              >
                新增
              </Button>
              {serviceConfig?.external_service_list.map(
                (serviceInfo, index) => {
                  return (
                    <>
                      <Row>
                        <Col span={12}>
                          <Form.Item
                            label="名称"
                            name={["external_service_list", index, "name"]}
                            required
                            rules={emptyValidatorRules}
                          >
                            <Input
                              style={{ width: "200px" }}
                              value={serviceInfo?.name}
                              onChange={(e) => {
                                this.onChangeExternalServiceList(index, {
                                  name: e.target.value,
                                });
                              }}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label="地址"
                            name={["external_service_list", index, "ip"]}
                          >
                            <Input
                              style={{ width: "200px" }}
                              value={serviceInfo?.ip}
                              onChange={(e) => {
                                this.onChangeExternalServiceList(index, {
                                  ip: e.target.value,
                                });
                              }}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label="端口"
                            name={["external_service_list", index, "port"]}
                            required
                            rules={portValidatorRules}
                          >
                            <div>
                              <InputNumber
                                style={{ width: "200px" }}
                                value={serviceInfo?.port}
                                onChange={(value) => {
                                  this.onChangeExternalServiceList(index, {
                                    port: value,
                                  });
                                }}
                              />
                              <QuestionCircleOutlined
                                style={{
                                  marginLeft: "6px",
                                }}
                                title={`将开放kafka的此端口。`}
                              />
                            </div>
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label="节点Base端口"
                            name={[
                              "external_service_list",
                              index,
                              "nodePortBase",
                            ]}
                            required
                            rules={portValidatorRules}
                          >
                            <div>
                              <InputNumber
                                style={{ width: "200px" }}
                                value={serviceInfo?.nodePortBase}
                                onChange={(value) => {
                                  this.onChangeExternalServiceList(index, {
                                    nodePortBase: value,
                                  });
                                }}
                              />
                              <QuestionCircleOutlined
                                style={{
                                  marginLeft: "6px",
                                }}
                                title={`每个节点都将开放一个端口，从此端口开始依次递增。`}
                              />
                            </div>
                          </Form.Item>
                        </Col>
                        <Col span={24}>
                          <Form.Item
                            label="TLS"
                            name={["external_service_list", index, "enableSSL"]}
                            required
                            rules={emptyValidatorRules}
                          >
                            <Radio.Group
                              value={serviceInfo?.enableSSL}
                              onChange={(e) => {
                                this.onChangeExternalServiceList(index, {
                                  enableSSL: e.target.value,
                                });
                              }}
                            >
                              <Radio value={true}>是</Radio>
                              <Radio value={false}>否</Radio>
                            </Radio.Group>
                            <Button
                              icon={<DeleteOutlined />}
                              type="link"
                              style={{ marginLeft: "16px", color: "black" }}
                              disabled={
                                serviceConfig?.external_service_list?.length <=
                                1
                              }
                              onClick={() =>
                                this.deleteExternalServiceList(index)
                              }
                            />
                          </Form.Item>
                        </Col>
                      </Row>
                      {index !==
                      serviceConfig.external_service_list.length - 1 ? (
                        <div className={"list-split"}></div>
                      ) : null}
                    </>
                  );
                },
              )}
            </Form.Item>
          ) : null}
          {[
            SERVICES.Kafka,
            SERVICES.Zookeeper,
            SERVICES.Prometheus,
            SERVICES.Grafana,
            SERVICES.ProtonNSQ,
            SERVICES.ProtonPolicyEngine,
          ].includes(this.props.service.key) ? (
            <>
              <Form.Item label="自定义配置资源限制:">
                <Radio.Group
                  value={!!serviceConfig?.resources}
                  onChange={(e) => {
                    this.onChangeServiceConfigResources(e.target.value);
                  }}
                >
                  <Radio value={true}>是</Radio>
                  <Radio value={false}>否</Radio>
                </Radio.Group>
              </Form.Item>
              {serviceConfig?.resources ? (
                <>
                  {serviceConfig?.resources?.limits ? (
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
                              value={serviceConfig?.resources?.limits?.cpu}
                              onChange={(e) => {
                                this.onChangeResource(
                                  "limits",
                                  "cpu",
                                  e.target.value,
                                );
                              }}
                            />
                            <QuestionCircleOutlined
                              style={{
                                marginLeft: "6px",
                              }}
                              title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
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
                              value={serviceConfig?.resources?.limits?.memory}
                              onChange={(e) => {
                                this.onChangeResource(
                                  "limits",
                                  "memory",
                                  e.target.value,
                                );
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
                  ) : null}
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label="Requests.CPU:"
                        name={["resources", "requests", "cpu"]}
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{ width: "200px" }}
                            value={serviceConfig?.resources?.requests?.cpu}
                            onChange={(e) => {
                              this.onChangeResource(
                                "requests",
                                "cpu",
                                e.target.value,
                              );
                            }}
                          />
                          {serviceConfig?.resources?.limits ? (
                            <QuestionCircleOutlined
                              style={{
                                marginLeft: "6px",
                              }}
                              title={`填写规则为整数或浮点数+单位，如(C,m)。\n为保证服务正常运行，请满足：Requests.CPU ≤ Limits.CPU`}
                            />
                          ) : null}
                        </div>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label="Requests.Memory:"
                        name={["resources", "requests", "memory"]}
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{ width: "200px" }}
                            value={serviceConfig?.resources?.requests?.memory}
                            onChange={(e) => {
                              this.onChangeResource(
                                "requests",
                                "memory",
                                e.target.value,
                              );
                            }}
                          />
                          {serviceConfig?.resources?.limits ? (
                            <QuestionCircleOutlined
                              style={{
                                marginLeft: "6px",
                              }}
                              title={`填写规则为整数或浮点数+单位，如(Mi,Gi,Ti,M,G,T)。\n为保证服务正常运行，请满足：Requests.Memory ≤ Limits.Memory`}
                            />
                          ) : null}
                        </div>
                      </Form.Item>
                    </Col>
                  </Row>
                </>
              ) : null}
            </>
          ) : null}
          {[SERVICES.Kafka].includes(this.props.service.key) ? (
            <>
              <Form.Item label="Exporter自定义配置资源限制:">
                <Radio.Group
                  value={!!serviceConfig?.exporter_resources}
                  onChange={(e) => {
                    this.onChangeServiceConfigComponentResources(
                      e.target.value,
                    );
                  }}
                >
                  <Radio value={true}>是</Radio>
                  <Radio value={false}>否</Radio>
                </Radio.Group>
              </Form.Item>
              {serviceConfig?.exporter_resources ? (
                <Row gutter={24}>
                  <Col span={12}>
                    <Form.Item
                      label="Requests.CPU:"
                      name={["exporter_resources", "requests", "cpu"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <Input
                        style={{ width: "200px" }}
                        value={serviceConfig?.exporter_resources?.requests?.cpu}
                        onChange={(e) => {
                          this.onChangeComponentResource("cpu", e.target.value);
                        }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      label="Requests.Memory:"
                      name={["exporter_resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <Input
                        style={{ width: "200px" }}
                        value={
                          serviceConfig?.exporter_resources?.requests?.memory
                        }
                        onChange={(e) => {
                          this.onChangeComponentResource(
                            "memory",
                            e.target.value,
                          );
                        }}
                      />
                    </Form.Item>
                  </Col>
                </Row>
              ) : null}
            </>
          ) : null}
        </Form>
      </div>
    );
  }
}

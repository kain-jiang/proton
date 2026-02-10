import * as React from "react";
import { DeleteOutlined } from "@aishutech/ui/icons";
import SelectNode from "../../SelectNode/component.view";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import { Form, Input, Row, Col, InputNumber, Radio } from "@aishutech/ui";
import OpensearchConfigBase from "./component.base";
import {
  DataBaseStorageType,
  booleanEmptyValidatorRules,
  emptyValidatorRules,
  replicaValidatorRules,
} from "../../helper";
import "./styles.view.scss";

export default class OpensearchConfig extends OpensearchConfigBase {
  render(): React.ReactNode {
    const { configData, dataBaseStorageType, service } = this.props;
    const { opensearchConfig, opensearchNodes } = this.state;

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
              onClick={this.props.onDeleteOpenSearchConfig.bind(this)}
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
            <Form>
              <Form.Item
                labelCol={{ span: 4 }}
                labelAlign="left"
                label="部署节点"
                required
              >
                <SelectNode
                  mode={false}
                  nodes={configData.nodesInfo}
                  selectedNodes={opensearchNodes}
                  onSelectedChange={(nodes) =>
                    this.onChangeOpensearchNode(nodes, opensearchConfig)
                  }
                />
              </Form.Item>
            </Form>
          ) : null}
          {dataBaseStorageType === DataBaseStorageType.DepositKubernetes ? (
            <Form
              labelCol={{ span: 4 }}
              labelAlign="left"
              name="opensearchForm_replicaForm"
              initialValues={opensearchConfig}
              validateTrigger="onBlur"
              ref={this.opensearchForm.replicaForm}
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
                  value={opensearchConfig?.replica_count}
                  onChange={(val) => {
                    this.onChangeOpensearch({
                      replica_count: val,
                    });
                  }}
                />
              </Form.Item>
            </Form>
          ) : null}
          <Form
            name="opensearchForm_modeForm"
            validateTrigger="onBlur"
            initialValues={opensearchConfig}
            ref={this.opensearchForm.modeForm}
          >
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="模式:"
                  name="mode"
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={opensearchConfig?.mode}
                    onChange={(e) => {
                      this.onChangeOpensearch({ mode: e.target.value });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="JVM配置:"
                  name={["config", "jvmOptions"]}
                  required
                  rules={emptyValidatorRules}
                >
                  <Input
                    style={{ width: "200px" }}
                    value={opensearchConfig?.config?.jvmOptions}
                    onChange={(e) => {
                      this.onChangeOpensearchConfig({
                        jvmOptions: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Form>
          <Form
            layout="horizontal"
            name="opensearchForm_dataForm"
            validateTrigger="onBlur"
            initialValues={opensearchConfig}
            ref={this.opensearchForm.dataForm}
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
                    value={opensearchConfig?.storage_capacity}
                    onChange={(e) => {
                      this.onChangeOpensearch({
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
                      value={opensearchConfig?.data_path}
                      onChange={(e) => {
                        this.onChangeOpensearch({ data_path: e.target.value });
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
                      value={opensearchConfig?.storageClassName}
                      onChange={(e) => {
                        this.onChangeOpensearch({
                          storageClassName: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                )}
              </Col>
            </Row>
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="远程词库:"
              name={["config", "hanlpRemoteextDict"]}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={opensearchConfig?.config?.hanlpRemoteextDict}
                onChange={(e) => {
                  this.onChangeOpensearchConfig({
                    hanlpRemoteextDict: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="去停词:"
              name={["config", "hanlpRemoteextStopwords"]}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={opensearchConfig?.config?.hanlpRemoteextStopwords}
                onChange={(e) => {
                  this.onChangeOpensearchConfig({
                    hanlpRemoteextStopwords: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="低警戒水位线:"
              name={[
                "settings",
                "cluster.routing.allocation.disk.watermark.low",
              ]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{
                    width: "95%",
                  }}
                  value={
                    opensearchConfig?.settings?.[
                      "cluster.routing.allocation.disk.watermark.low"
                    ]
                  }
                  onChange={(e) => {
                    this.onChangeSettings({
                      ["cluster.routing.allocation.disk.watermark.low"]:
                        e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title={
                    "控制磁盘使用的低警戒水位线。当设置为百分比时，OpenSearch不会将分片分配给使用率超过该百分比磁盘的节点。这也可以设置为比值，如0.85。最后，也可以设置为字节值，如400mb。此设置不会影响新创建索引的主分片，但会阻止分配其副本。默认值为85%。\n为保证服务正常运行，请设置合理的警戒水位线：低警戒水位线<高警戒水位线<洪泛警戒水位线"
                  }
                />
              </div>
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="高警戒水位线:"
              name={[
                "settings",
                "cluster.routing.allocation.disk.watermark.high",
              ]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{
                    width: "95%",
                  }}
                  value={
                    opensearchConfig?.settings?.[
                      "cluster.routing.allocation.disk.watermark.high"
                    ]
                  }
                  onChange={(e) => {
                    this.onChangeSettings({
                      ["cluster.routing.allocation.disk.watermark.high"]:
                        e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title={
                    "控制磁盘使用的高警戒水位线。当设置为百分比时，OpenSearch将尝试从磁盘使用率高于该百分比的节点重新迁移碎片。这也可以设置为比值，如0.85。最后，也可以设置为字节值，如400mb。此设置会影响所有碎片的分配。默认值为90%。\n为保证服务正常运行，请设置合理的警戒水位线：低警戒水位线<高警戒水位线<洪泛警戒水位线"
                  }
                />
              </div>
            </Form.Item>
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="洪泛警戒水位线:"
              name={[
                "settings",
                "cluster.routing.allocation.disk.watermark.flood_stage",
              ]}
              required
              rules={emptyValidatorRules}
            >
              <div>
                <Input
                  style={{
                    width: "95%",
                  }}
                  value={
                    opensearchConfig?.settings?.[
                      "cluster.routing.allocation.disk.watermark.flood_stage"
                    ]
                  }
                  onChange={(e) => {
                    this.onChangeSettings({
                      ["cluster.routing.allocation.disk.watermark.flood_stage"]:
                        e.target.value,
                    });
                  }}
                />
                <QuestionCircleOutlined
                  style={{
                    marginLeft: "6px",
                  }}
                  title={
                    "控制磁盘使用的洪泛警戒水位线。这是防止节点耗尽磁盘空间的最后手段。当有一块磁盘超过洪泛警戒水位线时，OpenSearch会强制将位于该节点上所有分片的所有索引置为只读模式。一旦磁盘利用率低于高水位线，索引块就被释放。这也可以设置为比值，如0.85。最后，也可以设置为字节值，如400mb。默认值为95%。\n为保证服务正常运行，请设置合理的警戒水位线：低警戒水位线<高警戒水位线<洪泛警戒水位线"
                  }
                />
              </div>
            </Form.Item>
            <Form.Item
              labelCol={{ span: 6 }}
              labelAlign="left"
              label="http.max_initial_line_length:"
              name={["settings", "http.max_initial_line_length"]}
              required
              rules={emptyValidatorRules}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={
                  opensearchConfig?.settings?.["http.max_initial_line_length"]
                }
                onChange={(e) => {
                  this.onChangeSettings({
                    ["http.max_initial_line_length"]: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelCol={{ span: 6 }}
              labelAlign="left"
              label="cluster.max_shards_per_node:"
              name={["settings", "cluster.max_shards_per_node"]}
              required
              rules={emptyValidatorRules}
            >
              <Input
                style={{
                  width: "95%",
                }}
                value={
                  opensearchConfig?.settings?.["cluster.max_shards_per_node"]
                }
                onChange={(e) => {
                  this.onChangeSettings({
                    ["cluster.max_shards_per_node"]: e.target.value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item
              labelAlign="left"
              label="内存锁定:"
              name={["settings", "bootstrap.memory_lock"]}
              required
              rules={booleanEmptyValidatorRules}
            >
              <Radio.Group
                value={opensearchConfig?.settings?.["bootstrap.memory_lock"]}
                onChange={(e) => {
                  this.onChangeSettings({
                    ["bootstrap.memory_lock"]: e.target.value,
                  });
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            <Form.Item
              labelAlign="left"
              label="开启NFS快照仓库:"
              name={["extraValues", "storage", "repo", "nfs", "enabled"]}
              required
              rules={booleanEmptyValidatorRules}
            >
              <Radio.Group
                value={
                  opensearchConfig?.extraValues?.storage?.repo?.nfs?.enabled
                }
                onChange={(e) => {
                  this.onChangeNFSConfig("nfs", { enabled: e.target.value });
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            {opensearchConfig?.extraValues?.storage?.repo?.nfs?.enabled ? (
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelAlign="left"
                    label="NFS快照仓库IP:"
                    name={["extraValues", "storage", "repo", "nfs", "server"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={
                        opensearchConfig?.extraValues?.storage?.repo?.nfs
                          ?.server
                      }
                      onChange={(e) => {
                        this.onChangeNFSConfig("nfs", {
                          server: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelAlign="left"
                    label="NFS快照仓库路径:"
                    name={["extraValues", "storage", "repo", "nfs", "path"]}
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={
                        opensearchConfig?.extraValues?.storage?.repo?.nfs?.path
                      }
                      onChange={(e) => {
                        this.onChangeNFSConfig("nfs", { path: e.target.value });
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            ) : null}
            <Form.Item
              labelAlign="left"
              label="开启HDFS快照仓库:"
              name={["extraValues", "storage", "repo", "hdfs", "enabled"]}
              required
              rules={booleanEmptyValidatorRules}
            >
              <Radio.Group
                value={
                  opensearchConfig?.extraValues?.storage?.repo?.hdfs?.enabled
                }
                onChange={(e) => {
                  this.onChangeNFSConfig("hdfs", { enabled: e.target.value });
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            <Form.Item label="自定义配置资源限制:">
              <Radio.Group
                value={!!opensearchConfig?.resources}
                onChange={(e) => {
                  this.onChangeOpensearchConfigResources(e.target.value);
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            {opensearchConfig?.resources ? (
              <>
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
                          value={opensearchConfig?.resources?.limits?.cpu}
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
                          value={opensearchConfig?.resources?.limits?.memory}
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
                          value={opensearchConfig?.resources?.requests?.cpu}
                          onChange={(e) => {
                            this.onChangeResource(
                              "requests",
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
                      label="Requests.Memory:"
                      name={["resources", "requests", "memory"]}
                      required
                      rules={emptyValidatorRules}
                    >
                      <div>
                        <Input
                          style={{ width: "200px" }}
                          value={opensearchConfig?.resources?.requests?.memory}
                          onChange={(e) => {
                            this.onChangeResource(
                              "requests",
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
              </>
            ) : null}
            <Form.Item label="Exporter自定义配置资源限制:">
              <Radio.Group
                value={!!opensearchConfig?.exporter_resources}
                onChange={(e) => {
                  this.onChangeOpensearchConfigComponentResources(
                    e.target.value,
                  );
                }}
              >
                <Radio value={true}>是</Radio>
                <Radio value={false}>否</Radio>
              </Radio.Group>
            </Form.Item>
            {opensearchConfig?.exporter_resources ? (
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
                      value={
                        opensearchConfig?.exporter_resources?.requests?.cpu
                      }
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
                        opensearchConfig?.exporter_resources?.requests?.memory
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
          </Form>
        </div>
      </div>
    );
  }
}

import * as React from "react";
import { Form, Input, Checkbox, Row, Col } from "@aishutech/ui";
import NetworkConfigBase from "./component.base";
import SelectNode from "../SelectNode/component.view";
import {
  CSPlugins,
  DataBaseStorageType,
  IP_Family,
  emptyValidatorRules,
} from "../helper";
import "./styles.view.scss";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";

export default class NetworkConfig extends NetworkConfigBase {
  render(): React.ReactNode {
    const { dataBaseStorageType, networkNodesValidateState } = this.props;
    return (
      <div className="network-contain">
        {dataBaseStorageType === DataBaseStorageType.Standard ? (
          <>
            <Form>
              <Form.Item label="Kubernetes Master节点" required>
                <SelectNode
                  mode={false}
                  nodes={this.props.configData.nodesInfo}
                  selectedNodes={this.state.nodes}
                  onSelectedChange={this.onChangeMasterNode.bind(this)}
                />
                {networkNodesValidateState ? (
                  <div style={{ color: "#FF4D4F" }}>请至少添加一个节点</div>
                ) : null}
              </Form.Item>
            </Form>
            <Form
              validateTrigger="onBlur"
              ref={this.networkForm.hostNetworkForm}
            >
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    label="docker IP："
                    name="bip"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={this.state.networkInfo.hostNetwork.bip}
                      onChange={(e) => {
                        this.onChangeNetworkConfig({ bip: e.target.value });
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    label="Pod 网段："
                    name="podNetworkCidr"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={this.state.networkInfo.hostNetwork.podNetworkCidr}
                      onChange={(e) => {
                        this.onChangeNetworkConfig({
                          podNetworkCidr: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    label="Serivce 网段："
                    name="serviceCidr"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={this.state.networkInfo.hostNetwork.serviceCidr}
                      onChange={(e) => {
                        this.onChangeNetworkConfig({
                          serviceCidr: e.target.value,
                        });
                      }}
                    />
                  </Form.Item>
                </Col>
                {this.state.networkInfo.ipFamilies[0] ===
                IP_Family.dualStack ? (
                  <>
                    <Col span={12}>
                      <Form.Item
                        label="IPv4 网卡："
                        name="ipv4Interface"
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{ width: "90%" }}
                            value={
                              this.state.networkInfo.hostNetwork.ipv4Interface
                            }
                            onChange={(e) => {
                              this.onChangeNetworkConfig({
                                ipv4Interface: e.target.value,
                              });
                            }}
                          />
                          <QuestionCircleOutlined
                            style={{
                              marginLeft: "6px",
                            }}
                            title="IPv4地址所在的网卡名"
                          />
                        </div>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label="IPv6 网卡："
                        name="ipv6Interface"
                        required
                        rules={emptyValidatorRules}
                      >
                        <div>
                          <Input
                            style={{ width: "90%" }}
                            value={
                              this.state.networkInfo.hostNetwork.ipv6Interface
                            }
                            onChange={(e) => {
                              this.onChangeNetworkConfig({
                                ipv6Interface: e.target.value,
                              });
                            }}
                          />
                          <QuestionCircleOutlined
                            style={{
                              marginLeft: "6px",
                            }}
                            title="IPv6地址所在的网卡名"
                          />
                        </div>
                      </Form.Item>
                    </Col>
                  </>
                ) : null}
              </Row>
            </Form>
            <Form
              layout="horizontal"
              validateTrigger="onBlur"
              ref={this.networkForm.networkInfoForm}
            >
              <Form.Item
                label="etcd 数据路径："
                name="etcdDataDir"
                required
                rules={emptyValidatorRules}
              >
                <Input
                  value={this.state.networkInfo.etcdDataDir}
                  onChange={(e) => {
                    this.onChangeDataDir({ etcdDataDir: e.target.value });
                  }}
                />
              </Form.Item>
              <Form.Item
                label="docker 数据路径："
                name="dockerDataDir"
                required
                rules={emptyValidatorRules}
              >
                <Input
                  value={this.state.networkInfo.dockerDataDir}
                  onChange={(e) => {
                    this.onChangeDataDir({ dockerDataDir: e.target.value });
                  }}
                />
              </Form.Item>
            </Form>
          </>
        ) : null}
        {dataBaseStorageType === DataBaseStorageType.DepositKubernetes ? (
          <Form>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="命名空间：">
                  <Input
                    value={this.state.deploy.namespace}
                    onChange={(e) => {
                      this.onChangeDeployConfig({ namespace: e.target.value });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="ServiceAccount：">
                  <Input
                    value={this.state.deploy.serviceaccount}
                    onChange={(e) => {
                      this.onChangeDeployConfig({
                        serviceaccount: e.target.value,
                      });
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        ) : null}
        <>
          <Row>
            <Col
              span={24}
              style={{
                color: "#000000",
                height: "30px",
                lineHeight: "30px",
                fontSize: "14px",
                fontWeight: "bold",
              }}
            >
              <span className="split"></span>
              <span>可选插件</span>
            </Col>
          </Row>
          <div
            style={{
              borderTop: "2px solid #EEEEEE",
              margin: "10px 0",
            }}
          ></div>
          <div>
            {CSPlugins.map((value) => {
              return (
                <div>
                  <Checkbox
                    checked={this.state.networkInfo?.addons?.includes(
                      value.key,
                    )}
                    onChange={(e) => this.onChangeCSPlugins(e, value.key)}
                  >
                    {value.key}
                  </Checkbox>
                  <QuestionCircleOutlined title={value.descritpion} />
                </div>
              );
            })}
          </div>
        </>
      </div>
    );
  }
}

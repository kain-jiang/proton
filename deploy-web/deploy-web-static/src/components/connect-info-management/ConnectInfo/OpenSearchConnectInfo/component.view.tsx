import * as React from "react";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Radio,
  Divider,
  Row,
  Col,
} from "@kweaver-ai/ui";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/icons";
import {
  OPENSEARCH_VERSION,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
  searchEngineType,
  searchEngineProtocolType,
} from "../../../component-management/helper";
import { OpenSearchConnectInfoBase } from "./component.base";
import styles from "./styles.module.less";
import __ from "../locale";
import { noop } from "lodash";

export class OpenSearchConnectInfo extends OpenSearchConnectInfoBase {
  render(): React.ReactNode {
    const { opensearch } = this.state;

    return (
      <div>
        <Divider orientation="left" orientationMargin="0">
          {__("连接配置")}
        </Divider>
        <Form
          layout="horizontal"
          name="opensearch"
          validateTrigger="onBlur"
          initialValues={opensearch}
          ref={this.form}
        >
          {opensearch?.source_type === SOURCE_TYPE.EXTERNAL ? (
            <>
              <div className={styles["component-title"]}>{__("账户信息")}</div>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("用户名")}
                    name="username"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      style={{ width: "200px" }}
                      value={opensearch?.username}
                      onChange={(e) => {
                        this.changeOpenSearchConnectInfo(
                          "username",
                          e.target.value,
                          opensearch
                        );
                      }}
                      disabled={
                        opensearch?.source_type === this.props.originSourceType
                      }
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("密码")}
                    name="password"
                    required
                    rules={
                      opensearch?.source_type === this.props.originSourceType
                        ? undefined
                        : emptyValidatorRules
                    }
                  >
                    <div>
                      <Input.Password
                        style={{ width: "200px" }}
                        value={
                          opensearch?.source_type ===
                          this.props.originSourceType
                            ? "******"
                            : opensearch?.password
                        }
                        onChange={(e) => {
                          this.changeOpenSearchConnectInfo(
                            "password",
                            e.target.value,
                            opensearch
                          );
                        }}
                        disabled={
                          opensearch?.source_type ===
                          this.props.originSourceType
                        }
                      />
                    </div>
                  </Form.Item>
                </Col>
              </Row>
              <div className={styles["component-title"]}>{__("连接信息")}</div>
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
                        value={opensearch?.hosts}
                        onChange={(e) => {
                          this.changeOpenSearchConnectInfo(
                            "hosts",
                            e.target.value,
                            opensearch
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
                    name="port"
                    required
                    rules={portValidatorRules}
                  >
                    <InputNumber
                      style={{ width: "200px" }}
                      value={opensearch?.port}
                      onChange={(val) => {
                        this.changeOpenSearchConnectInfo(
                          "port",
                          val,
                          opensearch
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("协议")}
                    name="protocol"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Select
                      showArrow
                      style={{ width: "200px" }}
                      optionLabelProp="label"
                      value={opensearch?.protocol}
                      getPopupContainer={(node) =>
                        node.parentElement || document.body
                      }
                      onSelect={(val) => {
                        this.changeOpenSearchConnectInfo(
                          "protocol",
                          val,
                          opensearch
                        );
                      }}
                    >
                      {Object.values(searchEngineProtocolType).map((key) => (
                        <Select.Option key={key} label={key}>
                          {key}
                        </Select.Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
            </>
          ) : null}

          <div className={styles["component-title"]}>{__("版本信息")}</div>
          <Form.Item
            name="distribution"
            labelCol={{ span: 2 }}
            labelAlign="left"
            label={
              <span className={styles["form-label"]} title={__("发行版")}>
                {__("发行版")}
              </span>
            }
            required
            rules={emptyValidatorRules}
          >
            <Select
              showArrow
              disabled={opensearch?.source_type === SOURCE_TYPE.INTERNAL}
              style={{ width: "200px" }}
              optionLabelProp="label"
              placeholder={__("请选择发行版")}
              value={opensearch?.distribution}
              getPopupContainer={(node) => node.parentElement || document.body}
              onSelect={(val) => {
                this.changeOpenSearchConnectInfo(
                  "distribution",
                  val,
                  opensearch
                );
              }}
            >
              {Object.values(searchEngineType).map((key) => (
                <Select.Option key={key} label={key}>
                  {key}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            labelCol={{ span: 2 }}
            labelAlign="left"
            label={__("版本")}
            required
          >
            <Select
              showArrow
              disabled={opensearch?.source_type === SOURCE_TYPE.INTERNAL}
              style={{ width: "200px" }}
              optionLabelProp="label"
              placeholder={__("请选择OpenSearch/ElasticSearch版本")}
              value={opensearch?.version}
              getPopupContainer={(node) => node.parentElement || document.body}
              onSelect={(val) => {
                this.changeOpenSearchConnectInfo("version", val, opensearch);
                this.props.updateConnectInfoValidateState({
                  OPENSEARCH_VERSION: ValidateState.Normal,
                });
              }}
            >
              {Object.keys(OPENSEARCH_VERSION).map((key) => (
                <Select.Option
                  key={OPENSEARCH_VERSION[key]}
                  label={OPENSEARCH_VERSION[key]}
                >
                  {OPENSEARCH_VERSION[key]}
                </Select.Option>
              ))}
            </Select>
            <QuestionCircleOutlined
              onPointerEnterCapture={noop}
              onPointerLeaveCapture={noop}
              style={{
                marginLeft: "6px",
              }}
              title={__(
                "使用OpenSearch时必须选择7.x.x； 使用ES时选择对应版本；其他情况下建议选择 7.x.x\n此版本信息可能和实际配置不一致。"
              )}
            />
            {this.props.connectInfoValidateState.OPENSEARCH_VERSION ? (
              <div style={{ color: "#FF4D4F" }}>{__("此项不允许为空。")}</div>
            ) : null}
          </Form.Item>
        </Form>
      </div>
    );
  }
}

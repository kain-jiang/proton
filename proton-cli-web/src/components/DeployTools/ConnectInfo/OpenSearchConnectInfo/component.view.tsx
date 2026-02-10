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
} from "@aishutech/ui";
import { Title } from "../../Title/component.view";
import { QuestionCircleOutlined } from "@aishutech/ui/icons";
import {
  OPENSEARCH_VERSION,
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
  ValidateState,
  searchEngineType,
} from "../../helper";
import { OpenSearchConnectInfoBase } from "./component.base";
import "./styles.view.scss";

export class OpenSearchConnectInfo extends OpenSearchConnectInfoBase {
  render(): React.ReactNode {
    const { opensearch } = this.state;

    return (
      <div className="service-box">
        <Title
          title={"SearchEngine" + "连接信息"}
          deleteCallback={
            opensearch?.source_type === SOURCE_TYPE.EXTERNAL &&
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
          value={opensearch?.source_type}
          onChange={(e) => {
            this.changeOpenSearchConnectInfo(
              "source_type",
              e.target.value,
              opensearch,
            );
          }}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.OPENSEARCH]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>第三方 搜索与分析引擎</Radio>
        </Radio.Group>
        <Form
          layout="horizontal"
          name="opensearch"
          validateTrigger="onBlur"
          initialValues={opensearch}
          ref={this.form}
        >
          {opensearch?.source_type === SOURCE_TYPE.EXTERNAL ? (
            <>
              <Divider orientation="left" orientationMargin="0">
                账户信息
              </Divider>
              <Row>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label="用户名:"
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
                          opensearch,
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label="密码:"
                    name="password"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input.Password
                      style={{ width: "200px" }}
                      value={opensearch?.password}
                      onChange={(e) => {
                        this.changeOpenSearchConnectInfo(
                          "password",
                          e.target.value,
                          opensearch,
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Divider orientation="left" orientationMargin="0">
                连接信息
              </Divider>
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
                        value={opensearch?.hosts}
                        onChange={(e) => {
                          this.changeOpenSearchConnectInfo(
                            "hosts",
                            e.target.value,
                            opensearch,
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
                          opensearch,
                        );
                      }}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </>
          ) : null}
          <Divider orientation="left" orientationMargin="0">
            版本信息
          </Divider>
          <Form.Item
            name="distribution"
            labelCol={{ span: 2 }}
            labelAlign="left"
            label="发行版:"
            required
            rules={emptyValidatorRules}
          >
            <Select
              showArrow
              disabled={opensearch?.source_type === SOURCE_TYPE.INTERNAL}
              style={{ width: "200px" }}
              optionLabelProp="label"
              placeholder="请选择发行版"
              value={opensearch?.distribution}
              onSelect={(val) => {
                this.changeOpenSearchConnectInfo(
                  "distribution",
                  val,
                  opensearch,
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
            label="版本:"
            required
          >
            <Select
              showArrow
              disabled={opensearch?.source_type === SOURCE_TYPE.INTERNAL}
              style={{ width: "200px" }}
              optionLabelProp="label"
              placeholder="请选择OpenSearch/ElasticSearch版本"
              value={opensearch?.version}
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
              style={{
                marginLeft: "6px",
              }}
              title={`使用OpenSearch时必须选择7.x.x； 使用ES时选择对应版本；其他情况下建议选择 7.x.x\n此版本信息可能和实际配置不一致。`}
            />
            {this.props.connectInfoValidateState.OPENSEARCH_VERSION ? (
              <div style={{ color: "#FF4D4F" }}>此项不允许为空。</div>
            ) : null}
          </Form.Item>
        </Form>
      </div>
    );
  }
}

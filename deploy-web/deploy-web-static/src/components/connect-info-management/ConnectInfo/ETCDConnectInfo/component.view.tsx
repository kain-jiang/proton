import * as React from "react";
import {
  Form,
  Input,
  InputNumber,
  Radio,
  Divider,
  Row,
  Col,
} from "@kweaver-ai/ui";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/icons";
import { ETCDConnectInfoBase } from "./component.base";
import {
  SOURCE_TYPE,
  emptyValidatorRules,
  portValidatorRules,
} from "../../../component-management/helper";
import styles from "./styles.module.less";
import __ from "../locale";
import { noop } from "lodash";

export class ETCDConnectInfo extends ETCDConnectInfoBase {
  render(): React.ReactNode {
    const { etcd } = this.state;

    return (
      <div>
        {etcd?.source_type === SOURCE_TYPE.EXTERNAL ? (
          <Form
            layout="horizontal"
            name="proton-etcd"
            validateTrigger="onBlur"
            initialValues={etcd}
            ref={this.form}
          >
            <Divider orientation="left" orientationMargin="0">
              {__("连接配置")}
            </Divider>
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
                      value={etcd?.hosts}
                      onChange={(e) => {
                        this.changeETCDConnectInfo(
                          "hosts",
                          e.target.value,
                          etcd
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
                    value={etcd?.port}
                    onChange={(val) => {
                      this.changeETCDConnectInfo("port", val, etcd);
                    }}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        ) : null}
      </div>
    );
  }
}

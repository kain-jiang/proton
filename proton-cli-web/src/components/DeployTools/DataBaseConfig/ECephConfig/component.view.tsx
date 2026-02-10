import * as React from "react";
import SelectNode from "../../SelectNode/component.view";
import { Form, Input, Row, Col } from "@aishutech/ui";
import { DeleteOutlined, QuestionCircleOutlined } from "@aishutech/ui/icons";
import { KeepalivedEnum } from "./index.d";
import ECephConfigBase from "./component.base";
import "./styles.view.scss";
import {
  vipValidatorRules,
} from "../../helper";

export default class ECephConfig extends ECephConfigBase {
  render(): React.ReactNode {
    const { configData, service } = this.props;
    const { ecephConfig, ecephNodes } = this.state;

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
              onClick={this.props.onDeleteECephConfig.bind(this)}
            />
          </Col>
        </Row>
        <div
          style={{
            borderTop: "2px solid #EEEEEE",
            margin: "10px 0",
          }}
        ></div>
        <Form.Item
          labelCol={{ span: 4 }}
          labelAlign="left"
          label="部署节点"
          required
        >
          <SelectNode
            mode={false}
            nodes={configData.nodesInfo}
            selectedNodes={ecephNodes}
            onSelectedChange={(nodes) =>
              this.onChangeECephNode(nodes, ecephConfig)
            }
          />
        </Form.Item>
        <div
          style={{
            marginTop: "10px",
          }}
        >
          <Form
            layout="horizontal"
            name="ecephForm_keepalivedForm"
            validateTrigger="onBlur"
            initialValues={ecephConfig?.keepalived}
            ref={this.ecephForm.keepalivedForm}
          >
            <Row>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="内部虚拟地址:"
                  name="internal"
                  rules={vipValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={ecephConfig?.keepalived?.internal}
                      placeholder="请输入IP地址及子网掩码"
                      onChange={(e) => {
                        this.onChangeECephKeepalived({
                          internal: e.target.value,
                        });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={
                        "当和AS融合部署且开启OSSGateway服务的情况下，OSSGateway连接该IP进行数据传输，缓解业务网络带宽压力。\n请按照以下规则输入内容:\nIPV4地址 + 掩码：\n虚拟地址格式形如 XXX.XXX.XXX.XXX/Y，每段必须是 0~255 之间的整数，掩码必须是 1~64 之间的整数。\nIPV6地址 + 掩码：\n虚拟地址格式形如 X:X:X:X:X:X:X:X/Y，其中X表示地址中的16b，以16进制表示；Y表示掩码，必须是 1~64 之间的整数。"
                      }
                    />
                  </div>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  labelCol={{ span: 8 }}
                  labelAlign="left"
                  label="外部虚拟地址:"
                  name="external"
                  rules={vipValidatorRules}
                >
                  <div>
                    <Input
                      style={{ width: "200px" }}
                      value={ecephConfig?.keepalived?.external}
                      placeholder="请输入IP地址及子网掩码"
                      onChange={(e) => {
                        this.onChangeECephKeepalived({
                          external: e.target.value,
                        });
                      }}
                    />
                    <QuestionCircleOutlined
                      style={{
                        marginLeft: "6px",
                      }}
                      title={
                        "外部网络（对象存储业务网络）的VIP，对接客户端进行数据读写。\n请按照以下规则输入内容:\nIPV4地址 + 掩码：\n虚拟地址格式形如 XXX.XXX.XXX.XXX/Y，每段必须是 0~255 之间的整数，掩码必须是 1~64 之间的整数。\nIPV6地址 + 掩码：\n虚拟地址格式形如 X:X:X:X:X:X:X:X/Y，其中X表示地址中的16b，以16进制表示；Y表示掩码，必须是 1~64 之间的整数。"
                      }
                    />
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </Form>
          <Form layout="horizontal">
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="Secret名称:"
            >
              <Input
                value={ecephConfig?.tls?.secret}
                onChange={(e) => {
                  this.onChangeECephTLS({ secret: e.target.value });
                }}
              />
            </Form.Item>
          </Form>
          <Form layout="horizontal">
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="数字证书:"
            >
              <Input
                value={ecephConfig?.tls?.["certificate-data"]}
                onChange={(e) => {
                  this.onChangeECephTLS({
                    ["certificate-data"]: e.target.value,
                  });
                }}
              />
            </Form.Item>
          </Form>
          <Form layout="horizontal">
            <Form.Item
              labelCol={{ span: 4 }}
              labelAlign="left"
              label="证书密钥:"
            >
              <Input
                value={ecephConfig?.tls?.["key-data"]}
                onChange={(e) => {
                  this.onChangeECephTLS({ ["key-data"]: e.target.value });
                }}
              />
            </Form.Item>
          </Form>
        </div>
      </div>
    );
  }
}

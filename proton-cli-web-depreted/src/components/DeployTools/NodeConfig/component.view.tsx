import * as React from "react";
import {
  Button,
  Card,
  Form,
  Input,
  Row,
  Col,
  Table,
  Typography,
  Popconfirm,
  Select,
  Tooltip,
  Radio,
  Space,
  Switch,
} from "@aishutech/ui";
import {
  validatorMessge,
  nodeInfoCheckResult,
  IP_Family_LIST,
  IP_Family,
  CHRONY_MODE,
  CHRONY_MODE_TEXT,
  emptyValidatorRules,
  FIREWALL_MODE,
  FIREWALL_MODE_TEXT,
} from "../helper";
import NodeConfigBase from "./component.base";
import { Title } from "../Title/component.view";
import "./styles.view.scss";

export default class NodeConfig extends NodeConfigBase {
  render(): React.ReactNode {
    const { CSIPFamily, chrony, firewall } = this.state;
    const columns = [
      {
        title: "节点名称(未填写时根据IP自动生成)",
        dataIndex: "name",
        key: "name",
        width: "20%",
        editable: true,
      },
      {
        title: CSIPFamily[0] !== IP_Family.ipv6 ? "IPv4(必填)" : "IPv4(非必填)",
        dataIndex: "ipv4",
        key: "ipv4",
        width: "20%",
        editable: true,
      },
      {
        title: CSIPFamily[0] !== IP_Family.ipv4 ? "IPv6(必填)" : "IPv6(非必填)",
        dataIndex: "ipv6",
        key: "ipv6",
        width: "20%",
        editable: true,
      },
      {
        title: "内部ip(非必填)",
        dataIndex: "internal_ip",
        key: "internal_ip",
        width: "20%",
        editable: true,
      },
      {
        title: "操作",
        dataIndex: "operation",
        width: "20%",
        render: (_, record, index) => {
          const editable = this.getNodeEditingStatus(index);
          return editable ? (
            <span>
              <Typography.Link
                style={{
                  marginRight: 10,
                }}
                onClick={() => {
                  this.onSaveEdit(record);
                }}
              >
                保存
              </Typography.Link>

              <Typography.Link
                onClick={this.onCancelEditing.bind(this)}
                style={{
                  marginRight: 10,
                }}
              >
                取消
              </Typography.Link>
              <span
                style={{
                  marginLeft: 10,
                  color: "red",
                }}
              >
                {this.state.validator != nodeInfoCheckResult.Normal
                  ? validatorMessge[this.state.validator].message
                  : ""}
              </span>
            </span>
          ) : (
            <span>
              <Typography.Link
                style={{
                  marginRight: 10,
                }}
                onClick={() => {
                  this.onClickEdit(record, index);
                }}
                disabled={
                  this.state.isEditingNodeIndex !== -1 &&
                  !this.getNodeEditingStatus(index)
                }
              >
                编辑
              </Typography.Link>
              <Popconfirm
                style={{
                  marginRight: 8,
                }}
                title="确定删除?"
                onConfirm={() => {
                  this.onDeleteNode(index);
                }}
                disabled={
                  this.state.isEditingNodeIndex !== -1 &&
                  !this.getNodeEditingStatus(index)
                }
              >
                <a
                  className={
                    this.state.isEditingNodeIndex !== -1 &&
                    !this.getNodeEditingStatus(index)
                      ? "pop-disabled"
                      : ""
                  }
                >
                  删除
                </a>
              </Popconfirm>
            </span>
          );
        },
      },
    ];

    const mergedColumns = columns.map((col) => {
      if (!col.editable) {
        return col;
      } else {
        return {
          ...col,
          onCell: (record, index) => ({
            record,
            inputType: "text",
            dataIndex: col.dataIndex,
            title: col.title,
            editing: this.getNodeEditingStatus(index),
          }),
        };
      }
    });

    const radioGroupStyle = { margin: "0 20px" };
    const formStyle = { marginLeft: "25px" };

    return (
      <div className="card-zone-box-wrap">
        <div className="card-zone-box">
          <Title title={"时间同步配置"} />
          <Row>
            <Col span={24}>
              <Radio.Group
                style={radioGroupStyle}
                value={chrony?.mode}
                disabled={this.state.isEditingNodeIndex !== -1}
                onChange={(e) => {
                  this.props.updateChrony({ mode: e.target.value });
                }}
              >
                <Space direction="vertical">
                  <Radio value={CHRONY_MODE.EXTERNAL_NTP}>
                    {CHRONY_MODE_TEXT[CHRONY_MODE.EXTERNAL_NTP]}
                  </Radio>
                  {chrony?.mode === CHRONY_MODE.EXTERNAL_NTP ? (
                    <Form
                      style={formStyle}
                      validateTrigger="onBlur"
                      ref={this.nodeForm.serverForm}
                      initialValues={this.props.configData.chrony}
                    >
                      <Form.Item
                        label="外部NTP服务器地址:"
                        name="server"
                        required
                        rules={emptyValidatorRules}
                        style={{ marginBottom: 0 }}
                      >
                        <Input
                          // placeholder="请输入时间服务器地址，多个地址请用英文逗号分隔开"
                          placeholder="请输入正确格式的时间服务器地址"
                          value={chrony?.server}
                          disabled={this.state.isEditingNodeIndex !== -1}
                          onChange={(e) =>
                            this.props.updateChrony({
                              server: e.target.value.split(","),
                            })
                          }
                          style={{
                            minWidth: "110px",
                            maxWidth: "500px",
                          }}
                        />
                      </Form.Item>
                    </Form>
                  ) : null}
                  <Radio value={CHRONY_MODE.LOCAL_MASTER}>
                    {CHRONY_MODE_TEXT[CHRONY_MODE.LOCAL_MASTER]}
                  </Radio>
                  <Radio value={CHRONY_MODE.USER_MANAGED}>
                    {CHRONY_MODE_TEXT[CHRONY_MODE.USER_MANAGED]}
                  </Radio>
                </Space>
              </Radio.Group>
            </Col>
          </Row>
        </div>
        <div className="card-zone-box">
          <Title title={"防火墙配置"} />
          <Row>
            <Col span={24}>
              <Radio.Group
                style={radioGroupStyle}
                value={firewall?.mode}
                disabled={this.state.isEditingNodeIndex !== -1}
                onChange={(e) => {
                  this.props.updateFirewall({ mode: e.target.value });
                }}
              >
                <Space direction="vertical">
                  <Radio value={FIREWALL_MODE.FIREWALLD}>
                    {FIREWALL_MODE_TEXT[FIREWALL_MODE.FIREWALLD]}
                  </Radio>
                  <Radio value={FIREWALL_MODE.USER_MANAGED}>
                    {FIREWALL_MODE_TEXT[FIREWALL_MODE.USER_MANAGED]}
                  </Radio>
                </Space>
              </Radio.Group>
            </Col>
          </Row>
        </div>
        <div className="card-zone-box">
          <Title
            title={"SSH远程连接配置"}
            tip="多节点部署时，所有SSH远程连接密码须保持一致。"
          />
          <Row>
            <Col span={24}>
              <Form
                layout="inline"
                initialValues={this.props.accountInfo}
                validateTrigger="onBlur"
                ref={this.nodeForm.acountInfoForm}
              >
                <Col span={12}>
                  <Form.Item
                    label="账号（管理员权限）:"
                    name="sshAccount"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input
                      value={this.props.accountInfo.sshAccount}
                      onChange={(e) =>
                        this.props.updateSSHInfo({
                          sshAccount: e.target.value,
                        })
                      }
                      style={{
                        minWidth: "110px",
                        maxWidth: "350px",
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="密码:"
                    name="sshPassword"
                    required
                    rules={emptyValidatorRules}
                  >
                    <Input.Password
                      value={this.props.accountInfo.sshPassword}
                      onChange={(e) =>
                        this.props.updateSSHInfo({
                          sshPassword: e.target.value,
                        })
                      }
                      style={{
                        maxWidth: "350px",
                        minWidth: "110px",
                      }}
                    />
                  </Form.Item>
                </Col>
              </Form>
            </Col>
          </Row>
        </div>
        <div className="card-zone-box">
          <Title title={"服务器节点"} />
          <div className="nodeIP">
            <Row>
              <Col span={12}>
                <span className="nodeIP-title">K8S网络协议栈:</span>
                <Select
                  disabled={
                    this.state.isEditingNodeIndex !== -1 ||
                    !!this.state.nodesInfo.length
                  }
                  value={this.state.CSIPFamily}
                  onChange={(value) => {
                    this.onChangeCSIPFamily([value]);
                  }}
                  getPopupContainer={(node) =>
                    node.parentElement || document.body
                  }
                >
                  {IP_Family_LIST.map((item) => (
                    <Select.Option key={item.value} label={item.label}>
                      <Tooltip placement="right">
                        <span>{item.label}</span>
                      </Tooltip>
                    </Select.Option>
                  ))}
                </Select>
              </Col>
              {this.state.CSIPFamily[0] === IP_Family.dualStack ? (
                <Col span={12} className="nodeIP-item">
                  <span className="nodeIP-title">开启双栈能力:</span>
                  <Switch
                    checked={this.state.enableDualStack}
                    onChange={(value) => this.onChangeEnableDualStack(value)}
                  />
                </Col>
              ) : null}
            </Row>
          </div>

          <Button
            onClick={this.addNodeInfoData.bind(this)}
            type="primary"
            style={{
              marginBottom: 16,
            }}
            disabled={this.state.isEditingNodeIndex !== -1}
          >
            添加节点
          </Button>
          <Form onValuesChange={this.onChangeRowData.bind(this)}>
            <Table
              components={{
                body: {
                  cell: this.editableCell.bind(this),
                },
              }}
              rowClassName={() => "editable-row"}
              bordered
              pagination={false}
              dataSource={this.state.nodesInfo}
              columns={mergedColumns}
              scroll={{
                y: 350,
              }}
            />
          </Form>
          {this.props.nodesValidateState ? (
            <div style={{ color: "#FF4D4F" }}>请至少添加一个节点。</div>
          ) : null}
        </div>
      </div>
    );
  }

  /**
   *
   * @param param0 当前记录的数据
   * @returns
   */
  editableCell({
    editing,
    dataIndex,
    title,
    inputType,
    record,
    index,
    children,
    ...restProps
  }) {
    return (
      <td {...restProps}>
        {editing ? (
          <Form.Item
            name={dataIndex}
            style={{
              margin: 0,
            }}
            initialValue={this.nodeInfo[dataIndex]}
            preserve={false}
          >
            <Input style={{ minWidth: "110px" }} />
          </Form.Item>
        ) : (
          children
        )}
      </td>
    );
  }
}

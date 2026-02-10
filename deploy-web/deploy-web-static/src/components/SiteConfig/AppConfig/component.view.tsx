import * as React from "react";
import { Form, Button, Input, Popover, Spin, Modal } from "antd";
import {
    QuestionCircleOutlined,
    ExclamationCircleOutlined,
    CloseCircleOutlined,
} from "@ant-design/icons";
import {
    AppServiceAccessingAddressStatus,
    WebHttpsValidateState,
    DialogStatus,
} from "./component.base";
import AppConfigBase from "./component.base";
import styles from "./styles.module.less";
import __ from "./locale";
import { CustomSpin } from "../../component-management/components/spin";

export default class AppConfig extends AppConfigBase {
    render() {
        const { dialogStatus } = this.state;

        const addressTooltipContent = (
            <div className={styles["icon-tips"]}>
                <p style={{ color: "#000" }}>{__("请按照以下规则输入内容:")}</p>
                <p className={styles["icon-text"]}>
                    <span>{__("IPV4地址：")}</span>
                    <span>
                        {__(
                            "IP地址格式形如 XXX.XXX.XXX.XXX，每段必须是 0~255 之间的整数。"
                        )}
                    </span>
                </p>
                <p className={styles["icon-text"]}>
                    <span>{__("IPV6地址：")}</span>
                    <span>
                        {__(
                            "IP地址格式形如 X:X:X:X:X:X:X:X，其中X表示地址中的16b，以16进制表示。"
                        )}
                    </span>
                </p>
                <p className={styles["icon-text"]}>
                    <span>{__("域名：")}</span>
                    <span>
                        {__(
                            "域名只能包含 英文、数字 及 -. 字符，每一级不能以“-”字符开头或结尾，每一级长度必需 1~63 个字符，且总长不能超过253个字符。"
                        )}
                    </span>
                </p>
            </div>
        );

        const typeTooltipContent = (
            <div className={styles["icon-tips"]}>
                <p className={styles["icon-text"]}>
                    <span>{__("internal：")}</span>
                    <span>
                        {__(
                            "class-443服务的nginx实例端口将和HTTPS端口保持一致。"
                        )}
                    </span>
                </p>
                <p className={styles["icon-text"]}>
                    <span>{__("external：")}</span>
                    <span>
                        {__(
                            "class-443服务的nginx实例将固定监听80/443端口，请注意配置负载均衡器。"
                        )}
                    </span>
                </p>
            </div>
        );

        const getValidateStatus = (status: number) => {
            if (
                status === AppServiceAccessingAddressStatus.Normal ||
                status === WebHttpsValidateState.Normal
            ) {
                return "";
            }
            return "error";
        };

        const getHelpMessage = (status: number, messages: any) => {
            if (
                status === AppServiceAccessingAddressStatus.Normal ||
                status === WebHttpsValidateState.Normal
            ) {
                return "";
            }
            return messages[status] || "";
        };

        return (
            <div className={styles["accessing-tabs"]}>
                <div className={styles["flexBox-container"]}>
                    <div style={{ display: "flex" }}>
                        <div style={{ flex: 1 }}>
                            <div>
                                <div>
                                    <Form
                                        className={styles["form"]}
                                        colon={false}
                                    >
                                        <Form.Item
                                            className={styles["card-form-item"]}
                                            label={
                                                <span
                                                    className={
                                                        styles["label-common"]
                                                    }
                                                >
                                                    {__("访问地址：")}
                                                    <Popover
                                                        content={
                                                            addressTooltipContent
                                                        }
                                                        trigger="hover"
                                                        placement="topLeft"
                                                    >
                                                        <QuestionCircleOutlined
                                                            className={
                                                                styles[
                                                                    "tool-icon"
                                                                ]
                                                            }
                                                            style={{
                                                                color: "#324673",
                                                                fontSize: 14,
                                                            }}
                                                        />
                                                    </Popover>
                                                </span>
                                            }
                                            required
                                            validateStatus={getValidateStatus(
                                                this.state
                                                    .appServiceAccessingAddressStatus
                                            )}
                                            help={getHelpMessage(
                                                this.state
                                                    .appServiceAccessingAddressStatus,
                                                {
                                                    [AppServiceAccessingAddressStatus.AppNodeEmpty]:
                                                        __(
                                                            "请设置管理控制台节点"
                                                        ),
                                                    [AppServiceAccessingAddressStatus.Empty]:
                                                        __(
                                                            "此输入项不允许为空"
                                                        ),
                                                    [AppServiceAccessingAddressStatus.ErrorEnglishLetter]:
                                                        __(
                                                            "域名只能包含 英文、数字 及 -. 字符，长度范围 3~20 个字符，请重新输入"
                                                        ),
                                                    [AppServiceAccessingAddressStatus.ErrorNoEnglish]:
                                                        __(
                                                            "IP地址格式形如 XXX.XXX.XXX.XXX，每段必须是 0~255 之间的整数，请重新输入"
                                                        ),
                                                }
                                            )}
                                        >
                                            <Input
                                                style={{ width: "200px" }}
                                                value={
                                                    this.state
                                                        .appServiceAccessingAddress
                                                }
                                                onChange={(e) =>
                                                    this.changeAppServiceAccessingAddress(
                                                        e.target.value
                                                    )
                                                }
                                            />
                                        </Form.Item>
                                        <Form.Item
                                            className={styles["card-form-item"]}
                                            label={
                                                <span
                                                    className={
                                                        styles["label-common"]
                                                    }
                                                >
                                                    {__("HTTPS端口：")}
                                                </span>
                                            }
                                            required
                                            validateStatus={getValidateStatus(
                                                this.state.webClientHttpsStatus
                                            )}
                                            help={getHelpMessage(
                                                this.state.webClientHttpsStatus,
                                                {
                                                    [WebHttpsValidateState.AppNodeEmpty]:
                                                        __(
                                                            "请设置管理控制台节点"
                                                        ),
                                                    [WebHttpsValidateState.Empty]:
                                                        __(
                                                            "此输入项不允许为空"
                                                        ),
                                                    [WebHttpsValidateState.InputError]:
                                                        __(
                                                            "端口号必须是 1~65535 之间的整数，请重新输入"
                                                        ),
                                                }
                                            )}
                                        >
                                            <Input
                                                style={{ width: "200px" }}
                                                value={
                                                    this.state.webClientPort
                                                        .https
                                                }
                                                onChange={(e) =>
                                                    this.changeWebHttps(
                                                        e.target.value
                                                    )
                                                }
                                            />
                                        </Form.Item>
                                        <Form.Item
                                            className={styles["card-form-item"]}
                                            label={
                                                <span
                                                    className={
                                                        styles["label-common"]
                                                    }
                                                >
                                                    {__("访问前缀：")}
                                                </span>
                                            }
                                        >
                                            <span
                                                className={
                                                    styles["field-common"]
                                                }
                                            >
                                                {this.state.path}
                                            </span>
                                        </Form.Item>
                                        <Form.Item
                                            className={styles["card-form-item"]}
                                            label={
                                                <span
                                                    className={
                                                        styles["label-common"]
                                                    }
                                                >
                                                    {__("访问地址类型：")}
                                                    <Popover
                                                        content={
                                                            typeTooltipContent
                                                        }
                                                        trigger="hover"
                                                        placement="topLeft"
                                                    >
                                                        <QuestionCircleOutlined
                                                            className={
                                                                styles[
                                                                    "tool-icon"
                                                                ]
                                                            }
                                                            style={{
                                                                color: "#324673",
                                                                fontSize: 14,
                                                            }}
                                                        />
                                                    </Popover>
                                                </span>
                                            }
                                        >
                                            <span
                                                className={
                                                    styles["field-common"]
                                                }
                                            >
                                                {this.state.type}
                                            </span>
                                        </Form.Item>
                                    </Form>
                                </div>
                                <div className={styles["button-position"]}>
                                    <Button
                                        className={styles["button-common"]}
                                        disabled={
                                            this.state
                                                .appServiceAccessingAddress ===
                                                this.lastAppServiceConfig
                                                    .appServiceAccessingAddress &&
                                            this.state.webClientPort.https ===
                                                this.lastAppServiceConfig
                                                    .webClientPort.https
                                        }
                                        onClick={this.completeAppService.bind(
                                            this
                                        )}
                                    >
                                        {__("保存")}
                                    </Button>
                                    <Button
                                        className={styles["button-common"]}
                                        onClick={this.cancelAppService.bind(
                                            this,
                                            DialogStatus.None
                                        )}
                                    >
                                        {__("取消")}
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <Modal
                    visible={dialogStatus === DialogStatus.ErrorDialogAppear}
                    title={
                        <span>
                            <CloseCircleOutlined
                                style={{ color: "#ff4d4f", marginRight: 8 }}
                            />
                            {__("访问配置修改失败，错误信息如下：")}
                        </span>
                    }
                    onOk={this.closeDialog.bind(this)}
                    onCancel={this.closeDialog.bind(this)}
                    okText={__("确定")}
                    cancelText={__("取消")}
                    width={520}
                    okButtonProps={{ danger: true }}
                >
                    <div
                        style={{
                            whiteSpace: "pre-wrap",
                            wordBreak: "break-word",
                        }}
                    >
                        {this.state.errorMessage}
                    </div>
                </Modal>
                {this.state.loadingStatus === true ? (
                    <CustomSpin text={__("正在保存...")} />
                ) : null}
                <Modal
                    visible={dialogStatus === DialogStatus.AlertBoxAppear}
                    title={
                        <span>
                            <ExclamationCircleOutlined
                                style={{ color: "#faad14", marginRight: 8 }}
                            />
                            {__("您确定要修改访问配置吗？")}
                        </span>
                    }
                    onOk={this.clickCompleteAppService.bind(this)}
                    onCancel={this.cancelAppService.bind(this, dialogStatus)}
                    okText={__("确定")}
                    cancelText={__("取消")}
                    width={520}
                >
                    <div className={styles["wrapper-detail"]}>
                        {__(
                            "完成修改后，您需要重新登录系统工作台并返回【访问配置】页面更新证书。"
                        )}
                    </div>
                </Modal>
            </div>
        );
    }
}

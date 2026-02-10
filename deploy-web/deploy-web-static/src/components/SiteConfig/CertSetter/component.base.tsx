import * as PropTypes from "prop-types";
import {
    SignType,
    Parts,
    PartsText,
    SignTypeText,
    isSelfSignCert,
} from "../helper";
import { AddrType } from "../../SiteConfig/helper";
import { Modal, message } from "antd";
import { manageLog } from "../../../core/log2";
import getErrorMessage from "../../../core/mediator/error";
import WebComponent from "../../webcomponent";
import { Props, State } from "./index.d";
import __ from "./locale";
import { Level, ManagementOps } from "../../../core/log";
import { accessAddr, cert, certInterface } from "../../../api/deploy-manager";
import { CertInfo } from "..";

export default class CertSetterBase extends WebComponent<Props, State> {
    state = {
        signType: SignType.Self,
        text: {
            key: "",
            serverCert: "",
        },
        validateState: {
            key: "",
            serverCert: "",
        },
    };

    static contextTypes = {
        toast: PropTypes.func,
    };

    keyNode: any = null;
    serverCertNode: any = null;

    /**
     * 切换证书类型下拉框触发
     * @param signType 签名类型
     */
    protected changeSignType(signType: string) {
        this.setState({
            signType,
            text: {
                key: "",
                serverCert: "",
            },
        });
        this.keyNode = null;
        this.serverCertNode = null;
    }

    /**
     * 点击生成证书按钮触发
     * @param certType 证书类型
     */
    protected async addSelfSignedCert(
        certType: string,
        signType: string,
        certInfo: CertInfo
    ) {
        if (!certInfo) {
            return null;
        }
        if (!certInfo.hasExpired && !isSelfSignCert(certInfo.issuer)) {
            return new Promise((resolve) => {
                Modal.confirm({
                    title: __("提示"),
                    content: __(
                        "检测到您配置的CA证书还未失效，生成自签名证书后将覆盖原有证书，是否生成自签名证书？"
                    ),
                    onOk: async () => {
                        resolve(true);
                    },
                    onCancel: () => {
                        resolve(false);
                    },
                });
            }).then(async (ret) => {
                if (!ret) {
                    return null;
                }
                await this.continueAddSelfSignedCert(
                    certType,
                    signType,
                    certInfo
                );
            });
        } else {
            await this.continueAddSelfSignedCert(certType, signType, certInfo);
        }
    }

    private async continueAddSelfSignedCert(
        certType: string,
        signType: string,
        certInfo: CertInfo
    ) {
        let timer = null;
        try {
            timer = setTimeout(() => location.reload(), 3000);
            const { host: hostname } = await accessAddr.get(
                certType === Parts.App ? AddrType.App : AddrType.Storage
            );

            await cert.setCert(certType, hostname);
            message.success(__("生成证书成功"));
            manageLog({
                level: Level.INFO,
                opType: ManagementOps.SET,
                msg: __("系统自动生成${certType}${signType} 成功", {
                    certType: PartsText[certType],
                    signType: SignTypeText[signType],
                }),
                exMsg: "",
            });
        } catch (err) {
            clearTimeout(timer as any);
            Modal.error({
                title: __("错误"),
                content: getErrorMessage("deploy", err),
            });
        }
    }

    /**
     * 点击上传触发
     * @param certType 证书类型
     */
    protected async uploadCert(
        certType: string,
        signType: string,
        certInfo: CertInfo
    ) {
        const keyNode = this.keyNode ? (this.keyNode as any).files[0] : null;
        const serverCertNode = this.serverCertNode
            ? (this.serverCertNode as any).files[0]
            : null;
        if (!keyNode || !serverCertNode) {
            if (!keyNode) {
                this.setState({
                    validateState: {
                        key: __("请先上传密钥。"),
                        serverCert: "",
                    },
                });
            } else {
                this.setState({
                    validateState: {
                        key: "",
                        serverCert: __("请先上传服务器证书。"),
                    },
                });
            }
        } else {
            if (!certInfo.hasExpired && !isSelfSignCert(certInfo.issuer)) {
                return new Promise((resolve) => {
                    Modal.confirm({
                        title: __("提示"),
                        content: __(
                            "检测到您配置的CA证书还未失效，上传后将覆盖原有证书，是否上传？"
                        ),
                        onOk: async () => {
                            resolve(true);
                        },
                        onCancel: () => {
                            resolve(false);
                        },
                    });
                }).then(async (ret) => {
                    if (!ret) {
                        return null;
                    }
                    await this.continueUploadCert(
                        certType,
                        signType,
                        certInfo,
                        keyNode,
                        serverCertNode
                    );
                });
            } else {
                await this.continueUploadCert(
                    certType,
                    signType,
                    certInfo,
                    keyNode,
                    serverCertNode
                );
            }
        }
    }

    private async continueUploadCert(
        certType: string,
        signType: string,
        certInfo: CertInfo,
        keyNode: File,
        serverCertNode: File
    ) {
        let data = new FormData();
        data.append("cert_key", keyNode);
        data.append("cert_crt", serverCertNode);
        let timer = null;
        try {
            timer = setTimeout(() => location.reload(), 3000);
            await certInterface.upload(certType, location.hostname, data);
            message.success(__("上传成功"));
            manageLog({
                level: Level.INFO,
                opType: ManagementOps.SET,
                msg: __("本地上传${certType} 成功", {
                    certType: PartsText[certType],
                    signType: SignTypeText[signType],
                }),
                exMsg: "",
            });
            setTimeout(() => location.reload(), 2000);
        } catch (err) {
            clearTimeout(timer as any);
            Modal.error({
                title: __("错误"),
                content: getErrorMessage("deploy", err),
            });
        }
    }

    /**
     * 切换文件触发
     * @param text 文本
     */
    protected changeFile(text: { key: string; serverCert: string }) {
        this.setState({
            text,
            validateState: {
                key: "",
                serverCert: "",
            },
        });
    }
}

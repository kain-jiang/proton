import { certDownload, getCertInfo } from "../../api/deploy-manager";
import { Modal, message } from "antd";
import * as React from "react";
import getErrorMessage from "../../core/mediator/error";
import styles from "./styles.module.less";
import WebComponent from "../webcomponent";
import { PageState, Parts } from "./helper";
import { CertInfo, Props, State } from "./index.d";
import { manageLog } from "../../core/log2";
import __ from "./locale";
import { Level, ManagementOps } from "../../core/log";
import { AppType, accessAddr } from "../../api/deploy-manager";

export default class SiteConfigBase extends WebComponent<Props, State> {
    state = {
        certsInfo: [],
        documentConfigInfo: {
            host: "",
            port: "",
            path: "",
            type: "internal",
        },
        pageState: {
            [Parts.App]: PageState.Info,
            [Parts.Config]: PageState.Info,
        },
        currentAppPorts: {
            webClientHttps: "",
            objStorageHttps: "",
        },
        oldAppPorts: {
            webClientHttps: "",
            objStorageHttps: "",
        },
        isCertDownload: true,
    };

    componentDidMount() {
        this.getHTTPSInfo();
    }

    /**
     * 获取 https 配置信息
     */
    private async getHTTPSInfo() {
        let certsInfo = await getCertInfo.get();
        const documentConfigInfo = await accessAddr.get(AppType.App);
        const isCertDownload: boolean = (await certDownload.get()).status;
        let pageState = {
            [Parts.App]: PageState.Info,
        };
        certsInfo.length &&
            certsInfo.map((certInfo: CertInfo) => {
                if (!certInfo.accepter) {
                    pageState[certInfo.certType] === PageState.Edit;
                }
            });
        this.setState({
            certsInfo,
            pageState,
            documentConfigInfo,
            isCertDownload,
        });
    }

    /**
     * 是否允许下载证书
     * @param isCertDownload 是否允许下载证书
     */
    protected async createCertConfig(isCertDownload: boolean) {
        if (isCertDownload) {
            try {
                await certDownload.post();
                manageLog({
                    level: Level.INFO,
                    opType: ManagementOps.SET,
                    msg: __("开启证书下载入口成功"),
                    exMsg: "",
                });
                this.setState({ isCertDownload });
            } catch (err) {
                Modal.error({
                    title: __("错误"),
                    content: getErrorMessage("deploy", err),
                });
            }
        } else {
            Modal.confirm({
                title: __("确定要关闭网页端证书下载入口吗？"),
                content: (
                    <div>
                        <div className={styles["infotips"]}>
                            {__(
                                "关闭后，终端用户将无法在网页端登录页下载证书及相关文档进行安装。"
                            )}
                        </div>
                    </div>
                ),
                onOk: async () => {
                    try {
                        await certDownload.del();
                        manageLog({
                            level: Level.WARN,
                            opType: ManagementOps.SET,
                            msg: __("关闭证书下载入口成功"),
                            exMsg: "",
                        });
                        this.setState({ isCertDownload });
                    } catch (err) {
                        Modal.error({
                            title: __("错误"),
                            content: getErrorMessage("deploy", err),
                        });
                    }
                },
            });
        }
    }

    /**
     * 变更页面状态
     * @param pageState 页面状态
     */
    protected changePageState(pageState: any) {
        this.setState({ pageState });
    }

    /**
     * 当 AppConfig 组件中 webClient https/http 更改成功后,向父组件 SiteConfig 抛出的事件
     * @param webClientHttps 最新状态的 webClientHttps 值
     * @param oldWebClientHttps 最新一次保存成功的 webClientHttps 值
     */
    protected async changeAppConfigWebClientPorts(
        webClientHttps: number | string,
        oldWebClientHttps: number | string
    ) {
        const { currentAppPorts, oldAppPorts } = this.state;

        this.setState({
            currentAppPorts: {
                ...currentAppPorts,
                webClientHttps: webClientHttps,
            },
            oldAppPorts: {
                ...oldAppPorts,
                webClientHttps: oldWebClientHttps,
            },
        });
    }

    /**
     * 当 AppConfig 组件中 对象存储 https/http 更改成功后,向父组件 SiteConfig 抛出的事件
     * @param objStorageHttps 最新状态的 对象存储 Https 值
     * @param objStorageHttp 最新状态的 对象存储 Http 值
     * @param oldObjStorageHttps 最新一次保存成功的 对象存储 Https 值
     * @param oldObjStorageHttp 最新一次保存成功的 对象存储 Http 值
     */
    protected changeAppConfigObjPorts(
        objStorageHttps: number | string,
        oldObjStorageHttps: number | string
    ) {
        const { currentAppPorts, oldAppPorts } = this.state;

        this.setState({
            currentAppPorts: {
                ...currentAppPorts,
                objStorageHttps: objStorageHttps,
            },
            oldAppPorts: {
                ...oldAppPorts,
                objStorageHttps: oldObjStorageHttps,
            },
        });
    }
}

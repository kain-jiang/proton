import * as React from "react";
import styles from "./styles.module.less";
import CertCard from "./CertCard/component.view";
import AppConfig from "./AppConfig/component.view";
import CertSetter from "./CertSetter/component.view";
import ConfigCard from "./ConfigCard/component.view";
import { Button, Checkbox } from "antd";
import { PartsText, PageState, Parts, isSelfSignCert } from "./helper";
import SiteConfigBase from "./component.base";
import __ from "./locale";

export default class HTTPSConfig extends SiteConfigBase {
    render() {
        const { certsInfo, pageState, documentConfigInfo, isCertDownload } =
            this.state;

        if (certsInfo.length < 1) {
            return null;
        }

        return (
            <div className={styles["container"]}>
                <div className={styles["wrap"]}>
                    <div className={styles["title"]}>
                        <div className={styles["split"]}></div>
                        {PartsText[Parts.Config]}
                        <span className={styles["tips"]}>
                            {__(
                                "当前配置对所有产品的访问地址生效，包含AnyShare的文档域地址，AnyBackup，AnyRobot，AnyDATA，AnyFabric的访问地址。"
                            )}
                        </span>
                    </div>
                    {!pageState[Parts.Config] ? (
                        [
                            <ConfigCard
                                documentConfigInfo={documentConfigInfo}
                            />,
                            <div className={styles["set-cert"]}>
                                <Button
                                    onClick={() =>
                                        this.changePageState({
                                            ...pageState,
                                            [Parts.Config]: PageState.Edit,
                                        })
                                    }
                                >
                                    {__("修改访问配置")}
                                </Button>
                            </div>,
                        ]
                    ) : (
                        <AppConfig
                            ref="appConfig"
                            changeAppConfigWebClientPorts={this.changeAppConfigWebClientPorts.bind(
                                this
                            )}
                            changeAppConfigObjPorts={this.changeAppConfigObjPorts.bind(
                                this
                            )}
                            changePageState={(newState) =>
                                this.changePageState({
                                    ...pageState,
                                    ["config"]: newState,
                                })
                            }
                        />
                    )}
                </div>
                {certsInfo.length &&
                    certsInfo.map((certInfo) => {
                        const { certType } = certInfo;
                        const { issuer } = certInfo;
                        return (
                            <div className={styles["wrap"]}>
                                <div className={styles["title"]}>
                                    <div className={styles["split"]}></div>
                                    {PartsText[certType]}
                                </div>
                                {!pageState[certType] ? (
                                    [
                                        <CertCard certInfo={certInfo} />,
                                        <div className={styles["set-cert"]}>
                                            <Button
                                                onClick={() =>
                                                    this.changePageState({
                                                        ...pageState,
                                                        [certType]:
                                                            PageState.Edit,
                                                    })
                                                }
                                            >
                                                {__("配置证书")}
                                            </Button>
                                        </div>,
                                        <div
                                            className={styles["set-cert"]}
                                            style={{
                                                display: isSelfSignCert(issuer)
                                                    ? "block"
                                                    : "none",
                                            }}
                                        >
                                            <Checkbox
                                                className={styles["checkbox1"]}
                                                checked={isCertDownload}
                                                onChange={(e) => {
                                                    this.createCertConfig(
                                                        e.target.checked
                                                    );
                                                }}
                                            ></Checkbox>
                                            <span className={styles["radio"]}>
                                                {__(
                                                    "勾选后，用户可在登录页下载证书及相关文档进行安装。"
                                                )}
                                            </span>
                                        </div>,
                                    ]
                                ) : (
                                    <CertSetter
                                        certInfo={certInfo}
                                        changePageState={(cerType, newState) =>
                                            this.changePageState({
                                                ...pageState,
                                                [cerType]: newState,
                                            })
                                        }
                                    />
                                )}
                            </div>
                        );
                    })}
            </div>
        );
    }
}

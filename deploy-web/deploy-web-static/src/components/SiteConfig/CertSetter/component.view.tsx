import * as React from "react";
import className from "classnames";
import styles from "./styles.module.less";
import { Select, Button } from "antd";
import { SignType, SignTypeText, PageState } from "../helper";
import CertSetterBase from "./component.base";
import __ from "./locale";

const SignTypes = [SignType.Self, SignType.CA];

export default class CertSetter extends CertSetterBase {
    render() {
        const { changePageState, certInfo } = this.props;
        const { signType, text, validateState } = this.state;
        const { certType } = certInfo;

        return (
            <form className={styles["form"]}>
                <div className={styles["row"]} key={"1"}>
                    <div className={className(styles["cell"], styles["label"])}>
                        {__("选择证书类型：")}
                    </div>
                    <div className={className(styles["cell"], styles["field"])}>
                        <Select
                            style={{ width: 323 }}
                            value={signType}
                            onChange={(value) => {
                                this.changeSignType(value);
                            }}
                        >
                            {SignTypes.map((signType) => {
                                return (
                                    <Select.Option
                                        key={signType}
                                        value={signType}
                                    >
                                        {SignTypeText[signType]}
                                    </Select.Option>
                                );
                            })}
                        </Select>
                    </div>
                </div>
                {signType === SignType.Self ? (
                    <div className={styles["row"]} key={"2"}>
                        <div
                            className={className(
                                styles["cell"],
                                styles["label"]
                            )}
                        ></div>
                        <div
                            className={className(
                                styles["cell"],
                                styles["field"]
                            )}
                        >
                            <Button
                                className={styles["btn-space"]}
                                onClick={() =>
                                    this.addSelfSignedCert(
                                        certType,
                                        signType,
                                        certInfo
                                    )
                                }
                            >
                                {__("生成自签名证书")}
                            </Button>
                            <Button
                                className={styles["btn-width"]}
                                onClick={() =>
                                    changePageState(certType, PageState.Info)
                                }
                            >
                                {__("取消")}
                            </Button>
                        </div>
                    </div>
                ) : (
                    [
                        <div className={styles["row"]} key={"3"}>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["label"]
                                )}
                            >
                                {__("密钥：")}
                            </div>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["field"]
                                )}
                            >
                                <span className={styles["required"]}>*</span>
                                <div className={styles["file-wrap"]}>
                                    <span>{text.key}</span>
                                    <div className={styles["file-button"]}>
                                        <span>{__("浏览")}</span>
                                        <input
                                            accept={".key"}
                                            type={"file"}
                                            name={"keyNode"}
                                            className={styles["file-input"]}
                                            ref={(keyNode) =>
                                                (this.keyNode = keyNode)
                                            }
                                            onChange={() =>
                                                this.changeFile({
                                                    ...text,
                                                    key: this.keyNode.files[0]
                                                        .name,
                                                })
                                            }
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>,
                        <div className={styles["row"]} key={"4"}>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["label"]
                                )}
                            >
                                {__("服务器证书：")}
                            </div>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["field"]
                                )}
                            >
                                <span className={styles["required"]}>*</span>
                                <div className={styles["file-wrap"]}>
                                    <span>{text.serverCert}</span>
                                    <div className={styles["file-button"]}>
                                        <span>{__("浏览")}</span>
                                        <input
                                            accept={".cer,.crt,.cert"}
                                            type={"file"}
                                            name={"serverCertNode"}
                                            className={styles["file-input"]}
                                            ref={(serverCertNode) =>
                                                (this.serverCertNode =
                                                    serverCertNode)
                                            }
                                            onChange={() =>
                                                this.changeFile({
                                                    ...text,
                                                    serverCert:
                                                        this.serverCertNode
                                                            .files[0].name,
                                                })
                                            }
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>,
                        <div className={styles["row"]} key={"5"}>
                            <div className={styles["cell"]}></div>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["field"]
                                )}
                            >
                                <span className={styles["tips"]}>
                                    {validateState.key ||
                                        validateState.serverCert}
                                </span>
                            </div>
                        </div>,
                        <div className={styles["row"]} key={"6"}>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["label"]
                                )}
                            ></div>
                            <div
                                className={className(
                                    styles["cell"],
                                    styles["field"]
                                )}
                            >
                                <Button
                                    className={className(
                                        styles["btn-space"],
                                        styles["btn-width"]
                                    )}
                                    onClick={() =>
                                        this.uploadCert(
                                            certType,
                                            signType,
                                            certInfo
                                        )
                                    }
                                >
                                    {__("上传")}
                                </Button>
                                <Button
                                    className={styles["btn-width"]}
                                    onClick={() =>
                                        changePageState(
                                            certType,
                                            PageState.Info
                                        )
                                    }
                                >
                                    {__("取消")}
                                </Button>
                            </div>
                        </div>,
                    ]
                )}
            </form>
        );
    }
}

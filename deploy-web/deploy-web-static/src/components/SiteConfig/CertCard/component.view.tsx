import * as React from "react";
import styles from "./styles.module.less";
import className from "classnames";
import { Card, Form, Row, Col } from "antd";
import { SignTypeText, SignType, isSelfSignCert } from "../helper";
import { Props } from "./index.d";
import __ from "./locale";

const CertCard: React.FunctionComponent<Props> = React.memo(({ certInfo }) => {
    const { accepter, certType, expireDate, hasExpired, issuer, startDate } =
        certInfo;

    if (!certInfo) {
        return null;
    }

    return (
        <>
            {hasExpired ? (
                <div className={styles["expired-title"]}>
                    {__("证书已过期")}
                </div>
            ) : null}
            <Card
                style={{ width: 363, minHeight: 159 }}
                className={className(styles["card"], {
                    [styles["sign-self-gray"]]:
                        isSelfSignCert(issuer) && hasExpired,
                    [styles["sign-ca-gray"]]:
                        !isSelfSignCert(issuer) && hasExpired,
                    [styles["sign-self"]]:
                        isSelfSignCert(issuer) && !hasExpired,
                    [styles["sign-ca"]]: !isSelfSignCert(issuer) && !hasExpired,
                    [styles["expired"]]: hasExpired,
                })}
            >
                <div className={styles["cert-type"]}>
                    <div className={className(styles["cell"])}>
                        {__("证书类型：") + "\u00a0"}
                    </div>
                    <div className={className(styles["cell"])}>
                        {
                            SignTypeText[
                                isSelfSignCert(issuer)
                                    ? SignType.Self
                                    : SignType.CA
                            ]
                        }
                    </div>
                </div>
                <Form>
                    <Form.Item
                        label={__("颁发者：")}
                        className={styles["card-form-item"]}
                    >
                        <span title={issuer} className={styles["text"]}>
                            {issuer}
                        </span>
                    </Form.Item>
                    <Form.Item
                        label={__("颁发给：")}
                        className={styles["card-form-item"]}
                    >
                        <span title={accepter} className={styles["text"]}>
                            {accepter}
                        </span>
                    </Form.Item>
                    <Form.Item
                        label={__("有效期：")}
                        className={styles["card-form-item"]}
                    >
                        <span className={styles["text"]}>
                            {`${startDate}-${expireDate}`}
                        </span>
                    </Form.Item>
                </Form>
            </Card>
        </>
    );
});

export default CertCard;

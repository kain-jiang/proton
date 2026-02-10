import * as React from "react";
import { Props } from "./index.d";
import styles from "./styles.module.less";
import className from "classnames";
import { Card, Form } from "antd";
import __ from "./locale";

const ConfigCard: React.FunctionComponent<Props> = React.memo(
    ({ documentConfigInfo }) => {
        const { host, port, path, type } = documentConfigInfo;

        if (!documentConfigInfo) {
            return null;
        }

        return (
            <Card
                style={{ width: 363, minHeight: 159 }}
                className={className(styles["card"], styles["config"])}
            >
                <Form>
                    <Form.Item
                        label={__("访问地址：")}
                        className={styles["card-form-item"]}
                    >
                        <span title={host} className={styles["text"]}>
                            {host}
                        </span>
                    </Form.Item>
                    <Form.Item
                        label={__("HTTPS端口：")}
                        className={styles["card-form-item"]}
                    >
                        <span title={port} className={styles["text"]}>
                            {port}
                        </span>
                    </Form.Item>
                    <Form.Item
                        label={__("访问前缀：")}
                        className={styles["card-form-item"]}
                    >
                        <span title={path} className={styles["text"]}>
                            {path}
                        </span>
                    </Form.Item>
                    <Form.Item
                        label={__("访问地址类型：")}
                        className={styles["card-form-item"]}
                    >
                        <span title={type} className={styles["text"]}>
                            {type}
                        </span>
                    </Form.Item>
                </Form>
            </Card>
        );
    }
);

export default ConfigCard;

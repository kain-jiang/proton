import React, { FC } from "react";
import { Table } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import {
    ApplicationItem,
    ServiceSchemaItem,
} from "../../../../../api/suite-management/suite-deploy/declare";
import styles from "./styles.module.less";
import __ from "./locale";

interface IProps {
    // 选中的服务信息
    serviceInfo: ApplicationItem;
    // 套件配置项信息
    suiteConfig: ServiceSchemaItem[];
}

export const ConfirmConfig: FC<IProps> = ({ serviceInfo, suiteConfig }) => {
    // 表格的列配置项
    const suiteColumns: TableColumnsType<ApplicationItem> = [
        {
            title: __("名称"),
            dataIndex: "title",
        },
        {
            title: __("版本"),
            dataIndex: "mversion",
        },
    ];

    // 表格的列配置项
    const serviceColumns = [
        {
            title: __("服务名称"),
            dataIndex: "title",
        },
        {
            title: __("版本"),
            dataIndex: "version",
        },
    ];

    const windowHeight = window.innerHeight > 720 ? "100vh" : "720px";

    return (
        <React.Fragment>
            <div className={styles["config-title"]}>{__("套件信息")}</div>
            <Table
                rowKey="name"
                className="service-table"
                dataSource={[{ ...serviceInfo }]}
                columns={suiteColumns}
                pagination={{ pageSize: 10, hideOnSinglePage: true }}
            />
            <div className={styles["config-title"]}>{__("部署服务信息")}</div>
            <Table
                rowKey="name"
                style={{ width: "650px" }}
                dataSource={suiteConfig}
                columns={serviceColumns}
                scroll={{
                    y: `calc(${windowHeight} - 504px)`,
                }}
                pagination={{
                    showQuickJumper: false,
                    showSizeChanger: true,
                    showTotal: (total) => __("共${total}条", { total }),
                }}
            />
        </React.Fragment>
    );
};

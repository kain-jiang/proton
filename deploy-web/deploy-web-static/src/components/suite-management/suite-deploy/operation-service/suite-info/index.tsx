import React, { FC } from "react";
import { Table, Row, Col } from "@kweaver-ai/ui";
import styles from "./styles.module.less";
import __ from "./locale";
import { Dot } from "../../../../common/components/text/dot";
import { safetyStr } from "../../../../common/utils/string";
import {
    AppsUploadStatusEnum,
    ComponentsInstallStatusEnum,
    appsUploadStatus,
    componentsInstallStatus,
} from "../../type.d";
import { ConnectInfoServices, componentsText } from "../type.d";

export interface ComponentsData {
    type: ConnectInfoServices;
    status: ComponentsInstallStatusEnum;
}

export interface AppsData {
    title: string;
    status: AppsUploadStatusEnum;
    version: string;
}

interface IProps {
    componentsData: ComponentsData[];
    appsData: AppsData[];
}

export const SuiteInfo: FC<IProps> = ({ componentsData, appsData }) => {
    const serviceColumns = [
        {
            title: __("服务名称"),
            dataIndex: "title",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value),
        },
        {
            title: __("上传状态"),
            dataIndex: "status",
            render: (value: AppsUploadStatusEnum) => {
                return (
                    <Dot color={appsUploadStatus[value].color}>
                        {appsUploadStatus[value].text}
                    </Dot>
                );
            },
            tooltip: (value: AppsUploadStatusEnum) =>
                appsUploadStatus[value].text,
        },
        {
            title: __("版本"),
            dataIndex: "version",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value),
        },
    ];

    const componentColumns = [
        {
            title: __("组件名称"),
            dataIndex: "type",
            render: (value: ConnectInfoServices) =>
                safetyStr(componentsText[value]),
            tooltip: (value: ConnectInfoServices) =>
                safetyStr(componentsText[value]),
        },
        {
            title: __("安装状态"),
            dataIndex: "status",
            render: (value: ComponentsInstallStatusEnum) => {
                return (
                    <Dot color={componentsInstallStatus[value].color}>
                        {componentsInstallStatus[value].text}
                    </Dot>
                );
            },
            tooltip: (value: ComponentsInstallStatusEnum) =>
                componentsInstallStatus[value].text,
        },
    ];

    const windowHeight = window.innerHeight > 720 ? "100vh" : "720px";

    return (
        <div className={styles["suite-info-container"]}>
            <div style={{ marginBottom: "14px" }}>{__("套件信息")}</div>
            <Row gutter={24}>
                <Col span={14}>
                    <Table
                        dataSource={appsData}
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
                </Col>
                <Col span={10}>
                    <Table
                        dataSource={componentsData}
                        columns={componentColumns}
                        scroll={{
                            y: `calc(${windowHeight} - 519px)`,
                        }}
                        pagination={{
                            pageSize: 1000,
                            hideOnSinglePage: true,
                        }}
                    />
                </Col>
            </Row>
        </div>
    );
};

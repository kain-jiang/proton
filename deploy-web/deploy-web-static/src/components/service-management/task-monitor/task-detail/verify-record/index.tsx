import React, { FC, useMemo, useState } from "react";
import { ContentLayout, Toolbar, Text } from "../../../../common/components";
import { VerifyDetail } from "./verify-detail";
import { Refresh, Button, Table, Space } from "@kweaver-ai/ui";
import { DownloadOutlined } from "@kweaver-ai/ui/icons";
import type { TableColumnsType } from "@kweaver-ai/ui";
import className from "classnames";
import {
    DataSchemaVerifyItem,
    FuncVerifyItem,
} from "../../../../../api/service-management/task-monitor/declare";
import { SERVICE_PREFIX } from "../../../config";
import { safetyStr } from "../../../../common/utils/string";
import { ActiveEnum, VerifyResultEnum, verifyResultStatus } from "./type.d";
import styles from "./styles.module.less";
import __ from "./locale";
import { formatDateTime } from "../../../utils/timer";
import { noop } from "lodash";

interface IProps {
    // 数据验证列表
    dataSchemaVerifyList: DataSchemaVerifyItem[];
    // 功能验证列表
    funcVerifyList: FuncVerifyItem[];
    // 刷新事件回调
    changeRefresh: (func: (refresh: boolean) => boolean) => void;
}

export const VerifyRecord: FC<IProps> = ({
    dataSchemaVerifyList,
    funcVerifyList,
    changeRefresh,
}) => {
    // 选中验证记录类型(功能验证和数据验证同时存在时有效)
    const [active, setActive] = useState<ActiveEnum>(ActiveEnum.FUNCVERIFY);
    // 是否展示滑窗
    const [open, setOpen] = useState<boolean>(false);
    // 验证记录id
    const [verifyId, setVerifyId] = useState<number>(0);

    // 当前展示表格验证记录的类型
    const verifyType = useMemo(() => {
        if (!funcVerifyList.length) {
            return ActiveEnum.DATASCHEMAVERIFY;
        } else if (!dataSchemaVerifyList.length) {
            return ActiveEnum.FUNCVERIFY;
        } else {
            return active;
        }
    }, [active, dataSchemaVerifyList, funcVerifyList]);

    // 表格的列配置项
    const columns: TableColumnsType<DataSchemaVerifyItem | FuncVerifyItem> = [
        {
            title: __("ID"),
            render: (_, record: any) => {
                return verifyType === ActiveEnum.DATASCHEMAVERIFY
                    ? record.did
                    : record.fid;
            },
            tooltip: (_, record: any) => {
                return verifyType === ActiveEnum.DATASCHEMAVERIFY
                    ? record.did
                    : record.fid;
            },
        },
        {
            title: __("验证结果"),
            dataIndex: "verifyResult",
            filters: Object.values(verifyResultStatus),
            onFilter: (value, record) => record.verifyResult === value,
            render: (value: string) => {
                return (
                    <Text textColor={verifyResultStatus[value].color}>
                        {verifyResultStatus[value].text}
                    </Text>
                );
            },
            tooltip: (value: string) => verifyResultStatus[value].text,
        },
        {
            title: __("结束时间"),
            dataIndex: "verifyEndTime",
            render: (value: string, record: any) => {
                if (record.verifyResult === VerifyResultEnum.PASS) {
                    return safetyStr(value, formatDateTime);
                } else {
                    // 失败的记录提供详情按钮
                    return (
                        <div className={styles["detail-td"]}>
                            <span className={styles["detail-span"]}>
                                {safetyStr(value, formatDateTime)}
                            </span>
                            <Button
                                type="link"
                                onClick={handleShowDetail(
                                    verifyType === ActiveEnum.FUNCVERIFY
                                        ? record.fid
                                        : record.did
                                )}
                            >
                                {__("详情")}
                            </Button>
                        </div>
                    );
                }
            },
            tooltip: (value: string) => safetyStr(value, formatDateTime),
        },
    ];

    // 点击详情
    const handleShowDetail = (id: number) => {
        return () => {
            setOpen(true);
            setVerifyId(id);
        };
    };

    const header = (
        <Toolbar
            left={
                <React.Fragment>
                    {dataSchemaVerifyList.length && funcVerifyList.length ? (
                        <Space size="middle" className={styles["tab-space"]}>
                            <Button
                                type="link"
                                className={
                                    active === ActiveEnum.DATASCHEMAVERIFY
                                        ? className(
                                              styles["unchecked-btn"],
                                              styles["skin-color-hover"]
                                          )
                                        : styles["skin-color"]
                                }
                                onClick={() => setActive(ActiveEnum.FUNCVERIFY)}
                            >
                                {__("功能验证")}
                            </Button>
                            <span>|</span>
                            <Button
                                type="link"
                                className={
                                    active === ActiveEnum.FUNCVERIFY
                                        ? className(
                                              styles["unchecked-btn"],
                                              styles["skin-color-hover"]
                                          )
                                        : styles["skin-color"]
                                }
                                onClick={() =>
                                    setActive(ActiveEnum.DATASCHEMAVERIFY)
                                }
                            >
                                {__("数据验证")}
                            </Button>
                        </Space>
                    ) : null}
                </React.Fragment>
            }
            right={
                <React.Fragment>
                    <DownloadOutlined
                        onPointerEnterCapture={noop}
                        onPointerLeaveCapture={noop}
                    />
                    <Refresh
                        onClick={() => changeRefresh((refresh) => !refresh)}
                    />
                </React.Fragment>
            }
            leftSize={12}
            rightSize={12}
            moduleName={SERVICE_PREFIX}
        />
    );
    return (
        <React.Fragment>
            <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
                <Table
                    key={verifyType}
                    columns={columns}
                    dataSource={
                        verifyType === ActiveEnum.FUNCVERIFY
                            ? funcVerifyList
                            : dataSchemaVerifyList
                    }
                    scroll={{
                        y: "calc(100vh - 300px)",
                    }}
                    pagination={{
                        showQuickJumper: true,
                        showSizeChanger: true,
                        showTotal: (total) => __("共${total}条", { total }),
                        pageSizeOptions: [5, 15, 30, 50, 100],
                        defaultPageSize: 15,
                    }}
                />
            </ContentLayout>
            {open && (
                <VerifyDetail
                    verifyId={verifyId}
                    open={open}
                    verifyType={verifyType}
                    onCancel={() => setOpen(false)}
                />
            )}
        </React.Fragment>
    );
};

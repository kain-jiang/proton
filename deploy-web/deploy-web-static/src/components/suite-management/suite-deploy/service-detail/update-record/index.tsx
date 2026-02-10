import React, { FC } from "react";
import { Table, Button, Modal } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { ReactComponent as UpdateRecordIcon } from "../../../assets/UpdateRecordIcon.svg";
import { formatTableResponse } from "../../../../common/utils/request";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import {
    ApplicationItem,
    IGetJobParams,
    IGetJobTableParams,
    SuiteItem,
} from "../../../../../api/suite-management/suite-deploy/declare";
import { IBaseProps } from "../../declare";
import { safetyTime } from "../../../utils/timer";
import styles from "./styles.module.less";
import __ from "./locale";
import { composeJob } from "../../../../../api/suite-management/suite-deploy";
import { safetyStr } from "../../../../common/utils/string";
import { JobType } from "../../../../../core/suite-management/suite-deploy";

interface IProps extends IBaseProps {
    // 服务id
    serviceId: number;
    // 服务名称
    serviceName: string;
    // 修改服务id
    changeServiceId: (id: number) => void;
    // 修改更新服务的信息
    changeUpdateServiceRecord: (record: ApplicationItem) => void;
    // 修改更新服务的jid
    changeJid: (jid: number) => void;
}
export const UpdateRecord: FC<IProps> = ({
    changeServiceMode,
    changeServiceId,
    changeJid,
    changeUpdateServiceRecord,
    serviceId,
    serviceName,
}) => {
    // 表格的基本设置
    const { state } = Table.useTable<
        SuiteItem,
        IGetJobTableParams,
        SuiteItem[]
    >({
        request: (params) => {
            const { current, pageSize } = params;
            const formatedParams: IGetJobParams = {
                offset: (current - 1) * pageSize,
                limit: pageSize,
                name: serviceName,
                type: JobType.Suite,
            };
            return composeJob.getJobList(formatedParams);
        },
        rowKey: "jid",
        pagination: {
            showTotal: (total) => __("共${total}条", { total }),
            pageSizeOptions: [10, 20, 50, 100],
            pageSize: 10,
        },
        ...formatTableResponse(),
    });

    // 表格的列配置项
    const columns: TableColumnsType<SuiteItem> = [
        {
            title: __("版本"),
            dataIndex: "mversion",
            render: (value: string, record: SuiteItem) => (
                <Button type="link" onClick={() => handleDetailClick(record)}>
                    {safetyStr(value)}
                </Button>
            ),
            tooltip: (value: string) => safetyStr(value),
        },
        {
            title: __("更新时间"),
            dataIndex: "startTime",
            render: (value: number) => safetyTime(value),
            tooltip: (value: number) => safetyTime(value),
        },
        {
            title: __("ID"),
            dataIndex: "jid",
            render: (value: number) => value,
            tooltip: (value: number) => value,
        },
        {
            title: __("操作"),
            render: (_, record: SuiteItem) => {
                return (
                    <Button
                        type="link"
                        icon={<UpdateRecordIcon />}
                        onClick={() => handleUpdateClick(record)}
                        disabled={record.jid === serviceId}
                    />
                );
            },
            tooltip: () => __("更新"),
        },
    ];
    /**
     * @description 查看历史版本
     * @param record 选中行表格数据信息
     */
    const handleDetailClick = (record: SuiteItem) => {
        changeServiceMode(ServiceMode.Service);
        changeServiceId(record.jid);
    };
    /**
     * @description 回退版本
     * @param record 选中行表格数据信息
     */
    const handleUpdateClick = (record: SuiteItem) => {
        Modal.confirm({
            title: __("您将安装此版本"),
            content: (
                <div className={styles["modal-content"]}>
                    {__("安装此版本将覆盖当前版本，是否继续安装")}
                </div>
            ),
            onOk: () => {
                // 跳转到更新页面
                changeServiceMode(ServiceMode.Revert);
                // 携带record参数
                changeUpdateServiceRecord({
                    mname: record.jname,
                    title: record.title!,
                    mversion: record.mversion,
                });
                // 携带任务id
                changeJid(record.jid);
            },
        });
    };
    return <Table {...state} columns={columns} />;
};

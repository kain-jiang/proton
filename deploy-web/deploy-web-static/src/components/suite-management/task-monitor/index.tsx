import React, { FC, useState } from "react";
import { ContentLayout, Toolbar, Text } from "../../common/components";
import { TaskDetail } from "./task-detail";
import { ChangeConfig } from "./change-config";
import { handleError } from "../utils/handleError";
import {
    IGetJobTableParams,
    IGetJobParams,
    SuiteItem,
} from "../../../api/suite-management/suite-deploy/declare";
import {
    taskConfigStatus,
    taskCategoryStatus,
    taskCategoryStatusItems,
    TaskCategoryStatusEnum,
} from "./type.d";
import { Button, Table, Space, Search, Refresh } from "@kweaver-ai/ui";
import type { TableColumnType, TableColumnsType } from "@kweaver-ai/ui";
import { formatTableResponse } from "../../common/utils/request";
import { safetyRunningTime, safetyTime } from "../utils/timer";
import { SERVICE_PREFIX } from "../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { composeJob } from "../../../api/suite-management/suite-deploy";
import { safetyStr } from "../../common/utils/string";
import { JobType } from "../../../core/suite-management/suite-deploy";

interface IProps {
    jobType: JobType;
}

export const SuiteTaskMonitor: FC<IProps> = ({ jobType }) => {
    // 输入框内容
    const [filter, setFilter] = useState<string>("");
    // 是否展示任务详情滑窗
    const [openDetail, setOpenDetail] = useState<boolean>(false);
    // 是否展示更改配置项滑窗
    const [openChangeConfig, setOpenChangeConfig] = useState<boolean>(false);
    // 任务id
    const [jid, setJid] = useState<number>(0);
    // 系统空间filter筛选项开关
    const [filterOpen, setFilterOpen] = useState(false);

    // 表格的基本设置
    const { state, api, data } = Table.useTable<
        SuiteItem,
        IGetJobTableParams,
        SuiteItem[]
    >({
        request: (params) => {
            const { _filter, current, pageSize, title } = params;
            const formatedParams: IGetJobParams = {
                offset: (current - 1) * pageSize,
                limit: pageSize,
                status: getFilterParams(
                    (_filter as any)?.status,
                    Object.values(taskCategoryStatus).length
                ),
                sid: [-1, undefined].includes((_filter as any)?.systemName?.[0])
                    ? undefined
                    : (_filter as any)?.systemName?.[0],
                title,
                type: jobType,
            };
            return composeJob.getJobList(formatedParams);
        },
        rowKey: "jid",
        rowSelection: {
            type: "checkbox",
        },
        pagination: {
            showTotal: (total) => __("共${total}条", { total }),
        },
        ...formatTableResponse(),
    });

    const { reload, setParams } = api;
    const { selectedRows } = data;

    const getFilterParams = (
        value: Array<any> | undefined,
        filterOptionCount: number
    ) => {
        if (value?.length) {
            if (value.length === filterOptionCount) {
                return undefined;
            } else {
                return value?.reduce((pre, val) => {
                    return pre.concat(taskCategoryStatusItems[val]);
                }, []);
            }
        }
        return undefined;
    };

    // 表格的列配置项
    const columns = [
        {
            title: __("状态"),
            dataIndex: "status",
            filters: Object.values(taskCategoryStatus),
            render: (value: number) => {
                return (
                    <Text textColor={taskConfigStatus[value].color}>
                        {taskConfigStatus[value].categoryText}
                    </Text>
                );
            },
            tooltip: (value: number) => {
                return taskConfigStatus[value].categoryText;
            },
        },
        {
            title: __("名称"),
            dataIndex: jobType === JobType.Suite ? "title" : "jname",
            render: (value: string, record: SuiteItem) => (
                <Button type="link" onClick={() => handleDetailClick(record)}>
                    {safetyStr(value)}
                </Button>
            ),
            tooltip: (value: string) => safetyStr(value),
        },
        ...(jobType === JobType.Suite
            ? [
                  {
                      title: __("标识"),
                      dataIndex: "jname",
                      render: (value: string) => safetyStr(value),
                      tooltip: (value: string) => safetyStr(value),
                  },
                  {
                      title: __("版本"),
                      dataIndex: "mversion",
                      render: (value: string) => safetyStr(value),
                      tooltip: (value: string) => safetyStr(value),
                  },
              ]
            : [
                  {
                      title: __("系统空间"),
                      dataIndex: "systemName",
                      render: (value: string, record: SuiteItem) =>
                          safetyStr(record.systemName!),
                      tooltip: (value: string, record: SuiteItem) =>
                          safetyStr(record.systemName!),
                  },
                  {
                      title: __("系统空间ID"),
                      render: (value: number, record: SuiteItem) =>
                          safetyStr(record.sid!.toString()),
                      tooltip: (value: number, record: SuiteItem) =>
                          safetyStr(record.sid!.toString()),
                  },
              ]),
        {
            title: __("类型"),
            dataIndex: "description",
            width: 100,
            render: (value: string) => {
                return value === "安装"
                    ? __("安装")
                    : value === "更新"
                    ? __("更新")
                    : "---";
            },
            tooltip: (value: string) => {
                return value === "安装"
                    ? __("安装")
                    : value === "更新"
                    ? __("更新")
                    : "---";
            },
        },
        {
            title: __("ID"),
            dataIndex: "jid",
            width: 70,
        },
        {
            title: __("开始时间"),
            dataIndex: "startTime",
            render: (value: number) => safetyTime(value),
            tooltip: (value: number) => safetyTime(value),
        },
        {
            title: __("结束时间"),
            dataIndex: "endTime",
            render: (value: number) => safetyTime(value),
            tooltip: (value: number) => safetyTime(value),
        },
        {
            title: __("运行时间"),
            render: (_: SuiteItem, record: SuiteItem) =>
                safetyRunningTime(record.endTime!, record.startTime!),
            tooltip: (_: SuiteItem, record: SuiteItem) =>
                safetyRunningTime(record.endTime!, record.startTime!),
        },
        {
            title: __("操作"),
            width: 300,
            render: (_: SuiteItem, record: SuiteItem) => {
                return (
                    <Space>
                        <Button
                            type="link"
                            onClick={() => handlePauseClick(record)}
                            disabled={
                                !taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.RUNNING
                                ].includes(record.status) &&
                                !taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.STOPPED
                                ].includes(record.status) &&
                                !taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.CONFIGCONFIRMED
                                ].includes(record.status)
                            }
                        >
                            {__("暂停/启动")}
                        </Button>
                        <Button
                            type="link"
                            onClick={() => handleRetryClick(record)}
                            disabled={
                                !taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.FAILED
                                ].includes(record.status)
                            }
                        >
                            {__("失败重试")}
                        </Button>
                        <Button
                            type="link"
                            onClick={() => handleChangeClick(record)}
                            disabled={
                                taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.RUNNING
                                ].includes(record.status) ||
                                taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.SUCCEEDED
                                ].includes(record.status)
                            }
                        >
                            {__("更改配置")}
                        </Button>
                    </Space>
                );
            },
            tooltip: () => undefined,
        },
    ].filter((item) => item);
    /**
     * @description 处理暂停/启动
     * @param {SuiteItem} record 选中数据信息
     */
    const handlePauseClick = async (record?: SuiteItem) => {
        const recordData = record ? record : selectedRows[0];
        try {
            if (
                !taskCategoryStatusItems[
                    TaskCategoryStatusEnum.RUNNING
                ].includes(recordData.status)
            ) {
                await composeJob.startJob(recordData.jid);
                reload();
            } else {
                await composeJob.pauseJob(recordData.jid);
                reload();
            }
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @description 处理失败重试
     * @param {SuiteItem} record 选中数据信息
     */
    const handleRetryClick = async (record?: SuiteItem) => {
        const recordData = record ? record : selectedRows[0];
        try {
            await composeJob.startJob(recordData.jid);
            reload();
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @description 处理更改配置
     * @param {SuiteItem} record 选中数据信息
     */
    const handleChangeClick = (record?: SuiteItem) => {
        const recordData = record ? record : selectedRows[0];
        setJid(recordData.jid);
        setOpenChangeConfig(true);
    };

    /**
     * @description 处理查看详情
     * @param {SuiteItem} record 选中数据信息
     */
    const handleDetailClick = (record: SuiteItem) => {
        setJid(record.jid);
        setOpenDetail(true);
    };
    /**
     * @description 关闭更改配置项滑窗
     */
    const onChangeConfigCancel = () => {
        setOpenChangeConfig(false);
        reload();
    };
    /**
     * @description 关闭详情滑窗
     */
    const onDetailCancel = () => {
        setOpenDetail(false);
    };
    /**
     * @description 更新输入框数据
     * @param e 输入框change事件
     */
    const onFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFilter(e.target.value);
        setParams({
            title: e.target.value,
        });
    };

    const header = (
        <Toolbar
            left={
                <Space>
                    {/* 暂停/启动 */}
                    <Button
                        type="default"
                        onClick={() => handlePauseClick()}
                        disabled={
                            selectedRows.length !== 1 ||
                            (!taskCategoryStatusItems[
                                TaskCategoryStatusEnum.RUNNING
                            ].includes(selectedRows[0].status) &&
                                !taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.STOPPED
                                ].includes(selectedRows[0].status) &&
                                !taskCategoryStatusItems[
                                    TaskCategoryStatusEnum.CONFIGCONFIRMED
                                ].includes(selectedRows[0].status))
                        }
                    >
                        {__("暂停/启动")}
                    </Button>
                    {/* 失败重试 */}
                    <Button
                        type="default"
                        onClick={() => handleRetryClick()}
                        disabled={
                            selectedRows.length !== 1 ||
                            !taskCategoryStatusItems[
                                TaskCategoryStatusEnum.FAILED
                            ].includes(selectedRows[0].status)
                        }
                    >
                        {__("失败重试")}
                    </Button>
                    {/* 更改配置 */}
                    <Button
                        type="default"
                        onClick={() => handleChangeClick()}
                        disabled={
                            selectedRows.length !== 1 ||
                            taskCategoryStatusItems[
                                TaskCategoryStatusEnum.RUNNING
                            ].includes(selectedRows[0].status) ||
                            taskCategoryStatusItems[
                                TaskCategoryStatusEnum.SUCCEEDED
                            ].includes(selectedRows[0].status)
                        }
                    >
                        {__("更改配置")}
                    </Button>
                </Space>
            }
            right={
                <React.Fragment>
                    <Search
                        placeholder={__("搜索任务名称")}
                        value={filter}
                        onChange={onFilterChange}
                        className={styles["task-input"]}
                        debounce
                    />
                    <Refresh onClick={() => reload()} />
                </React.Fragment>
            }
            cols={[{ span: 12 }, { span: 12 }]}
            moduleName={SERVICE_PREFIX}
        />
    );

    return (
        <React.Fragment>
            <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
                <Table
                    {...state}
                    columns={columns as TableColumnsType<SuiteItem>}
                />
            </ContentLayout>
            {openDetail && (
                <TaskDetail
                    open={openDetail}
                    onCancel={onDetailCancel}
                    jid={jid}
                />
            )}
            {openChangeConfig && (
                <ChangeConfig
                    open={openChangeConfig}
                    onCancel={onChangeConfigCancel}
                    jid={jid}
                />
            )}
        </React.Fragment>
    );
};

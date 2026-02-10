import React, { FC, useState } from "react";
import { ContentLayout, Toolbar, Text } from "../../common/components";
import { TaskDetail } from "./task-detail";
import { ChangeConfig } from "./change-config";
import { handleError } from "../utils/handleError";
import { serviceJob } from "../../../api/service-management/service-deploy";
import {
  IGetJobTableParams,
  JobItem,
  IGetJobParams,
} from "../../../api/service-management/service-deploy/declare";
import {
  taskConfigStatus,
  taskCategoryStatus,
  taskCategoryStatusItems,
  TaskCategoryStatusEnum,
  jobOperateTypeStatus,
  JobOperateType,
} from "./type.d";
import { Button, Table, Space, Search, Refresh } from "@kweaver-ai/ui";
import type { TableColumnType, TableColumnsType } from "@kweaver-ai/ui";
import { formatTableResponse } from "../../common/utils/request";
import { safetyRunningTime, safetyTime } from "../utils/timer";
import { SERVICE_PREFIX } from "../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { safetyStr } from "../../common/utils/string";

export const TaskMonitor: FC = () => {
  // 输入框内容
  const [filter, setFilter] = useState<string>("");
  // 是否展示任务详情滑窗
  const [openDetail, setOpenDetail] = useState<boolean>(false);
  // 是否展示更改配置项滑窗
  const [openChangeConfig, setOpenChangeConfig] = useState<boolean>(false);
  // 任务id
  const [jid, setJid] = useState<number>(0);

  // 表格的基本设置
  const { state, api, data } = Table.useTable<
    JobItem,
    IGetJobTableParams,
    JobItem[]
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
        jtype:
          (_filter as any)?.operateType?.length &&
          (_filter as any)?.operateType?.length !==
            Object.values(jobOperateTypeStatus).length
            ? (_filter as any)?.operateType
            : undefined,
        sid: [-1, undefined].includes((_filter as any)?.systemName?.[0])
          ? undefined
          : (_filter as any)?.systemName?.[0],
        title,
      };
      return serviceJob.get(formatedParams);
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
  const columns: TableColumnsType<JobItem> = [
    {
      title: __("状态"),
      dataIndex: "status",
      filters: Object.values(taskCategoryStatus),
      render: (_, record: JobItem) => {
        return (
          <Text textColor={taskConfigStatus[record.target.status].color}>
            {taskConfigStatus[record.target.status].categoryText}
          </Text>
        );
      },
      tooltip: (_, record: JobItem) => {
        return taskConfigStatus[record.target.status].categoryText;
      },
    },
    {
      title: __("名称"),
      render: (value: string, record: JobItem) => (
        <Button type="link" onClick={() => handleDetailClick(record)}>
          {record.target.application.title}
        </Button>
      ),
      tooltip: (value: string, record: JobItem) =>
        record.target.application.title,
    },
    {
      title: __("标识"),
      render: (value: string, record: JobItem) =>
        record.target.application.name,
      tooltip: (value: string, record: JobItem) =>
        record.target.application.name,
    },
    {
      title: __("版本"),
      render: (value: string, record: JobItem) =>
        record.target.application.version,
      tooltip: (value: string, record: JobItem) =>
        record.target.application.version,
    },
    {
      title: __("系统空间"),
      dataIndex: "systemName",
      render: (value: string, record: JobItem) =>
        safetyStr(record.target.systemName!),
      tooltip: (value: string, record: JobItem) =>
        safetyStr(record.target.systemName!),
    },
    {
      title: __("系统空间ID"),
      render: (value, record: JobItem) =>
        safetyStr(record.target.sid!.toString()),
      tooltip: (value, record: JobItem) =>
        safetyStr(record.target.sid!.toString()),
    },
    Table.getOperation<JobItem>((record) => ({
      menu: {
        onClick: (e) => {
          if (e.key === "pause") {
            handlePauseClick(record);
          } else if (e.key === "retry") {
            handleRetryClick(record);
          } else if (e.key === "change") {
            handleChangeClick(record);
          } else if (e.key === "delete") {
            handleDeleteClick(record);
          }
        },
        items: [
          {
            key: "pause",
            label: __("暂停/启动"),
            disabled:
              !taskCategoryStatusItems[TaskCategoryStatusEnum.RUNNING].includes(
                record.target.status
              ) &&
              !taskCategoryStatusItems[TaskCategoryStatusEnum.STOPPED].includes(
                record.target.status
              ) &&
              !taskCategoryStatusItems[
                TaskCategoryStatusEnum.CONFIGCONFIRMED
              ].includes(record.target.status),
          },
          {
            key: "retry",
            label: __("失败重试"),
            disabled: !taskCategoryStatusItems[
              TaskCategoryStatusEnum.FAILED
            ].includes(record.target.status),
          },
          {
            key: "change",
            label: __("更改配置"),
            disabled:
              record?.target?.operateType === JobOperateType.Uninstall ||
              taskCategoryStatusItems[TaskCategoryStatusEnum.RUNNING].includes(
                record.target.status
              ),
          },
          {
            key: "delete",
            label: __("删除任务"),
            disabled: true,
          },
        ],
      },
    })),
    {
      title: __("类型"),
      dataIndex: "operateType",
      filters: Object.values(jobOperateTypeStatus),
      width: 100,
      render: (_, record: JobItem) => {
        return jobOperateTypeStatus[record.target.operateType].text;
      },
      tooltip: (_, record: JobItem) => {
        return jobOperateTypeStatus[record.target.operateType].text;
      },
    },
    {
      title: __("ID"),
      dataIndex: "jid",
      width: 70,
    },
    {
      title: __("开始时间"),
      render: (_, record: JobItem) => safetyTime(record.target.startTime),
      tooltip: (_, record: JobItem) => safetyTime(record.target.startTime),
    },
    {
      title: __("结束时间"),
      render: (_, record: JobItem) => safetyTime(record.target.endTime),
      tooltip: (_, record: JobItem) => safetyTime(record.target.endTime),
    },
    {
      title: __("运行时间"),
      render: (_, record: JobItem) =>
        safetyRunningTime(record.target.endTime, record.target.startTime),
      tooltip: (_, record: JobItem) =>
        safetyRunningTime(record.target.endTime, record.target.startTime),
    },
    {
      title: __("备注"),
      render: (_, record: JobItem) => safetyStr(record.target.comment),
      tooltip: (_, record: JobItem) => safetyStr(record.target.comment),
    },
  ];
  /**
   * @description 处理暂停/启动
   * @param {JobItem} record 选中数据信息
   */
  const handlePauseClick = async (record?: JobItem) => {
    const recordData = record ? record : selectedRows[0];
    try {
      if (
        !taskCategoryStatusItems[TaskCategoryStatusEnum.RUNNING].includes(
          recordData.target.status
        )
      ) {
        await serviceJob.executeJob(recordData.jid);
        reload();
      } else {
        await serviceJob.pauseJob(recordData.jid);
        reload();
      }
    } catch (error: any) {
      handleError(error);
    }
  };
  /**
   * @description 处理失败重试
   * @param {JobItem} record 选中数据信息
   */
  const handleRetryClick = async (record?: JobItem) => {
    const recordData = record ? record : selectedRows[0];
    try {
      await serviceJob.executeJob(recordData.jid);
      reload();
    } catch (error: any) {
      handleError(error);
    }
  };
  /**
   * @description 处理更改配置
   * @param {JobItem} record 选中数据信息
   */
  const handleChangeClick = (record?: JobItem) => {
    const recordData = record ? record : selectedRows[0];
    setJid(recordData.target.id);
    setOpenChangeConfig(true);
  };
  /**
   * @description 处理删除任务(暂不支持)
   * @param {JobItem} record 选中数据信息
   */
  const handleDeleteClick = (record?: JobItem) => {};
  /**
   * @description 处理查看详情
   * @param {JobItem} record 选中数据信息
   */
  const handleDetailClick = (record: JobItem) => {
    setJid(record.target.id);
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
              ].includes(selectedRows[0].target.status) &&
                !taskCategoryStatusItems[
                  TaskCategoryStatusEnum.STOPPED
                ].includes(selectedRows[0].target.status) &&
                !taskCategoryStatusItems[
                  TaskCategoryStatusEnum.CONFIGCONFIRMED
                ].includes(selectedRows[0].target.status))
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
              !taskCategoryStatusItems[TaskCategoryStatusEnum.FAILED].includes(
                selectedRows[0].target.status
              )
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
              selectedRows[0].target.operateType === JobOperateType.Uninstall ||
              taskCategoryStatusItems[TaskCategoryStatusEnum.RUNNING].includes(
                selectedRows[0].target.status
              )
            }
          >
            {__("更改配置")}
          </Button>
          {/* 删除任务 */}
          <Button type="default" onClick={() => handleDeleteClick()} disabled>
            {__("删除任务")}
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
        <Table {...state} columns={columns} />
      </ContentLayout>
      {openDetail && (
        <TaskDetail open={openDetail} onCancel={onDetailCancel} jid={jid} />
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

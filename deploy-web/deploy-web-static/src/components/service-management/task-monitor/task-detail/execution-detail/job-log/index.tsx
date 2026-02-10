import React, { FC, useState } from "react";
import { Table, Button, Input, Modal } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { ContentLayout } from "../../../../../common/components";
import { safetyStr } from "../../../../../common/utils/string";
import {
  IGetLogParams,
  IGetLogTableParams,
  JobLogItem,
} from "../../../../../../api/service-management/service-deploy/declare";
import { formatTableResponse } from "../../../../../common/utils/request";
import { serviceJob } from "../../../../../../api/service-management/service-deploy";
import { safetyTime } from "../../../../utils/timer";
import { SERVICE_PREFIX } from "../../../../config";
import __ from "./locale";

interface IProps {
  // 任务id
  jid?: number;
  // 组件实例id
  cid?: number;
}
export const JobLog: FC<IProps> = ({ jid = -1, cid = -1 }) => {
  const [openLogInfo, setOpenLogInfo] = useState<boolean>(false);
  const [logInfo, setLogInfo] = useState<string>("");

  // 表格的基本设置
  const { state } = Table.useTable<
    JobLogItem,
    IGetLogTableParams,
    JobLogItem[]
  >({
    request: (params) => {
      const { current, pageSize } = params;
      const formatedParams: IGetLogParams = {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        jid,
        cid,
      };
      return serviceJob.getLog(formatedParams);
    },
    rowKey: "jlid",
    pagination: {
      showTotal: (total) => __("共${total}条", { total }),
      pageSizeOptions: [10, 20, 50, 100],
      pageSize: 10,
    },
    ...formatTableResponse(),
  });

  // 查看日志信息
  const handleCheckLog = (record: JobLogItem) => {
    setOpenLogInfo(true);
    setLogInfo(record.msg);
  };

  // 表格的列配置项
  const columns: TableColumnsType<JobLogItem> = [
    {
      title: __("应用名称"),
      dataIndex: "title",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("组件名称"),
      dataIndex: "cname",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("错误码"),
      dataIndex: "code",
      render: (value: number) => safetyStr(value.toString()),
      tooltip: (value: number) => safetyStr(value.toString()),
    },
    {
      title: __("错误码描述"),
      dataIndex: "description",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("日志时间"),
      dataIndex: "time",
      render: (value: number) => safetyTime(value),
      tooltip: (value: number) => safetyTime(value),
    },
    {
      title: __("操作"),
      width: 170,
      render: (value: string, record: JobLogItem) => {
        return (
          <Button
            type="link"
            onClick={() => {
              handleCheckLog(record);
            }}
          >
            {__("查看日志信息")}
          </Button>
        );
      },
      tooltip: () => __("查看日志信息"),
    },
  ];
  return (
    <ContentLayout moduleName={SERVICE_PREFIX}>
      <Table {...state} columns={columns} />
      <Modal
        key={Date.now()}
        title={__("日志信息")}
        open={openLogInfo}
        onOk={() => setOpenLogInfo(false)}
        onCancel={() => setOpenLogInfo(false)}
        wrapClassName={`${SERVICE_PREFIX}-log-modal`}
      >
        <Input.TextArea readOnly value={logInfo} rows={10} />
      </Modal>
    </ContentLayout>
  );
};

import React, { FC, useState } from "react";
import { Text } from "../../../../../common/components";
import { JobLog } from "../job-log";
import { Table, Button, Drawer } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { TaskConfigStatusEnum, taskConfigStatus } from "../../../type.d";
import { safetyStr } from "../../../../../common/utils/string";
import { ServiceTableType } from "../../../../utils/formatTable";
import __ from "./locale";

interface IProps {
  // 表格数据
  dataSource: ServiceTableType[];
  // 任务id
  jid: number;
}
export const ServiceExecution: FC<IProps> = ({ dataSource, jid }) => {
  const [componentInfo, setComponentInfo] = useState<ServiceTableType>(
    {} as ServiceTableType
  );
  const [openComponentLog, setOpenComponentLog] = useState<boolean>(false);

  // 表格的列配置项
  const columns: TableColumnsType<ServiceTableType> = [
    {
      title: __("名称"),
      dataIndex: "name",
    },
    {
      title: __("状态"),
      dataIndex: "status",
      render: (value: TaskConfigStatusEnum) => {
        return (
          <Text textColor={taskConfigStatus[value].color}>
            {taskConfigStatus[value].categoryText}
          </Text>
        );
      },
      tooltip: (value: TaskConfigStatusEnum) =>
        taskConfigStatus[value].categoryText,
    },
    {
      title: __("版本"),
      dataIndex: "version",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("执行信息"),
      render: (_, record: ServiceTableType) => {
        return (
          <Button
            type="link"
            onClick={() => {
              setComponentInfo(record);
              setOpenComponentLog(true);
            }}
          >
            {__("查看详情")}
          </Button>
        );
      },
      tooltip: (_) => __("查看详情"),
    },
  ];
  return (
    <>
      <Table
        dataSource={dataSource}
        columns={columns}
        scroll={{
          y: "calc(100vh - 515px)",
        }}
        pagination={{
          showQuickJumper: true,
          showSizeChanger: true,
          showTotal: (total) => __("共${total}条", { total }),
        }}
      />
      <Drawer
        title={componentInfo.name}
        width={900}
        onClose={() => {
          setOpenComponentLog(false);
        }}
        open={openComponentLog}
        destroyOnClose
        maskStyle={{ backgroundColor: "rgba(0,0,0,0.1)" }}
        showFooter={false}
      >
        <JobLog jid={jid} cid={componentInfo.cid} />
      </Drawer>
    </>
  );
};

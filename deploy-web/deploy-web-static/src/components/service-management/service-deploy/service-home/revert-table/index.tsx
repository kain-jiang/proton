import React, { FC, useEffect, useState } from "react";
import {
  Button,
  Modal,
  Table,
  TableColumnsType,
  message,
} from "@kweaver-ai/ui";
import {
  IGetJobParams,
  IGetJobTableParams,
  JobItem,
} from "../../../../../api/service-management/service-deploy/declare";
import { serviceJob } from "../../../../../api/service-management/service-deploy";
import { safetyTime } from "../../../utils/timer";
import { formatTableResponse } from "../../../../common/utils/request";
import __ from "./locale";
import { SERVICE_PREFIX } from "../../../config";
import {
  JobOperateType,
  TaskConfigStatusEnum,
} from "../../../task-monitor/type.d";
import styles from "./styles.module.less";
import { handleError } from "../../../utils/handleError";
import { ElementText } from "../../../../common/components";
import elementStyles from "../../../../common/components/element-text/style.module.less";
import { assignTo } from "../../../../../tools/browser";
import { deployMiniPathname } from "../../../../../core/path";

interface IProps {
  // 选中的服务名称
  serviceName: string;
  // 控制弹窗开关
  changeRevertTableOpen: (item: boolean) => void;
  // 系统空间id
  sid: number;
}
export const RevertTable: FC<IProps> = ({
  serviceName,
  changeRevertTableOpen,
  sid,
}) => {
  const [selectedRecord, setSelectedRecord] = useState<JobItem>({} as JobItem);

  // 表格的基本设置
  const { state } = Table.useTable<JobItem, IGetJobTableParams, JobItem[]>({
    request: (params) => {
      const { current, pageSize } = params;
      const formatedParams: IGetJobParams = {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        name: serviceName,
        status: [TaskConfigStatusEnum.SUCCESS],
        sid,
      };
      return serviceJob.get(formatedParams);
    },
    rowKey: "jid",
    pagination: {
      showTotal: (total) => __("共${total}条", { total }),
      pageSizeOptions: [10, 20, 50, 100],
      pageSize: 10,
    },
    rowSelection: {
      type: "radio",
      columnWidth: 33,
      onChange: (selectedRowKeys: React.Key[], selectedRows: JobItem[]) => {
        setSelectedRecord(selectedRows[0]);
      },
    },
    ...formatTableResponse(),
  });

  // 表格的列配置项
  const columns: TableColumnsType<JobItem> = [
    {
      title: __("版本"),
      render: (_, record: JobItem) => record.target.application.version,
      tooltip: (_, record: JobItem) => record.target.application.version,
    },
    {
      title: __("更新时间"),
      render: (_, record: JobItem) => safetyTime(record.target.startTime),
      tooltip: (_, record: JobItem) => safetyTime(record.target.startTime),
    },
    {
      title: __("ID"),
      render: (_, record: JobItem) => record.target.id,
      tooltip: (_, record: JobItem) => record.target.id,
    },
  ];

  // 二次确认弹窗
  const handleOKComfirm = () => {
    Modal.confirm({
      title: __("您将安装此版本"),
      content: (
        <div className={styles["modal-content"]}>
          {__("安装此版本将覆盖当前版本，是否继续安装")}
        </div>
      ),
      onOk: () => revertService(),
    });
  };

  // 回滚服务
  const revertService = async () => {
    try {
      const { formData } = await serviceJob.getJSONSchema(selectedRecord.jid);
      await serviceJob.createAndExecuteJob({
        formData: formData!,
        aid: selectedRecord?.target?.application?.aid,
        operateType: JobOperateType.Revert,
        sid,
      });
      // 成功提示
      message.success(
        <ElementText
          text={__("回滚服务任务创建成功，前往-查看")}
          insert={
            <a className={elementStyles["target-herf"]} onClick={clickHerf}>
              {__("【任务监控】")}
            </a>
          }
        />
      );
    } catch (error: any) {
      handleError(error);
    } finally {
      changeRevertTableOpen(false);
    }
  };

  /**
   * @description 跳转到任务监控
   */
  const clickHerf = () => {
    assignTo(deployMiniPathname.taskMonitorPathname);
  };

  return (
    <Modal
      title={__("选择回滚版本")}
      open={true}
      onOk={() => handleOKComfirm()}
      onCancel={() => changeRevertTableOpen(false)}
      okButtonProps={{ disabled: !selectedRecord.jid }}
      wrapClassName={`${SERVICE_PREFIX}-log-modal`}
    >
      <div style={{ minHeight: "400px", maxHeight: "400px" }}>
        <Table {...state} columns={columns} />
      </div>
    </Modal>
  );
};

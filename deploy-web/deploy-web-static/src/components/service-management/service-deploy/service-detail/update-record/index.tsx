import React, { FC } from "react";
import { Table, Button, Modal } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { ReactComponent as UpdateRecordIcon } from "../../../assets/UpdateRecordIcon.svg";
import { formatTableResponse } from "../../../../common/utils/request";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import { serviceJob } from "../../../../../api/service-management/service-deploy";
import {
  IGetJobParams,
  JobItem,
  IGetJobTableParams,
  ApplicationItem,
} from "../../../../../api/service-management/service-deploy/declare";
import { IBaseProps } from "../../declare";
import { safetyTime } from "../../../utils/timer";
import styles from "./styles.module.less";
import __ from "./locale";

interface IProps extends IBaseProps {
  // 服务id
  serviceId: number;
  // 服务名称
  serviceName: string;
  // 修改服务id
  changeServiceId: (id: number) => void;
  // 修改主服务id
  changeMainServiceId: (id: number) => void;
  // 修改更新服务的信息
  changeUpdateServiceRecord: (record: ApplicationItem) => void;
  // 修改更新服务的jid
  changeJid: (jid: number) => void;
  // 系统空间id
  sid: number;
  // 修改系统空间id
  changeSid: (sid: number) => void;
}
export const UpdateRecord: FC<IProps> = ({
  changeServiceMode,
  changeServiceId,
  changeMainServiceId,
  changeJid,
  changeUpdateServiceRecord,
  serviceId,
  serviceName,
  sid,
  changeSid,
}) => {
  // 表格的基本设置
  const { state } = Table.useTable<JobItem, IGetJobTableParams, JobItem[]>({
    request: (params) => {
      const { current, pageSize } = params;
      const formatedParams: IGetJobParams = {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        name: serviceName,
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
    ...formatTableResponse(),
  });

  // 表格的列配置项
  const columns: TableColumnsType<JobItem> = [
    {
      title: __("版本"),
      render: (_, record: JobItem) => (
        <Button type="link" onClick={() => handleDetailClick(record)}>
          {record.target.application.version}
        </Button>
      ),
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
    {
      title: __("操作"),
      render: (_, record: JobItem) => {
        return (
          <Button
            type="link"
            icon={<UpdateRecordIcon />}
            onClick={() => handleUpdateClick(record)}
            disabled={record.target.id === serviceId}
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
  const handleDetailClick = (record: JobItem) => {
    changeServiceMode(ServiceMode.Service);
    changeServiceId(record.target.id);
    changeMainServiceId(record.target.id);
  };
  /**
   * @description 回退版本
   * @param record 选中行表格数据信息
   */
  const handleUpdateClick = (record: JobItem) => {
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
          aid: record.target.application.aid,
          name: record.target.application.name,
          title: record.target.application.title,
          version: record.target.application.version,
        });
        // 携带任务id
        changeJid(record.jid);
        changeSid(record.target.sid!);
      },
    });
  };
  return <Table {...state} columns={columns} />;
};

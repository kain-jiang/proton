import React, { FC, useState } from "react";
import { ContentLayout, Toolbar, Text } from "../../../../common/components";
import {
  JobOperateType,
  jobOperateTypeStatus,
  taskConfigStatus,
} from "../../type.d";
import { ServiceExecution } from "./service-execution";
import { JobLog } from "./job-log";
import { Refresh, Form, Tabs } from "@kweaver-ai/ui";
import { DownloadOutlined } from "@kweaver-ai/ui/icons";
import { ServiceJSONSchemaItem } from "../../../../../api/service-management/service-deploy/declare";
import { formatTable } from "../../../utils/formatTable";
import { safetyRunningTime, safetyTime } from "../../../utils/timer";
import { SERVICE_PREFIX } from "../../../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";
import { safetyStr } from "../../../../common/utils/string";

interface IProps {
  // 任务信息
  taskInfo: ServiceJSONSchemaItem;
  // 任务id
  jid: number;
  // 刷新事件回调
  changeRefresh: (func: (refresh: boolean) => boolean) => void;
}
export const ExecutionDetail: FC<IProps> = ({
  taskInfo,
  jid,
  changeRefresh,
}) => {
  const [active, setActive] = useState<string>("1");
  /**
   * @description 切换tab事件
   * @param value
   */
  const onTabChange = (value: string): void => {
    setActive(value);
  };

  const items = [
    {
      label: __("组件"),
      key: "1",
      children: (
        <ServiceExecution
          jid={jid}
          dataSource={
            taskInfo?.formData ? formatTable(taskInfo?.formData.components) : []
          }
        />
      ),
    },
    {
      label: __("任务执行详情"),
      key: "2",
      children: <JobLog jid={jid} />,
    },
  ];
  const header = (
    <Toolbar
      right={
        <React.Fragment>
          <DownloadOutlined
            onPointerEnterCapture={noop}
            onPointerLeaveCapture={noop}
          />
          <Refresh onClick={() => changeRefresh((refresh) => !refresh)} />
        </React.Fragment>
      }
      rightSize={24}
      wrapperClass={styles["execution-detail-toolbar"]}
      moduleName={SERVICE_PREFIX}
    />
  );
  return (
    <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
      <Form labelCol={{ span: 3 }} wrapperCol={{ span: 16 }} labelAlign="left">
        <Form.Item label={__("任务类型")} className={styles["form-item"]}>
          {safetyStr(
            jobOperateTypeStatus[taskInfo?.operateType as JobOperateType]?.text
          )}
        </Form.Item>
        <Form.Item label={__("开始时间")} className={styles["form-item"]}>
          {safetyTime(taskInfo?.startTime)}
        </Form.Item>
        <Form.Item label={__("结束时间")} className={styles["form-item"]}>
          {safetyTime(taskInfo?.endTime)}
        </Form.Item>
        <Form.Item label={__("运行时间")} className={styles["form-item"]}>
          {safetyRunningTime(taskInfo?.endTime, taskInfo?.startTime)}
        </Form.Item>
        <Form.Item label={__("状态")} className={styles["form-item"]}>
          <Text textColor={taskConfigStatus[taskInfo?.status]?.color}>
            {taskConfigStatus[taskInfo?.status]?.categoryText}
          </Text>
        </Form.Item>
        <Form.Item label={__("执行阶段")} className={styles["form-item"]}>
          {taskConfigStatus[taskInfo?.status]?.text}
        </Form.Item>
      </Form>
      <Tabs
        className="service-inner-tab"
        defaultActiveKey={String(active) || "1"}
        onChange={onTabChange}
        items={items}
        destroyInactiveTabPane
      />
    </ContentLayout>
  );
};

import React, { FC, useEffect, useState } from "react";
import { Detail } from "./detail";
import { ExecutionDetail } from "./execution-detail";
import { VerifyRecord } from "./verify-record";
import { Drawer, Tabs } from "@kweaver-ai/ui";
import { handleError } from "../../utils/handleError";
import { serviceJob } from "../../../../api/service-management/service-deploy";
import { verify } from "../../../../api/service-management/task-monitor";
import { ServiceJSONSchemaItem } from "../../../../api/service-management/service-deploy/declare";
import {
  DataSchemaVerifyItem,
  FuncVerifyItem,
} from "../../../../api/service-management/task-monitor/declare";
import __ from "./locale";

interface IProps {
  // 滑窗状态
  open: boolean;
  // 任务id
  jid: number;
  // 关闭滑窗
  onCancel: () => void;
}
export const TaskDetail: FC<IProps> = ({ open, jid, onCancel }) => {
  // 当前Tabs组件key
  const [active, setActive] = useState<string>("1");
  // 任务信息
  const [taskInfo, setTaskInfo] = useState<ServiceJSONSchemaItem>({
    status: 0,
  } as ServiceJSONSchemaItem);
  // 数据验证列表
  const [dataSchemaVerifyList, setDataSchemaVerifyList] = useState<
    DataSchemaVerifyItem[]
  >([]);
  // 功能验证列表
  const [funcVerifyList, setFuncVerifyList] = useState<FuncVerifyItem[]>([]);
  // 控制刷新数据
  const [refresh, setRefresh] = useState<boolean>(false);

  useEffect(() => {
    getTaskDetailInfo();
  }, [jid, refresh]);
  /**
   * @description 获取任务信息和验证记录
   */
  const getTaskDetailInfo = async () => {
    try {
      const taskRes = await serviceJob.getJSONSchema(jid);
      setTaskInfo(taskRes);
      const verifyRes = await verify.get(jid);
      setDataSchemaVerifyList(verifyRes.dataSchemaVerifyList || []);
      setFuncVerifyList(verifyRes.funcVerifyList || []);
    } catch (error: any) {
      handleError(error);
    }
  };
  /**
   * @description 切换tab事件
   * @param value
   */
  const onTabChange = (value: string): void => {
    setActive(value);
  };
  /**
   * @description 获取tab组件
   */
  const getTabItems = () => {
    const items = [
      {
        label: __("详情"),
        key: "1",
        children: <Detail taskInfo={taskInfo} />,
      },
      {
        label: __("执行详情"),
        key: "2",
        children: (
          <ExecutionDetail
            taskInfo={taskInfo}
            jid={jid}
            changeRefresh={setRefresh}
          />
        ),
      },
    ];
    if (dataSchemaVerifyList.length || funcVerifyList.length) {
      return [
        ...items,
        {
          label: __("验证记录"),
          key: "3",
          children: (
            <VerifyRecord
              dataSchemaVerifyList={dataSchemaVerifyList}
              funcVerifyList={funcVerifyList}
              changeRefresh={setRefresh}
            />
          ),
        },
      ];
    } else {
      return items;
    }
  };

  return (
    <Drawer
      title={taskInfo.title}
      width={1000}
      onClose={onCancel}
      open={open}
      push={false}
      showFooter={false}
      destroyOnClose
    >
      <Tabs
        className="service-tab-container"
        defaultActiveKey={String(active) || "1"}
        onChange={onTabChange}
        items={getTabItems()}
        destroyInactiveTabPane
      />
    </Drawer>
  );
};

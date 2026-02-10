import React, { FC, useState } from "react";
import { Tabs } from "@kweaver-ai/ui";
import { TaskMonitor } from "../task-monitor";
import { SuiteTaskMonitor } from "../../suite-management/task-monitor";
import { JobType } from "../../../core/suite-management/suite-deploy";
import __ from "./locale";

export const ServiceTaskMonitor: FC = () => {
  // 当前Tabs组件key
  const [active, setActive] = useState<string>("1");

  const onTabChange = (value: string): void => {
    setActive(value);
  };

  const items = [
    {
      label: __("单任务"),
      key: "1",
      children: <TaskMonitor />,
    },
    {
      label: __("批量任务"),
      key: "2",
      children: <SuiteTaskMonitor jobType={JobType.Batch} />,
    },
  ];

  return (
    <Tabs
      className="service-tab-container"
      defaultActiveKey={String(active) || "1"}
      onChange={onTabChange}
      items={items}
      destroyInactiveTabPane
    />
  );
};

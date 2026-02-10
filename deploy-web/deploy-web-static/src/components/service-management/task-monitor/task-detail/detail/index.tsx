import React, { FC } from "react";
import { ServiceFramework } from "../../../service-framework";
import { Space, Form } from "@kweaver-ai/ui";
import { DownOutlined } from "@kweaver-ai/ui/icons";
import { ServiceJSONSchemaItem } from "../../../../../api/service-management/service-deploy/declare";
import { safetyStr } from "../../../../common/utils/string";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";

interface IProps {
  // 任务信息
  taskInfo: ServiceJSONSchemaItem;
}
export const Detail: FC<IProps> = ({ taskInfo }) => {
  return (
    <div className={styles["detail"]}>
      <Space className={styles["detail-title"]}>
        <span>{__("服务信息")}</span>
        <DownOutlined
          onPointerEnterCapture={noop}
          onPointerLeaveCapture={noop}
        />
      </Space>
      <Form labelCol={{ span: 3 }} wrapperCol={{ span: 16 }} labelAlign="left">
        <Form.Item label={__("服务名称")} className={styles["form-item"]}>
          {taskInfo.title}
        </Form.Item>
        <Form.Item label={__("版本")} className={styles["form-item"]}>
          {safetyStr(taskInfo.version)}
        </Form.Item>
        <Form.Item label={__("备注")} className={styles["form-item"]}>
          {safetyStr(taskInfo.comment)}
        </Form.Item>
      </Form>
      <Space className={styles["config-title"]}>
        <span>{__("配置项")}</span>
        <DownOutlined
          onPointerEnterCapture={noop}
          onPointerLeaveCapture={noop}
        />
      </Space>
      <ServiceFramework
        formData={taskInfo.formData}
        schema={taskInfo.schema}
        uiSchema={taskInfo.uiSchema || {}}
        isReadOnly={true}
      />
    </div>
  );
};

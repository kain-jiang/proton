import React, { FC } from "react";
import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ServiceFramework } from "../../../service-framework";
import { Space, Table, Form, Input } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { DownOutlined } from "@kweaver-ai/ui/icons";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import { ApplicationItem } from "../../../../../api/service-management/service-deploy/declare";
import { OperationType } from "../type.d";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";
import { ServiceSchemaItem } from "../../../../../api/suite-management/suite-deploy/declare";

interface IProps {
  // 备注
  comment: string;
  // 配置项 formData
  formData: RJSFSchema;
  // 配置项 schema
  schema: RJSFSchema;
  // 配置项 UIschema
  uiSchema: UiSchema;
  // 操作服务类型（安装或更新）
  operationType: OperationType;
  // 修改备注
  changeComment: (comment: string) => void;
  // 批量任务 服务配置项 / 单服务任务 服务信息
  serviceConfig: ServiceSchemaItem[] | ApplicationItem[];
  // 批量任务名称
  batchJobName: string;
}
export const ConfirmConfig: FC<IProps> = ({
  formData,
  schema,
  uiSchema,
  comment,
  operationType,
  changeComment,
  serviceConfig,
  batchJobName,
}) => {
  // 表格的列配置项
  const columns: TableColumnsType<ApplicationItem | ServiceSchemaItem> = [
    {
      title: __("服务名称"),
      dataIndex: "title",
    },
    {
      title: __("版本"),
      dataIndex: "version",
    },
  ];
  /**
   * @description 修改备注
   * @param e 输入框改变时间
   */
  const handleCommentChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    changeComment(e.target.value);
  };

  const windowHeight = window.innerHeight > 720 ? "100vh" : "720px";

  return (
    <React.Fragment>
      {serviceConfig.length > 1 ? (
        <div style={{ marginBottom: "20px" }}>
          {__("批量任务名称：${batchJobName}", { batchJobName })}
        </div>
      ) : (
        <Space className={styles["config-title"]}>
          <span>{__("服务信息")}</span>
          <DownOutlined
            onPointerEnterCapture={noop}
            onPointerLeaveCapture={noop}
          />
        </Space>
      )}
      <Table
        rowKey="aid"
        className="service-table"
        dataSource={serviceConfig}
        columns={columns}
        scroll={{ y: `calc(${windowHeight} - 400px)` }}
        pagination={{
          showQuickJumper: true,
          showSizeChanger: true,
          showTotal: (total) => __("共${total}条", { total }),
          hideOnSinglePage: serviceConfig.length > 1 ? false : true,
        }}
      />
      {serviceConfig.length > 1 ? null : (
        <>
          <Space className={styles["config-title"]}>
            <span>{__("配置项")}</span>
            <DownOutlined
              onPointerEnterCapture={noop}
              onPointerLeaveCapture={noop}
            />
          </Space>
          <ServiceFramework
            uiSchema={uiSchema}
            formData={formData}
            schema={schema}
            isReadOnly={true}
          />
        </>
      )}
      {operationType === ServiceMode.Install && serviceConfig.length === 1 ? (
        <Form className={styles["comment-form"]}>
          <Form.Item label={__("备注")}>
            <Input
              className={styles["comment-input"]}
              value={comment}
              onChange={handleCommentChange}
            />
          </Form.Item>
        </Form>
      ) : null}
    </React.Fragment>
  );
};

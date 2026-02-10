import React, { FC } from "react";
import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ServiceFramework } from "../../../service-framework";
import { Space } from "@kweaver-ai/ui";
import { DownOutlined } from "@kweaver-ai/ui/icons";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";

interface IProps {
  // 配置项formData
  formData: RJSFSchema;
  // 配置项schema
  schema: RJSFSchema;
  // 配置项UIschema
  uiSchema: UiSchema;
  // 修改formData
  onChangeFormData: (formData: any) => void;
  // 修改是否完成配置项验证
  changeIsFormValidator: (isFormValidator: boolean) => void;
}
export const CreateConfig: FC<IProps> = ({
  formData,
  schema,
  uiSchema,
  onChangeFormData,
  changeIsFormValidator,
}) => {
  return (
    <React.Fragment>
      <Space className={styles["config-title"]}>
        <span>{__("配置项")}</span>
        <DownOutlined
          onPointerEnterCapture={noop}
          onPointerLeaveCapture={noop}
        />
      </Space>
      <ServiceFramework
        formData={formData}
        uiSchema={uiSchema}
        schema={schema}
        onChangeFormData={onChangeFormData}
        isReadOnly={false}
        changeIsFormValidator={changeIsFormValidator}
      />
    </React.Fragment>
  );
};

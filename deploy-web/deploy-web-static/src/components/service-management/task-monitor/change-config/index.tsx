import React, { FC, useEffect, useState } from "react";
import { Drawer, Space } from "@kweaver-ai/ui";
import { DownOutlined } from "@kweaver-ai/ui/icons";
import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ServiceFramework } from "../../service-framework";
import { formValidatorComfirm, handleError } from "../../utils/handleError";
import { serviceJob } from "../../../../api/service-management/service-deploy";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";

interface IProps {
  // 是否展示滑窗
  open: boolean;
  // 任务id
  jid: number;
  // 关闭滑窗
  onCancel: () => void;
}
export const ChangeConfig: FC<IProps> = ({ open, jid, onCancel }) => {
  // 配置项formData
  const [formData, setFormData] = useState<RJSFSchema>({});
  // 配置项schema
  const [schema, setSchema] = useState<RJSFSchema>({});
  // 配置项uiSchema
  const [uiSchema, setUiSchema] = useState<UiSchema>({});
  // 配置项是否通过验证
  const [isFormValidator, setIsFormValidator] = useState<boolean>(false);

  useEffect(() => {
    getFormerJSONSchema(jid);
  }, [jid]);
  /**
   * @description 获取先前配置项
   * @param {number} jid 任务id
   */
  const getFormerJSONSchema = async (jid: number) => {
    try {
      const res = await serviceJob.getJSONSchema(jid);
      setFormData(res.formData!);
      setSchema(res.schema!);
      setUiSchema(res.uiSchema || {});
    } catch (error: any) {
      handleError(error);
    }
  };
  /**
   * @description 确定更改配置项
   */
  const onOk = async () => {
    if (!isFormValidator) {
      formValidatorComfirm();
      return;
    }
    try {
      await serviceJob.configJob(jid, {
        formData,
        jid,
      });
      await serviceJob.executeJob(jid);
      onCancel();
    } catch (error: any) {
      handleError(error);
    }
  };
  return (
    <Drawer
      title={__("更改配置")}
      onOk={onOk}
      onClose={onCancel}
      open={open}
      width={1000}
    >
      <div className={styles["config-drawer-body"]}>
        <Space className={styles["config-title"]}>
          <span>{__("配置项")}</span>
          <DownOutlined
            onPointerEnterCapture={noop}
            onPointerLeaveCapture={noop}
          />
        </Space>
        <ServiceFramework
          formData={formData}
          schema={schema}
          uiSchema={uiSchema}
          onChangeFormData={setFormData}
          isReadOnly={false}
          changeIsFormValidator={setIsFormValidator}
        />
      </div>
    </Drawer>
  );
};

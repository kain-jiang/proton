import React, { FC, useEffect, useState } from "react";
import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ContentLayout, ElementText } from "../../../common/components";
import { ServiceFramework } from "../../service-framework";
import elementStyles from "../../../common/components/element-text/style.module.less";
import {
  Breadcrumb,
  Space,
  Button,
  Form,
  Select,
  message,
} from "@kweaver-ai/ui";
import { LeftOutlined } from "@kweaver-ai/ui/icons";
import { formValidatorComfirm, handleError } from "../../utils/handleError";
import { IBaseProps } from "../declare";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import { serviceJob } from "../../../../api/service-management/service-deploy";
import { ApplicationItem } from "../../../../api/service-management/service-deploy/declare";
import { assignTo } from "../../../../tools/browser";
import { deployMiniPathname } from "../../../../core/path";
import { SERVICE_PREFIX } from "../../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";

interface IProps extends IBaseProps {
  // 更新服务原始任务id
  jid: number;
  // 更新服务信息
  updateServiceRecord: ApplicationItem;
  // 系统空间id
  sid: number;
}

export const RevertService: FC<IProps> = ({
  changeServiceMode,
  updateServiceRecord,
  jid,
  sid,
}) => {
  // 配置项formData
  const [formData, setFormData] = useState<RJSFSchema>();
  // 配置项schema
  const [schema, setSchema] = useState<RJSFSchema>({});
  // 配置项uiSchema
  const [uiSchema, setUiSchema] = useState<UiSchema>({});
  // 配置项是否验证
  const [isFormValidator, setIsFormValidator] = useState<boolean>(false);

  // 更新不同版本服务时不执行，回退历史版本执行
  useEffect(() => {
    getJSONSchema(jid);
  }, [jid]);

  /**
   * @description 获取配置项
   * @param {number} jid 任务id
   */
  const getJSONSchema = async (jid: number) => {
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
   * @description 确定更新
   */
  const handleUpdateClick = async () => {
    if (!isFormValidator) {
      formValidatorComfirm();
      return;
    }
    try {
      await serviceJob.createAndExecuteJob({
        formData: formData!,
        aid: updateServiceRecord.aid,
        sid,
      });
      // 成功提示
      message.success(
        <ElementText
          text={__("更新服务任务创建成功，前往-查看")}
          insert={
            <a className={elementStyles["target-herf"]} onClick={clickHerf}>
              {__("【任务监控】")}
            </a>
          }
        />
      );
      changeServiceMode(ServiceMode.Home);
    } catch (error: any) {
      handleError(error);
    }
  };

  /**
   * @description 跳转到任务监控
   */
  const clickHerf = () => {
    assignTo(deployMiniPathname.taskMonitorPathname);
  };
  const breadcrumb = (
    <Breadcrumb className={styles["update-breadcrumb"]}>
      <div onClick={() => changeServiceMode(ServiceMode.Home)}>
        <LeftOutlined
          className={styles["breadcrumb-icon"]}
          onPointerEnterCapture={noop}
          onPointerLeaveCapture={noop}
        />
      </div>
      <Breadcrumb.Item onClick={() => changeServiceMode(ServiceMode.Home)}>
        <span className={styles["breadcrumb-title"]}>{__("更新服务")}</span>
      </Breadcrumb.Item>
    </Breadcrumb>
  );
  const footer = (
    <Space>
      <Button type="primary" disabled={!formData} onClick={handleUpdateClick}>
        {__("确定")}
      </Button>
      <Button
        type="default"
        onClick={() => changeServiceMode(ServiceMode.Home)}
      >
        {__("取消")}
      </Button>
    </Space>
  );
  return (
    <ContentLayout
      breadcrumb={breadcrumb}
      footer={footer}
      moduleName={SERVICE_PREFIX}
    >
      <span className={styles["config-title"]}>{__("配置项")}</span>
      <Form>
        <Form.Item label={__("版本")} required={true}>
          <Select
            defaultValue={updateServiceRecord.version}
            style={{ width: 350, marginLeft: "20px" }}
            disabled
          />
        </Form.Item>
      </Form>
      {formData && (
        <ServiceFramework
          formData={formData}
          uiSchema={uiSchema}
          schema={schema}
          isReadOnly={false}
          onChangeFormData={setFormData}
          changeIsFormValidator={setIsFormValidator}
        />
      )}
    </ContentLayout>
  );
};

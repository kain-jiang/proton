import React, { FC, useEffect, useState } from "react";
import { ContentLayout, ElementText } from "../../../common/components";
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
import {
    ApplicationItem,
    ICreateComposeJobParams,
    ServiceSchemaItem,
} from "../../../../api/suite-management/suite-deploy/declare";
import { assignTo } from "../../../../tools/browser";
import { suiteTaskMonitorPathname } from "../../../../core/path";
import { SERVICE_PREFIX } from "../../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";
import { SuiteConfig } from "../../suite-config";
import { composeJob } from "../../../../api/suite-management/suite-deploy";
import { ConfigEditStatusEnum } from "../../suite-config/helper";
import { SchemaOperateType } from "../../../../core/suite-management/suite-deploy";

interface IProps extends IBaseProps {
    // 更新服务原始任务id
    jid: number;
    // 更新服务信息
    updateServiceRecord: ApplicationItem;
}
export const RevertService: FC<IProps> = ({
    changeServiceMode,
    updateServiceRecord,
    jid,
}) => {
    // 套件配置项（apps）
    const [suiteConfig, setSuiteConfig] = useState<ServiceSchemaItem[]>([]);

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
            const res = await composeJob.getJobInfo(jid);
            const apps = res?.config?.apps || [];
            setSuiteConfig(
                apps.map((app) => {
                    return {
                        ...app,
                        editStatus: ConfigEditStatusEnum.Unsubmitted,
                    };
                })
            );
        } catch (error: any) {
            handleError(error);
        }
    };

    /**
     * @description 确定更新
     */
    const handleUpdateClick = async () => {
        if (
            suiteConfig.some(
                (config) =>
                    config?.editStatus === ConfigEditStatusEnum.Unsubmitted
            )
        ) {
            formValidatorComfirm();
            return;
        }
        const appsPayload = suiteConfig.map((config) => {
            return {
                name: config.name,
                version: config.version,
                formData: config.formData,
            };
        });
        const payload: ICreateComposeJobParams = {
            jname: updateServiceRecord.mname,
            mversion: updateServiceRecord.mversion,
            description: "更新",
            config: {
                apps: appsPayload,
                pcomponents: null,
            },
        };
        try {
            await composeJob.createJob(payload);
            // 成功提示
            message.success(
                <ElementText
                    text={__("更新服务任务创建成功，前往-查看")}
                    insert={
                        <a
                            className={elementStyles["target-herf"]}
                            onClick={clickHerf}
                        >
                            {__("【套件任务监控】")}
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
        assignTo(suiteTaskMonitorPathname);
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
            <Breadcrumb.Item
                onClick={() => changeServiceMode(ServiceMode.Home)}
            >
                <span className={styles["breadcrumb-title"]}>
                    {__("更新服务")}
                </span>
            </Breadcrumb.Item>
        </Breadcrumb>
    );
    const footer = (
        <Space>
            <Button type="primary" onClick={handleUpdateClick}>
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
                        defaultValue={updateServiceRecord.mversion}
                        style={{ width: 350, marginLeft: "20px" }}
                        disabled
                    />
                </Form.Item>
            </Form>
            {suiteConfig.length ? (
                <SuiteConfig
                    suiteConfig={suiteConfig}
                    isSynchronousUpdate={false}
                    changeSuiteConfig={setSuiteConfig}
                    operateType={SchemaOperateType.Revert}
                />
            ) : null}
        </ContentLayout>
    );
};

import React, { FC, useState } from "react";
import { ContentLayout, ElementText } from "../../../common/components";
import { ChooseService } from "./choose-service";
import { ConfirmConfig } from "./confirm-config";
import elementStyles from "../../../common/components/element-text/style.module.less";
import {
    Breadcrumb,
    Steps,
    Space,
    Button,
    Modal,
    message,
    Tooltip,
} from "@kweaver-ai/ui";
import { LeftOutlined } from "@kweaver-ai/ui/icons";
import { formValidatorComfirm, handleError } from "../../utils/handleError";
import { IBaseProps } from "../declare";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import {
    ApplicationItem,
    ServiceSchemaItem,
} from "../../../../api/suite-management/suite-deploy/declare";
import { OperationServiceStepsEnum, OperationType } from "./type.d";
import { assignTo } from "../../../../tools/browser";
import { suiteTaskMonitorPathname } from "../../../../core/path";
import { SERVICE_PREFIX } from "../../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";
import { SuiteConfig } from "../../suite-config";
import { ConfigEditStatusEnum } from "../../suite-config/helper";
import { composeJob } from "../../../../api/suite-management/suite-deploy";

const initailServiceInfo: ApplicationItem = {
    mname: "",
    title: "",
    mversion: "",
};
interface IProps extends IBaseProps {
    // 更新服务信息
    updateServiceRecord?: ApplicationItem;
    // 操作服务类型（安装或更新）
    operationType: OperationType;
}
export const OperationService: FC<IProps> = ({
    changeServiceMode,
    updateServiceRecord,
    operationType,
}) => {
    const [current, setCurrent] = useState<OperationServiceStepsEnum>(
        OperationServiceStepsEnum.CHOOSESERVICE
    );
    // 选择的套件信息
    const [serviceInfo, setServiceInfo] = useState<ApplicationItem>(
        updateServiceRecord || initailServiceInfo
    );
    // 不同版本服务列表
    const [serviceInfos, setServiceInfos] = useState<ApplicationItem[]>([]);
    // 套件服务上传和组件安装是否正确
    const [suiteConfigCorrect, setSuiteConfigCorrect] =
        useState<boolean>(false);
    // 套件配置项(apps)
    const [suiteConfig, setSuiteConfig] = useState<ServiceSchemaItem[]>([]);
    // 是否同步更新
    const [isSynchronousUpdate, setIsSynchronousUpdate] =
        useState<boolean>(false);
    // 确认按钮是否加载中
    const [loading, setLoading] = useState(false);

    /**
     * @description 点击下一步
     */
    const handeleNextClick = async () => {
        if (current === OperationServiceStepsEnum.CHOOSESERVICE) {
            setCurrent(current + 1);
        } else if (current === OperationServiceStepsEnum.CREATECONFIG) {
            if (
                suiteConfig.every(
                    (config) =>
                        config.editStatus === ConfigEditStatusEnum.Submitted
                )
            ) {
                setCurrent(current + 1);
            } else {
                formValidatorComfirm();
            }
        }
    };

    /**
     * @description 确定安装或更新
     */
    const handleOperationClick = async () => {
        try {
            setLoading(true);
            const appsPayload = suiteConfig.map((config) => {
                return {
                    name: config.name,
                    version: config.version,
                    formData: config.formData,
                };
            });
            const payload = {
                config: { apps: appsPayload, pcomponents: null },
                jname: serviceInfo.mname,
                mversion: serviceInfo.mversion,
                description:
                    operationType === ServiceMode.Install ? "安装" : "更新",
            };
            // 配置任务并执行
            await composeJob.createJob(payload);
            // 成功提示
            message.success(
                <ElementText
                    text={
                        operationType === ServiceMode.Install
                            ? __("安装套件任务创建成功，前往-查看")
                            : __("更新套件任务创建成功，前往-查看")
                    }
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
        } finally {
            setLoading(false);
        }
    };
    /**
     * @description 跳转到任务监控
     */
    const clickHerf = () => {
        assignTo(suiteTaskMonitorPathname);
    };
    const items = [
        {
            title:
                operationType === ServiceMode.Install
                    ? __("选择套件")
                    : __("选择版本"),
        },
        {
            title: __("填写配置项"),
        },
        {
            title: __("确认信息"),
        },
    ];
    /**
     * @description 获取操作服务的内容
     * @return 操作服务当前步骤的组件
     */
    const getContent = () => {
        if (current === OperationServiceStepsEnum.CHOOSESERVICE) {
            return (
                <ChooseService
                    operationType={operationType}
                    serviceInfo={serviceInfo}
                    serviceInfos={serviceInfos}
                    isSynchronousUpdate={isSynchronousUpdate}
                    changeServiceInfo={setServiceInfo}
                    changeServiceInfos={setServiceInfos}
                    changeSuiteConfigCorrect={setSuiteConfigCorrect}
                    changeSuiteConfig={setSuiteConfig}
                    changeIsSynchronousUpdate={setIsSynchronousUpdate}
                />
            );
        } else if (current === OperationServiceStepsEnum.CREATECONFIG) {
            return (
                <SuiteConfig
                    suiteConfig={suiteConfig}
                    isSynchronousUpdate={isSynchronousUpdate}
                    changeSuiteConfig={setSuiteConfig}
                />
            );
        } else if (current === OperationServiceStepsEnum.CONFIRMCONFIG) {
            return (
                <ConfirmConfig
                    serviceInfo={serviceInfo as any}
                    suiteConfig={suiteConfig}
                />
            );
        }
    };
    const breadcrumb = (
        <Breadcrumb className={styles["operation-breadcrumb"]}>
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
                    {operationType === ServiceMode.Install
                        ? __("安装套件")
                        : __("更新套件")}
                </span>
            </Breadcrumb.Item>
        </Breadcrumb>
    );
    const footer = (
        <Space>
            {current > OperationServiceStepsEnum.CHOOSESERVICE && (
                <Button type="default" onClick={() => setCurrent(current - 1)}>
                    {__("上一步")}
                </Button>
            )}
            {current < items.length - 1 && (
                <Tooltip
                    placement="topLeft"
                    title={
                        !current
                            ? !serviceInfo.mversion
                                ? __("必填项异常")
                                : !suiteConfigCorrect
                                ? __("服务安装包未上传或所需组件未安装")
                                : ""
                            : ""
                    }
                >
                    <div style={{ display: "inline-block" }}>
                        <Button
                            type="primary"
                            disabled={
                                !current &&
                                (!serviceInfo.mversion || !suiteConfigCorrect)
                            }
                            onClick={handeleNextClick}
                        >
                            {__("下一步")}
                        </Button>
                    </div>
                </Tooltip>
            )}
            {current === items.length - 1 && (
                <Button
                    type="primary"
                    onClick={handleOperationClick}
                    loading={loading}
                    key={loading ? "true" : "false"}
                >
                    {__("确定")}
                </Button>
            )}
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
            contentClassName={styles["service-operation-content"]}
            moduleName={SERVICE_PREFIX}
        >
            <div className={styles["operation-steps"]}>
                <Steps current={current} items={items} />
            </div>
            {getContent()}
        </ContentLayout>
    );
};

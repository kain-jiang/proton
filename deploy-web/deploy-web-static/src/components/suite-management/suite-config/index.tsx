import React, { FC, useEffect, useState } from "react";
import { Row, Col, Button, Tag, Modal } from "@kweaver-ai/ui";
import { VerticalAlignTopOutlined } from "@kweaver-ai/ui/icons";
import styles from "./styles.module.less";
import { ContentLayout, Toolbar } from "../../common/components";
import { SERVICE_PREFIX } from "../config";
import className from "classnames";
import __ from "./locale";
import { ConfigEditStatus, ConfigEditStatusEnum } from "./helper";
import { ServiceFramework } from "../service-framework";
import { ServiceSchemaItem } from "../../../api/suite-management/suite-deploy/declare";
import { serviceJob } from "../../../api/service-management/service-deploy";
import { handleError } from "../utils/handleError";
import { mergeMap, schemaCompose } from "../utils/schemaCompose";
import { SchemaOperateType } from "../../../core/suite-management/suite-deploy";
import { noop } from "lodash";

interface IProps {
    // 套件配置项
    suiteConfig: ServiceSchemaItem[];
    // 修改套件配置项
    changeSuiteConfig: (item: ServiceSchemaItem[]) => void;
    // 是否同步更新
    isSynchronousUpdate: boolean;
    // 操作类型
    operateType?: SchemaOperateType;
    // 系统空间id（仅用于批量任务监控）
    sid?: number;
}

// 用于控制jsonschema触发提交操作
let submitCallback = () => {};
export const SuiteConfig: FC<IProps> = ({
    suiteConfig,
    isSynchronousUpdate,
    changeSuiteConfig,
    operateType,
    sid,
}) => {
    const [currentServiceConfig, setCurrentServiceConfig] =
        useState<ServiceSchemaItem>({} as ServiceSchemaItem);

    const [isFormValidator, setIsFormValidator] = useState<boolean>(false);

    // 默认选中第一个不是禁用状态的服务
    useEffect(() => {
        setCurrentServiceConfig(
            suiteConfig.find(
                (config) => config.editStatus !== ConfigEditStatusEnum.Disabled
            ) || ({} as ServiceSchemaItem)
        );
    }, []);

    // 点击其他服务时，获取配置项
    useEffect(() => {
        if (currentServiceConfig.name) {
            getCurrentSchema();
        }
    }, [currentServiceConfig.name]);

    // isFormValidator为true，代表提交配置项成功
    useEffect(() => {
        if (isFormValidator) {
            changeSuiteConfig(
                suiteConfig.map((config) => {
                    if (config.name === currentServiceConfig.name) {
                        return {
                            ...config,
                            formData: currentServiceConfig.formData,
                            editStatus: ConfigEditStatusEnum.Submitted,
                        };
                    }
                    return config;
                })
            );
        }
    }, [isFormValidator]);

    const getCurrentSchema = async () => {
        // 第一次修改该服务
        if (!currentServiceConfig?.schema) {
            try {
                // 获取套件清单配置项
                const suiteSchemaConfig =
                    await serviceJob.getJSONSchemaSnapshotByName(
                        {
                            name: currentServiceConfig.name,
                            version: currentServiceConfig.version,
                        },
                        sid ? { sid } : undefined
                    );
                // 合并后的配置项
                let composedConfig: any = null;
                if (
                    operateType !== SchemaOperateType.Revert &&
                    operateType !== SchemaOperateType.ChangeConfig
                ) {
                    try {
                        // 获取系统配置项
                        const systemSchemaConfig =
                            await serviceJob.getJSONSchemaSnapshotByName({
                                name: currentServiceConfig.name,
                            });
                        composedConfig = schemaCompose(
                            currentServiceConfig,
                            suiteSchemaConfig,
                            systemSchemaConfig,
                            isSynchronousUpdate
                        );
                    } catch (error: any) {
                        if (error.status === 404) {
                            composedConfig = schemaCompose(
                                currentServiceConfig,
                                suiteSchemaConfig,
                                null,
                                isSynchronousUpdate
                            );
                        } else {
                            throw error;
                        }
                    }
                } else {
                    composedConfig = suiteSchemaConfig;
                }
                setCurrentServiceConfig({
                    ...currentServiceConfig,
                    formData:
                        operateType !== SchemaOperateType.Revert &&
                        operateType !== SchemaOperateType.ChangeConfig
                            ? composedConfig.formData
                            : currentServiceConfig.formData,
                    schema: composedConfig.schema,
                    uiSchema: composedConfig.uiSchema,
                    version: composedConfig.version,
                });
                changeSuiteConfig(
                    suiteConfig.map((config) => {
                        if (config.name === currentServiceConfig.name) {
                            return {
                                ...config,
                                formData:
                                    operateType !== SchemaOperateType.Revert &&
                                    operateType !==
                                        SchemaOperateType.ChangeConfig
                                        ? composedConfig.formData
                                        : currentServiceConfig.formData,
                                schema: composedConfig.schema,
                                uiSchema: composedConfig.uiSchema,
                                version: composedConfig.version,
                            };
                        }
                        return config;
                    })
                );
            } catch (error) {
                handleError(error);
            }
        }
    };

    // 提交配置项
    const handleSubmit = () => {
        submitCallback();
    };

    // 点击其他服务
    const handleChangeService = (serviceInfo: ServiceSchemaItem) => {
        if (
            serviceInfo.name !== currentServiceConfig.name &&
            serviceInfo.editStatus !== ConfigEditStatusEnum.Disabled
        ) {
            if (operateType !== SchemaOperateType.ChangeConfig) {
                setIsFormValidator(false);
                setCurrentServiceConfig(
                    suiteConfig.find(
                        (config) => config.name === serviceInfo.name
                    )!
                );
            } else {
                if (
                    suiteConfig.find(
                        (config) => config.name === currentServiceConfig.name
                    )?.editStatus === ConfigEditStatusEnum.Unsubmitted
                ) {
                    Modal.info({
                        title: __("提示"),
                        content: (
                            <div className={styles["modal-content"]}>
                                {__("当前更改服务未提交配置项，请提交")}
                            </div>
                        ),
                        closable: true,
                        okButtonProps: {
                            style: {
                                display: "none",
                            },
                        },
                    });
                } else {
                    setIsFormValidator(false);
                    setCurrentServiceConfig(
                        suiteConfig.find(
                            (config) => config.name === serviceInfo.name
                        )!
                    );
                }
            }
        }
    };

    const serviceListHeader = (
        <Toolbar
            left={
                <span className={styles["header-title"]}>{__("服务类型")}</span>
            }
            leftSize={24}
            moduleName={SERVICE_PREFIX}
        />
    );
    const serviceConfigHeader = (
        <Toolbar
            left={
                <span className={styles["header-title"]}>
                    {__("服务配置项")}
                </span>
            }
            right={
                <Button
                    type="default"
                    icon={
                        <VerticalAlignTopOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                        />
                    }
                    style={{ marginRight: "16px" }}
                    onClick={handleSubmit}
                >
                    {__("提交配置项")}
                </Button>
            }
            cols={[{ span: 16 }, { span: 8 }]}
            moduleName={SERVICE_PREFIX}
        />
    );

    const getClassName = (
        serviceInfo: ServiceSchemaItem,
        currentServiceConfig: ServiceSchemaItem
    ) => {
        if (serviceInfo.name === currentServiceConfig.name) {
            return className(
                styles["service-content"],
                styles["service-content-choosed"],
                styles["skin-color"]
            );
        } else if (serviceInfo.editStatus === ConfigEditStatusEnum.Disabled) {
            return className(
                styles["service-content"],
                styles["service-content-disabled"]
            );
        } else {
            return styles["service-content"];
        }
    };
    return (
        <Row className={styles["config-container"]}>
            <Col span={6} className={styles["col-list"]}>
                <ContentLayout
                    header={serviceListHeader}
                    moduleName={SERVICE_PREFIX}
                >
                    <div className={styles["service-content-list"]}>
                        {suiteConfig.map((serviceInfo) => {
                            return (
                                <div
                                    className={getClassName(
                                        serviceInfo,
                                        currentServiceConfig
                                    )}
                                    onClick={() =>
                                        handleChangeService(serviceInfo)
                                    }
                                >
                                    <div
                                        className={
                                            styles["service-content-name"]
                                        }
                                    >
                                        <span title={serviceInfo.title}>
                                            {serviceInfo.title}
                                        </span>
                                    </div>
                                    {[
                                        ConfigEditStatusEnum.Submitted,
                                        ConfigEditStatusEnum.Unsubmitted,
                                    ].includes(serviceInfo.editStatus!) ? (
                                        <Tag
                                            color={
                                                ConfigEditStatus[
                                                    serviceInfo.editStatus!
                                                ].color
                                            }
                                            style={{
                                                margin: "9px 0 0 16px",
                                                height: "22px",
                                            }}
                                        >
                                            {
                                                ConfigEditStatus[
                                                    serviceInfo.editStatus!
                                                ].text
                                            }
                                        </Tag>
                                    ) : null}
                                </div>
                            );
                        })}
                    </div>
                </ContentLayout>
            </Col>
            <Col span={18}>
                <ContentLayout
                    header={serviceConfigHeader}
                    moduleName={SERVICE_PREFIX}
                >
                    <ServiceFramework
                        formData={currentServiceConfig.formData}
                        uiSchema={currentServiceConfig.uiSchema}
                        schema={currentServiceConfig.schema}
                        isReadOnly={false}
                        setCallback={(callback: any) => {
                            submitCallback = callback;
                        }}
                        onChangeFormData={(formData) => {
                            setCurrentServiceConfig({
                                ...currentServiceConfig,
                                formData,
                            });
                            changeSuiteConfig(
                                suiteConfig.map((config) => {
                                    if (
                                        config.name ===
                                            currentServiceConfig.name &&
                                        config.editStatus !==
                                            ConfigEditStatusEnum.Submitted
                                    ) {
                                        return {
                                            ...config,
                                            editStatus:
                                                ConfigEditStatusEnum.Unsubmitted,
                                        };
                                    }
                                    return config;
                                })
                            );
                        }}
                        changeIsFormValidator={setIsFormValidator}
                    />
                </ContentLayout>
            </Col>
        </Row>
    );
};

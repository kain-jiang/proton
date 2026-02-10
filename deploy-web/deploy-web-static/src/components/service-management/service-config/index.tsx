import React, { FC, useContext, useEffect, useMemo, useState } from "react";
import { Row, Col, Button, Tag } from "@kweaver-ai/ui";
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
import { noop } from "lodash";
import { DndContext } from "@dnd-kit/core";
import type { DragEndEvent } from "@dnd-kit/core";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import {
    arrayMove,
    SortableContext,
    useSortable,
    verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { SyntheticListenerMap } from "@dnd-kit/core/dist/hooks/utilities";
import { CSS } from "@dnd-kit/utilities";

interface IProps {
    // 服务配置项
    serviceConfig: ServiceSchemaItem[];
    // 修改服务配置项
    changeServiceConfig: (item: ServiceSchemaItem[] | any) => void;
    // 系统空间id
    sid: number;
}

// 用于控制jsonschema触发提交操作
let submitCallback = () => {};

interface RowContextProps {
    setActivatorNodeRef?: (element: HTMLElement | null) => void;
    listeners?: SyntheticListenerMap;
}

const RowContext = React.createContext<RowContextProps>({});

const SortableItem: React.FC<{
    serviceInfo: ServiceSchemaItem;
    currentServiceConfig: ServiceSchemaItem;
    handleChangeService: any;
}> = ({ serviceInfo, currentServiceConfig, handleChangeService }) => {
    const {
        attributes,
        listeners,
        setNodeRef,
        setActivatorNodeRef,
        transform,
        transition,
        isDragging,
    } = useSortable({ id: serviceInfo.aid });

    const style: React.CSSProperties = {
        transform: CSS.Translate.toString(transform),
        // transition,
        ...(isDragging
            ? {
                  position: "relative",
                  zIndex: 9999,
                  border: "1px dashed #dcdcdc",
              }
            : {}),
    };

    const contextValue = useMemo(
        () => ({ setActivatorNodeRef, listeners }),
        [setActivatorNodeRef, listeners]
    );

    const getClassName = (
        serviceInfo: ServiceSchemaItem,
        currentServiceConfig: ServiceSchemaItem
    ) => {
        if (serviceInfo.name === currentServiceConfig.name) {
            return className(
                styles["service-content"],
                styles["service-content-choosed"]
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
        <RowContext.Provider value={contextValue}>
            <div
                className={getClassName(serviceInfo, currentServiceConfig)}
                onClick={() => handleChangeService(serviceInfo)}
                ref={setNodeRef}
                style={style}
                {...attributes}
            >
                <div className={styles["service-content-name"]}>
                    <span title={serviceInfo.title}>{serviceInfo.title}</span>
                </div>
                {[
                    ConfigEditStatusEnum.Submitted,
                    ConfigEditStatusEnum.Unsubmitted,
                ].includes(serviceInfo.editStatus!) ? (
                    <Tag
                        color={ConfigEditStatus[serviceInfo.editStatus!].color}
                        style={{
                            margin: "9px 0 0 16px",
                            height: "22px",
                        }}
                    >
                        {ConfigEditStatus[serviceInfo.editStatus!].text}
                    </Tag>
                ) : null}
            </div>
        </RowContext.Provider>
    );
};

export const ServiceConfig: FC<IProps> = ({
    serviceConfig,
    changeServiceConfig,
    sid,
}) => {
    const [currentServiceConfig, setCurrentServiceConfig] =
        useState<ServiceSchemaItem>({} as ServiceSchemaItem);

    const [isFormValidator, setIsFormValidator] = useState<boolean>(false);

    // 默认选中第一个不是禁用状态的服务
    useEffect(() => {
        setCurrentServiceConfig(
            serviceConfig.find(
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
            changeServiceConfig(
                serviceConfig.map((config) => {
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
                const result = await serviceJob.getJSONSchemaSnapshot(
                    currentServiceConfig.aid,
                    { sid }
                );
                setCurrentServiceConfig({
                    ...currentServiceConfig,
                    formData: result.formData,
                    schema: result.schema,
                    uiSchema: result.uiSchema,
                });
                changeServiceConfig(
                    serviceConfig.map((config) => {
                        if (config.name === currentServiceConfig.name) {
                            return {
                                ...config,
                                formData: result.formData,
                                schema: result.schema,
                                uiSchema: result.uiSchema,
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
            setIsFormValidator(false);
            setCurrentServiceConfig(
                serviceConfig.find(
                    (config) => config.name === serviceInfo.name
                )!
            );
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

    // 拖拽后回调
    const onDragEnd = ({ active, over }: DragEndEvent) => {
        if (active.id !== over?.id) {
            changeServiceConfig((prevState: ServiceSchemaItem[]) => {
                const activeIndex = prevState.findIndex(
                    (record) => record.aid === active.id
                );
                const overIndex = prevState.findIndex(
                    (record) => record.aid === over?.id
                );
                return arrayMove(prevState, activeIndex, overIndex);
            });
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
                        <DndContext
                            modifiers={[restrictToVerticalAxis]}
                            onDragEnd={onDragEnd}
                        >
                            <SortableContext
                                items={serviceConfig.map((i) => i.aid)}
                                strategy={verticalListSortingStrategy}
                            >
                                {serviceConfig.map((serviceInfo) => {
                                    return (
                                        <SortableItem
                                            serviceInfo={serviceInfo}
                                            currentServiceConfig={
                                                currentServiceConfig
                                            }
                                            handleChangeService={
                                                handleChangeService
                                            }
                                        />
                                    );
                                })}
                            </SortableContext>
                        </DndContext>
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
                            changeServiceConfig(
                                serviceConfig.map((config) => {
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

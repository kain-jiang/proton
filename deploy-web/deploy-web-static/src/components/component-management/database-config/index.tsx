import React, { FC, useState, useEffect, useMemo, useRef } from "react";
import {
    Divider,
    Form,
    FormInstance,
    Input,
    Radio,
    RadioChangeEvent,
} from "@kweaver-ai/ui";
import __ from "./locale";
import {
    ComponentName,
    ConfigData,
    DefaultConfigData,
    OperationType,
    SERVICES,
    SOURCE_TYPE,
    buildInComponentsText,
    componentDependenciesMap,
    filterEmptyKey,
    nameValidatorRules,
} from "../helper";
import { componentManage } from "../../../api/component-manage";
import { handleError } from "../../service-management/utils/handleError";
import { ServiceConfig } from "../service-config";

interface IProps {
    // 组件类型
    component: SERVICES;
    // 操作类型
    operationType: OperationType;
    // 组件配置信息
    componentConfigData: ConfigData;
    // 修改组件配置信息
    setComponentConfigData: (componentConfigData: ConfigData | any) => void;
    // 修改组件表单ref
    setComponentForm: (item: any) => void;
    // 是否展示title
    showTitle: boolean;
    // 组件名称
    componentName: string;
}

export const DataBaseConfig: FC<IProps> = ({
    component,
    operationType,
    componentConfigData,
    setComponentConfigData,
    setComponentForm,
    showTitle,
    componentName,
}) => {
    // 存储方式：本地数据路径、存储类
    const [sourceType, setSourceType] = useState(SOURCE_TYPE.INTERNAL);
    // 已部署的节点
    const [nodeOptions, setNodeOptions] = useState<string[]>([]);
    // zookeeper已部署的节点
    const [zookeeperNodeOptions, setZookeeperNodeOptions] = useState<string[]>(
        []
    );
    // zookeeper操作类型
    const [zookeeperOperationType, setZookeeperOperationType] =
        useState(operationType);
    // 原始副本数，用于校验规则
    const [originReplicaCount, setOriginReplicaCount] = useState<number>(0);
    // zookeeper原始副本数
    const [originZookeeperReplicaCount, setOriginZookeeperReplicaCount] =
        useState<number>(0);
    const [hasDatabaseConnectInfo, setHasDatabaseConnectInfo] =
        useState<boolean>(false);

    const form = useRef<FormInstance>(null);

    const getAddKafkaComponentInfo = async () => {
        try {
            // 没有kafka，有zookeeper
            const dependencyResult = await componentManage.getComponentInfo(
                SERVICES.Zookeeper,
                SERVICES.Zookeeper
            );
            setComponentConfigData({
                [SERVICES.Kafka]: {
                    params: DefaultConfigData[SERVICES.Kafka],
                    name: ComponentName[SERVICES.Kafka],
                    dependencies: {
                        zookeeper: dependencyResult.name,
                    },
                },
                [SERVICES.Zookeeper]: dependencyResult,
            });
            setZookeeperNodeOptions(dependencyResult?.params?.hosts || []);
            setOriginZookeeperReplicaCount(
                dependencyResult?.params?.replica_count || 0
            );
            setZookeeperOperationType(OperationType.Edit);
            setSourceType(
                dependencyResult?.params?.storageClassName
                    ? SOURCE_TYPE.EXTERNAL
                    : SOURCE_TYPE.INTERNAL
            );
        } catch (error: any) {
            if (error.status === 404) {
                // 没有kafka和zookeeper
                setComponentConfigData({
                    [SERVICES.Kafka]: {
                        params: DefaultConfigData[SERVICES.Kafka],
                        name: ComponentName[SERVICES.Kafka],
                        dependencies: {
                            zookeeper: ComponentName[SERVICES.Zookeeper],
                        },
                    },
                    [SERVICES.Zookeeper]: {
                        params: DefaultConfigData[SERVICES.Zookeeper],
                    },
                });
                setSourceType(SOURCE_TYPE.INTERNAL);
            } else {
                handleError(error);
            }
        }
    };

    useEffect(() => {
        setComponentForm((componentForm: any) => {
            return {
                ...componentForm,
                nameForm: form,
            };
        });
        if (operationType === OperationType.Add) {
            if (component === SERVICES.Kafka) {
                getAddKafkaComponentInfo();
            } else {
                setComponentConfigData({
                    [component]: {
                        params: {
                            ...DefaultConfigData[component],
                            data_path: DefaultConfigData[component]?.data_path,
                        },
                    },
                });
                setSourceType(SOURCE_TYPE.INTERNAL);
            }
        } else {
            getComponentInfo(component);
        }
    }, []);

    // 单实例模式 强制设置内置组件名称 (注：由于kafka获取初始值是异步操作，因此不生效)
    useEffect(() => {
        if (operationType === OperationType.Add) {
            setComponentConfigData((componentConfigData: ConfigData) => {
                return {
                    ...componentConfigData,
                    [component]: {
                        ...componentConfigData[component],
                        name: ComponentName[component],
                        ...(component === SERVICES.Kafka
                            ? {
                                  dependencies: {
                                      ...componentConfigData[component]
                                          ?.dependencies,
                                      zookeeper:
                                          ComponentName[SERVICES.Zookeeper],
                                  },
                              }
                            : {}),
                    },
                };
            });
        }
    }, []);

    useEffect(() => {
        form.current?.setFieldsValue({
            ...componentConfigData[component],
        });
    }, [
        componentConfigData[component]?.name,
        componentConfigData[component]?.dependencies?.zookeeper,
    ]);

    const getComponentInfo = async (component: SERVICES) => {
        try {
            let result = await componentManage.getComponentInfo(
                component,
                componentName!
            );
            if (component === SERVICES.Kafka) {
                // 存在kafka，一定会安装zookeeper
                const dependencyResult = await componentManage.getComponentInfo(
                    SERVICES.Zookeeper,
                    result?.dependencies?.zookeeper || ""
                );
                result = {
                    ...result,
                    params: {
                        ...result?.params,
                        external_service_list: result?.params
                            ?.external_service_list?.length
                            ? result?.params?.external_service_list
                            : DefaultConfigData[SERVICES.Kafka]
                                  .external_service_list,
                    },
                };
                setComponentConfigData({
                    [SERVICES.Kafka]: result,
                    [SERVICES.Zookeeper]: dependencyResult,
                });
                setZookeeperNodeOptions(dependencyResult?.params?.hosts || []);

                setOriginZookeeperReplicaCount(
                    dependencyResult?.params?.replica_count || 0
                );
            } else if (component === SERVICES.OpenSearch) {
                if (!result?.params?.extraValues?.storage?.repo?.hdfs) {
                    result = {
                        ...result,
                        params: {
                            ...result?.params,
                            extraValues: {
                                ...result?.params?.extraValues,
                                storage: {
                                    ...result?.params?.extraValues?.storage,
                                    repo: {
                                        ...DefaultConfigData[component]
                                            .extraValues?.storage?.repo,
                                    },
                                },
                            },
                        },
                    };
                }
                setComponentConfigData({
                    [component]: result,
                });
                setHasDatabaseConnectInfo(!!result?.params?.username);
            } else {
                // 非kafka组件
                setComponentConfigData({
                    [component]: result,
                });
                setHasDatabaseConnectInfo(!!result?.params?.username);
            }
            setOriginReplicaCount(result?.params?.replica_count || 0);
            setSourceType(
                result?.params?.storageClassName
                    ? SOURCE_TYPE.EXTERNAL
                    : SOURCE_TYPE.INTERNAL
            );
            setNodeOptions(result?.params?.hosts || []);
        } catch (error: any) {
            handleError(error);
        }
    };

    // 只有添加组件才能修改sourcetype
    const handleChangeSourceType = (e: RadioChangeEvent) => {
        setSourceType(e.target.value);
        if (component === SERVICES.Kafka) {
            setComponentConfigData({
                [SERVICES.Kafka]: {
                    params: filterEmptyKey(
                        DefaultConfigData[SERVICES.Kafka],
                        e.target.value
                    ),
                    name: componentConfigData[component]?.name,
                    dependencies: componentConfigData[component]?.dependencies,
                },
                [SERVICES.Zookeeper]: {
                    params: filterEmptyKey(
                        DefaultConfigData[SERVICES.Zookeeper],
                        e.target.value
                    ),
                },
            });
        } else {
            setComponentConfigData({
                [component]: {
                    params: filterEmptyKey(
                        DefaultConfigData[component],
                        e.target.value
                    ),
                    name: componentConfigData[component]?.name,
                },
            });
        }
    };

    return (
        <>
            <Divider orientation="left" orientationMargin="0">
                {__("组件名称")}
            </Divider>
            <Form
                labelAlign="left"
                initialValues={componentConfigData[component]}
                validateTrigger="onBlur"
                ref={form}
            >
                <Form.Item
                    // labelCol={{ span: 4 }}
                    label={__("组件名称")}
                    name="name"
                    required
                    rules={nameValidatorRules}
                >
                    <Input
                        style={{ width: "200px" }}
                        value={componentConfigData[component]?.name}
                        disabled={true}
                        onChange={(e) => {
                            setComponentConfigData({
                                ...componentConfigData,
                                [component]: {
                                    ...componentConfigData[component],
                                    name: e.target.value,
                                },
                            });
                        }}
                    />
                </Form.Item>
                {[SERVICES.Kafka, SERVICES.PolicyEngine].includes(component) ? (
                    <Form.Item
                        label={
                            buildInComponentsText[
                                componentDependenciesMap[component]
                            ] + __("组件名称")
                        }
                        name={[
                            "dependencies",
                            componentDependenciesMap[component],
                        ]}
                        required
                        rules={nameValidatorRules}
                    >
                        <Input
                            style={{ width: "200px" }}
                            value={
                                componentConfigData[component]?.dependencies?.[
                                    componentDependenciesMap[component]
                                ]
                            }
                            disabled={true}
                            onChange={(e) => {
                                setComponentConfigData({
                                    ...componentConfigData,
                                    [component]: {
                                        ...componentConfigData[component],
                                        dependencies: {
                                            ...componentConfigData[component]
                                                ?.dependencies,
                                            [componentDependenciesMap[
                                                component
                                            ]]: e.target.value,
                                        },
                                    },
                                });
                            }}
                        />
                    </Form.Item>
                ) : null}
            </Form>
            <Divider orientation="left" orientationMargin="0">
                {__("存储方式")}
            </Divider>
            <Radio.Group
                style={{
                    margin: "10px 0",
                }}
                value={sourceType}
                disabled={
                    operationType === OperationType.Edit ||
                    component === SERVICES.Nebula ||
                    zookeeperOperationType === OperationType.Edit
                }
                onChange={handleChangeSourceType}
            >
                <Radio value={SOURCE_TYPE.INTERNAL}>{__("本地路径存储")}</Radio>
                <Radio value={SOURCE_TYPE.EXTERNAL}>{__("存储类")}</Radio>
            </Radio.Group>
            <Divider orientation="left" orientationMargin="0">
                {__("服务配置")}
            </Divider>
            <ServiceConfig
                component={component}
                operationType={operationType}
                componentConfigData={componentConfigData}
                setComponentConfigData={setComponentConfigData}
                sourceType={sourceType}
                nodeOptions={nodeOptions}
                setComponentForm={setComponentForm}
                showTitle={showTitle}
                originReplicaCount={originReplicaCount}
                hasDatabaseConnectInfo={hasDatabaseConnectInfo}
            />
            {component === SERVICES.Kafka ? (
                <ServiceConfig
                    component={SERVICES.Zookeeper}
                    operationType={zookeeperOperationType}
                    componentConfigData={componentConfigData}
                    setComponentConfigData={setComponentConfigData}
                    sourceType={sourceType}
                    nodeOptions={zookeeperNodeOptions}
                    setComponentForm={setComponentForm}
                    showTitle={true}
                    originReplicaCount={originZookeeperReplicaCount}
                />
            ) : null}
        </>
    );
};

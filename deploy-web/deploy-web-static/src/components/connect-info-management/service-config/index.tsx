import React, { FC, useEffect, useMemo, useState } from "react";
import {
    Button,
    Divider,
    Form,
    Modal,
    Radio,
    RadioChangeEvent,
    Select,
} from "@kweaver-ai/ui";
import __ from "./locale";
import {
    ConnectInfoServices,
    serviceMapComponent,
    servicesInternalSourceTypeText,
    servicesSourceTypeText,
    transMongoOptions2String,
} from "../helper";
import { componentManage } from "../../../api/component-manage";
import {
    CONNECT_SERVICES,
    DefaultConfigData,
    MQ_TYPE,
    OperationType,
    SERVICES,
    SOURCE_TYPE,
    changeResourceConnectServiceTypeByService,
    getETCDInfoByResourceType,
    getMQInfoByResourceType,
    getMongoDBInfoByResourceType,
    getOpenSearchInfoByResourceType,
    getPolicyEngineInfoByResourceType,
    getRDSInfoByResourceType,
    getRedisInfoByResourceType,
} from "../../component-management/helper";
import { handleError } from "../../service-management/utils/handleError";
import { InformationOutlined } from "@kweaver-ai/ui/lib/icons";
import { noop } from "lodash";
import { DataBaseConfig } from "../../component-management/database-config";
import { ConfigData as ComponentConfigData } from "../../component-management/helper";
import { ConnectInfo } from "../ConnectInfo/index.view";
import { ConnectInfoValidateState } from "../index.d";
import { ComponentListItem } from "../../../api/component-manage/declare";

interface IProps {
    // 服务类型
    service: ConnectInfoServices;
    // 操作类型
    operationType: OperationType;
    // 服务连接信息数据
    serviceConfigData: any;
    // 修改服务连接信息
    setServiceConfigData: (item: any) => void;
    // 内置组件数据
    componentConfigData: ComponentConfigData;
    // 修改内置组件数据
    setComponentConfigData: (item: ComponentConfigData) => void;
    // 修改内置组件表单校验
    setComponentForm: (item: any) => void;
    // 修改连接信息表单校验
    setConnectInfoForm: (item: any) => void;
    // 修改部分单独配置信息校验状态
    updateConnectInfoValidateState: (item: any) => void;
    // 部分单独配置信息校验状态
    connectInfoValidateState: ConnectInfoValidateState;
    // 系统空间id
    sid: number;
    // 资源类型为内置时，是否选择内置组件
    setIsBuildinComponentChosen: (isChosen: boolean) => void;
}
interface OriginConnectInfoData {
    internal: any;
    external: any;
}

export const ServiceConfig: FC<IProps> = ({
    service,
    operationType,
    serviceConfigData,
    setServiceConfigData,
    componentConfigData,
    setComponentConfigData,
    setComponentForm,
    setConnectInfoForm,
    updateConnectInfoValidateState,
    connectInfoValidateState,
    sid,
    setIsBuildinComponentChosen,
}) => {
    // 资源类型（内置、第三方）
    const [sourceType, setSourceType] = useState<string>();
    // 连接信息的原始数据
    const [originConnectInfoData, setOriginConnectInfoData] =
        useState<OriginConnectInfoData>({
            internal: {},
            external: {},
        });
    // 控制连接信息是否可编辑（noop都可编辑，internal外置可编辑，external内置可编辑）
    const [originSourceType, setOriginSourceType] = useState(
        SOURCE_TYPE.INTERNAL
    );
    // 记录第三方连接信息当前类型（mq、redis和rds）
    const [originConnectInfoType, setOriginConnectInfoType] =
        useState<string>("");
    // 内置组件的操作类型
    const [componentOperationType, setComponentOperationType] =
        useState<OperationType>(OperationType.Edit);
    const [mqConfig, setmqConfig] = useState({
        mq_type: MQ_TYPE.KAFKA,
    });
    // 内置组件名称
    const [componentName, setComponentName] = useState("");
    // 原始内置组件名称
    const [originComponentName, setOriginComponentName] = useState("");
    // 可选内置组件列表
    const [componentList, setComponentList] = useState<ComponentListItem[]>([]);

    useEffect(() => {
        getConnectInfo();
    }, []);

    // 修改isBuildinComponentChosen，控制是否禁用【确定】
    useEffect(() => {
        if (
            sourceType === SOURCE_TYPE.INTERNAL &&
            !componentName &&
            componentOperationType === OperationType.Edit &&
            mqConfig.mq_type === MQ_TYPE.KAFKA
        ) {
            setIsBuildinComponentChosen(false);
        } else {
            setIsBuildinComponentChosen(true);
        }
    }, [componentName, sourceType, componentOperationType, mqConfig]);

    // 单实例模式下 直接调用一次，根据componentList元素个数判断是否可以新建内置组件
    useEffect(() => {
        onDropdownVisibleChange(true);
    }, []);

    const getConnectInfo = async () => {
        try {
            if (operationType === OperationType.Add) {
                try {
                    // await componentManage.getComponentInfo(
                    //     serviceMapComponent[service],
                    //     "要改"
                    // );

                    setSourceType(SOURCE_TYPE.INTERNAL);
                    setOriginSourceType(SOURCE_TYPE.NOOP);
                    // setComponentOperationType(OperationType.Edit);

                    // 根据内置组件获取默认连接信息
                    setOriginConnectInfoData({
                        internal: changeResourceConnectServiceTypeByService(
                            false,
                            serviceMapComponent[service]
                        ),
                        external: changeResourceConnectServiceTypeByService(
                            true,
                            serviceMapComponent[service]
                        ),
                    });
                    setServiceConfigData({
                        resource_connect_info: {
                            [service]:
                                changeResourceConnectServiceTypeByService(
                                    false,
                                    serviceMapComponent[service]
                                ),
                        },
                    });
                } catch (err: any) {
                    if (err.status === 404) {
                        if (service === ConnectInfoServices.ETCD) {
                            // etcd没有第三方，没有内置组件时也默认使用内置
                            setSourceType(SOURCE_TYPE.INTERNAL);
                            setServiceConfigData({
                                resource_connect_info: {
                                    [service]:
                                        changeResourceConnectServiceTypeByService(
                                            false,
                                            serviceMapComponent[service]
                                        ),
                                },
                            });
                        } else {
                            setSourceType(SOURCE_TYPE.EXTERNAL);
                            setServiceConfigData({
                                resource_connect_info: {
                                    [service]:
                                        changeResourceConnectServiceTypeByService(
                                            true,
                                            serviceMapComponent[service]
                                        ),
                                },
                            });
                        }
                        setOriginSourceType(SOURCE_TYPE.NOOP);
                        setComponentOperationType(OperationType.Add);
                        // 获取默认连接信息
                        setOriginConnectInfoData({
                            internal: changeResourceConnectServiceTypeByService(
                                false,
                                serviceMapComponent[service]
                            ),
                            external: changeResourceConnectServiceTypeByService(
                                true,
                                serviceMapComponent[service]
                            ),
                        });
                    } else {
                        handleError(err);
                    }
                }
            } else {
                let result = await componentManage.getConnectInfo(
                    service,
                    service,
                    sid
                );
                setSourceType(result?.info?.source_type);
                setOriginSourceType(result?.info?.source_type);

                if (result?.info?.source_type === SOURCE_TYPE.INTERNAL) {
                    if (
                        service !== ConnectInfoServices.MQ ||
                        result?.info.mq_type === SERVICES.Kafka
                    ) {
                        setComponentName(result?.instance?.name);
                        setOriginComponentName(result?.instance?.name);
                        setOriginConnectInfoData({
                            internal: result.info,
                            external: changeResourceConnectServiceTypeByService(
                                true,
                                serviceMapComponent[service]
                            ),
                        });
                        setServiceConfigData({
                            resource_connect_info: {
                                [service]: result.info,
                            },
                        });
                        setComponentOperationType(OperationType.Edit);
                    } else {
                        // 特殊处理内置nsq
                        setOriginConnectInfoData({
                            internal: result.info,
                            external: changeResourceConnectServiceTypeByService(
                                true,
                                serviceMapComponent[service]
                            ),
                        });
                        setServiceConfigData({
                            resource_connect_info: {
                                [service]: result.info,
                            },
                        });
                        setmqConfig({ mq_type: MQ_TYPE.NSQ });
                        // try {
                        //     await componentManage.getComponentInfo(
                        //         SERVICES.Kafka,
                        //         "要改"
                        //     );
                        //     setComponentOperationType(OperationType.Edit);
                        // } catch (err: any) {
                        //     if (err.status === 404) {
                        //         setComponentOperationType(OperationType.Add);
                        //     } else {
                        //         handleError(err);
                        //     }
                        // }
                    }
                } else {
                    if (service === ConnectInfoServices.MongoDB) {
                        result = {
                            ...result,
                            info: transMongoOptions2String(result.info),
                        };
                    }
                    if (
                        result?.info?.mq_type ||
                        result?.info?.rds_type ||
                        result?.info?.connect_type
                    ) {
                        setOriginConnectInfoType(
                            result?.info?.mq_type ||
                                result?.info?.rds_type ||
                                result?.info?.connect_type
                        );
                    }
                    setOriginComponentName("");
                    setComponentName("");
                    setOriginConnectInfoData({
                        internal: changeResourceConnectServiceTypeByService(
                            false,
                            serviceMapComponent[service]
                        ),
                        external: result.info,
                    });
                    setServiceConfigData({
                        resource_connect_info: {
                            [service]: result.info,
                        },
                    });

                    // try {
                    //     await componentManage.getComponentInfo(
                    //         serviceMapComponent[service],
                    //         "要改"
                    //     );
                    //     setComponentOperationType(OperationType.Edit);
                    // } catch (err: any) {
                    //     if (err.status === 404) {
                    //         setComponentOperationType(OperationType.Add);
                    //     } else {
                    //         handleError(err);
                    //     }
                    // }
                }
            }
        } catch (error) {
            handleError(error);
        }
    };
    const handleChangeSourceType = (e: RadioChangeEvent) => {
        Modal.confirm({
            title: __("提示"),
            okText: __("确定"),
            content: __(
                "您确定要切换此资源类型吗？若未完成数据迁移，可能会导致业务服务异常。"
            ),
            onOk: () => handleCustomOk(e.target.value),
            icon: (
                <InformationOutlined
                    style={{ color: "#126EE3" }}
                    onPointerEnterCapture={noop}
                    onPointerLeaveCapture={noop}
                />
            ),
        });
    };

    const handleCustomOk = (val: string) => {
        setSourceType(val);
        if (val === SOURCE_TYPE.INTERNAL) {
            setServiceConfigData({
                resource_connect_info: {
                    [service]: originConnectInfoData.internal,
                },
            });
            setComponentName(originComponentName);
            if (originConnectInfoData.internal.mq_type === MQ_TYPE.NSQ) {
                setmqConfig({ mq_type: MQ_TYPE.NSQ });
            } else {
                setmqConfig({ mq_type: MQ_TYPE.KAFKA });
            }
        } else {
            setServiceConfigData({
                resource_connect_info: {
                    [service]: originConnectInfoData.external,
                },
            });
            setComponentForm({});
            setComponentConfigData({});
            setComponentName("");
        }
    };

    const handleChangeMQType = (val: string) => {
        setmqConfig((mqConfig) => {
            return { ...mqConfig, mq_type: val };
        });
        // val取值需满足SERVICES的值
        setServiceConfigData({
            resource_connect_info: {
                [service]: changeResourceConnectServiceTypeByService(
                    false,
                    val
                ),
            },
        });
        if (val === MQ_TYPE.NSQ) {
            setComponentForm({});
            setComponentConfigData({});
            setComponentName("");
        } else {
            setComponentName(originComponentName);
        }
    };
    /**
     * 更新连接信息
     * @param key 连接信息键
     * @param curInfo 连接信息新值
     */
    const onUpdateConnectInfo = (key: string, curInfo: any, type?: string) => {
        let info: any;
        if (key === CONNECT_SERVICES.RDS) {
            if (type) {
                if (type === originConnectInfoType) {
                    info = {
                        [key]: originConnectInfoData.external,
                    };
                } else {
                    info = {
                        [key]: changeResourceConnectServiceTypeByService(
                            true,
                            SERVICES.MariaDB,
                            type
                        ),
                    };
                }
            } else {
                info = {
                    [key]: getRDSInfoByResourceType(curInfo),
                };
            }
        } else if (key === CONNECT_SERVICES.MONGODB) {
            info = {
                [key]: getMongoDBInfoByResourceType(curInfo),
            };
        } else if (key === CONNECT_SERVICES.REDIS) {
            if (type) {
                if (type === originConnectInfoType) {
                    info = {
                        [key]: originConnectInfoData.external,
                    };
                } else {
                    info = {
                        [key]: changeResourceConnectServiceTypeByService(
                            true,
                            SERVICES.Redis,
                            type
                        ),
                    };
                }
            } else {
                info = {
                    [key]: getRedisInfoByResourceType(curInfo),
                };
            }
        } else if (key === CONNECT_SERVICES.MQ) {
            if (type) {
                if (type === originConnectInfoType) {
                    info = {
                        [key]: originConnectInfoData.external,
                    };
                } else {
                    info = {
                        [key]: changeResourceConnectServiceTypeByService(
                            true,
                            SERVICES.Kafka,
                            type
                        ),
                    };
                }
            } else {
                info = {
                    [key]: getMQInfoByResourceType(curInfo),
                };
            }
        } else if (key === CONNECT_SERVICES.OPENSEARCH) {
            info = {
                [key]: getOpenSearchInfoByResourceType(curInfo),
            };
        } else if (key === CONNECT_SERVICES.POLICY_ENGINE) {
            info = {
                [key]: getPolicyEngineInfoByResourceType(curInfo),
            };
        } else if (key === CONNECT_SERVICES.ETCD) {
            info = {
                [key]: getETCDInfoByResourceType(curInfo),
            };
        }
        setServiceConfigData({
            resource_connect_info: info,
        });
    };

    const onDropdownVisibleChange = async (open: boolean) => {
        if (open) {
            try {
                const { data } = (await componentManage.getComponentList({
                    offset: 0,
                    limit: 10000,
                    type: [serviceMapComponent[service]],
                    nobind: true,
                })) as any;
                setComponentList(data);
            } catch (e) {
                handleError(e);
            }
        }
    };

    // 修改绑定的内置组件
    const handleChangeComponentName = (name: string) => {
        setComponentName(name);
        setComponentForm({});
        setComponentConfigData({});
        setComponentOperationType(OperationType.Edit);
    };

    // 新建内置组件
    const handleAddComponent = () => {
        setComponentName("");
        setComponentForm({});
        setComponentConfigData({});
        setComponentOperationType(OperationType.Add);
    };

    return (
        <>
            <Divider orientation="left" orientationMargin="0">
                {__("资源类型")}
            </Divider>
            <Radio.Group
                style={{
                    margin: "10px 0",
                }}
                value={sourceType}
                onChange={handleChangeSourceType}
                disabled={service === ConnectInfoServices.ETCD}
            >
                <Radio value={SOURCE_TYPE.INTERNAL}>{`${__("内置")}${
                    servicesInternalSourceTypeText[service]
                }`}</Radio>
                <Radio value={SOURCE_TYPE.EXTERNAL}>{`${__("第三方")}${
                    servicesSourceTypeText[service]
                }`}</Radio>
            </Radio.Group>
            {service === ConnectInfoServices.MQ &&
            sourceType === SOURCE_TYPE.INTERNAL ? (
                <Form.Item
                    labelCol={{ span: 2 }}
                    labelAlign="left"
                    label={__("MQ类型")}
                    required
                >
                    <Select
                        showArrow
                        style={{
                            width: "200px",
                        }}
                        optionLabelProp="label"
                        placeholder={__("请选择 MQ 类型")}
                        value={mqConfig.mq_type}
                        onSelect={(val) => handleChangeMQType(val)}
                        getPopupContainer={(node) =>
                            node.parentElement || document.body
                        }
                    >
                        <Select.Option
                            key={MQ_TYPE.KAFKA}
                            label={MQ_TYPE.KAFKA}
                        >
                            {MQ_TYPE.KAFKA}
                        </Select.Option>
                        <Select.Option key={MQ_TYPE.NSQ} label={MQ_TYPE.NSQ}>
                            {MQ_TYPE.NSQ}
                        </Select.Option>
                    </Select>
                </Form.Item>
            ) : null}
            {sourceType === SOURCE_TYPE.INTERNAL &&
            mqConfig.mq_type === MQ_TYPE.KAFKA ? (
                <Form.Item
                    labelCol={{ span: 2 }}
                    labelAlign="left"
                    label={__("内置组件")}
                >
                    <Select
                        style={{
                            width: "200px",
                        }}
                        disabled={!!originComponentName}
                        placeholder={__("请选择内置组件")}
                        getPopupContainer={(node) =>
                            node.parentElement || document.body
                        }
                        value={componentName || undefined}
                        onChange={(val) => handleChangeComponentName(val)}
                        onDropdownVisibleChange={onDropdownVisibleChange}
                        options={componentList.map((component) => ({
                            label: component.name,
                            value: component.name,
                        }))}
                    />
                    <Button
                        style={{ marginLeft: "6px" }}
                        type="link"
                        disabled={
                            !!originComponentName || !!componentList.length
                        }
                        onClick={handleAddComponent}
                    >
                        {__("新建内置组件")}
                    </Button>
                </Form.Item>
            ) : null}

            {sourceType === SOURCE_TYPE.INTERNAL &&
            mqConfig.mq_type === MQ_TYPE.KAFKA &&
            ((componentName && componentOperationType === OperationType.Edit) ||
                componentOperationType === OperationType.Add) ? (
                <DataBaseConfig
                    key={componentName}
                    component={serviceMapComponent[service]}
                    operationType={componentOperationType}
                    componentConfigData={componentConfigData}
                    setComponentConfigData={setComponentConfigData}
                    setComponentForm={setComponentForm}
                    showTitle={true}
                    componentName={componentName}
                />
            ) : null}
            <ConnectInfo
                key={sourceType}
                configData={serviceConfigData}
                onUpdateConnectInfo={onUpdateConnectInfo}
                updateConnectInfoForm={setConnectInfoForm}
                updateConnectInfoValidateState={updateConnectInfoValidateState}
                connectInfoValidateState={connectInfoValidateState}
                originSourceType={originSourceType}
                originConnectInfoType={originConnectInfoType}
            />
        </>
    );
};

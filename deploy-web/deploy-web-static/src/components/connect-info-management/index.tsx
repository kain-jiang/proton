import React, { FC, useState, useEffect, useMemo } from "react";
import {
    Dropdown,
    Button,
    Space,
    Drawer,
    message,
    Table,
    Divider,
    Form,
    Select,
    Input,
} from "@kweaver-ai/ui";
import { CaretDownOutlined, PlusOutlined } from "@kweaver-ai/ui/icons";
import type {
    MenuProps,
    TableColumnType,
    TableColumnsType,
} from "@kweaver-ai/ui";
import { componentManage } from "../../api/component-manage";
import { handleError } from "../service-management/utils/handleError";
import styles from "./styles.module.less";
import __ from "./locale";
import {
    ConfigData,
    ConnectInfoData,
    ConnectInfoServices,
    connectInfoServices,
    defaultAddableServices,
    serviceMapComponent,
    servicesText,
    transMongoOptions2Object,
    transMongoOptions2String,
    updateServiceLog,
} from "./helper";
import {
    ConfigData as ComponentConfigData,
    ComponentName,
    DefaultConnectInfoValidateState,
    MQ_AUTH_MACHANISM,
    MQ_TYPE,
    OPENSEARCH_VERSION,
    RDS_TYPE,
    REDIS_CONNECT_TYPE,
    SERVICES,
    SOURCE_TYPE,
    ValidateState,
    OperationType,
} from "../component-management/helper";
import { ServiceConfig } from "./service-config";
import { ConnectInfoValidateState } from "./index.d";
import { deployMiniPathname } from "../../core/path";
import { assignTo } from "../../tools/browser";
import { noop } from "lodash";
import { CustomSpin } from "../component-management/components/spin";
import { ContentLayout } from "../common/components";
import { SERVICE_PREFIX } from "../component-management/config";
import { formatTableResponse } from "../common/utils/request";
import {
    ConnectInfoListItem,
    IGetConnectInfoListParams,
    IGetConnectInfoListTableParams,
} from "../../api/component-manage/declare";
import { safetyStr } from "../common/utils/string";
import { SystemConfig } from "../../api/tenant-management/declare";
import { system } from "../../api/tenant-management";

export const ConnectInfoManagement: FC = () => {
    // 操作类型
    const [operationType, setOperationType] = useState<OperationType>(
        OperationType.Edit
    );
    // 服务类型
    const [serviceType, setServiceType] = useState<ConnectInfoServices>(
        ConnectInfoServices.RDS
    );
    // 控制抽屉开关
    const [open, setOpen] = useState<boolean>(false);
    // 服务连接信息数据
    const [serviceConfigData, setServiceConfigData] = useState<{
        resource_connect_info: any;
    }>({
        resource_connect_info: {},
    });
    // 服务对应内置组件数据
    const [componentConfigData, setComponentConfigData] =
        useState<ComponentConfigData>({} as ComponentConfigData);
    // 内置组件表单校验ref
    const [componentForm, setComponentForm] = useState({});
    // 连接信息表单校验ref
    const [connectInfoForm, setConnectInfoForm] = useState({});
    // 连接信息部分单独信息校验
    const [connectInfoValidateState, setConnectInfoValidateState] = useState(
        DefaultConnectInfoValidateState
    );
    // 是否正在保存配置
    const [isLoading, setIsLoading] = useState<boolean>(false);
    // 系统空间filter筛选项开关
    const [filterOpen, setFilterOpen] = useState(false);
    // 添加/更新连接信息时绑定的系统空间id
    const [sid, setSid] = useState(0);
    // 系统空间列表
    const [systemList, setSystemList] = useState<SystemConfig[]>([]);
    // 当资源类型是内置时，是否选择内置组件
    const [isBuildinComponentChosen, setIsBuildinComponentChosen] =
        useState(false);
    // 原始操作类型，表示操作入口
    const [originOperationType, setOriginOperationType] =
        useState<OperationType>(OperationType.Add);

    // 表格的基本设置
    const { state, api } = Table.useTable<
        ConnectInfoListItem,
        IGetConnectInfoListTableParams,
        ConnectInfoListItem[]
    >({
        request: (params) => {
            const { _filter, current, pageSize } = params;
            const formatedParams: IGetConnectInfoListParams = {
                offset: (current - 1) * pageSize,
                limit: pageSize,
                type: getFilterParams(
                    (_filter as any)?.type,
                    Object.values(connectInfoServices).length
                ),
                sid: [-1, undefined].includes((_filter as any)?.systemName?.[0])
                    ? undefined
                    : (_filter as any)?.systemName?.[0],
            };
            return componentManage.getConnectInfoList(formatedParams);
        },
        rowKey: (record) => `${record.name}_${record.sid}`,
        pagination: {
            showTotal: (total) => __("共${total}条", { total }),
        },
        ...formatTableResponse(),
    });

    const { reload } = api;

    const getFilterParams = (
        value: Array<any> | undefined,
        filterOptionCount: number
    ) => {
        if (value?.length) {
            if (value.length === filterOptionCount) {
                return undefined;
            } else {
                return value;
            }
        }
        return undefined;
    };

    // 表格的列配置项
    const columns: TableColumnsType<ConnectInfoListItem> = [
        {
            title: __("连接信息名称"),
            dataIndex: "name",
        },
        {
            title: __("连接信息类型"),
            dataIndex: "type",
            filters: connectInfoServices.map((connectInfoService) => {
                return {
                    value: connectInfoService,
                    text: servicesText[connectInfoService],
                };
            }),
            render: (value) => servicesText[value],
            tooltip: (value) => servicesText[value],
        },
        {
            title: __("系统空间"),
            dataIndex: "systemName",
            render: (text) => safetyStr(text),
            tooltip: (text) => safetyStr(text),
        },
        {
            title: __("系统空间ID"),
            dataIndex: "sid",
            render: (value) => safetyStr(value),
            tooltip: (value) => safetyStr(value),
        },
        {
            title: __("操作"),
            render: (_, record: ConnectInfoListItem) => (
                <Button
                    type="link"
                    onClick={() => handleEditConnectInfo(record)}
                >
                    {__("编辑")}
                </Button>
            ),
            tooltip: () => __("编辑"),
        },
    ];

    /**
     * 更新连接配置校验状态
     */
    const updateConnectInfoValidateState = (
        value: Partial<ConnectInfoValidateState>
    ) => {
        setConnectInfoValidateState((preState) => {
            return {
                ...preState,
                ...value,
            };
        });
    };

    const clickHerf = () => {
        assignTo(deployMiniPathname.serviceDeployPathname);
    };

    const handleUpdateService = () => {
        // 校验内置组件表单
        const componentFormCheck = Object.values(componentForm).map(
            async (form: any) => {
                return await form?.current?.validateFields();
            }
        );
        // 校验连接信息表单
        const connectInfoFormCheck = Object.values(connectInfoForm).map(
            async (form: any) => {
                return await form?.current?.validateFields();
            }
        );
        // 校验部分单独配置信息
        const configCheck = new Promise<void>((resolve, reject) => {
            const resource_connect_info =
                serviceConfigData?.resource_connect_info;
            // rds类型
            if (
                resource_connect_info?.rds &&
                resource_connect_info?.rds.source_type ===
                    SOURCE_TYPE.EXTERNAL &&
                !Object.values(RDS_TYPE).includes(
                    resource_connect_info?.rds?.rds_type
                )
            ) {
                updateConnectInfoValidateState({
                    RDS_TYPE: ValidateState.Empty,
                });
                reject();
            }
            // mongodb ssl
            if (
                resource_connect_info?.mongodb &&
                resource_connect_info?.mongodb.source_type ===
                    SOURCE_TYPE.EXTERNAL &&
                ![true, false].includes(resource_connect_info?.mongodb?.ssl)
            ) {
                updateConnectInfoValidateState({
                    MONGODB_SSL: ValidateState.Empty,
                });
                reject();
            }
            // redis连接模式
            if (
                resource_connect_info?.redis &&
                resource_connect_info?.redis.source_type ===
                    SOURCE_TYPE.EXTERNAL &&
                !Object.values(REDIS_CONNECT_TYPE).includes(
                    resource_connect_info?.redis?.connect_type
                )
            ) {
                updateConnectInfoValidateState({
                    REDIS_CONNECT_TYPE: ValidateState.Empty,
                });
                reject();
            }
            // mq资源类型
            if (
                resource_connect_info?.mq &&
                !Object.values(SOURCE_TYPE).includes(
                    resource_connect_info?.mq?.source_type
                )
            ) {
                updateConnectInfoValidateState({
                    MQ_RADIO: ValidateState.Empty,
                });
                reject();
            }
            // mq类型
            if (
                resource_connect_info?.mq &&
                !Object.values(MQ_TYPE).includes(
                    resource_connect_info?.mq?.mq_type
                )
            ) {
                updateConnectInfoValidateState({
                    MQ_TYPE: ValidateState.Empty,
                });
                reject();
            }
            // mq认证机制
            if (
                resource_connect_info?.mq?.auth &&
                !Object.values(MQ_AUTH_MACHANISM).includes(
                    resource_connect_info?.mq?.auth?.mechanism
                )
            ) {
                updateConnectInfoValidateState({
                    MQ_AUTH_MACHANISM: ValidateState.Empty,
                });
                reject();
            }
            // opensearch版本
            if (
                resource_connect_info?.opensearch &&
                !Object.values(OPENSEARCH_VERSION).includes(
                    resource_connect_info?.opensearch?.version
                )
            ) {
                updateConnectInfoValidateState({
                    OPENSEARCH_VERSION: ValidateState.Empty,
                });
                reject();
            }
            resolve();
        });

        Promise.all([
            ...componentFormCheck,
            ...connectInfoFormCheck,
            configCheck,
        ])
            .then(async () => {
                let payload: ConnectInfoData = {
                    name: serviceType,
                    info: serviceConfigData.resource_connect_info?.[
                        serviceType
                    ],
                    sid,
                };
                if (
                    serviceType === ConnectInfoServices.MongoDB &&
                    payload?.info.source_type === SOURCE_TYPE.EXTERNAL
                ) {
                    payload = {
                        ...payload,
                        info: transMongoOptions2Object(payload?.info),
                    };
                }
                if (
                    serviceType === ConnectInfoServices.RDS &&
                    payload?.info.source_type === SOURCE_TYPE.EXTERNAL
                ) {
                    // 将admin_user和admin_passwd转换为base64
                    const newInfo = {
                        ...payload?.info,
                        admin_key: payload?.info?.auto_create_database
                            ? btoa(
                                  unescape(
                                      encodeURIComponent(
                                          `${payload?.info?.admin_user}:${payload?.info?.admin_passwd}`
                                      )
                                  )
                              )
                            : undefined,
                    };
                    delete newInfo.auto_create_database;
                    delete newInfo.admin_user;
                    delete newInfo.admin_passwd;
                    payload = {
                        ...payload,
                        info: newInfo,
                    };
                }
                if (
                    [
                        ConnectInfoServices.RDS,
                        ConnectInfoServices.MongoDB,
                    ].includes(serviceType)
                ) {
                    if (payload?.info.source_type === SOURCE_TYPE.INTERNAL) {
                        const instance = {
                            ...componentConfigData[
                                serviceMapComponent[serviceType]
                            ],
                            // name: ComponentName[
                            //     serviceMapComponent[serviceType]
                            // ],
                            type: serviceMapComponent[serviceType],
                        };
                        const info = {
                            ...payload.info,
                            username: instance?.params?.username,
                            password: instance?.params?.password,
                        };
                        payload = { ...payload, instance, info };
                    }
                } else if (serviceType !== ConnectInfoServices.MQ) {
                    if (payload?.info.source_type === SOURCE_TYPE.INTERNAL) {
                        const instance = {
                            ...componentConfigData[
                                serviceMapComponent[serviceType]
                            ],
                            // name: ComponentName[
                            //     serviceMapComponent[serviceType]
                            // ],
                            type: serviceMapComponent[serviceType],
                        };
                        payload = { ...payload, instance };
                    }
                } else {
                    if (
                        payload?.info.source_type === SOURCE_TYPE.INTERNAL &&
                        payload?.info.mq_type === MQ_TYPE.KAFKA
                    ) {
                        const kafkaInstance = {
                            ...componentConfigData[SERVICES.Kafka],
                            // name: ComponentName[SERVICES.Kafka],
                            type: SERVICES.Kafka,
                        };
                        const zookeeperInstance = {
                            ...componentConfigData[SERVICES.Zookeeper],
                            name: componentConfigData[SERVICES.Kafka]
                                ?.dependencies?.zookeeper,
                            type: SERVICES.Zookeeper,
                        };
                        payload = {
                            ...payload,
                            instance: kafkaInstance,
                            zookeeper: zookeeperInstance,
                        };
                    }
                }
                try {
                    setIsLoading(true);
                    await componentManage.putConnectInfo(serviceType, payload);
                    updateServiceLog(operationType, payload?.info, serviceType);
                    message.success(
                        <div style={{ display: "inline-block" }}>
                            {__(
                                "修改成功，配置将在服务更新后生效。请前往【服务管理-服务部署】页面更新服务。"
                            )}
                            <a
                                style={{
                                    marginLeft: "10px",
                                    color: "#126EE3",
                                }}
                                onClick={clickHerf}
                            >
                                {__("立即前往")}
                            </a>
                        </div>
                    );
                    reload();
                    handleClose();
                } catch (error) {
                    handleError(error);
                } finally {
                    setIsLoading(false);
                }
            })
            .catch(() => {});
    };

    const handleClose = () => {
        setOpen(false);
        setServiceConfigData({ resource_connect_info: {} });
        setComponentConfigData({});
        setComponentForm({});
        setConnectInfoForm({});
        setConnectInfoValidateState(DefaultConnectInfoValidateState);
        setSid(0);
        setOriginOperationType(OperationType.Add);
    };

    // 处理切换系统空间
    const handleSidChange = async (sid: number) => {
        try {
            const { totalNum } = (await componentManage.getConnectInfoList({
                offset: 0,
                limit: 10,
                type: [serviceType],
                sid: sid,
            })) as any;
            setOperationType(totalNum ? OperationType.Edit : OperationType.Add);
        } catch (e) {
            handleError(e);
        }

        setSid(sid);
        setServiceConfigData({ resource_connect_info: {} });
        setComponentConfigData({});
        setComponentForm({});
        setConnectInfoForm({});
        setConnectInfoValidateState(DefaultConnectInfoValidateState);
    };

    const handleMenuClick: MenuProps["onClick"] = (e) => {
        // setOperationType(OperationType.Add);
        setServiceType(e.key as ConnectInfoServices);
        setOpen(true);
        getSystemList();
    };

    // 编辑连接信息
    const handleEditConnectInfo = (record: ConnectInfoListItem) => {
        setOperationType(OperationType.Edit);
        setServiceType(record.type as ConnectInfoServices);
        setOpen(true);
        setSid(record.sid);
        setOriginOperationType(OperationType.Edit);
        getSystemList();
    };

    // 获取系统空间列表
    const getSystemList = async () => {
        try {
            const data = await system.get({
                offset: 0,
                limit: 10000,
                mode: true,
            });
            setSystemList(data);
        } catch (e) {
            handleError(e);
        }
    };

    const items: MenuProps["items"] = useMemo(() => {
        return defaultAddableServices.map((addableService) => {
            return {
                label: servicesText[addableService],
                key: addableService,
            };
        });
    }, []);

    const menuProps = {
        items,
        onClick: handleMenuClick,
    };

    const header = (
        <div className={styles["dropdown"]}>
            <Dropdown menu={menuProps} trigger={["click"]}>
                <Button type="primary">
                    <Space>
                        <PlusOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                        />
                        {__("添加或更新服务")}
                        <CaretDownOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                        />
                    </Space>
                </Button>
            </Dropdown>
        </div>
    );

    return (
        <>
            <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
                <Table {...state} columns={columns} />
            </ContentLayout>
            <Drawer
                title={
                    operationType === OperationType.Add
                        ? __("添加或更新-${service}", {
                              service: servicesText[serviceType],
                          })
                        : __("编辑-${service}", {
                              service: servicesText[serviceType],
                          })
                }
                width={1000}
                onClose={handleClose}
                onOk={handleUpdateService}
                okButtonProps={{ disabled: !sid || !isBuildinComponentChosen }}
                open={open}
                destroyOnClose
            >
                <Divider orientation="left" orientationMargin="0">
                    {__("实例信息")}
                </Divider>
                <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("连接信息名称")}
                    required
                >
                    <Input
                        style={{ width: "200px" }}
                        value={serviceType}
                        disabled
                    />
                </Form.Item>
                <Form.Item
                    labelCol={{ span: 4 }}
                    labelAlign="left"
                    label={__("系统空间")}
                    required
                >
                    <Select
                        style={{
                            width: "200px",
                        }}
                        disabled={originOperationType === OperationType.Edit}
                        placeholder={__("请选择系统空间")}
                        value={sid || undefined}
                        onChange={(val) => handleSidChange(val)}
                        getPopupContainer={(node) =>
                            node.parentElement || document.body
                        }
                        options={systemList.map((system) => ({
                            label: system.systemName,
                            value: system.sid,
                        }))}
                    />
                </Form.Item>
                {sid ? (
                    <ServiceConfig
                        key={sid}
                        sid={sid}
                        service={serviceType}
                        operationType={operationType}
                        serviceConfigData={serviceConfigData}
                        setServiceConfigData={setServiceConfigData}
                        componentConfigData={componentConfigData}
                        setComponentConfigData={setComponentConfigData}
                        setComponentForm={setComponentForm}
                        setConnectInfoForm={setConnectInfoForm}
                        updateConnectInfoValidateState={
                            updateConnectInfoValidateState
                        }
                        connectInfoValidateState={connectInfoValidateState}
                        setIsBuildinComponentChosen={
                            setIsBuildinComponentChosen
                        }
                    />
                ) : null}
            </Drawer>
            {isLoading ? <CustomSpin text={__("保存配置中...")} /> : null}
        </>
    );
};

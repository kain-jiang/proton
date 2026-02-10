import React, { FC, useEffect, useMemo, useState } from "react";
import { ContentLayout, Toolbar } from "../../../../common/components";
import {
    Form,
    Input,
    Button,
    Drawer,
    Search,
    Table,
    Select,
    Switch,
} from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { handleError } from "../../../utils/handleError";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import {
    ApplicationItem,
    ServiceSchemaItem,
} from "../../../../../api/suite-management/suite-deploy/declare";
import { OperationType } from "../type.d";
import styles from "./styles.module.less";
import { SERVICE_PREFIX } from "../../../config";
import __ from "./locale";
import { AppsData, ComponentsData, SuiteInfo } from "../suite-info";
import { suiteManifests } from "../../../../../api/suite-management/suite-deploy";
import { componentManage } from "../../../../../api/component-manage";
import {
    AppsUploadStatusEnum,
    ComponentsInstallStatusEnum,
} from "../../type.d";
import { serviceApplication } from "../../../../../api/service-management/service-deploy";
import { ConfigEditStatusEnum } from "../../../suite-config/helper";
import { QuestionCircleOutlined } from "@kweaver-ai/ui/lib/icons";
import { noop } from "lodash";
import { FormLabel } from "../../../../common/components/form-label";

interface IProps {
    // 操作服务的类型（安装或更新）
    operationType: OperationType;
    // 选中的服务信息
    serviceInfo: ApplicationItem;
    // 选中名称服务所有版本的信息
    serviceInfos: ApplicationItem[];
    // 是否同步更新
    isSynchronousUpdate: boolean;
    // 修改选中的服务
    changeServiceInfo: (item: ApplicationItem) => void;
    // 修改选中名称服务所有版本的信息
    changeServiceInfos: (item: ApplicationItem[]) => void;
    // 修改套件配置状态
    changeSuiteConfigCorrect: (item: boolean) => void;
    // 修改套件配置项
    changeSuiteConfig: (item: ServiceSchemaItem[]) => void;
    // 修改同步更新开关
    changeIsSynchronousUpdate: (item: boolean) => void;
}
export const ChooseService: FC<IProps> = ({
    operationType,
    serviceInfo,
    serviceInfos,
    isSynchronousUpdate,
    changeServiceInfo,
    changeServiceInfos,
    changeSuiteConfigCorrect,
    changeSuiteConfig,
    changeIsSynchronousUpdate,
}) => {
    // 控制滑窗
    const [open, setOpen] = useState<boolean>(false);
    // 滑窗输入框过滤
    const [filter, setFilter] = useState<string>("");
    // 滑窗中点击的套件名称
    const [selectedServiceInfo, setSelectedServiceInfo] = useState<{
        title: string;
        name: string;
    }>({ title: "", name: "" });
    // 完整应用名称表格数据
    const [dataSource, setDataSource] = useState<ApplicationItem[]>([]);
    // 当前表格页码
    const [current, setCurrent] = useState<number>(1);
    // 组件安装table数据
    const [componentsData, setComponentsData] = useState<ComponentsData[]>([]);
    // 服务上传table数据
    const [appsData, setAppsData] = useState<AppsData[]>([]);

    // 过滤后的表格数据
    const filteredDataSource = useMemo(() => {
        return dataSource.filter((item) => {
            return item?.title?.includes(filter);
        });
    }, [dataSource, filter]);
    // 格式化服务列表数据
    const formatServiceInfos = useMemo(() => {
        return serviceInfos.map((serviceInfo) => {
            return {
                value: serviceInfo.mversion,
                label: serviceInfo.mversion,
            };
        });
    }, [serviceInfos]);
    // 更新服务时获取版本列表
    useEffect(() => {
        if (operationType === ServiceMode.Update) {
            getServiceInfo(serviceInfo.mname);
        }
    }, [serviceInfo.mname]);

    // 选中套件版本变化时，获取服务上传状态和组件安装状态
    useEffect(() => {
        if (serviceInfo.mversion) {
            getSuiteInfo(serviceInfo);
        }
    }, [serviceInfo.mversion]);

    const getSuiteInfo = async (serviceInfo: ApplicationItem) => {
        try {
            changeSuiteConfigCorrect(false);
            let hasError = false;
            const suiteManifestsData = await suiteManifests.getSuiteManifests(
                serviceInfo.mname,
                serviceInfo.mversion
            );
            const apps = suiteManifestsData?.config?.apps || [];
            changeSuiteConfig(
                apps.map((app) => {
                    return {
                        ...app,
                        editStatus: ConfigEditStatusEnum.Unsubmitted,
                    };
                })
            );
            const pcomponents = suiteManifestsData?.config?.pcomponents || [];
            // 获取服务是否上传
            const appFormatData = await Promise.all(
                apps.map(async (app) => {
                    try {
                        await serviceApplication.getApplicationUploadStatus(
                            app.name,
                            app.version
                        );
                        return {
                            title: app.title,
                            status: AppsUploadStatusEnum.UPLOADED,
                            version: app.version,
                        };
                    } catch (error: any) {
                        hasError = true;
                        if (error.status === 404) {
                            return {
                                title: app.title,
                                status: AppsUploadStatusEnum.UNUPLOADED,
                                version: app.version,
                            };
                        } else {
                            throw error;
                        }
                    }
                })
            );

            // 获取组件是否安装
            const componentFormatData = await Promise.all(
                pcomponents.map(async (pcomponent) => {
                    try {
                        await componentManage.getConnectInfo(
                            pcomponent.type,
                            pcomponent.type
                        );
                        return {
                            type: pcomponent.type,
                            status: ComponentsInstallStatusEnum.INSTALLED,
                        };
                    } catch (error: any) {
                        hasError = true;
                        if (error.status === 404) {
                            return {
                                type: pcomponent.type,
                                status: ComponentsInstallStatusEnum.UNINSTALLED,
                            };
                        } else {
                            throw error;
                        }
                    }
                })
            );

            setAppsData(appFormatData);
            setComponentsData(componentFormatData);
            changeSuiteConfigCorrect(!hasError);
        } catch (error) {
            handleError(error);
            changeSuiteConfigCorrect(false);
        }
    };

    // 表格的列配置项
    const columns: TableColumnsType<ApplicationItem> = [
        {
            title: __("名称"),
            dataIndex: "title",
        },
    ];
    /**
     * @description 确定滑窗
     */
    const onOk = () => {
        if (selectedServiceInfo.title !== serviceInfo.title) {
            changeServiceInfo({
                title: selectedServiceInfo.title,
                mname: selectedServiceInfo.name,
                mversion: "",
            });
            getServiceInfo(selectedServiceInfo.name);
        }
        // 重置滑窗数据
        onCancel();
    };
    /**
     * @description 获取同一名称不同版本服务列表
     * @param {string} name 服务名称
     */
    const getServiceInfo = async (name: string) => {
        try {
            const res = await suiteManifests.getApplication({
                offset: 0,
                limit: 10000,
                name,
            });
            changeServiceInfos(res.data);
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @description 取消滑窗
     */
    const onCancel = () => {
        setOpen(false);
        setSelectedServiceInfo({ title: "", name: "" });
        setFilter("");
        setCurrent(1);
    };
    /**
     * @description 打开滑窗
     */
    const handleChooseClick = () => {
        setOpen(true);
        getApplicationItems();
    };
    /**
     * @description 获取滑窗内不同名称的应用列表
     */
    const getApplicationItems = async () => {
        try {
            const res = await suiteManifests.getApplication({
                offset: 0,
                limit: 10000,
                nowork: true,
            });
            setDataSource(res.data);
        } catch (error: any) {
            handleError(error);
        }
    };

    const onFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFilter(e.target.value);
        setCurrent(1);
    };

    const rowSelection = {
        onChange: (_: React.Key[], selectedRows: ApplicationItem[]) => {
            setSelectedServiceInfo({
                name: selectedRows[0].mname,
                title: selectedRows[0].title,
            });
        },
    };

    /**
     * @description 更改下拉表单
     * @param value
     * @param selectedItem 单选选中的服务
     */
    const handleSelectChange = (value: any, selectedItem: any) => {
        changeServiceInfo({
            ...serviceInfo,
            mversion: selectedItem.label,
        });
        changeIsSynchronousUpdate(false);
    };

    const header = (
        <Toolbar
            right={<Search value={filter} onChange={onFilterChange} debounce />}
            rightSize={24}
            moduleName={SERVICE_PREFIX}
        />
    );

    return (
        <React.Fragment>
            <Form labelCol={{ flex: "130px" }} labelAlign="left">
                {operationType === ServiceMode.Install ? (
                    <Form.Item
                        label={<FormLabel text={__("选择安装的套件")} />}
                        required={true}
                    >
                        <Input
                            value={serviceInfo.title}
                            className={styles["install-form-input"]}
                            disabled
                        />
                        <Button type="default" onClick={handleChooseClick}>
                            {__("选择")}
                        </Button>
                    </Form.Item>
                ) : null}
                <Form.Item
                    label={
                        <FormLabel
                            text={
                                operationType === ServiceMode.Install
                                    ? __("选择安装的版本")
                                    : __("选择版本")
                            }
                        />
                    }
                    required={true}
                >
                    <Select
                        value={serviceInfo.mversion || ""}
                        style={{ width: 350 }}
                        onChange={handleSelectChange}
                        listItemHeight={32}
                        options={formatServiceInfos}
                    />
                </Form.Item>
                {operationType === ServiceMode.Update ? (
                    <Form.Item
                        label={<FormLabel text={__("是否同步更新")} />}
                        required={true}
                    >
                        <Switch
                            checked={isSynchronousUpdate}
                            onChange={(val) => changeIsSynchronousUpdate(val)}
                        />
                        <QuestionCircleOutlined
                            onPointerEnterCapture={noop}
                            onPointerLeaveCapture={noop}
                            style={{
                                marginLeft: "6px",
                                verticalAlign: "middle",
                            }}
                            title={__("同步更新当前系统内最新运行的配置内容")}
                        />
                    </Form.Item>
                ) : null}
            </Form>
            {serviceInfo.mversion ? (
                <SuiteInfo
                    componentsData={componentsData}
                    appsData={appsData}
                />
            ) : null}
            {operationType === ServiceMode.Install && (
                <Drawer
                    title={__("选择应用")}
                    open={open}
                    onOk={onOk}
                    onClose={onCancel}
                    okButtonProps={{
                        disabled: !selectedServiceInfo.name,
                    }}
                    destroyOnClose
                >
                    <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
                        <Table
                            rowKey={"mname"}
                            dataSource={filteredDataSource}
                            columns={columns}
                            scroll={{
                                y: "calc(100vh - 300px)",
                            }}
                            rowSelection={{
                                type: "radio",
                                ...rowSelection,
                                columnWidth: 33,
                            }}
                            pagination={{
                                current,
                                showQuickJumper: true,
                                showSizeChanger: true,
                                showTotal: (total) =>
                                    __("共${total}条", { total }),
                                onChange: (page) => setCurrent(page),
                            }}
                        />
                    </ContentLayout>
                </Drawer>
            )}
        </React.Fragment>
    );
};

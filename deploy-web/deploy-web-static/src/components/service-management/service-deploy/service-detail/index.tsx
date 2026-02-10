import React, { FC, useEffect, useState } from "react";
import { ServiceTitle } from "./service-title";
import { ContentLayout, Toolbar } from "../../../common/components";
import { ServiceTable } from "./service-table";
import { UpdateRecord } from "./update-record";
import { ServiceFramework } from "../../service-framework";
import { Button, Breadcrumb, Tabs } from "@kweaver-ai/ui";
import className from "classnames";
import { handleError } from "../../utils/handleError";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import {
    serviceComponent,
    serviceJob,
} from "../../../../api/service-management/service-deploy";
import {
    ComponentItem,
    ComponentType,
    ServiceJSONSchemaItem,
} from "../../../../api/service-management/service-deploy/declare";
import type { ServiceModeType } from "./type";
import { ApplicationItem } from "../../../../api/service-management/service-deploy/declare";
import { IBaseProps } from "../declare";
import { formatTable, formatServiceTitle } from "../../utils/formatTable";
import { SERVICE_PREFIX } from "../../config";
import styles from "./styles.module.less";
import __ from "./locale";

interface IProps extends IBaseProps {
    // 主服务id
    mainServiceId: number;
    // 当前服务id
    serviceId: number;
    // 更改当前服务id
    changeServiceId: (id: number) => void;
    // 修改更新服务记录
    changeUpdateServiceRecord?: (record: ApplicationItem) => void;
    // 修改更新任务jid
    changeJid?: (jid: number) => void;
    // 修改主服务id
    changeMainServiceId?: (id: number) => void;
    // 当前服务类型
    serviceModeType: ServiceModeType;
    // 修改系统空间id
    changeSid?: (sid: number) => void;
}
// 初始化数据
const initialServiceDetailInfo: ServiceJSONSchemaItem = {
    id: 0,
    name: "",
    title: "",
    version: "",
    status: 1,
    comment: "",
    aid: 0,
    startTime: -1,
    endTime: -1,
    createTime: -1,
};
const initialComponentInfo: ComponentItem = {
    formData: {
        component: null,
    } as ComponentType,
    schema: {},
};
export const ServiceDetail: FC<IProps> = ({
    changeServiceMode,
    changeServiceId,
    changeUpdateServiceRecord,
    changeJid,
    changeMainServiceId,
    mainServiceId,
    serviceId,
    serviceModeType,
    changeSid,
}) => {
    const [active, setActive] = useState<string>("1");
    // 服务详细信息
    const [serviceDetailInfo, setServiceDetailInfo] =
        useState<ServiceJSONSchemaItem>(initialServiceDetailInfo);
    // 微服务详细信息
    const [componentInfo, setComponentInfo] =
        useState<ComponentItem>(initialComponentInfo);
    // 微服务依赖信息
    const [componentDependenceInfos, setComponentDependenceInfos] = useState<
        ComponentType[]
    >([]);
    // 获取数据
    useEffect(() => {
        if (serviceModeType === ServiceMode.Service) {
            getServiceDetailInfo();
        } else {
            getComponentInfo();
            getComponentDependenceInfos();
        }
    }, [serviceId, serviceModeType]);
    /**
     * @description 获取服务详情信息
     */
    const getServiceDetailInfo = async () => {
        try {
            const res = await serviceJob.getJSONSchema(serviceId);
            setServiceDetailInfo(res);
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @description 获取微服务详细信息
     */
    const getComponentInfo = async () => {
        try {
            const res = await serviceComponent.getComponentInfo(serviceId);
            setComponentInfo(res);
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @description 获取微服务依赖信息
     */
    const getComponentDependenceInfos = async () => {
        try {
            const res = await serviceComponent.getComponentDependence(
                serviceId
            );
            setComponentDependenceInfos(res);
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @function
     * @description 切换tab事件
     * @param {string} value
     * @return {void}
     */
    const onTabChange = (value: string): void => {
        setActive(value);
    };
    /**
     * @function
     * @description 面包屑跳转事件
     * @param {ServiceMode} value
     * @return {void}
     */
    const handleBreadcrumbChange = (value: ServiceMode) => {
        return () => {
            if (value === ServiceMode.Home) {
                // 返回服务管理页面
                changeServiceMode(ServiceMode.Home);
            } else {
                // 返回服务详情页面
                changeServiceMode(ServiceMode.Service);
                changeServiceId(mainServiceId);
            }
        };
    };
    /**
     * @description 跳转到更新页面
     */
    const handleUpdateClick = () => {
        changeServiceMode(ServiceMode.Update);
        changeUpdateServiceRecord!({
            name: serviceDetailInfo.name,
            title: serviceDetailInfo.title,
            version: serviceDetailInfo.version,
            aid: serviceDetailInfo.aid,
        });
        changeSid!(serviceDetailInfo.sid!);
    };

    const breadcrumb = (
        <Breadcrumb separator=">" className={styles["service-breadcrumb"]}>
            <Breadcrumb.Item
                className={styles["service-breadcrumb-item"]}
                onClick={handleBreadcrumbChange(ServiceMode.Home)}
            >
                {__("服务管理")}
            </Breadcrumb.Item>
            {serviceModeType === ServiceMode.Service ? (
                <Breadcrumb.Item
                    className={className(
                        styles["breadcrumb-current"],
                        styles["skin-color"]
                    )}
                >
                    {__("服务详情")}
                </Breadcrumb.Item>
            ) : (
                <React.Fragment>
                    <Breadcrumb.Item
                        onClick={handleBreadcrumbChange(ServiceMode.Service)}
                        className={styles["service-breadcrumb-item"]}
                    >
                        {__("服务详情")}
                    </Breadcrumb.Item>
                    <Breadcrumb.Item
                        className={className(
                            styles["breadcrumb-current"],
                            styles["skin-color"]
                        )}
                    >
                        {__("组件详情")}
                    </Breadcrumb.Item>
                </React.Fragment>
            )}
        </Breadcrumb>
    );

    const header =
        serviceModeType === ServiceMode.Service ? (
            <Toolbar
                left={
                    <ServiceTitle
                        serviceDetailInfo={serviceDetailInfo}
                        serviceModeType={ServiceMode.Service}
                    />
                }
                right={
                    <Button type="default" onClick={handleUpdateClick}>
                        {__("更新")}
                    </Button>
                }
                cols={[{ span: 22 }, { span: 2 }]}
                moduleName={SERVICE_PREFIX}
            />
        ) : (
            <Toolbar
                left={
                    <ServiceTitle
                        serviceDetailInfo={formatServiceTitle(
                            componentInfo.formData
                        )}
                        serviceModeType={ServiceMode.MicroService}
                    />
                }
                leftSize={24}
                moduleName={SERVICE_PREFIX}
            />
        );
    /**
     * @description 获取Tabs内容
     * @param {ServiceModeType} serviceModeType 当前服务为主服务还是微服务
     * @return Tabs组件的items数组
     */
    const getTabItems = (serviceModeType: ServiceModeType) => {
        const tabItems = [
            {
                label:
                    serviceModeType === ServiceMode.Service
                        ? __("组件")
                        : __("依赖服务"),
                key: "1",
                children: (
                    <ServiceTable
                        changeServiceMode={changeServiceMode}
                        changeServiceId={changeServiceId}
                        serviceModeType={serviceModeType}
                        dataSource={
                            serviceModeType === ServiceMode.Service
                                ? serviceDetailInfo.formData
                                    ? formatTable(
                                          serviceDetailInfo.formData.components
                                      )
                                    : []
                                : formatTable(componentDependenceInfos)
                        }
                    />
                ),
            },
            {
                label: __("配置项"),
                key: "2",
                children: (
                    <ContentLayout moduleName={SERVICE_PREFIX}>
                        <ServiceFramework
                            formData={
                                serviceModeType === ServiceMode.Service
                                    ? serviceDetailInfo.formData!
                                    : componentInfo.formData
                            }
                            schema={
                                serviceModeType === ServiceMode.Service
                                    ? serviceDetailInfo.schema!
                                    : componentInfo.schema
                            }
                            uiSchema={
                                serviceModeType === ServiceMode.Service
                                    ? serviceDetailInfo.uiSchema || {}
                                    : componentInfo.uiSchema || {}
                            }
                            isReadOnly={true}
                        />
                    </ContentLayout>
                ),
            },
        ];
        // 服务详情页面中包括更新记录
        if (serviceModeType === ServiceMode.Service) {
            tabItems.push({
                label: __("更新记录"),
                key: "3",
                children: (
                    <UpdateRecord
                        changeServiceMode={changeServiceMode}
                        changeServiceId={changeServiceId}
                        changeMainServiceId={changeMainServiceId!}
                        changeUpdateServiceRecord={changeUpdateServiceRecord!}
                        changeJid={changeJid!}
                        serviceId={serviceId}
                        serviceName={serviceDetailInfo.name}
                        sid={serviceDetailInfo.sid!}
                        changeSid={changeSid!}
                    />
                ),
            });
        }
        return tabItems;
    };

    return (
        <ContentLayout
            header={header}
            breadcrumb={breadcrumb}
            moduleName={SERVICE_PREFIX}
        >
            <Tabs
                className="service-tab-container"
                defaultActiveKey={String(active) || "1"}
                onChange={onTabChange}
                items={getTabItems(serviceModeType)}
                destroyInactiveTabPane
            />
        </ContentLayout>
    );
};

import React, { FC, useEffect, useState } from "react";
import { ServiceTitle } from "./service-title";
import { ContentLayout, Toolbar } from "../../../common/components";
import { ServiceTable } from "./service-table";
import { UpdateRecord } from "./update-record";
import { Button, Breadcrumb, Tabs } from "@kweaver-ai/ui";
import className from "classnames";
import { handleError } from "../../utils/handleError";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import {
    ApplicationItem,
    SuiteItem,
} from "../../../../api/suite-management/suite-deploy/declare";
import { IBaseProps } from "../declare";
import { SERVICE_PREFIX } from "../../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { composeJob } from "../../../../api/suite-management/suite-deploy";

interface IProps extends IBaseProps {
    // 当前服务id
    serviceId: number;
    // 更改当前服务id
    changeServiceId: (id: number) => void;
    // 修改更新服务记录
    changeUpdateServiceRecord?: (record: ApplicationItem) => void;
    // 修改更新任务jid
    changeJid?: (jid: number) => void;
}
// 初始化数据
const initialServiceDetailInfo: SuiteItem = {
    jid: 0,
    jname: "",
    title: "",
    mversion: "",
    status: 1,
    startTime: -1,
    endTime: -1,
    createTime: -1,
    config: {
        apps: [],
        pcomponents: [],
    },
};

export const ServiceDetail: FC<IProps> = ({
    changeServiceMode,
    changeServiceId,
    changeUpdateServiceRecord,
    changeJid,
    serviceId,
}) => {
    const [active, setActive] = useState<string>("1");
    // 套件详细信息
    const [serviceDetailInfo, setServiceDetailInfo] = useState<SuiteItem>(
        initialServiceDetailInfo
    );

    // 获取数据
    useEffect(() => {
        getServiceDetailInfo();
    }, [serviceId]);
    /**
     * @description 获取服务详情信息
     */
    const getServiceDetailInfo = async () => {
        try {
            const res = await composeJob.getJobInfo(serviceId);
            setServiceDetailInfo(res);
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
    const handleBreadcrumbChange = () => {
        // 返回服务管理页面
        changeServiceMode(ServiceMode.Home);
    };
    /**
     * @description 跳转到更新页面
     */
    const handleUpdateClick = () => {
        changeServiceMode(ServiceMode.Update);
        changeUpdateServiceRecord!({
            mname: serviceDetailInfo.jname,
            title: serviceDetailInfo.title!,
            mversion: serviceDetailInfo.mversion,
        });
    };

    const breadcrumb = (
        <Breadcrumb separator=">" className={styles["service-breadcrumb"]}>
            <Breadcrumb.Item
                className={styles["service-breadcrumb-item"]}
                onClick={handleBreadcrumbChange}
            >
                {__("套件管理")}
            </Breadcrumb.Item>
            <Breadcrumb.Item
                className={className(
                    styles["breadcrumb-current"],
                    styles["skin-color"]
                )}
            >
                {__("套件详情")}
            </Breadcrumb.Item>
        </Breadcrumb>
    );

    const header = (
        <Toolbar
            left={<ServiceTitle serviceDetailInfo={serviceDetailInfo} />}
            right={
                <Button type="default" onClick={handleUpdateClick}>
                    {__("更新")}
                </Button>
            }
            cols={[{ span: 20 }, { span: 4 }]}
            moduleName={SERVICE_PREFIX}
        />
    );

    /**
     * Tabs内容
     */
    const tabItems = [
        {
            label: __("服务信息"),
            key: "1",
            children: (
                <ServiceTable
                    dataSource={serviceDetailInfo?.config?.apps || []}
                />
            ),
        },
        {
            label: __("更新记录"),
            key: "2",
            children: (
                <UpdateRecord
                    changeServiceMode={changeServiceMode}
                    changeServiceId={changeServiceId}
                    changeUpdateServiceRecord={changeUpdateServiceRecord!}
                    changeJid={changeJid!}
                    serviceId={serviceId}
                    serviceName={serviceDetailInfo.jname}
                />
            ),
        },
    ];

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
                items={tabItems}
                destroyInactiveTabPane
            />
        </ContentLayout>
    );
};

import __ from "../../locale";
import React from "react";
import {
    accessConfigurationPathname,
    deployMiniPathname,
    componentManagementPathname,
    connectInfoManagementPathname,
} from "../path";
// import { Site } from "mediator";
import { ReactComponent as ServiceDeployIcon } from "./logo-assets/serviceDeploy.svg";
import { ReactComponent as TaskMonitorIcon } from "./logo-assets/taskMonitor.svg";
import { ServiceDeploy } from "../../components/service-management/service-deploy";
import { ComponentManagement } from "../../components/component-management";
import { ConnectInfoManagement } from "../../components/connect-info-management";
import { ServiceTaskMonitor } from "../../components/service-management/service-task-monitor";
import { getIcon } from "./method";
import HTTPSConfig from "../../components/SiteConfig/component.view";

/**
 * warning: 新增|修改 一级路由名称之后，nginx/default.conf需要做相应的修改。
 */
/**
 * 动态获取 defaultMeunsItem
 * @param moduleConfigs 模块化配置
 * @returns
 */
export const getMenusItems = () => {
    return [
        {
            label: __("环境与资源"),
            key: "source",
            type: "group",
            children: [
                {
                    label: __("访问配置"),
                    key: accessConfigurationPathname,
                    path: accessConfigurationPathname,
                    icon: getIcon(
                        "//ip:port/deploy/static/media/access-configuration.svg"
                    ),
                    slave: true,
                    app: () => <HTTPSConfig />,
                },
                {
                    label: __("连接信息管理"),
                    key: connectInfoManagementPathname,
                    path: connectInfoManagementPathname,
                    icon: getIcon(
                        "//ip:port/deploy/static/media/connectInfoManagement.svg"
                    ),
                    slave: true,
                    app: () => <ConnectInfoManagement />,
                },
                {
                    label: __("内置组件管理"),
                    key: componentManagementPathname,
                    path: componentManagementPathname,
                    icon: getIcon(
                        "//ip:port/deploy/static/media/componentManagement.svg"
                    ),
                    slave: true,
                    app: () => <ComponentManagement />,
                },
            ],
        },
        {
            label: __("服务管理"),
            key: "service-management",
            type: "group",
            children: [
                {
                    label: __("服务部署"),
                    key: deployMiniPathname.serviceDeployPathname,
                    path: deployMiniPathname.serviceDeployPathname,
                    icon: getIcon(
                        "//ip:port/deploy/static/media/serviceDeploy.svg"
                    ),
                    slave: true,
                    app: () => <ServiceDeploy />,
                },
                {
                    label: __("任务监控"),
                    key: deployMiniPathname.taskMonitorPathname,
                    path: deployMiniPathname.taskMonitorPathname,
                    icon: getIcon(
                        "//ip:port/deploy/static/media/taskMonitor.svg"
                    ),
                    slave: true,
                    app: () => <ServiceTaskMonitor />,
                },
            ],
        },
    ];
};

/**
 * 获取最小安装框架
 * @returns
 */
export const getMenusItemsForDeployMini = () => {
    return [
        {
            label: __("服务管理"),
            key: "service-management",
            type: "group",
            children: [
                {
                    label: __("服务部署"),
                    key: deployMiniPathname.serviceDeployPathname,
                    path: deployMiniPathname.serviceDeployPathname,
                    icon: <ServiceDeployIcon />,
                    slave: true,
                    app: () => <ServiceDeploy />,
                },
                {
                    label: __("任务监控"),
                    key: deployMiniPathname.taskMonitorPathname,
                    path: deployMiniPathname.taskMonitorPathname,
                    icon: <TaskMonitorIcon />,
                    slave: true,
                    app: () => <ServiceTaskMonitor />,
                },
            ],
        },
    ];
};

import cookie from "js-cookie";
import { URLPrefixFormatter, URL_PREFIX_MODE } from "../url";
import {
    environment,
    accessConfiguration,
    serviceManagement,
    serviceDeploy,
    taskMonitor,
    componentManagement,
    connectInfoManagement,
    suiteManagement,
    suiteTaskMonitor,
} from "../../locale";
import { session } from "../mediator";
import { UserInfo } from "../../api/oauth/declare";
import { SystemRoleType } from "../roles";

const prefix = "deploy";
// 增加前缀
let customPrefix =
    cookie.get("X-Forwarded-Prefix") || session.get("X-Forwarded-Prefix") || "";
customPrefix = URLPrefixFormatter(customPrefix, URL_PREFIX_MODE.tail);

// 获取涉及deploy-mini的URI路径(套件管理和服务管理)
class DeployMiniPathname {
    get deployMiniCustomPrefix() {
        const deployMiniPrefix =
            cookie.get("X-Forwarded-Prefix") ||
            session.get("X-Forwarded-Prefix") ||
            "";
        return URLPrefixFormatter(deployMiniPrefix, URL_PREFIX_MODE.tail);
    }

    /**
     * 服务管理路径
     */
    get serviceManagementPathname() {
        return `${this.deployMiniCustomPrefix}/${prefix}/${space2connector(
            serviceManagement[2]
        )}`;
    }

    get serviceDeployPathname() {
        return `${this.serviceManagementPathname}/${space2connector(
            serviceDeploy[2]
        )}`;
    }

    get taskMonitorPathname() {
        return `${this.serviceManagementPathname}/${space2connector(
            taskMonitor[2]
        )}`;
    }
}

export const deployMiniPathname = new DeployMiniPathname();

/**
 * 格式化路径
 * @param str
 * @returns
 */
export function space2connector(str: string) {
    // 1. 清理特殊字符 2. 空格替换为连接符 3. 转小写
    return str.replace(/&/, "").replace(/\s+/g, "-").toLowerCase();
}

export const consolePath = "/console/";

export let defaultPathList = [`/${prefix}`, `/${prefix}/`];

export const environmentPathname = `${customPrefix}/${prefix}/${space2connector(
    environment[2]
)}`;
export const suiteManagementPathname = `${customPrefix}/${prefix}/${space2connector(
    suiteManagement[2]
)}`;

// 环境与资源
export const accessConfigurationPathname = `${environmentPathname}/${space2connector(
    accessConfiguration[2]
)}`;
export const connectInfoManagementPathname = `${environmentPathname}/${space2connector(
    connectInfoManagement[2]
)}`;
export const componentManagementPathname = `${environmentPathname}/${space2connector(
    componentManagement[2]
)}`;

// 套件管理
export const suiteTaskMonitorPathname = `${suiteManagementPathname}/${space2connector(
    suiteTaskMonitor[2]
)}`;

export const setupDefaultPath = function (cb: (path: string) => any) {
    defaultPathList = defaultPathList.map(cb);
};

export const getFirstPathname = (userInfo: UserInfo) => {
    if (
        userInfo?.user?.roles?.some((role: any) => {
            return [SystemRoleType.Supper, SystemRoleType.Admin].includes(
                role?.id
            );
        })
    ) {
        return `${customPrefix}/deploy/information-security/auth/user-org`;
    } else if (
        userInfo?.user?.roles?.some((role: any) => {
            return [SystemRoleType.Securit, SystemRoleType.OrgManager].includes(
                role?.id
            );
        })
    ) {
        return `${customPrefix}/deploy/information-security/auth/user-org`;
    } else {
        return `${customPrefix}/deploy/information-security/audit/auditlog`;
    }
};

export const deployMiniList = [
    deployMiniPathname.serviceDeployPathname,
    deployMiniPathname.taskMonitorPathname,
];

import { request } from "../../tools/request";
import { LogItemType } from "./declare.d";

/**
 * 日志类型
 */
export const enum LogEnumType {
    // 管理日志
    Management = "management",
    // 登录日志
    Login = "login",
    // 操作日志
    Operation = "operation",
}

/**
 * 用户类型
 */
export const enum UserType {
    // 实名用户
    AuthenticatedUser = "authenticated_user",
    // 匿名用户
    AnonymousUser = "anonymous_user",
    // 应用账户
    App = "app",
    // 内部服务
    InternalService = "internal_service",
}

class AuditLog {
    url = "/api/deployweb/audit-log/v1/log";

    /**
     * 记录审计日志
     */
    log(category: LogEnumType, configs: LogItemType): Promise<null> {
        return request.post(`${this.url}/${category}`, configs);
    }
}

export const auditLog = new AuditLog();

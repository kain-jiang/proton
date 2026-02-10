import { getUTCTime } from "../mediator/date";
import { session } from "../mediator";
import { LogEnumType, UserType, auditLog } from "../../api/audit-log";
import { v4 as uuidv4 } from "uuid";
import { request } from "../../tools/request";

/**
 * 记录日志
 * @param logType 日志类别
 * @param level 日志级别
 * @param opType 操作类型
 * @param msg 内容
 * @param exMsg 附加信息
 * @param userId 用户id
 */
async function log({ logType, level, opType, msg = "", exMsg = "" }: any) {
    const headers = await request.head("/interface/deployweb/meta");

    auditLog.log(logType, {
        user_id: session.get("deploy.userid"),
        user_name: session.get("deploy.username"),
        user_type: UserType.AuthenticatedUser,
        level,
        op_type: opType,
        date: getUTCTime(headers["x-server-time"]) * 1000,
        ip: headers["x-tclient-addr"],
        msg: msg.trim(),
        ex_msg: exMsg.trim(),
        out_biz_id: uuidv4(),
    });
}

export const loginLog = ({ ...params }) =>
    log({ ...params, logType: LogEnumType.Login });

// 管理日志
export const manageLog = ({ ...params }) =>
    log({ ...params, logType: LogEnumType.Management });

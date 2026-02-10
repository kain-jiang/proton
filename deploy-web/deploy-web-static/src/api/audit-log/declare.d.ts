/**
 * 审计日志信息类型
 */
export interface LogItemType {
    // 用户ID
    user_id: string;
    // 用户名
    user_name?: string;
    // 用户类型
    user_type: string;
    // 日志级别
    level: number;
    // 日志生成时间（微秒，16位时间戳）
    date: number;
    // 操作者IP
    ip?: string;
    // 操作者设备地址
    mac?: string;
    // 日志描述
    msg: string;
    // 附加信息
    ex_msg?: string;
    // 用户终端类型
    user_agent?: string;
    // 额外信息
    additional_info?: string;
    // 日志类型
    op_type: number;
    // 外部业务ID
    out_biz_id: string;
    // 用户所在部门
    dept_paths?: string;
}

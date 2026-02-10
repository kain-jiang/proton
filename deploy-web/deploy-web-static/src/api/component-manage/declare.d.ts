// 获取组件列表表格参数
export interface IGetComponentListTableParams {
    type?: string[];
    sid?: number;
    // 组件未被绑定
    nobind?: boolean;
}

// 获取组件列表接口参数
export interface IGetComponentListParams extends IGetComponentListTableParams {
    offset: number;
    limit: number;
}

export interface ComponentListItem {
    sid: number;
    namespace: string;
    systemName: string;
    type: string;
    name: string;
}

// 获取连接信息列表表格参数
export interface IGetConnectInfoListTableParams {
    // 连接信息类型
    type?: string[];
    // 系统空间id
    sid?: number;
    // 连接信息名称
    name?: string;
}

// 获取连接信息列表接口参数
export interface IGetConnectInfoListParams
    extends IGetConnectInfoListTableParams {
    offset: number;
    limit: number;
}

// 连接信息数据
export interface ConnectInfoListItem extends ComponentListItem {
    config: object;
}

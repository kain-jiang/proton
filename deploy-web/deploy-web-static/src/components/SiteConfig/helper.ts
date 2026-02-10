import __ from "./locale";

// 类型
export enum Parts {
    App = "app",
    Config = "config",
}

// 证书分类
export const PartsText = {
    [Parts.App]: __("证书配置"),
    [Parts.Config]: __("访问配置"),
};

// 签名类型
export enum SignType {
    Self = "AnyShare",
    CA = "CA",
}

// 证书类型
export const SignTypeText = {
    [SignType.Self]: __("自签名证书"),
    [SignType.CA]: __("CA证书"),
};

// 页面状态
export enum PageState {
    Info,
    Edit,
}

// 是否是自签名证书
export const isSelfSignCert = (issuer: any) => {
    return issuer === SignType.Self || issuer.indexOf("aishu.cn") !== -1;
};

// 地址类型
export const enum AddrType {
    App = "app",
    Storage = "storage",
}

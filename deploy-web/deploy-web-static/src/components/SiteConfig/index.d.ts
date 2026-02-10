export interface Props extends React.ClassAttributes<void> {
    // 修改页面状态
    // changePageState: () => any;
}

export interface State {
    // 证书信息
    certsInfo: ReadonlyArray<CertInfo>;

    // 文档域访问配置信息
    documentConfigInfo: {
        // 访问地址
        host: string;
        // 端口
        port: string;
    };

    /**
     * 当前最新状态的web客户端和对象存储的 http https 端口值
     */
    currentAppPorts: {
        /**
         * Web客户端访问https端口输入值
         */
        webClientHttps: number | string;

        /**
         * 对象存储https端口输入值
         */
        objStorageHttps: number | string;
    };

    /**
     * 旧的 web客户端和对象存储的 http https 端口值
     */
    oldAppPorts: {
        /**
         * Web客户端访问https端口输入值
         */
        webClientHttps: number | string;

        /**
         * 对象存储https端口输入值
         */
        objStorageHttps: number | string;
    };

    // 页面状态
    pageState: PageState;

    //是否允许下载证书
    isCertDownload: boolean;
}

export type CertInfo = {
    // 颁发者
    accepter: string;

    // 对象存储还是app
    certType: string;

    // 过期时间
    expireDate: string;

    // 是否已经过期
    hasExpired: boolean;

    // 颁发给
    issuer: string;

    // 开始时间
    startDate: string;
};

export type PageState = {
    // 应用
    app: number;
};

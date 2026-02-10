export interface Props extends React.ClassAttributes<void> {
    /**
     * 当应用组件中的web客户端 http/https 端口改变时调用该函数，用于更改最新状态的端口值
     */
    changeAppConfigWebClientPorts: (
        webClientHttps: number | string,
        oldWebClientHttps: number | string
    ) => void;

    /**
     * 当应用组件中的对象存储 http/https 端口改变时调用该函数，用于更改最新状态的端口值
     */
    changeAppConfigObjPorts: (
        objStorageHttps: number | string,
        oldObjStorageHttps: number | string
    ) => void;

    // 修改页面状态
    changePageState: (newState) => any;
}

export interface State {
    /**
     * 应用服务中访问地址输入值
     */
    appServiceAccessingAddress: string;

    /**
     * Web客户端访问端口输入值
     */
    webClientPort: {
        /**
         * Web客户端访问https端口输入值
         */
        https: number | string;
    };

    /**
     * 前缀
     */
    path: string;

    /**
     * 访问地址类型
     */
    type: string;

    /**
     * 应用服务中访问地址ValidateBox验证状态
     */
    appServiceAccessingAddressStatus: number;

    /**
     * web客户端访问https端口ValidateBox验证状态
     */
    webClientHttpsStatus: number;

    /**
     * 弹窗状态
     */
    dialogStatus: number;

    /**
     * 若应用服务被改变则 保存/取消 按钮出现
     */
    isAppServiceChanged: boolean;

    /**
     * ErrorDialog 中的错误提示
     */
    errorMessage: string;

    /**
     * 转圈圈组件状态
     */
    loadingStatus: boolean;
}

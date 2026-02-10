import * as PropTypes from "prop-types";
import getErrorMessage from "../../../core/mediator/error";
import WebComponent from "../../webcomponent";
import { isIP } from "validator";
import { defaultPathList } from "../../../core/path";
import { PageState } from "../helper";
import { signup } from "../../../core/auth";
import { Props, State } from "./index.d";
import __ from "./locale";
import { accessAddr } from "../../../api/deploy-manager";
import { AppType } from "../../../api/deploy-manager";
import { message } from "antd";

export enum AppServiceAccessingAddressStatus {
    /**
     * 验证合法,没有气泡提示
     */
    Normal,

    /**
     * 没有启用应用节点
     */
    AppNodeEmpty,

    /**
     * 内外网映射地址为空
     */
    Empty,

    /**
     * 内外网映射地址不合法,包含英文字母,认为输入是 域名
     */
    ErrorEnglishLetter,

    /**
     * 内外网映射地址不合法,不包含英文字母,认为输入是 IP
     */
    ErrorNoEnglish,
}

export enum WebHttpValidateState {
    /**
     * 验证合法,没有气泡提示
     */
    Normal,

    /**
     * 没有启用应用节点
     */
    AppNodeEmpty,

    /**
     * web客户端访问端口http或者https输入为空
     */
    Empty,

    /**
     * 输入不合法
     */
    InputError,
}

export enum WebHttpsValidateState {
    Normal,
    AppNodeEmpty,
    Empty,
    InputError,
}

export enum DialogStatus {
    /**
     * 合法
     */
    None,

    /**
     * 不合法,ErrorDialog 弹窗出现
     */
    ErrorDialogAppear,

    /**
     * AlertBox 弹窗出现
     */
    AlertBoxAppear,
}

export enum VipSys {
    /**
     * vip 中 sys 值为1
     */
    VipSys1 = 1,

    /**
     * vip 中 sys 值为2
     */
    VipSys2 = 2,
}

export default class AppConfigBase extends WebComponent<Props, State> {
    static contextTypes = {
        toast: PropTypes.any,
    };

    static defaultProps = {};

    /**
     * 最后一次保存生效的应用服务值
     */
    lastAppServiceConfig = {
        appServiceAccessingAddress: "",
        webClientPort: {
            https: "",
        },
    };

    /**
     * 获取 appIp
     * 根据 aPPIP 判断应用服务是否可用
     */
    appIp = location.hostname;

    /**
     * 获取应用服务和对象存储中小黄框的默认状态
     */
    defaultValidateBoxStatus = {
        appServiceAccessingAddressStatus: -1,
        webClientHttps: -1,
    };

    state = {
        appServiceAccessingAddress: "",
        webClientPort: {
            https: "",
        },
        path: "",
        type: "internal",
        appServiceAccessingAddressStatus:
            AppServiceAccessingAddressStatus.Normal,
        webClientHttpsStatus: WebHttpsValidateState.Normal,
        isAppServiceChanged: false,
        dialogStatus: DialogStatus.None,
        errorMessage: "",
        loadingStatus: false,
    };

    async componentDidMount() {
        this.initAppService();
    }

    /**
     *  应用服务访问地址部分初始化
     */
    private async initAppService() {
        // 调取接口获得返回的应用节点信息,返回数组非空代表应用节点开启
        const { host, port, path, type } = await accessAddr.get(AppType.App);
        this.setState({
            appServiceAccessingAddress: host,
            webClientPort: {
                https: port,
            },
            path,
            type,
            webClientHttpsStatus: WebHttpsValidateState.Normal,
            appServiceAccessingAddressStatus:
                AppServiceAccessingAddressStatus.Normal,
        });

        this.lastAppServiceConfig = {
            ...this.lastAppServiceConfig,
            webClientPort: this.state.webClientPort,
            appServiceAccessingAddress: this.state.appServiceAccessingAddress,
        };

        this.defaultValidateBoxStatus = {
            ...this.defaultValidateBoxStatus,
            webClientHttps: this.state.webClientHttpsStatus,
            appServiceAccessingAddressStatus:
                this.state.appServiceAccessingAddressStatus,
        };

        this.props.changeAppConfigWebClientPorts(
            Number(this.state.webClientPort.https),
            Number(this.lastAppServiceConfig.webClientPort.https)
        );
    }

    /**
     * 获取输入的应用服务访问地址
     */
    protected async changeAppServiceAccessingAddress(value: string) {
        this.setState({
            appServiceAccessingAddress: value.trim(),
            appServiceAccessingAddressStatus:
                AppServiceAccessingAddressStatus.Normal,
            isAppServiceChanged: true,
        });
    }
    /**
     * 检验输入地址中是否包含英文字母
     */
    private isAddressIncludeEnLetter(address: string) {
        return /[a-z]/gi.test(address);
    }

    /**
     * 验证域名是否合法
     */
    private isDomainNameValidate(domain: string) {
        return /^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/.test(
            domain
        );
    }

    /**
     * 应用服务点击保存触发事件
     */
    protected async completeAppService() {
        if (
            !(
                (isIP(this.state.appServiceAccessingAddress, 4) ||
                    isIP(this.state.appServiceAccessingAddress, 6) ||
                    this.isDomainNameValidate(
                        this.state.appServiceAccessingAddress
                    )) &&
                this.isPortValidate(this.state.webClientPort.https)
            )
        ) {
            let appServiceAccessingAddressStatusTemp =
                this.state.appServiceAccessingAddressStatus;
            let webClientHttpsStatusTemp = this.state.webClientHttpsStatus;
            // 验证应用服务中访问地址合法性
            if (
                !(
                    isIP(this.state.appServiceAccessingAddress, 4) ||
                    isIP(this.state.appServiceAccessingAddress, 6)
                )
            ) {
                if (this.state.appServiceAccessingAddress.length === 0) {
                    appServiceAccessingAddressStatusTemp =
                        AppServiceAccessingAddressStatus.Empty;
                } else {
                    if (
                        this.isAddressIncludeEnLetter(
                            this.state.appServiceAccessingAddress
                        )
                    ) {
                        if (
                            this.isDomainNameValidate(
                                this.state.appServiceAccessingAddress
                            )
                        ) {
                            appServiceAccessingAddressStatusTemp =
                                AppServiceAccessingAddressStatus.Normal;
                        } else {
                            appServiceAccessingAddressStatusTemp =
                                AppServiceAccessingAddressStatus.ErrorEnglishLetter;
                        }
                    } else {
                        appServiceAccessingAddressStatusTemp =
                            AppServiceAccessingAddressStatus.ErrorNoEnglish;
                    }
                }
            }
            // 验证应用服务中 Web客户端 https 端口合法性
            if (!this.isPortValidate(this.state.webClientPort.https)) {
                this.state.webClientPort.https === ""
                    ? (webClientHttpsStatusTemp = WebHttpsValidateState.Empty)
                    : (webClientHttpsStatusTemp =
                          WebHttpsValidateState.InputError);
            }

            this.setState({
                appServiceAccessingAddressStatus:
                    appServiceAccessingAddressStatusTemp,
                webClientHttpsStatus: webClientHttpsStatusTemp,
            });
        } else {
            this.setState({
                dialogStatus: DialogStatus.AlertBoxAppear,
            });
        }
    }

    /**
     * 应用服务点击警告框中确认按钮
     */
    protected async clickCompleteAppService() {
        const {
            appServiceAccessingAddress: host,
            webClientPort: { https: port },
        } = this.state;
        let playload = {};

        if (host !== this.lastAppServiceConfig.appServiceAccessingAddress) {
            playload["host"] = host;
        }
        if (port !== this.lastAppServiceConfig.webClientPort.https) {
            playload["port"] = port;
        }

        try {
            this.setState({ dialogStatus: DialogStatus.None });
            this.setState({ loadingStatus: true });
            await accessAddr.put(AppType.App, playload);
            this.setState({ isAppServiceChanged: false });
            message.success(__("保存成功"));

            if (host !== this.lastAppServiceConfig.appServiceAccessingAddress) {
                // manageLog({
                //   level: Level.INFO,
                //   opType: ManagementOps.SET,
                //   msg: __("设置 访问地址 成功，"),
                //   exMsg: __("访问地址：${appServiceAccessingAddress}", {
                //     appServiceAccessingAddress: host,
                //   }),
                // });
            }
            if (port !== this.lastAppServiceConfig.webClientPort.https) {
                await this.props.changeAppConfigWebClientPorts(
                    Number(port),
                    Number(this.lastAppServiceConfig.webClientPort.https)
                );
                // manageLog({
                //   level: Level.INFO,
                //   opType: ManagementOps.SET,
                //   msg: __("修改 访问端口 成功"),
                //   exMsg: __("${port}", {
                //     port: port,
                //   }),
                // });
            }

            this.lastAppServiceConfig = {
                appServiceAccessingAddress: host,
                webClientPort: { https: port },
            };

            message.info(__("即将跳转至登录页，请稍等..."));
            setTimeout(() => {
                // 去除通过服务检查，因为火狐不支持
                signup(
                    location.protocol +
                        "//" +
                        host +
                        ":" +
                        port +
                        defaultPathList[1]
                );
            }, 5000);
        } catch (ex: any) {
            this.setState({
                errorMessage: getErrorMessage("deploy", ex) || ex.message,
                dialogStatus: DialogStatus.ErrorDialogAppear,
            });
        } finally {
            this.setState({ loadingStatus: false });
        }
    }

    /**
     * 获取输入的web客户端访问https端口
     */
    protected async changeWebHttps(value: string) {
        const { webClientPort } = this.state;

        this.setState({
            webClientPort: {
                ...webClientPort,
                https: value.trim(),
            },
            webClientHttpsStatus: WebHttpValidateState.Normal,
            isAppServiceChanged: true,
        });
    }

    /**
     * 验证输入的http或者https端口号是是否合法
     */
    private isPortValidate(value: any) {
        const port = Number(value);
        if (port !== Math.floor(value)) {
            return false;
        }
        return port >= 1 && port <= 65535;
    }

    /**
     * 关闭弹窗触发事件
     */
    protected closeDialog() {
        this.setState({
            dialogStatus: DialogStatus.None,
        });
    }

    /**
     * 应用服务取消按钮点击时触发事件
     */
    protected cancelAppService(dialogStatus: DialogStatus) {
        if (dialogStatus === DialogStatus.AlertBoxAppear) {
            this.setState({
                dialogStatus: DialogStatus.None,
            });
        }
        this.setState({
            appServiceAccessingAddress:
                this.lastAppServiceConfig.appServiceAccessingAddress,
            webClientPort: this.lastAppServiceConfig.webClientPort,
            appServiceAccessingAddressStatus:
                this.defaultValidateBoxStatus.appServiceAccessingAddressStatus,
            webClientHttpsStatus: this.defaultValidateBoxStatus.webClientHttps,
            isAppServiceChanged: false,
        });
        this.props.changePageState(PageState.Info);
    }
}

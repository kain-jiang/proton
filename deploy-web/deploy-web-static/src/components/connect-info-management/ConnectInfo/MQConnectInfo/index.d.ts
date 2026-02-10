import { ConnectInfoValidateState, MQConnectInfo } from "../../index.d";

export interface Props {
    // 所有数据
    configData: any;

    // 连接配置校验状态
    connectInfoValidateState: ConnectInfoValidateState;

    // 更新MariaDB的数据
    updateConnectInfo: (o: MQConnectInfo, type?: string) => void;

    // 更新连接配置form实例
    updateConnectInfoForm: (form) => void;

    // 更新连接配置校验状态
    updateConnectInfoValidateState: (
        value: Partial<ConnectInfoValidateState>
    ) => void;

    // 连接信息的可编辑状态
    originSourceType: string;

    // 已存在第三方连接信息的类型
    originConnectInfoType: string;
}

export interface State {
    // 编辑中的数据
    mq: MQConnectInfo;
    // mq列表
    mqTypeList: Array<string>;
}

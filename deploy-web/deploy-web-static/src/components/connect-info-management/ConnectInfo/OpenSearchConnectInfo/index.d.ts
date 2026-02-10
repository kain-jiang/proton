import { ConnectInfoValidateState, OpensearchConnectInfo } from "../../index.d";

export interface Props {
    // 所有数据
    configData: any;

    // 连接配置校验状态
    connectInfoValidateState: ConnectInfoValidateState;

    // 更新MariaDB的数据
    updateConnectInfo: (o: OpensearchConnectInfo) => void;

    // 更新连接配置form实例
    updateConnectInfoForm: (form) => void;

    // 更新连接配置校验状态
    updateConnectInfoValidateState: (
        value: Partial<ConnectInfoValidateState>
    ) => void;

    // 连接信息的可编辑状态
    originSourceType: string;
}

export interface State {
    // 编辑中的数据
    opensearch: OpensearchConnectInfo;
}

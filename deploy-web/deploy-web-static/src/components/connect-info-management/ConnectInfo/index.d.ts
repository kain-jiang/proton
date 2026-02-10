import { ConnectInfoValidateState } from "../index.d";
import { Services } from "../../component-management/index.d";

declare namespace DataBaseConfigType {
    interface Props {
        // 所有数据
        configData: any;

        // 连接配置部分配置项校验状态
        connectInfoValidateState: ConnectInfoValidateState;

        // 更新源数据
        onUpdateConnectInfo: (key, o, type?) => void;

        // 更新连接配置form实例
        updateConnectInfoForm: (value) => void;

        // 更新连接配置校验状态
        updateConnectInfoValidateState: (
            value: Partial<ConnectInfoValidateState>
        ) => void;

        // 连接信息的可编辑状态
        originSourceType: string;

        // 已存在第三方连接信息的类型
        originConnectInfoType: string;
    }

    // interface State extends Services {}
}

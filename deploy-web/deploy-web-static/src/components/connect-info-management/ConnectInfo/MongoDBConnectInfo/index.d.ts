import { ConnectInfoValidateState, MongoDBConnectInfo } from "../../index.d";
import { Services } from "../../../component-management/index.d";

export interface Props {
    // 所有数据
    configData: any;

    // 连接配置校验状态
    connectInfoValidateState: ConnectInfoValidateState;

    // 更新数据
    updateConnectInfo: (o: MongoDBConnectInfo) => void;

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
    mongodb: MongoDBConnectInfo;
}

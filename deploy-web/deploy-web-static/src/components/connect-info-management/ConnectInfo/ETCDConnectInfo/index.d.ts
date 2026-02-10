import { ETCDConnectInfo } from "../../index.d";

export interface Props {
    // 所有数据
    configData: any;

    // 更新MariaDB的数据
    updateConnectInfo: (o: ETCDConnectInfo) => void;

    // 更新连接配置form实例
    updateConnectInfoForm: (form) => void;
}

export interface State {
    // 编辑中的数据
    etcd: ETCDConnectInfo;
}

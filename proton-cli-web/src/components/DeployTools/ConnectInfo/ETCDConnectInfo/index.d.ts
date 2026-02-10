import { ETCDConnectInfo, ConfigData } from "../../index";

export interface Props {
  // 模板类型
  dataBaseStorageType: string;

  // 所有数据
  configData: ConfigData;
}

export interface State {
  // 编辑中的数据
  etcd: ETCDConnectInfo;
}

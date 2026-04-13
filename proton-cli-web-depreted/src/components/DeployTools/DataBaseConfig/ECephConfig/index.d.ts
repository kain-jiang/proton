import { ProtonECeph, NodeInfo, ConfigData } from "../../index";

declare namespace ECephType {
  interface Props {
    // 模板类型
    dataBaseStorageType: string;

    // 所有数据
    configData: ConfigData;

    // title
    service: {
      key: string;
      name: string;
    };

    // 更新ECeph的数据
    onUpdateECephData: (config: ProtonECeph, dataBaseStorageType) => void;

    // 删除ECeph
    onDeleteECephConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;
  }

  interface State {
    // ECeph编辑中的数据
    ecephConfig: ProtonECeph;

    // ECeph节点
    ecephNodes: Array<NodeInfo>;
  }
}

// 枚举虚拟地址类型
export enum KeepalivedEnum {
  Internal = "internal",
  External = "external",
}

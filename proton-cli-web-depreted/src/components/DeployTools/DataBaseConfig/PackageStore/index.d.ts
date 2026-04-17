import { PackageStoreConfig, NodeInfo, ConfigData } from "../../index";

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

  // 更新MariaDB的数据
  onUpdatePackageStore: (
    config: PackageStoreConfig,
    dataBaseStorageType: string,
  ) => void;

  // 删除当前组件
  onDeletePackageStoreConfig: () => void;

  // 更新基础服务配置form实例
  updateDataBaseForm: (value) => void;
}

interface State {
  // MariaDB 编辑中的数据
  packageStoreConfig: PackageStoreConfig;

  // 部署节点
  packageStoreNodes: Array<NodeInfo>;
}

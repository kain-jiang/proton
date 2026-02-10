import { ProtonMariadb, NodeInfo, ConfigData } from "../../index";

declare namespace MariaDBType {
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
    onUpdateMariDBData: (config: ProtonMariadb, dataBaseStorageType) => void;

    // 删除当前组件
    onDeleteMariaDBConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;
  }

  interface State {
    // MariaDB 编辑中的数据
    mariaDBConfig: ProtonMariadb;

    // 部署节点
    mariaDBNodes: Array<NodeInfo>;
  }
}

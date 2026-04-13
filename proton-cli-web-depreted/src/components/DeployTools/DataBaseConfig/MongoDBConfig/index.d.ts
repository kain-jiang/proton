import { ProtonDataBase, NodeInfo, ConfigData } from "../../index";

declare namespace MongoDBType {
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
    onUpdateMongoDBData: (config: ProtonDataBase, dataBaseStorageType) => void;

    // 删除MongoDB
    onDeleteMongoDBConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;
  }

  interface State {
    // MariaDB编辑中的数据
    mongoDBConfig: ProtonDataBase;

    // mongoDBNodes 节点
    mongoDBNodes: Array<NodeInfo>;
  }
}

import { ProtonDataBase, NodeInfo, ConfigData } from "../../index";

declare namespace RedisType {
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
    onUpdateRedisData: (config: ProtonDataBase, dataBaseStorageType) => void;

    // 删除Redis
    onDeleteRedisConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;
  }

  interface State {
    // MariaDB编辑中的数据
    redisConfig: ProtonDataBase;

    // MariaDB节点
    redisNodes: Array<NodeInfo>;
  }
}

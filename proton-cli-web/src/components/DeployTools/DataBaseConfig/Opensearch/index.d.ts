import { Opensearch, NodeInfo, ConfigData } from "../../index";

declare namespace OpensearchType {
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

    // 更新opensearch的数据
    onUpdateOpensearchData: (config: Opensearch, dataBaseStorageType) => void;

    // 删除服务
    onDeleteOpenSearchConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;
  }

  interface State {
    // MariaDB编辑中的数据
    opensearchConfig: Opensearch;

    // opensearch节点
    opensearchNodes: Array<NodeInfo>;
  }
}

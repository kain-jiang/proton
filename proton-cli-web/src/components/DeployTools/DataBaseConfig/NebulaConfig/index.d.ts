import { ConfigData, ProtonNebula, NodeInfo } from "../../index";

declare namespace NebulaType {
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

    // 更新数据
    onUpdateNebulaData: (config: ProtonNebula, dataBaseStorageType) => void;

    // 删除当前组件
    onDeleteNebulaConfig: () => void;

    // 更新基础服务配置form实例
    updateDataBaseForm: (value) => void;
  }

  interface State {
    // nebula 编辑中的数据
    nebulaConfig: ProtonNebula;

    // 部署节点
    nebulaNodes: Array<NodeInfo>;
  }
}

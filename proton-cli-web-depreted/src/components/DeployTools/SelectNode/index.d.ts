import { NodeInfo } from '../index';

declare namespace SelectNodeType {
    interface Props {
        // 所有可选节点
        nodes: Array<NodeInfo>;

        // 默认选中的节点
        selectedNodes: Array<NodeInfo>;

        // 数据变更事件
        onSelectedChange: (nodes: Array<NodeInfo>) => void;

        //select模式
        mode:boolean;
    }

    interface State {

        // 选中的节点
        selectedNodes: Array<NodeInfo>;
    }
}
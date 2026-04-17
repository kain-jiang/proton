import * as React from "react";
import { SelectNodeType } from "./index";
import { message } from "@aishutech/ui";

export default class SelectNodeBase extends React.Component<
  SelectNodeType.Props,
  SelectNodeType.State
> {
  state = {
    selectedNodes: [],
  };

  componentDidMount(): void {
    this.setState({
      selectedNodes: this.props.selectedNodes,
    });
  }

  componentDidUpdate(
    prevProps: Readonly<SelectNodeType.Props>,
    prevState: Readonly<SelectNodeType.State>,
    snapshot?: any,
  ): void {
    if (this.props.selectedNodes != prevProps.selectedNodes) {
      this.setState({
        selectedNodes: this.props.selectedNodes,
      });
    }
  }
  /**
   * 保存选中的节点
   */
  public onSelectNode(value) {
    if (this.props.mode) {
      this.setState(
        {
          selectedNodes: [
            ...this.props.nodes.filter((node) => node.name === value),
          ],
        },
        () => {
          this.props.onSelectedChange(this.state.selectedNodes);
        },
      );
    } else {
      this.setState(
        {
          selectedNodes: [
            ...this.state.selectedNodes,
            ...this.props.nodes.filter((node) => node.name === value),
          ],
        },
        () => {
          this.props.onSelectedChange(this.state.selectedNodes);
        },
      );
    }
  }

  /**
   * 取消单个节点的选中
   * @param value
   */
  public onDeselectNode(value) {
    if (this.state.selectedNodes.length === 1) {
      message.info(`请至少保留一个节点`);
    } else {
      this.setState(
        {
          selectedNodes: this.state.selectedNodes.filter(
            (node) => node.name != value,
          ),
        },
        () => {
          this.props.onSelectedChange(this.state.selectedNodes);
        },
      );
    }
  }
}

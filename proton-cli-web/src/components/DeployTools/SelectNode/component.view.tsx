import * as React from "react";
import { NodeInfo } from "../index";
import {
  Checkbox,
  Row,
  Col,
  Form,
  Button,
  Drawer,
  Tag,
  Space,
  Tooltip,
  Select,
} from "@aishutech/ui";
import SelectNodeBase from "./component.base";
import "./styles.view.scss";

export default class SelectNode extends SelectNodeBase {
  render(): React.ReactNode {
    return (
      <Select
        mode={this.props.mode ? "tags" : "multiple"}
        showArrow
        style={{
          width: "100%",
        }}
        optionLabelProp="label"
        placeholder="请选择部署节点"
        value={this.state.selectedNodes.map((node) => node.name)}
        onSelect={(value) => {
          this.onSelectNode(value);
        }}
        onDeselect={this.onDeselectNode.bind(this)}
        getPopupContainer={(node) => node.parentElement || document.body}
      >
        {this.props.nodes.map((value) => (
          <Select.Option key={value.name} label={value.name}>
            <Tooltip
              placement="right"
              title={`ipv4:${value.ipv4} ipv6:${value.ipv6} `}
            >
              <span>{value.name}</span>
            </Tooltip>
          </Select.Option>
        ))}
      </Select>
    );
  }
}

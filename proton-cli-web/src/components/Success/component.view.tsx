import * as React from "react";
import { Button, Result } from "@aishutech/ui";
import "./styles.view.scss";

export default class Success extends React.Component {
  render(): React.ReactNode {
    return (
      <div className="success-contain">
        <Result status="success" title="集群初始化成功" />
      </div>
    );
  }
}

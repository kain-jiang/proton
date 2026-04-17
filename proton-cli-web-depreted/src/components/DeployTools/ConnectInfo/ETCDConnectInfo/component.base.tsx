import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";

export class ETCDConnectInfoBase extends React.Component<Props, State> {
  state = {
    etcd: null,
  };

  componentDidMount(): void {
    const { configData } = this.props;
    this.initConfig(configData?.resource_connect_info?.etcd);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.etcd,
        configData?.resource_connect_info?.etcd,
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.etcd);
    }
  }

  /**
   * 初始化配置
   * @param etcd 连接信息类型
   */
  private initConfig(etcd) {
    this.setState({
      etcd,
    });
  }
}

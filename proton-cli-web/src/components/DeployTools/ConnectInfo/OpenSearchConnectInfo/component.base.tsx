import * as React from "react";
import { isEqual } from "lodash";
import { CONNECT_SERVICES } from "../../helper";
import { Props, State } from "./index";
import { FormInstance } from "@aishutech/ui";

export class OpenSearchConnectInfoBase extends React.Component<Props, State> {
  state = {
    opensearch: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData?.resource_connect_info?.opensearch);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.opensearch,
        configData?.resource_connect_info?.opensearch,
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.opensearch);
    }
  }

  /**
   * 初始化配置
   * @param opensearch 连接信息类型
   */
  private initConfig(opensearch) {
    this.setState(
      {
        opensearch,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.opensearch,
        });
      },
    );
  }

  /**
   * changeOpenSearchConnectInfo
   */
  public changeOpenSearchConnectInfo(key, val, opensearch) {
    const cur = {
      ...opensearch,
      [key]: val,
    };
    this.setState(
      {
        opensearch: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }
}

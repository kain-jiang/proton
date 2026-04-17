import * as React from "react";
import { isEqual } from "lodash";
import { CONNECT_SERVICES } from "../../helper";
import { Props, State } from "./index";
import { FormInstance } from "@aishutech/ui";

export class MongoDBConnectInfoBase extends React.Component<Props, State> {
  state = {
    mongodb: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData?.resource_connect_info?.mongodb);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.mongodb,
        configData?.resource_connect_info?.mongodb,
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.mongodb);
    }
  }

  /**
   * 初始化配置
   * @param mongodb 连接信息类型
   */
  private initConfig(mongodb) {
    this.setState(
      {
        mongodb,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.mongodb,
        });
      },
    );
  }

  /**
   * changeMongoDBConnectInfo
   */
  public changeMongoDBConnectInfo(key, val, mongodb) {
    const cur = {
      ...mongodb,
      [key]: val,
    };
    this.setState(
      {
        mongodb: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }
}

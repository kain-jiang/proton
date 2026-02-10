import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";
import { FormInstance } from "@kweaver-ai/ui";

export class MongoDBConnectInfoBase extends React.Component<Props, State> {
  state = {
    mongodb: null as any,
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
        configData?.resource_connect_info?.mongodb
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.mongodb);
    }
  }

  /**
   * 初始化配置
   * @param mongodb 连接信息类型
   */
  private initConfig(mongodb: any) {
    this.setState(
      {
        mongodb,
      },
      () => {
        this.form.current?.setFieldsValue({
          ...this.state.mongodb,
        });
      }
    );
  }

  /**
   * changeMongoDBConnectInfo
   */
  public changeMongoDBConnectInfo(key: string, val: any, mongodb: any) {
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
      }
    );
  }
}

import * as React from "react";
import { isEqual } from "lodash";
import { CONNECT_SERVICES } from "../../helper";
import { Props, State } from "./index";
import { FormInstance } from "@aishutech/ui";

export class PolicyEngineConnectInfoBase extends React.Component<Props, State> {
  state = {
    policy_engine: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData?.resource_connect_info?.policy_engine);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.policy_engine,
        configData?.resource_connect_info?.policy_engine,
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.policy_engine);
    }
  }

  /**
   * 初始化配置
   * @param policy_engine 连接信息类型
   */
  private initConfig(policy_engine) {
    this.setState(
      {
        policy_engine,
      },
      () => {
        this.form.current?.setFieldsValue({
          ...this.state.policy_engine,
        });
      },
    );
  }

  /**
   * changePolicyEngineConnectInfo
   */
  public changePolicyEngineConnectInfo(key, val, policy_engine) {
    const cur = {
      ...policy_engine,
      [key]: val,
    };
    this.setState(
      {
        policy_engine: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }
}

import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";
import { FormInstance } from "@kweaver-ai/ui";
import { CONNECT_SERVICES } from "../../../component-management/helper";

export class PolicyEngineConnectInfoBase extends React.Component<Props, State> {
  state = {
    policyengine: null as any,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(
      configData?.resource_connect_info?.[CONNECT_SERVICES.POLICY_ENGINE]
    );
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.[
          CONNECT_SERVICES.POLICY_ENGINE
        ],
        configData?.resource_connect_info?.[CONNECT_SERVICES.POLICY_ENGINE]
      )
    ) {
      this.initConfig(
        configData?.resource_connect_info?.[CONNECT_SERVICES.POLICY_ENGINE]
      );
    }
  }

  /**
   * 初始化配置
   * @param policyengine 连接信息类型
   */
  private initConfig(policyengine: any) {
    this.setState(
      {
        policyengine,
      },
      () => {
        this.form.current?.setFieldsValue({
          ...this.state.policyengine,
        });
      }
    );
  }

  /**
   * changePolicyEngineConnectInfo
   */
  public changePolicyEngineConnectInfo(
    key: string,
    val: any,
    policyengine: any
  ) {
    const cur = {
      ...policyengine,
      [key]: val,
    };
    this.setState(
      {
        policyengine: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      }
    );
  }
}

import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";
import { FormInstance } from "@kweaver-ai/ui";
import { ETCDConnectInfo } from "../../index.d";
import { CONNECT_SERVICES } from "../../../component-management/helper";

export class ETCDConnectInfoBase extends React.Component<Props, State> {
  state = {
    etcd: null as any,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData?.resource_connect_info?.[CONNECT_SERVICES.ETCD]);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.[CONNECT_SERVICES.ETCD],
        configData?.resource_connect_info?.[CONNECT_SERVICES.ETCD]
      )
    ) {
      this.initConfig(
        configData?.resource_connect_info?.[CONNECT_SERVICES.ETCD]
      );
    }
  }

  /**
   * 初始化配置
   * @param etcd 连接信息类型
   */
  private initConfig(etcd: ETCDConnectInfo) {
    this.setState(
      {
        etcd,
      },
      () => {
        this.form.current?.setFieldsValue({
          ...this.state.etcd,
        });
      }
    );
  }

  /**
   * changePolicyEngineConnectInfo
   */
  public changeETCDConnectInfo(
    key: string,
    val: string,
    etcd: ETCDConnectInfo
  ) {
    const cur = {
      ...etcd,
      [key]: val,
    };
    this.setState(
      {
        etcd: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      }
    );
  }
}

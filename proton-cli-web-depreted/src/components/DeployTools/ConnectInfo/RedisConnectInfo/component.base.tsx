import * as React from "react";
import { isEqual } from "lodash";
import { CONNECT_SERVICES } from "../../helper";
import { Props, State } from "./index";
import { FormInstance } from "@aishutech/ui";

export class RedisConnectInfoBase extends React.Component<Props, State> {
  state = {
    redis: null,
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData?.resource_connect_info?.redis);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData?.resource_connect_info?.redis,
        configData?.resource_connect_info?.redis,
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.redis);
    }
  }

  /**
   * 初始化配置
   * @param redis 连接信息类型
   */
  private initConfig(redis) {
    this.setState(
      {
        redis,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.redis,
        });
      },
    );
  }

  /**
   * changeRedisConnectInfo
   */
  public changeRedisConnectInfo(key, val, redis) {
    const cur = {
      ...redis,
      [key]: val,
    };
    this.setState(
      {
        redis: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }
}

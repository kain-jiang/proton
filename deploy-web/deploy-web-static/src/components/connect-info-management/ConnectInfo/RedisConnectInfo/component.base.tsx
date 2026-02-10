import * as React from "react";
import { isEqual } from "lodash";
import { Props, State } from "./index";
import { FormInstance } from "@kweaver-ai/ui";
import { RedisConnectInfo } from "../../index.d";
import { SOURCE_TYPE } from "../../../component-management/helper";

export class RedisConnectInfoBase extends React.Component<Props, State> {
  state = {
    redis: null as any,
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
        configData?.resource_connect_info?.redis
      )
    ) {
      this.initConfig(configData?.resource_connect_info?.redis);
    }
  }

  /**
   * 初始化配置
   * @param redis 连接信息类型
   */
  private initConfig(redis: any) {
    this.setState(
      {
        redis,
      },
      () => {
        this.form.current?.setFieldsValue({
          ...this.state.redis,
        });
      }
    );
  }

  /**
   * changeRedisConnectInfo
   */
  public changeRedisConnectInfo(key: string, val: any, redis: any) {
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
      }
    );
  }

  // 是否禁用不可修改信息
  protected getIsDisabled(redis: RedisConnectInfo) {
    const { originConnectInfoType, originSourceType } = this.props;
    if (redis?.source_type === SOURCE_TYPE.INTERNAL) {
      return originSourceType === SOURCE_TYPE.INTERNAL;
    } else {
      if (originSourceType !== SOURCE_TYPE.EXTERNAL) {
        return false;
      } else {
        return originConnectInfoType === redis?.connect_type;
      }
    }
  }
}

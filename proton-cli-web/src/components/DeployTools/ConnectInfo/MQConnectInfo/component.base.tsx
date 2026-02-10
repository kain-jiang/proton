import * as React from "react";
import { isEqual } from "lodash";
import { CONNECT_SERVICES, MQ_TYPE, SERVICES, SOURCE_TYPE } from "../../helper";
import { Props, State } from "./index";
import { FormInstance } from "@aishutech/ui";

export class MQConnectInfoBase extends React.Component<Props, State> {
  state = {
    mq: null,
    mqTypeList: [],
  };

  form = React.createRef<FormInstance>();

  componentDidMount(): void {
    const { configData, updateConnectInfoForm } = this.props;
    updateConnectInfoForm(this.form);
    this.initConfig(configData);
  }

  componentDidUpdate(prevProps: Readonly<Props>): void {
    const { configData } = this.props;
    if (
      !isEqual(
        prevProps.configData[SERVICES.ProtonNSQ],
        configData[SERVICES.ProtonNSQ],
      ) ||
      !isEqual(
        prevProps.configData[SERVICES.Kafka],
        configData[SERVICES.Kafka],
      ) ||
      !isEqual(
        prevProps.configData?.resource_connect_info?.mq,
        configData?.resource_connect_info?.mq,
      )
    ) {
      this.initConfig(configData);
    }
  }

  /**
   * 初始化配置
   * @param mq 连接信息类型
   */
  private initConfig(configData) {
    let mq = configData?.resource_connect_info?.mq,
      mqTypeList;
    if (mq?.source_type === SOURCE_TYPE.INTERNAL) {
      mqTypeList = [
        configData[SERVICES.ProtonNSQ] ? MQ_TYPE.NSQ : null,
        configData[SERVICES.Kafka] ? MQ_TYPE.KAFKA : null,
      ].filter((item) => item);
    } else {
      mqTypeList = Object.keys(MQ_TYPE).map((key) => MQ_TYPE[key]);
    }
    this.setState(
      {
        mq,
        mqTypeList,
      },
      () => {
        this.form.current.setFieldsValue({
          ...this.state.mq,
        });
      },
    );
  }

  /**
   * 修改mq信息
   * @param key 键
   * @param val 值
   * @param mq mq信息
   */
  public changeMQConnectInfo(key, val, mq) {
    let cur = mq;
    // 当是kafka切换source type类型时直接清空mq type
    if (key === "source_type") {
      cur = {
        ...mq,
        [key]: val,
        mq_type: undefined,
      };
    } else {
      cur = {
        ...mq,
        [key]: val,
      };
    }
    this.setState(
      {
        mq: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }

  /**
   * 修改mq auth信息
   * @param key 键
   * @param val 值
   * @param mq mq信息
   */
  public changeMQAuthConnectInfo(key, val, mq) {
    const cur = {
      ...mq,
      auth: {
        ...mq.auth,
        [key]: val,
      },
    };
    this.setState(
      {
        mq: cur,
      },
      () => {
        this.props.updateConnectInfo(cur);
      },
    );
  }
}

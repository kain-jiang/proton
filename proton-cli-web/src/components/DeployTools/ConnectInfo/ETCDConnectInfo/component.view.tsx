import * as React from "react";
import { Radio, Divider } from "@aishutech/ui";
import { Title } from "../../Title/component.view";
import { ETCDConnectInfoBase } from "./component.base";
import {
  CONNECT_SERVICES,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
} from "../../helper";
import "./styles.view.scss";

export class ETCDConnectInfo extends ETCDConnectInfoBase {
  render(): React.ReactNode {
    const { etcd } = this.state;

    return (
      <div className="service-box">
        <Title
          title={CONNECT_SERVICES_TEXT[CONNECT_SERVICES.ETCD] + "连接信息"}
        />
        <Divider orientation="left" orientationMargin="0">
          资源类型
        </Divider>
        <Radio.Group
          style={{
            margin: "10px 0",
          }}
          disabled
          value={etcd?.source_type}
        >
          <Radio value={SOURCE_TYPE.INTERNAL}>
            本地 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.ETCD]}
          </Radio>
          <Radio value={SOURCE_TYPE.EXTERNAL}>
            第三方 {CONNECT_SERVICES_TEXT[CONNECT_SERVICES.ETCD]}
          </Radio>
        </Radio.Group>
      </div>
    );
  }
}

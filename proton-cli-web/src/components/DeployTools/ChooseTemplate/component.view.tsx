import * as React from "react";
import { Props } from "./index.d";
import { Card, Form, Select } from "@aishutech/ui";
import { DEVICESPECSMAP, DataBaseStorageList } from "../../DeployTools/helper";
import "./styles.view.scss";

export const ChooseTemplate: React.FC<Props> = ({
  changeDataBaseStorageType,
  configData,
  updateDeploy,
}) => {
  return (
    <div className="wrap">
      <div className="tips">
        请选择产品型号（部署AnyShare场景选择，其他产品使用默认值即可）
      </div>
      <Form>
        <Form.Item>
          <Select
            style={{
              width: "350px",
            }}
            value={configData.deploy.devicespec}
            placeholder="默认值"
            onChange={(value) =>
              updateDeploy({
                devicespec: value,
              })
            }
            getPopupContainer={(node) => node.parentElement || document.body}
            options={Object.keys(DEVICESPECSMAP).map((deviceType) => {
              return {
                label: deviceType,
                options: DEVICESPECSMAP[deviceType].map((deviceSpec) => {
                  return {
                    label: deviceSpec,
                    value: deviceSpec,
                  };
                }),
              };
            })}
          />
        </Form.Item>
      </Form>
      <div className="tips">请选择部署模式：</div>
      {DataBaseStorageList.map(({ type, title, style, content }) => {
        return (
          <Card
            className="card"
            style={style}
            key={type}
            onClick={() => changeDataBaseStorageType(type)}
          >
            <div className="title">{title}</div>
            <div className="content">{content}</div>
          </Card>
        );
      })}
    </div>
  );
};

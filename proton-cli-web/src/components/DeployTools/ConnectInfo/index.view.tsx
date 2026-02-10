import * as React from "react";
import { Menu, Dropdown, Button } from "@aishutech/ui";
import { MQConnectInfo } from "./MQConnectInfo/component.view";
import { DownOutlined, PlusOutlined } from "@aishutech/ui/icons";
import { RDSConnectInfo } from "./RDSConnectInfo/component.view";
import { ETCDConnectInfo } from "./ETCDConnectInfo/component.view";
import { RedisConnectInfo } from "./RedisConnectInfo/component.view";
import { MongoDBConnectInfo } from "./MongoDBConnectInfo/component.view";
import { OpenSearchConnectInfo } from "./OpenSearchConnectInfo/component.view";
import { PolicyEngineConnectInfo } from "./PolicyEngineConnectInfo/component.view";
import {
  CONNECT_SERVICES,
  DEGAULT_CONNECT_SERVICE,
  CONNECT_SERVICES_TEXT,
  SOURCE_TYPE,
} from "../helper";
import { DataBaseConfigType } from "./index";
import "./styles.view.scss";

export const ConnectInfo: React.FC<DataBaseConfigType.Props> = React.memo(
  ({
    configData,
    dataBaseStorageType,
    connectInfoValidateState,
    onUpdateConnectInfo,
    onDeleteResource,
    onAddResource,
    updateConnectInfoForm,
    updateConnectInfoValidateState,
  }) => {
    const updateConnectInfo = (key, value) => {
      onUpdateConnectInfo(key, value);
    };

    const getConnectInfoTemplates = () => {
      const items = DEGAULT_CONNECT_SERVICE.filter(
        (service) =>
          !Object.keys(configData.resource_connect_info).includes(service),
      ).map((service) => ({
        label: CONNECT_SERVICES_TEXT[service],
        key: service,
      }));

      return (
        <>
          <div
            style={{
              paddingBottom: "20px",
            }}
          >
            <Dropdown
              overlay={
                <Menu
                  items={items}
                  onClick={(item) => {
                    onAddResource(item.key);
                  }}
                />
              }
              disabled={!items.length}
            >
              <a onClick={(e) => e.preventDefault()}>
                <Button disabled={!Boolean(items.length)} type="primary">
                  <PlusOutlined />
                  添加可选服务
                  <DownOutlined />
                </Button>
              </a>
            </Dropdown>
          </div>
          {configData?.resource_connect_info?.rds ? (
            <RDSConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              connectInfoValidateState={connectInfoValidateState}
              onDeleteResource={() => onDeleteResource(CONNECT_SERVICES.RDS)}
              updateConnectInfo={(o) =>
                updateConnectInfo(CONNECT_SERVICES.RDS, o)
              }
              updateConnectInfoForm={(form) =>
                updateConnectInfoForm({ [CONNECT_SERVICES.RDS]: form })
              }
              updateConnectInfoValidateState={updateConnectInfoValidateState}
            />
          ) : null}
          {configData?.resource_connect_info?.mongodb ? (
            <MongoDBConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              connectInfoValidateState={connectInfoValidateState}
              onDeleteResource={() =>
                onDeleteResource(CONNECT_SERVICES.MONGODB)
              }
              updateConnectInfo={(o) =>
                updateConnectInfo(CONNECT_SERVICES.MONGODB, o)
              }
              updateConnectInfoForm={(form) =>
                updateConnectInfoForm({ [CONNECT_SERVICES.MONGODB]: form })
              }
              updateConnectInfoValidateState={updateConnectInfoValidateState}
            />
          ) : null}
          {configData?.resource_connect_info?.redis ? (
            <RedisConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              connectInfoValidateState={connectInfoValidateState}
              onDeleteResource={() => onDeleteResource(CONNECT_SERVICES.REDIS)}
              updateConnectInfo={(o) =>
                updateConnectInfo(CONNECT_SERVICES.REDIS, o)
              }
              updateConnectInfoForm={(form) =>
                updateConnectInfoForm({ [CONNECT_SERVICES.REDIS]: form })
              }
              updateConnectInfoValidateState={updateConnectInfoValidateState}
            />
          ) : null}
          {configData?.resource_connect_info?.mq ? (
            <MQConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              connectInfoValidateState={connectInfoValidateState}
              onDeleteResource={() => onDeleteResource(CONNECT_SERVICES.MQ)}
              updateConnectInfo={(o) =>
                updateConnectInfo(CONNECT_SERVICES.MQ, o)
              }
              updateConnectInfoForm={(form) =>
                updateConnectInfoForm({ [CONNECT_SERVICES.MQ]: form })
              }
              updateConnectInfoValidateState={updateConnectInfoValidateState}
            />
          ) : null}
          {configData?.resource_connect_info?.opensearch ? (
            <OpenSearchConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              connectInfoValidateState={connectInfoValidateState}
              onDeleteResource={() =>
                onDeleteResource(CONNECT_SERVICES.OPENSEARCH)
              }
              updateConnectInfo={(o) =>
                updateConnectInfo(CONNECT_SERVICES.OPENSEARCH, o)
              }
              updateConnectInfoForm={(form) =>
                updateConnectInfoForm({ [CONNECT_SERVICES.OPENSEARCH]: form })
              }
              updateConnectInfoValidateState={updateConnectInfoValidateState}
            />
          ) : null}

          {configData?.resource_connect_info?.policy_engine ? (
            <PolicyEngineConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onDeleteResource={() =>
                onDeleteResource(CONNECT_SERVICES.POLICY_ENGINE)
              }
              updateConnectInfo={(o) =>
                updateConnectInfo(CONNECT_SERVICES.POLICY_ENGINE, o)
              }
              updateConnectInfoForm={(form) =>
                updateConnectInfoForm({
                  [CONNECT_SERVICES.POLICY_ENGINE]: form,
                })
              }
            />
          ) : null}
          {configData?.resource_connect_info?.etcd ? (
            <ETCDConnectInfo
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
            />
          ) : null}
        </>
      );
    };

    return <div className="service-contain">{getConnectInfoTemplates()}</div>;
  },
);

import * as React from "react";
import { Title } from "../Title/component.view";
import RedisConfig from "./RedisConfig/component.view";
import NebulaConfig from "./NebulaConfig/component.view";
import PackageStore from "./PackageStore/component.view";
import MariaDBConfig from "./MaridbConfig/component.view";
import MongoDBConfig from "./MongoDBConfig/component.view";
import ServiceConfig from "./ServiceConfig/component.view";
import OpensearchConfig from "./Opensearch/component.view";
import ECephConfig from "./ECephConfig/component.view";
import { Menu, Dropdown, Button, Checkbox } from "@aishutech/ui";
import { DownOutlined, PlusOutlined } from "@aishutech/ui/icons";
import {
  ConfigOtherService,
  filterEmptyKey,
  SERVICES,
  AddableServices,
  DataBaseStorageType,
} from "../helper";
import { DataBaseConfigType } from "./index";
import "./styles.view.scss";
import MonitorConfig from "./MonitorConfig/component.view";

export const DataBaseConfig: React.FC<DataBaseConfigType.Props> = React.memo(
  ({
    dataBaseStorageType,
    configData,
    addableServices,
    selectableServices,
    grafanaNodesValidateState,
    prometheusNodesValidateState,
    monitorNodesValidateState,
    onDeleteService,
    onUpdateDataBaseConfig,
    onAddService: addService,
    onDeleteService: deleteService,
    updateDataBaseForm,
    updateGrafanaNodesValidateState,
    updatePrometheusNodesValidateState,
    updateMonitorNodesValidateState,
  }) => {
    const onAddService = (value) => {
      addService(value.key);
    };
    const onChangeOtherService = (e, service) => {
      if (e.target.checked) {
        addService(service);
      } else {
        deleteService(service);
      }
    };

    const updateResource = (key, value, dataBaseStorageType) => {
      const val = filterEmptyKey(
        {
          ...configData[key],
          ...value,
        },
        dataBaseStorageType,
      );
      onUpdateDataBaseConfig({
        [key]: val,
      });
    };

    const getServiceTemplates = (service) => {
      switch (service.key) {
        case SERVICES.ProtonMonitor:
          if (dataBaseStorageType === DataBaseStorageType.DepositKubernetes)
            return null;
          return (
            <MonitorConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              monitorNodesValidateState={monitorNodesValidateState}
              onUpdateMonitorData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteMonitorConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
              updateMonitorNodesValidateState={updateMonitorNodesValidateState}
            />
          );
        case SERVICES.ProtonMariadb:
          return (
            <MariaDBConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdateMariDBData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteMariaDBConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.ProtonMongodb:
          return (
            <MongoDBConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdateMongoDBData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteMongoDBConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.ProtonRedis:
          return (
            <RedisConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdateRedisData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteRedisConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.Opensearch:
          return (
            <OpensearchConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdateOpensearchData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteOpenSearchConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.Nebula:
          return (
            <NebulaConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdateNebulaData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteNebulaConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.PackageStore:
          return (
            <PackageStore
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdatePackageStore={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeletePackageStoreConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.ECeph:
          return (
            <ECephConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              onUpdateECephData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteECephConfig={() => onDeleteService(service.key)}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
            />
          );
        case SERVICES.ProtonNSQ:
        case SERVICES.ProtonPolicyEngine:
        case SERVICES.ProtonEtcd:
        case SERVICES.Kafka:
        case SERVICES.Zookeeper:
        case SERVICES.Prometheus:
        case SERVICES.Grafana:
          return (
            <ServiceConfig
              key={service.key}
              service={service}
              configData={configData}
              dataBaseStorageType={dataBaseStorageType}
              grafanaNodesValidateState={grafanaNodesValidateState}
              prometheusNodesValidateState={prometheusNodesValidateState}
              onUpdateServiceData={(value, dataBaseStorageType) =>
                updateResource(service.key, value, dataBaseStorageType)
              }
              onDeleteServiceConfig={() => {
                onDeleteService(service.key);
              }}
              updateDataBaseForm={(form) =>
                updateDataBaseForm({ [service.key]: form })
              }
              updatePrometheusNodesValidateState={
                updatePrometheusNodesValidateState
              }
              updateGrafanaNodesValidateState={updateGrafanaNodesValidateState}
            />
          );
      }
    };

    const getOtherServiceTemplates = () => {
      return (
        <div className="service-box">
          <Title title={"其他服务"} />
          <div>
            {ConfigOtherService.map((value) => {
              return (
                <div>
                  <Checkbox
                    defaultChecked
                    // disabled
                    onChange={(e) => onChangeOtherService(e, value.serviceKey)}
                  >
                    {value.key}
                  </Checkbox>
                </div>
              );
            })}
          </div>
        </div>
      );
    };

    return (
      <div className="database-contain">
        <div
          style={{
            paddingBottom: "20px",
          }}
        >
          <Dropdown
            overlay={
              <Menu
                items={addableServices
                  .filter((item) => {
                    return !AddableServices.includes(item.key);
                  })
                  .map((value) => ({
                    label: value.name,
                    key: value.key,
                  }))}
                onClick={(value) => {
                  onAddService(value);
                }}
              />
            }
            disabled={!addableServices.length}
          >
            <a onClick={(e) => e.preventDefault()}>
              <Button
                disabled={!Boolean(addableServices.length)}
                type="primary"
              >
                <PlusOutlined />
                添加可选服务
                <DownOutlined />
              </Button>
            </a>
          </Dropdown>
        </div>
        {/* 基础服务功能配置 */}
        {selectableServices.map((value) => getServiceTemplates(value))}
        {/* 其他服务功能配置 */}
        {getOtherServiceTemplates()}
      </div>
    );
  },
);

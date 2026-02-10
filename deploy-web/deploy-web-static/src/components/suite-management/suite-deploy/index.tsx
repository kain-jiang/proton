import React, { FC, useState } from "react";
import { ServiceMode } from "../../../core/service-management/service-deploy";
import { ServiceHome } from "./service-home";
import { ServiceDetail } from "./service-detail";
import { OperationService } from "./operation-service";
import { RevertService } from "./revert-service";
import { ApplicationItem } from "../../../api/suite-management/suite-deploy/declare";

export const SuiteDeploy: FC = () => {
  // 套件部署功能类型
  const [serviceMode, setServiceMode] = useState<ServiceMode>(ServiceMode.Home);
  // 当前服务id
  const [serviceId, setServiceId] = useState<number>(0);
  // 更新服务信息
  const [updateServiceRecord, setUpdateServiceRecord] =
    useState<ApplicationItem>({ mname: "", mversion: "", title: "" });
  // 更新服务jid
  const [updateRecordJid, setUpdateRecordJid] = useState<number>(0);

  const getComponents = () => {
    switch (serviceMode) {
      // 安装
      case ServiceMode.Install:
        return (
          <OperationService
            changeServiceMode={setServiceMode}
            operationType={ServiceMode.Install}
          />
        );
      // 更新
      case ServiceMode.Update:
        return (
          <OperationService
            changeServiceMode={setServiceMode}
            operationType={ServiceMode.Update}
            updateServiceRecord={updateServiceRecord}
          />
        );
      // 回退版本
      case ServiceMode.Revert:
        return (
          <RevertService
            jid={updateRecordJid}
            changeServiceMode={setServiceMode}
            updateServiceRecord={updateServiceRecord}
          />
        );
      // 服务详情
      case ServiceMode.Service:
        return (
          <ServiceDetail
            key={Date.now()}
            changeServiceMode={setServiceMode}
            changeServiceId={setServiceId}
            changeUpdateServiceRecord={setUpdateServiceRecord}
            changeJid={setUpdateRecordJid}
            serviceId={serviceId}
          />
        );
      // 服务管理首页
      default:
        return (
          <ServiceHome
            changeServiceMode={setServiceMode}
            changeServiceId={setServiceId}
            changeUpdateServiceRecord={setUpdateServiceRecord}
          />
        );
    }
  };

  return getComponents();
};

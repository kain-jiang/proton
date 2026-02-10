import React, { FC, useState } from "react";
import { ServiceMode } from "../../../core/service-management/service-deploy";
import { ServiceHome } from "./service-home";
import { ServiceDetail } from "./service-detail";
import { OperationService } from "./operation-service";
import { RevertService } from "./revert-service";
import { ApplicationItem } from "../../../api/service-management/service-deploy/declare";

export const ServiceDeploy: FC = () => {
  // 服务部署功能类型
  const [serviceMode, setServiceMode] = useState<ServiceMode>(ServiceMode.Home);
  // 上一次主服务id
  const [mainServiceId, setMainServiceId] = useState<number>(0);
  // 当前服务id
  const [serviceId, setServiceId] = useState<number>(0);
  // 更新服务信息
  const [updateServiceRecord, setUpdateServiceRecord] =
    useState<ApplicationItem>({ aid: 0, name: "", version: "", title: "" });
  // 更新服务jid
  const [updateRecordJid, setUpdateRecordJid] = useState<number>(0);
  // 系统空间id
  const [sid, setSid] = useState(0);

  const getComponents = () => {
    switch (serviceMode) {
      // 安装
      case ServiceMode.Install:
        return (
          <OperationService
            changeServiceMode={setServiceMode}
            operationType={ServiceMode.Install}
            sid={sid}
            changeSid={setSid}
          />
        );
      // 更新
      case ServiceMode.Update:
        return (
          <OperationService
            changeServiceMode={setServiceMode}
            operationType={ServiceMode.Update}
            updateServiceRecord={[{ ...updateServiceRecord }]}
            sid={sid}
            changeSid={setSid}
          />
        );
      // 回退版本
      case ServiceMode.Revert:
        return (
          <RevertService
            jid={updateRecordJid}
            changeServiceMode={setServiceMode}
            updateServiceRecord={updateServiceRecord}
            sid={sid}
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
            changeMainServiceId={setMainServiceId}
            mainServiceId={mainServiceId}
            serviceId={serviceId}
            serviceModeType={ServiceMode.Service}
            changeSid={setSid}
          />
        );
      // 微服务详情
      case ServiceMode.MicroService:
        return (
          <ServiceDetail
            key={Date.now()}
            changeServiceMode={setServiceMode}
            changeServiceId={setServiceId}
            mainServiceId={mainServiceId}
            serviceId={serviceId}
            serviceModeType={ServiceMode.MicroService}
          />
        );
      // 服务管理首页
      default:
        return (
          <ServiceHome
            changeServiceMode={setServiceMode}
            changeServiceId={setServiceId}
            changeUpdateServiceRecord={setUpdateServiceRecord}
            changeMainServiceId={setMainServiceId}
            changeSid={setSid}
          />
        );
    }
  };

  return getComponents();
};

import React, { FC, useState } from "react";
import { ContentLayout, ElementText } from "../../../common/components";
import { ChooseService } from "./choose-service";
import { CreateConfig } from "./create-config";
import { ConfirmConfig } from "./confirm-config";
import elementStyles from "../../../common/components/element-text/style.module.less";
import {
  Breadcrumb,
  Steps,
  Space,
  Button,
  Modal,
  message,
  Table,
} from "@kweaver-ai/ui";
import { LeftOutlined } from "@kweaver-ai/ui/icons";
import { formValidatorComfirm, handleError } from "../../utils/handleError";
import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { IBaseProps } from "../declare";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import {
  ApplicationItem,
  DependenciesListItem,
  ISortServiceParams,
} from "../../../../api/service-management/service-deploy/declare";
import {
  serviceApplication,
  serviceJob,
} from "../../../../api/service-management/service-deploy";
import { OperationServiceStepsEnum, OperationType } from "./type.d";
import { assignTo } from "../../../../tools/browser";
import { deployMiniPathname } from "../../../../core/path";
import { SERVICE_PREFIX } from "../../config";
import styles from "./styles.module.less";
import __ from "./locale";
import { noop } from "lodash";
import { ServiceSchemaItem } from "../../../../api/suite-management/suite-deploy/declare";
import { ServiceConfig } from "../../service-config";
import { ConfigEditStatusEnum } from "../../service-config/helper";
import { formValidatorComfirm as batchFormValidatorComfirm } from "../../../suite-management/utils/handleError";
import { composeJob } from "../../../../api/suite-management/suite-deploy";
import {
  DefaultServiceDeployValidateState,
  ValidateState,
} from "../../utils/validator";

interface IProps extends IBaseProps {
  // 更新服务信息
  updateServiceRecord?: ApplicationItem[];
  // 操作服务类型（安装、更新或批量更新）
  operationType: OperationType;
  // 系统空间id
  sid: number;
  // 修改系统空间id
  changeSid: (sid: number) => void;
}
export const OperationService: FC<IProps> = ({
  changeServiceMode,
  updateServiceRecord,
  operationType,
  sid,
  changeSid,
}) => {
  const [current, setCurrent] = useState<OperationServiceStepsEnum>(
    OperationServiceStepsEnum.CHOOSESERVICE
  );
  // 配置项formData
  const [formData, setFormData] = useState<RJSFSchema>({});
  // 配置项schema
  const [schema, setSchema] = useState<RJSFSchema>({});
  // 配置项UIschema
  const [uiSchema, setUISchema] = useState<UiSchema>({});
  // 备注内容
  const [comment, setComment] = useState<string>("");
  // 配置项表单是否通过验证
  const [isFormValidator, setIsFormValidator] = useState<boolean>(false);
  // 控制模板展示开关
  const [templateOpen, setTemplateOpen] = useState<boolean>(false);
  // 选中的配置模板id
  const [selectedTid, setSelectedTid] = useState<number[]>([]);
  // 当应用版本无法被后端解析报错时，用户可以自定义获取配置模板的版本
  const [customVersion, setCustomVersion] = useState<string>("");
  // 批量任务名称
  const [batchJobName, setBatchJobName] = useState("");
  // 选中的服务列表
  const [selectedServiceList, setSelectedServiceList] = useState<
    DependenciesListItem[]
  >((updateServiceRecord as DependenciesListItem[]) || []);
  // 批量任务各服务配置项
  const [serviceConfig, setServiceConfig] = useState<ServiceSchemaItem[]>([]);
  // 确定按钮是否加载中
  const [loading, setLoading] = useState(false);
  const [checkDependenciesCallback, setCheckDependenciesCallback] =
    useState<any>(noop);
  // 校验状态
  const [validateState, setValidateState] = useState(
    DefaultServiceDeployValidateState
  );

  // 排序数据并跳转下一步
  const handleSortServiceList = (sorted: ISortServiceParams[]) => {
    const sortedServiceList = sorted.map((service) => {
      return selectedServiceList.find((item) => item.name === service.name)!;
    });
    setServiceConfig(
      sortedServiceList.map((service) => ({
        ...service,
        editStatus: ConfigEditStatusEnum.Unsubmitted,
      })) as any
    );
    setCurrent(current + 1);
  };

  /**
   * @description 点击下一步
   */
  const handeleNextClick = async () => {
    if (current === OperationServiceStepsEnum.CHOOSESERVICE) {
      if (operationType === ServiceMode.Install) {
        if (selectedServiceList.some((service) => !service.select)) {
          checkDependenciesCallback(true);
          return;
        } else {
          checkDependenciesCallback(false);
        }
      }
      if (selectedServiceList.length > 1) {
        if (!batchJobName) {
          setValidateState({
            ...validateState,
            BatchJobName: ValidateState.Empty,
          });
          return;
        }
        try {
          const { sorted, outer } = await serviceApplication.sortService(
            sid,
            selectedServiceList
          );
          if (outer?.length) {
            missDependenceConfirm(outer, true, sorted);
            return;
          }
          handleSortServiceList(sorted);
        } catch (error: any) {
          handleError(error);
        }
      } else {
        // 判断是否已安装依赖服务
        try {
          const res = await serviceJob.getJSONSchemaSnapshot(
            selectedServiceList[0]?.aid,
            {
              tid: [...selectedTid].reverse(),
              sid,
            }
          );
          setFormData(res.formData!);
          setSchema(res.schema!);
          setUISchema(res.uiSchema || {});
          setComment("");
          setCurrent(current + 1);
        } catch (error: any) {
          handleError(error);
        }
      }
    } else if (current === OperationServiceStepsEnum.CREATECONFIG) {
      if (selectedServiceList.length > 1) {
        if (
          serviceConfig.every(
            (config) => config.editStatus === ConfigEditStatusEnum.Submitted
          )
        ) {
          setCurrent(current + 1);
        } else {
          batchFormValidatorComfirm();
        }
      } else {
        if (isFormValidator) {
          setCurrent(current + 1);
        } else {
          formValidatorComfirm();
        }
      }
    }
  };

  // 【下一步】禁用逻辑
  const getIsDisabled = () => {
    if (current === 0) {
      switch (true) {
        // // 存在未选版本的服务
        // case selectedServiceList.some((service) => !service.version):
        //   return true;
        // 未选任何服务
        case selectedServiceList.length === 0:
          return true;
        default:
          return false;
      }
    }
  };

  // 表格的列配置项
  const columns = [
    {
      title: __("依赖服务名称"),
      dataIndex: "title",
      ellipsis: true,
    },
    {
      title: __("依赖服务标识"),
      dataIndex: "name",
      ellipsis: true,
    },
  ];

  const missDependenceConfirm = (
    message: any,
    isBatchJob: boolean,
    sorted?: ISortServiceParams[]
  ) => {
    // 缺失依赖服务错误提示
    Modal.confirm({
      title: __("依赖服务缺失"),
      wrapClassName: `${SERVICE_PREFIX}-warning-modal`,
      content: (
        <React.Fragment>
          <div>
            {__(
              "您正在安装的服务存在依赖服务未安装，会导致当前服务无法正常安装，请先安装依赖服务。"
            )}
          </div>
          {isBatchJob ? (
            <Table
              dataSource={message}
              columns={columns}
              scroll={{
                y: message.length > 5 ? "200px" : undefined,
              }}
              pagination={{
                pageSize: 1000,
                hideOnSinglePage: true,
              }}
              rowKey="name"
            />
          ) : (
            <div>{__("依赖服务名称：${message}", { message })}</div>
          )}
        </React.Fragment>
      ),
      onOk: isBatchJob
        ? () => {
            handleSortServiceList(sorted!);
          }
        : () => {
            changeServiceMode(ServiceMode.Home);
          },
      onCancel: () => {
        changeServiceMode(ServiceMode.Home);
      },
      cancelButtonProps: isBatchJob
        ? undefined
        : {
            style: {
              display: "none",
            },
          },
      cancelText:
        operationType === ServiceMode.Install ? __("取消安装") : __("取消更新"),
    });
  };

  /**
   * @description 确定安装或更新
   */
  const handleOperationClick = async () => {
    try {
      setLoading(true);
      if (serviceConfig.length > 1) {
        const appsPayload = serviceConfig.map((config) => {
          return {
            name: config.name,
            version: config.version,
            formData: config.formData,
            aid: config.aid,
          };
        });
        const payload = {
          config: { apps: appsPayload, pcomponents: null },
          jname: batchJobName,
          description: operationType === ServiceMode.Install ? "安装" : "更新",
          sid,
        };
        await composeJob.createJob(payload);
      } else {
        // 配置任务并执行
        await serviceJob.createAndExecuteJob(
          operationType === ServiceMode.Install
            ? {
                formData,
                aid: selectedServiceList[0]?.aid,
                comment,
                sid,
              }
            : {
                formData,
                aid: selectedServiceList[0]?.aid,
                sid,
              }
        );
      }
      // 成功提示
      message.success(
        <ElementText
          text={
            operationType === ServiceMode.Install
              ? __("安装更新服务任务创建成功，前往-查看")
              : __("更新服务任务创建成功，前往-查看")
          }
          insert={
            <a className={elementStyles["target-herf"]} onClick={clickHerf}>
              {__("【任务监控】")}
            </a>
          }
        />
      );
      changeServiceMode(ServiceMode.Home);
    } catch (error: any) {
      if (error?.status === 412 && error?.code === 16) {
        missDependenceConfirm(JSON.parse(error?.message)?.Detail, false);
      } else {
        handleError(error);
      }
    } finally {
      setLoading(false);
    }
  };
  /**
   * @description 跳转到任务监控
   */
  const clickHerf = () => {
    assignTo(deployMiniPathname.taskMonitorPathname);
  };
  const items = [
    {
      title:
        operationType === ServiceMode.Install ? __("选择服务") : __("选择版本"),
    },
    {
      title: __("填写配置项"),
    },
    {
      title: __("确认信息"),
    },
  ];
  /**
   * @description 获取操作服务的内容
   * @return 操作服务当前步骤的组件
   */
  const getContent = () => {
    if (current === OperationServiceStepsEnum.CHOOSESERVICE) {
      return (
        <ChooseService
          operationType={operationType}
          templateOpen={templateOpen}
          changeTemplateOpen={setTemplateOpen}
          selectedTid={selectedTid}
          changeSelectedTid={setSelectedTid}
          customVersion={customVersion}
          changeCustomVersion={setCustomVersion}
          batchJobName={batchJobName}
          changeBatchJobName={setBatchJobName}
          selectedServiceList={selectedServiceList as DependenciesListItem[]}
          changeSelectedServiceList={setSelectedServiceList}
          sid={sid}
          changeSid={changeSid}
          setCheckDependenciesCallback={setCheckDependenciesCallback}
          validateState={validateState}
          changeValidateState={setValidateState}
        />
      );
    } else if (current === OperationServiceStepsEnum.CREATECONFIG) {
      return (
        <>
          {selectedServiceList.length > 1 ? (
            <ServiceConfig
              serviceConfig={serviceConfig}
              changeServiceConfig={setServiceConfig}
              sid={sid}
            />
          ) : (
            <CreateConfig
              formData={formData}
              uiSchema={uiSchema}
              schema={schema}
              onChangeFormData={setFormData}
              changeIsFormValidator={setIsFormValidator}
            />
          )}
        </>
      );
    } else if (current === OperationServiceStepsEnum.CONFIRMCONFIG) {
      return (
        <ConfirmConfig
          comment={comment}
          formData={formData}
          schema={schema}
          uiSchema={uiSchema}
          operationType={operationType}
          changeComment={setComment}
          serviceConfig={
            selectedServiceList.length > 1 ? serviceConfig : selectedServiceList
          }
          batchJobName={batchJobName}
        />
      );
    }
  };
  const breadcrumb = (
    <Breadcrumb className={styles["operation-breadcrumb"]}>
      <div onClick={() => changeServiceMode(ServiceMode.Home)}>
        <LeftOutlined
          className={styles["breadcrumb-icon"]}
          onPointerEnterCapture={noop}
          onPointerLeaveCapture={noop}
        />
      </div>
      <Breadcrumb.Item onClick={() => changeServiceMode(ServiceMode.Home)}>
        <span className={styles["breadcrumb-title"]}>
          {operationType === ServiceMode.Install
            ? __("安装更新服务")
            : __("更新服务")}
        </span>
      </Breadcrumb.Item>
    </Breadcrumb>
  );
  const footer = (
    <Space>
      {current > OperationServiceStepsEnum.CHOOSESERVICE && (
        <Button type="default" onClick={() => setCurrent(current - 1)}>
          {__("上一步")}
        </Button>
      )}
      {current < items.length - 1 && (
        <Button
          type="primary"
          disabled={getIsDisabled()}
          onClick={handeleNextClick}
        >
          {__("下一步")}
        </Button>
      )}
      {current === items.length - 1 && (
        <Button
          type="primary"
          onClick={handleOperationClick}
          loading={loading}
          key={loading ? "true" : "false"}
        >
          {__("确定")}
        </Button>
      )}
      <Button
        type="default"
        onClick={() => changeServiceMode(ServiceMode.Home)}
      >
        {__("取消")}
      </Button>
    </Space>
  );
  return (
    <ContentLayout
      breadcrumb={breadcrumb}
      footer={footer}
      contentClassName={styles["service-operation-content"]}
      moduleName={SERVICE_PREFIX}
    >
      <div className={styles["operation-steps"]}>
        <Steps current={current} items={items} />
      </div>
      {getContent()}
    </ContentLayout>
  );
};

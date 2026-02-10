import React, { FC, useEffect, useState } from "react";
import {
  Dropdown,
  Button,
  Space,
  Drawer,
  message,
  Table,
} from "@kweaver-ai/ui";
import { CaretDownOutlined, PlusOutlined } from "@kweaver-ai/ui/icons";
import type {
  MenuProps,
  TableColumnType,
  TableColumnsType,
} from "@kweaver-ai/ui";
import { componentManage } from "../../api/component-manage";
import {
  ConfigData,
  OperationType,
  SERVICES,
  buildInComponents,
  buildInComponentsText,
  defaultAddableComponents,
  updateComponentLog,
} from "./helper";
import { handleError } from "../service-management/utils/handleError";
import styles from "./styles.module.less";
import __ from "./locale";
import { DataBaseConfig } from "./database-config";
import { assignTo } from "../../tools/browser";
import { deployMiniPathname } from "../../core/path";
import { noop } from "lodash";
import { CustomSpin } from "./components/spin";
import { ContentLayout } from "../common/components";
import { SERVICE_PREFIX } from "./config";
import { formatTableResponse } from "../common/utils/request";
import { safetyStr } from "../common/utils/string";
import {
  ComponentListItem,
  IGetComponentListParams,
  IGetComponentListTableParams,
} from "../../api/component-manage/declare";

export const ComponentManagement: FC = () => {
  // 操作类型
  const [operationType, setOperationType] = useState<OperationType>(
    OperationType.Edit
  );
  // 组件类型
  const [componentType, setComponentType] = useState<SERVICES>(SERVICES.Kafka);
  // 控制抽屉开关
  const [open, setOpen] = useState<boolean>(false);
  // 修改后的内置组件数据
  const [componentConfigData, setComponentConfigData] = useState<ConfigData>(
    {} as ConfigData
  );
  // 内置组件表单校验
  const [componentForm, setComponentForm] = useState({});
  // 是否正在保存配置
  const [isLoading, setIsLoading] = useState<boolean>(false);
  // 编辑的组件实例名称
  const [componentName, setComponentName] = useState("");
  // 系统空间filter筛选项开关
  const [filterOpen, setFilterOpen] = useState(false);
  // 单实例模式下 可添加的内置组件
  const [addableComponentsList, setAddableComponentsList] = useState<
    SERVICES[]
  >(defaultAddableComponents);

  useEffect(() => {
    getAddableComponents();
  }, []);

  // 表格的基本设置
  const { state, api } = Table.useTable<
    ComponentListItem,
    IGetComponentListTableParams,
    ComponentListItem[]
  >({
    request: (params) => {
      const { _filter, current, pageSize } = params;
      const formatedParams: IGetComponentListParams = {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        type: getFilterParams(
          (_filter as any)?.type,
          Object.values(buildInComponents).length
        ),
        sid: [-1, undefined].includes((_filter as any)?.systemName?.[0])
          ? undefined
          : (_filter as any)?.systemName?.[0],
      };
      return componentManage.getComponentList(formatedParams);
    },
    rowKey: "name",
    pagination: {
      showTotal: (total) => __("共${total}条", { total }),
    },
    ...formatTableResponse(),
  });

  const { reload } = api;

  const getFilterParams = (
    value: Array<any> | undefined,
    filterOptionCount: number
  ) => {
    if (value?.length) {
      if (value.length === filterOptionCount) {
        return undefined;
      } else {
        return value;
      }
    }
    return undefined;
  };

  // 表格的列配置项
  const columns: TableColumnsType<ComponentListItem> = [
    {
      title: __("组件名称"),
      dataIndex: "name",
    },
    {
      title: __("组件类型"),
      dataIndex: "type",
      filters: buildInComponents.map((buildInComponent) => {
        return {
          value: buildInComponent,
          text: buildInComponentsText[buildInComponent],
        };
      }),
      render: (value) => buildInComponentsText[value],
      tooltip: (value) => buildInComponentsText[value],
    },
    {
      title: __("系统空间"),
      dataIndex: "systemName",
      render: (text) => safetyStr(text),
      tooltip: (text) => safetyStr(text),
    },
    {
      title: __("系统空间ID"),
      dataIndex: "sid",
      render: (value) => safetyStr(value),
      tooltip: (value) => safetyStr(value),
    },
    {
      title: __("操作"),
      render: (_, record: ComponentListItem) => (
        <Button type="link" onClick={() => handleEditComponent(record)}>
          {__("编辑")}
        </Button>
      ),
      tooltip: () => __("编辑"),
    },
  ];

  const handleEditComponent = (record: ComponentListItem) => {
    setOperationType(OperationType.Edit);
    setComponentType(record.type as SERVICES);
    setComponentName(record.name);
    setOpen(true);
  };

  const clickHerf = () => {
    assignTo(deployMiniPathname.serviceDeployPathname);
  };

  const handleUpdateComponent = () => {
    // 表单校验
    const dataBaseFormCheck = Object.values(componentForm).map(
      async (form: any) => {
        return await form?.current?.validateFields();
      }
    );
    Promise.all([...dataBaseFormCheck])
      .then(async () => {
        try {
          setIsLoading(true);
          if (componentType !== SERVICES.Kafka) {
            let payload;
            if (operationType === OperationType.Add) {
              payload = {
                ...componentConfigData[componentType],
                type: componentType,
              };
            } else {
              payload = componentConfigData[componentType];
            }
            await componentManage.putComponentInfo(componentType, payload!);
            updateComponentLog(operationType, payload?.params, componentType);
          } else {
            // 特殊处理kafka，同时修改kafka和zookeeper组件信息
            let kafkaPayload, zookeeperPayload;
            if (operationType === OperationType.Add) {
              kafkaPayload = {
                ...componentConfigData[SERVICES.Kafka],
                type: SERVICES.Kafka,
              };
              zookeeperPayload = {
                ...componentConfigData[SERVICES.Zookeeper],
                name: componentConfigData[SERVICES.Kafka]?.dependencies
                  ?.zookeeper,
                type: SERVICES.Zookeeper,
              };
            } else {
              kafkaPayload = componentConfigData[SERVICES.Kafka];
              zookeeperPayload = componentConfigData[SERVICES.Zookeeper];
            }
            await componentManage.putComponentInfo(
              SERVICES.Zookeeper,
              zookeeperPayload!
            );
            await componentManage.putComponentInfo(
              SERVICES.Kafka,
              kafkaPayload!
            );
            updateComponentLog(
              operationType,
              kafkaPayload?.params,
              SERVICES.Kafka
            );
            updateComponentLog(
              operationType,
              zookeeperPayload?.params,
              SERVICES.Zookeeper
            );
          }
          message.success(
            <div style={{ display: "inline-block" }}>
              {__(
                "修改成功，配置将在服务更新后生效。请前往【服务管理-服务部署】页面更新服务。"
              )}
              <a
                style={{
                  marginLeft: "10px",
                  color: "#126EE3",
                }}
                onClick={clickHerf}
              >
                {__("立即前往")}
              </a>
            </div>
          );
          // 更新成功后重新获取组件信息，渲染组件card
          reload();
          getAddableComponents();
          handleClose();
        } catch (error) {
          handleError(error);
        } finally {
          setIsLoading(false);
        }
      })
      .catch(() => {});
  };

  // 关闭抽屉
  const handleClose = () => {
    setOpen(false);
    setComponentConfigData({});
    setComponentForm({});
  };

  // 获取可添加组件列表
  const getAddableComponents = async () => {
    const newAddableComponentsList = await Promise.all(
      defaultAddableComponents.map(async (service) => {
        try {
          const { totalNum } = (await componentManage.getComponentList({
            offset: 0,
            limit: 1,
            type: [service],
          })) as any;
          if (totalNum === 0) {
            return service;
          }
        } catch (e) {
          handleError(e);
        }
      })
    );

    setAddableComponentsList(
      newAddableComponentsList.filter((item) => item) as SERVICES[]
    );
  };

  // 点击添加可选组件
  const handleMenuClick: MenuProps["onClick"] = (e) => {
    setOperationType(OperationType.Add);
    setComponentType(e.key as SERVICES);
    setOpen(true);
  };

  // 下拉组件item
  const items: MenuProps["items"] = defaultAddableComponents.map(
    (addableComponent) => {
      if (addableComponentsList.includes(addableComponent)) {
        return {
          label: buildInComponentsText[addableComponent],
          key: addableComponent,
        };
      } else {
        return null;
      }
    }
  );

  const menuProps = {
    items,
    onClick: handleMenuClick,
  };

  const header = (
    <div className={styles["dropdown"]}>
      <Dropdown
        menu={menuProps}
        trigger={["click"]}
        disabled={!addableComponentsList.length}
      >
        <Button type="primary">
          <Space>
            <PlusOutlined
              onPointerEnterCapture={noop}
              onPointerLeaveCapture={noop}
            />
            {__("添加可选组件")}
            <CaretDownOutlined
              onPointerEnterCapture={noop}
              onPointerLeaveCapture={noop}
            />
          </Space>
        </Button>
      </Dropdown>
    </div>
  );
  return (
    <>
      <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
        <Table {...state} columns={columns} />
      </ContentLayout>
      <Drawer
        title={
          operationType === OperationType.Add
            ? __("添加-${component}", {
                component: buildInComponentsText[componentType],
              })
            : __("编辑-${component}", {
                component: buildInComponentsText[componentType],
              })
        }
        width={1000}
        onClose={handleClose}
        onOk={handleUpdateComponent}
        open={open}
        destroyOnClose
      >
        <DataBaseConfig
          component={componentType}
          operationType={operationType}
          componentConfigData={componentConfigData}
          setComponentConfigData={setComponentConfigData}
          setComponentForm={setComponentForm}
          showTitle={false}
          componentName={componentName}
        />
      </Drawer>
      {isLoading ? <CustomSpin text={__("保存配置中...")} /> : null}
    </>
  );
};

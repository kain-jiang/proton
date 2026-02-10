import React, { FC, useEffect, useMemo, useState } from "react";
import { ContentLayout } from "../../../../common/components";
import {
  Form,
  Input,
  Button,
  Drawer,
  Table,
  Select,
  Switch,
  message,
  Space,
  Tag,
  Tooltip,
  Popover,
} from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { ConfigTemplate } from "../config-template";
import { handleError } from "../../../utils/handleError";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import { serviceApplication } from "../../../../../api/service-management/service-deploy";
import {
  ApplicationItem,
  DependenciesListItem,
  DependenciesServiceItem,
} from "../../../../../api/service-management/service-deploy/declare";
import { OperationType } from "../type.d";
import styles from "./styles.module.less";
import { SERVICE_PREFIX } from "../../../config";
import __ from "./locale";
import { PlusOutlined } from "@kweaver-ai/ui/icons";
import { noop } from "lodash";
import { FormLabel } from "../../../../common/components/form-label";
import { system } from "../../../../../api/tenant-management";
import { selectStatus } from "../../type.d";
import { Dot } from "../../../../common/components/text/dot";
import { CommonEnum } from "../../../utils/common.type";
import {
  ServiceDeployValidateState,
  ValidateState,
} from "../../../utils/validator";

interface IProps {
  // 操作服务的类型（安装、更新或批量更新）
  operationType: OperationType;
  // 配置模板开关
  templateOpen: boolean;
  // 修改配置模板开关
  changeTemplateOpen: (item: boolean) => void;
  // 选中的配置模板id
  selectedTid: number[];
  // 修改选中的配置模板id
  changeSelectedTid: (item: number[]) => void;
  // 自定义版本
  customVersion: string;
  // 修改自定义版本
  changeCustomVersion: (item: string) => void;
  // 批量任务名称
  batchJobName: string;
  // 修改批量任务名称
  changeBatchJobName: (name: string) => void;
  // 选中的所有服务列表
  selectedServiceList: DependenciesListItem[];
  // 修改选中的所有服务列表
  changeSelectedServiceList: (selectedServiceList: any) => void;
  // 系统空间id
  sid: number;
  // 修改系统空间id
  changeSid: (sid: number) => void;
  // 检查依赖服务是否均已确认
  setCheckDependenciesCallback: (fn: any) => void;
  // 字段校验状态
  validateState: ServiceDeployValidateState;
  // 修改校验状态
  changeValidateState: (validateState: ServiceDeployValidateState) => void;
}
export const ChooseService: FC<IProps> = ({
  operationType,
  templateOpen,
  changeTemplateOpen,
  selectedTid,
  changeSelectedTid,
  customVersion,
  changeCustomVersion,
  batchJobName,
  changeBatchJobName,
  selectedServiceList,
  changeSelectedServiceList,
  sid,
  changeSid,
  setCheckDependenciesCallback,
  validateState,
  changeValidateState,
}) => {
  // 控制滑窗
  const [open, setOpen] = useState<boolean>(false);
  // 滑窗选中的服务列表
  const [tempSelectedServiceList, setTempSelectedServiceList] = useState<
    ApplicationItem[]
  >([]);
  // 完整应用名称表格数据
  const [dataSource, setDataSource] = useState<ApplicationItem[]>([]);
  // 各服务可选版本（共用一个字段）
  const [versionOptions, setVersionOptions] = useState<
    {
      label: string;
      value: string;
    }[]
  >([]);
  // 系统空间列表选项
  const [systemOptions, setSystemOptions] = useState<
    { label: string; value: number }[]
  >([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  // 是否有未确认的服务
  const [existUnconfirmedService, setExistUnconfirmedService] = useState(false);

  // 滑窗过滤掉已经选择过的服务
  const filteredDataSource = useMemo(() => {
    return dataSource.filter((item) => {
      return !selectedServiceList.some((service) => service.name === item.name);
    });
  }, [dataSource, selectedServiceList]);

  // 更新单服务获取版本下拉列表
  useEffect(() => {
    if (operationType === ServiceMode.Update) {
      onDropdownVisibleChange(true, selectedServiceList[0]);
    }
  }, [selectedServiceList[0]?.name]);

  useEffect(() => {
    if (operationType !== ServiceMode.Update) {
      getSystemOptions();
    }
  }, []);

  useEffect(() => {
    setCheckDependenciesCallback(() => {
      return (isExist: boolean) => setExistUnconfirmedService(isExist);
    });
  }, [setCheckDependenciesCallback, setExistUnconfirmedService]);

  // 表格的列配置项
  const columns: TableColumnsType<ApplicationItem> = [
    {
      title: __("名称"),
      dataIndex: "title",
    },
  ];

  // 表格的列配置项
  const columnsForSelectedTable: TableColumnsType<DependenciesListItem> = [
    {
      title: __("名称"),
      dataIndex: "title",
    },
    {
      title: __("版本"),
      dataIndex: "version",
      width: 290,
      render: (value: string, record: DependenciesListItem) => {
        return (
          <Select
            style={{ width: "250px" }}
            disabled={record.select}
            value={record.version}
            onChange={(value, option) =>
              onDependenciesVersionChange(value, option, record)
            }
            listItemHeight={32}
            onDropdownVisibleChange={(open) =>
              onDropdownVisibleChange(open, record)
            }
            options={versionOptions}
          />
        );
      },
      tooltip: () => null,
    },
    {
      title: __("是否确认使用当前版本"),
      dataIndex: "select",
      render: (value: any) => {
        return (
          <Dot color={selectStatus[value].color}>
            {selectStatus[value].text}
          </Dot>
        );
      },
      tooltip: (value: any) => {
        return selectStatus[value].text;
      },
    },
    {
      title: __("执行内容"),
      dataIndex: "installed",
      render: (value: boolean) =>
        value === true
          ? __("更新")
          : value === false
          ? __("安装")
          : CommonEnum.PLACEHOLDER,
      tooltip: (value: boolean) =>
        value === true
          ? __("更新")
          : value === false
          ? __("安装")
          : CommonEnum.PLACEHOLDER,
    },
    {
      title: __("依赖服务"),
      dataIndex: "dependencies",
      width: 400,
      render: (value: DependenciesServiceItem[]) => {
        return renderDependenciesTags(value);
      },
      tooltip: () => null,
    },
    {
      title: __("操作"),
      width: 140,
      render: (_, record: DependenciesListItem) => {
        return (
          <Space>
            <Button
              type="link"
              disabled={!!record.select || !record.version}
              onClick={() => handleConfirmService([record])}
            >
              {__("确认")}
            </Button>
            <Button type="link" onClick={() => handleDeleteService(record)}>
              {__("删除")}
            </Button>
          </Space>
        );
      },
      tooltip: () => null,
    },
  ];

  // 确认服务获取新依赖数据列表
  const handleConfirmService = async (record?: DependenciesListItem[]) => {
    if (!record) {
      try {
        const result = await serviceApplication.getDependenciesList(
          sid,
          selectedServiceList.map((service) => ({
            name: service.name,
            version: service.version,
            select: true,
          }))
        );

        changeSelectedServiceList([...Object.values(result)]);
        setExistUnconfirmedService(false);
        setSelectedRowKeys([]);
      } catch (e: any) {
        if (e?.status === 412 && e?.code === 48) {
          handleError({
            ...e,
            message: __("缺失依赖服务：${service}", {
              service: JSON.parse(e?.message)?.Detail,
            }),
          });
        } else {
          handleError(e);
        }
      }
      return;
    }
    const payload = [
      ...selectedServiceList
        .filter((service) => service.select)
        .map((service) => ({
          name: service.name,
          version: service.version,
          select: true,
        })),
      ...record.map((service) => ({
        name: service.name,
        version: service.version,
        select: true,
      })),
    ];
    try {
      const result = await serviceApplication.getDependenciesList(sid, payload);
      const newServicesList = Object.values(result);
      const unselectedList = selectedServiceList
        .filter((service) => !service.select)
        .filter(
          (service) =>
            !newServicesList.some((item) => item.name === service.name)
        );
      changeSelectedServiceList([...newServicesList, ...unselectedList]);
      setExistUnconfirmedService(false);
      setSelectedRowKeys([]);
    } catch (e: any) {
      if (e?.status === 412 && e?.code === 48) {
        handleError({
          ...e,
          message: __("缺失依赖服务：${service}", {
            service: JSON.parse(e?.message)?.Detail,
          }),
        });
      } else {
        handleError(e);
      }
    }
  };

  const renderDependenciesTags = (
    dependenciesList: DependenciesServiceItem[]
  ) => {
    if (!dependenciesList?.length) return CommonEnum.PLACEHOLDER;
    if (dependenciesList.length <= 2) {
      return (
        <Space>
          {dependenciesList.map((item) => {
            return (
              <Tag className={styles["tag-item"]} title={item.name}>
                {item.name}
              </Tag>
            );
          })}
        </Space>
      );
    }
    return (
      <Space>
        <Tag className={styles["tag-item"]} title={dependenciesList[0].name}>
          {dependenciesList[0].name}
        </Tag>
        <Tag className={styles["tag-item"]} title={dependenciesList[1].name}>
          {dependenciesList[1].name}
        </Tag>
        <Popover
          content={
            <Space>
              {dependenciesList.map((item, index) => {
                if (index < 2) return null;
                return (
                  <Tag className={styles["tag-item"]} title={item.name}>
                    {item.name}
                  </Tag>
                );
              })}
            </Space>
          }
        >
          <Tag>{`+${dependenciesList.length - 2}`}</Tag>
        </Popover>
      </Space>
    );
  };

  const rowSelectionForSelectTable = {
    selectedRowKeys: selectedRowKeys,
    onChange: (
      selectedRowKeys: React.Key[],
      selectedRows: DependenciesListItem[]
    ) => {
      setSelectedRowKeys(selectedRowKeys);
    },
    getCheckboxProps: (record: DependenciesListItem) => ({
      disabled: record.select === true, // Column configuration not to be checked
    }),
  };

  // 获取系统空间列表选项
  const getSystemOptions = async () => {
    try {
      const data = await system.get({
        offset: 0,
        limit: 10000,
        mode: true,
      });
      setSystemOptions(
        data.map((item) => ({
          label: item.systemName,
          value: item.sid!,
        }))
      );
    } catch (e) {
      handleError(e);
    }
  };

  /**
   * @description 确定滑窗
   */
  const onOk = () => {
    const newSelectedServiceList: DependenciesListItem[] =
      tempSelectedServiceList.map((item) => {
        return {
          ...item,
          select: false,
          installed: undefined,
          versions: [],
          dependencies: [],
        };
      });
    changeSelectedServiceList([
      ...selectedServiceList,
      ...newSelectedServiceList,
    ]);
    setExistUnconfirmedService(false);

    changeSelectedTid([]);
    changeTemplateOpen(false);
    changeCustomVersion("");
    // 重置滑窗数据
    onCancel();
  };

  /**
   * @description 取消滑窗
   */
  const onCancel = () => {
    setOpen(false);
    setTempSelectedServiceList([]);
  };
  /**
   * @description 打开滑窗
   */
  const handleChooseClick = () => {
    setOpen(true);
    getApplicationItems();
  };
  /**
   * @description 获取滑窗内不同名称的服务列表
   */
  const getApplicationItems = async () => {
    try {
      const res = await serviceApplication.getApplication({
        offset: 0,
        limit: 10000,
        sid,
      });
      setDataSource(res || []);
    } catch (error: any) {
      handleError(error);
    }
  };

  const rowSelection = {
    onChange: (
      selectedRowKeys: React.Key[],
      selectedRows: ApplicationItem[]
    ) => {
      setTempSelectedServiceList(selectedRows);
    },
  };

  // 修改系统空间后，清空所有state
  const onSystemChange = (value: number) => {
    setSelectedRowKeys([]);
    changeSelectedServiceList([]);
    setExistUnconfirmedService(false);
    changeSid(value);
    changeSelectedTid([]);
    changeTemplateOpen(false);
    changeCustomVersion("");
    changeBatchJobName("");
  };

  // 修改安装/批量更新某服务版本
  const onVersionChange = (
    value: any,
    option: any,
    selectedservice: ApplicationItem
  ) => {
    changeSelectedServiceList(
      selectedServiceList.map((service) => {
        if (service.name === selectedservice.name) {
          return {
            ...service,
            version: option.label,
            aid: option.aid,
          };
        } else {
          return service;
        }
      })
    );
    setExistUnconfirmedService(false);

    changeSelectedTid([]);
    changeCustomVersion("");
  };

  // 修改安装/批量更新某服务版本
  const onDependenciesVersionChange = (
    value: string,
    option: any,
    record: DependenciesListItem
  ) => {
    changeSelectedServiceList(
      selectedServiceList.map((item) => {
        if (item.name === record.name) {
          return { ...item, version: value };
        } else {
          return { ...item };
        }
      })
    );
    setExistUnconfirmedService(false);

    changeSelectedTid([]);
    changeCustomVersion("");
  };

  const onDropdownVisibleChange = async (
    open: boolean,
    service: DependenciesListItem | ApplicationItem
  ) => {
    if (open) {
      if ((service as DependenciesListItem)?.versions?.length) {
        setVersionOptions(
          (service as DependenciesListItem)?.versions.map(
            (version: string) => ({
              value: version,
              label: version,
            })
          )
        );
        return;
      }
      try {
        const res = await serviceApplication.getApplication({
          offset: 0,
          limit: 10000,
          name: service.name,
        });
        setVersionOptions(
          res.map((item) => {
            return {
              label: item.version,
              value: item.version,
              aid: item.aid,
            };
          })
        );
      } catch (error: any) {
        handleError(error);
      }
    } else {
      setVersionOptions([]);
    }
  };

  // 删除某个服务
  const handleDeleteService = (service: ApplicationItem) => {
    if (selectedServiceList.length === 2) {
      changeBatchJobName("");
      changeValidateState({
        ...validateState,
        BatchJobName: ValidateState.Normal,
      });
    }
    changeSelectedServiceList(
      selectedServiceList.filter((item) => item.name !== service.name)
    );
    setExistUnconfirmedService(false);
    setSelectedRowKeys(selectedRowKeys.filter((key) => key !== service.aid));

    changeSelectedTid([]);
    changeTemplateOpen(false);
    changeCustomVersion("");
  };

  /**
   * @description 修改开关状态
   * @param checked
   */
  const handleSwitchChange = (checked: boolean) => {
    if (selectedServiceList[0]?.version) {
      changeTemplateOpen(checked);
      changeSelectedTid([]);
    } else {
      message.info(__("请先选择服务及其版本。"));
    }
  };

  const windowHeight = window.innerHeight > 720 ? "100vh" : "720px";

  return (
    <>
      <Form labelCol={{ flex: "130px" }} labelAlign="left">
        {operationType !== ServiceMode.Update ? (
          <Form.Item label={__("选择系统空间")} required={true}>
            <Select
              value={sid || ""}
              style={{ width: 490 }}
              onChange={(value) => onSystemChange(value as number)}
              listItemHeight={32}
              options={systemOptions}
            />
          </Form.Item>
        ) : null}
        {operationType === ServiceMode.Update ? (
          <Form.Item label={__("版本")} required={true}>
            <Select
              value={selectedServiceList[0]?.version || ""}
              style={{ width: 350 }}
              onChange={(value, option) =>
                onVersionChange(value, option, selectedServiceList[0])
              }
              listItemHeight={32}
              options={versionOptions}
            />
          </Form.Item>
        ) : null}
        {selectedServiceList.length > 1 ? (
          <Form.Item label={__("批量任务名称")} required={true}>
            <Input
              value={batchJobName}
              onChange={(e) => {
                changeBatchJobName(e.target.value);
                changeValidateState({
                  ...validateState,
                  BatchJobName: ValidateState.Normal,
                });
              }}
              style={{ width: 490 }}
            />
            {validateState.BatchJobName ? (
              <div style={{ color: "red" }}>{__("此项不允许为空。")}</div>
            ) : null}
          </Form.Item>
        ) : null}
        {operationType !== ServiceMode.Update && sid ? (
          <>
            <Form.Item
              label={<FormLabel text={__("选择的服务")} />}
              required={true}
            >
              <Space>
                <Button
                  icon={
                    <PlusOutlined
                      onPointerEnterCapture={noop}
                      onPointerLeaveCapture={noop}
                    />
                  }
                  onClick={handleChooseClick}
                >
                  {__("选择服务")}
                </Button>
                <Button
                  disabled={
                    !selectedRowKeys.length ||
                    selectedRowKeys.some(
                      (key) =>
                        !selectedServiceList.find(
                          (service) => key === service.aid
                        )?.version
                    )
                  }
                  onClick={() =>
                    handleConfirmService(
                      selectedServiceList.filter((service) =>
                        selectedRowKeys.includes(service.aid)
                      )
                    )
                  }
                >
                  {__("批量确认")}
                </Button>
                <Button
                  disabled={
                    !selectedServiceList.some((service) => !service.select) ||
                    selectedServiceList.some((service) => !service.version) ||
                    (selectedRowKeys.length > 0 &&
                      selectedRowKeys.length !==
                        selectedServiceList.filter((service) => !service.select)
                          .length)
                  }
                  onClick={() => handleConfirmService()}
                >
                  {__("全部确认")}
                </Button>
              </Space>
            </Form.Item>
            <Table
              rowKey="aid"
              dataSource={selectedServiceList}
              columns={columnsForSelectedTable}
              scroll={{ y: `calc(${windowHeight} - 500px)` }}
              rowSelection={{
                ...rowSelectionForSelectTable,
              }}
              pagination={{
                pageSize: 1000,
                hideOnSinglePage: true,
              }}
            />
            {existUnconfirmedService ? (
              <span style={{ color: "red" }}>
                {__("有未确认的服务，请先确认")}
              </span>
            ) : null}
          </>
        ) : null}

        {operationType === ServiceMode.Update ||
        (selectedServiceList.length === 1 && selectedServiceList[0].select) ? (
          <Form.Item label={<FormLabel text={__("使用配置模板")} />}>
            <Switch checked={templateOpen} onChange={handleSwitchChange} />
          </Form.Item>
        ) : null}
      </Form>
      {templateOpen && (
        <ConfigTemplate
          key={selectedServiceList[0]?.version}
          aname={selectedServiceList[0]?.name}
          aversion={selectedServiceList[0]?.version}
          selectedTid={selectedTid}
          changeSelectedTid={changeSelectedTid}
          operationType={operationType}
          customVersion={customVersion}
          changeCustomVersion={changeCustomVersion}
        />
      )}
      {operationType === ServiceMode.Install && (
        <Drawer
          title={__("选择服务")}
          open={open}
          onOk={onOk}
          okButtonProps={{
            disabled: !tempSelectedServiceList.length,
          }}
          onClose={onCancel}
          destroyOnClose
        >
          <ContentLayout moduleName={SERVICE_PREFIX}>
            <Table
              rowKey={"title"}
              dataSource={filteredDataSource}
              columns={columns}
              scroll={{
                y: "calc(100vh - 300px)",
              }}
              rowSelection={{
                ...rowSelection,
              }}
              pagination={{
                showQuickJumper: true,
                showSizeChanger: true,
                showTotal: (total) => __("共${total}条", { total }),
              }}
            />
          </ContentLayout>
        </Drawer>
      )}
    </>
  );
};

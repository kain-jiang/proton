import React, { FC, useEffect, useState } from "react";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import {
  ContentLayout,
  ElementText,
  Toolbar,
} from "../../../common/components";
import { Dot } from "../../../common/components/text/dot";
import {
  ServiceItem,
  IGetServiceTableParams,
  IGetServiceParams,
  ApplicationItem,
} from "../../../../api/service-management/service-deploy/declare";
import {
  serviceApplication,
  serviceJob,
} from "../../../../api/service-management/service-deploy";
import {
  serviceCategoryStatus,
  serviceCategoryStatusItems,
  serviceConfigStatus,
  ServiceConfigStatusEnum,
} from "../type.d";
import { IBaseProps } from "../declare";
import {
  Button,
  Table,
  Space,
  Search,
  Refresh,
  message,
  Modal,
} from "@kweaver-ai/ui";
import type { TableColumnType, TableColumnsType } from "@kweaver-ai/ui";
import { EyeOutlined } from "@kweaver-ai/ui/icons";
import { formatTableResponse } from "../../../common/utils/request";
import { safetyStr } from "../../../common/utils/string";
import { SERVICE_PREFIX } from "../../config";
import __ from "./locale";
import { RevertTable } from "./revert-table";
import { handleError } from "../../utils/handleError";
import elementStyles from "../../../common/components/element-text/style.module.less";
import { assignTo } from "../../../../tools/browser";
import { deployMiniPathname } from "../../../../core/path";
import styles from "./styles.module.less";
import { noop } from "lodash";

interface IProps extends IBaseProps {
  // 修改服务id
  changeServiceId: (id: number) => void;
  // 修改更新服务信息
  changeUpdateServiceRecord: (record: ApplicationItem) => void;
  // 修改主服务id
  changeMainServiceId: (id: number) => void;
  // 修改系统空间id
  changeSid: (sid: number) => void;
}

let modal: any = null;
export const ServiceHome: FC<IProps> = ({
  changeServiceMode,
  changeServiceId,
  changeUpdateServiceRecord,
  changeMainServiceId,
  changeSid,
}) => {
  // 输入框过滤
  const [filter, setFilter] = useState<string>("");
  // 控制回滚table是否开启
  const [revertTableOpen, setRevertTableOpen] = useState<boolean>(false);
  // 选中的服务名称
  const [serviceName, setServiceName] = useState<string>("");
  // 卸载错误中是否显示依赖服务信息
  const [showDependence, setShowDependence] = useState(false);
  // 卸载错误中依赖服务的列表
  const [serviceErrList, setServiceErrList] = useState([]);
  // 系统空间filter筛选项开关
  const [filterOpen, setFilterOpen] = useState(false);
  // 回滚服务的系统空间id
  const [revertServiceSid, setRevertServiceSid] = useState(0);

  // 卸载错误弹窗响应式更新
  useEffect(() => {
    if (modal) {
      modal.update((prevConfig: any) => ({
        ...prevConfig,
        content: getModalContent(showDependence, serviceErrList),
      }));
    }
  }, [showDependence, serviceErrList]);

  const getModalContent = (showDependence: boolean, serviceArr: any[]) => {
    return (
      <div style={{ marginTop: "15px" }}>
        <div>
          {__(
            "此服务与其他服务存在相互依赖关系，为保证卸载顺利请先卸载关联服务。"
          )}
          <EyeOutlined
            className={styles["icon"]}
            onClick={() =>
              setShowDependence((showDependence) => !showDependence)
            }
            onPointerEnterCapture={noop}
            onPointerLeaveCapture={noop}
          />
        </div>
        {showDependence ? (
          <div className={styles["dependence-service-list"]}>
            {serviceArr.map((item: any) => {
              return (
                <div>
                  <span
                    className={styles["service-item"]}
                    title={item?.From?.name}
                  >
                    {item?.From?.name}
                  </span>
                  <span
                    style={{
                      margin: "0 2px",
                      verticalAlign: "top",
                    }}
                  >
                    {"->"}
                  </span>
                  <span
                    className={styles["service-item"]}
                    title={item?.To?.name}
                  >
                    {item?.To?.name}
                  </span>
                </div>
              );
            })}
          </div>
        ) : null}
      </div>
    );
  };

  // 表格的基本设置
  const { state, api, data } = Table.useTable<
    ServiceItem,
    IGetServiceTableParams,
    ServiceItem[]
  >({
    request: (params) => {
      const { _filter, current, pageSize, title } = params;
      const formatedParams: IGetServiceParams = {
        offset: (current - 1) * pageSize,
        limit: pageSize,
        status: getFilterParams(
          (_filter as any)?.status,
          Object.values(serviceCategoryStatus).length
        ),
        sid: [-1, undefined].includes((_filter as any)?.systemName?.[0])
          ? undefined
          : (_filter as any)?.systemName?.[0],
        title,
      };
      return serviceApplication.get(formatedParams);
    },
    rowKey: "id",
    rowSelection: {
      type: "checkbox",
    },
    pagination: {
      showTotal: (total) => __("共${total}条", { total }),
    },
    ...formatTableResponse(),
  });

  const { reload, setParams } = api;
  const { selectedRows } = data;

  const getFilterParams = (
    value: Array<any> | undefined,
    filterOptionCount: number
  ) => {
    if (value?.length) {
      if (value.length === filterOptionCount) {
        return undefined;
      } else {
        return value.reduce((pre, val) => {
          return pre.concat(serviceCategoryStatusItems[val]);
        }, []);
      }
    }
    return undefined;
  };

  // 表格的列配置项
  const columns: TableColumnsType<ServiceItem> = [
    {
      title: __("名称"),
      dataIndex: "title",
      render: (value: string, record: ServiceItem) => (
        <Button type="link" onClick={() => handleDetailClick(record)}>
          {value}
        </Button>
      ),
    },
    {
      title: __("状态"),
      dataIndex: "status",
      filters: Object.values(serviceCategoryStatus),
      render: (value: ServiceConfigStatusEnum) => {
        return (
          <Dot color={serviceConfigStatus[value].color}>
            {serviceConfigStatus[value].text}
          </Dot>
        );
      },
      tooltip: (value: ServiceConfigStatusEnum) => {
        return serviceConfigStatus[value].text;
      },
    },
    {
      title: __("版本"),
      dataIndex: "version",
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
      title: __("组件"),
      render: (_, record: ServiceItem) => (
        <Button type="link" onClick={() => handleDetailClick(record)}>
          {__("查看详情")}
        </Button>
      ),
      tooltip: () => __("查看详情"),
    },
    {
      title: __("备注"),
      dataIndex: "comment",
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
    {
      title: __("操作"),
      width: 220,
      render: (_, record: ServiceItem) => {
        return (
          <Space>
            <Button type="link" onClick={() => handleUpdateClick(record)}>
              {__("更新")}
            </Button>
            <Button type="link" onClick={() => handleRevertClick(record)}>
              {__("回滚")}
            </Button>
            {["DeploymentStudio", "IdentifyAndAuthentication"].includes(
              record.name
            ) ? null : (
              <Button type="link" onClick={() => handleUninstallClick(record)}>
                {__("卸载")}
              </Button>
            )}
          </Space>
        );
      },
      tooltip: () => null,
    },
  ];

  // 处理安装/批量安装
  const handleInstallClick = () => {
    changeServiceMode(ServiceMode.Install);
    changeSid(0);
  };

  // 处理更新
  const handleUpdateClick = (record: ServiceItem) => {
    changeServiceMode(ServiceMode.Update);
    changeUpdateServiceRecord({
      name: record.name,
      title: record.title,
      version: record.version,
      aid: record.aid,
    });
    changeSid(record.sid!);
  };

  // 处理回滚
  const handleRevertClick = (record: ServiceItem) => {
    setRevertTableOpen(true);
    setServiceName(record.name);
    setRevertServiceSid(record.sid!);
  };

  // 卸载服务二次确认弹窗
  const handleUninstallClick = (record: ServiceItem) => {
    Modal.confirm({
      title: __("确认要卸载${title}吗？", { title: record.title }),
      content: (
        <div style={{ marginTop: "15px" }}>
          {__("卸载后此服务将不存在，如需使用要重新安装。")}
        </div>
      ),
      onOk: () => {
        handleUninstallConfirm(record, false);
      },
    });
  };

  // 处理卸载
  const handleUninstallConfirm = async (
    record: ServiceItem,
    force: boolean
  ) => {
    try {
      await serviceJob.uninstallService({
        name: record.name,
        force,
        sid: record.sid,
      });
      // 成功提示
      message.success(
        <ElementText
          text={__("卸载服务任务创建成功，前往-查看")}
          insert={
            <a className={elementStyles["target-herf"]} onClick={clickHerf}>
              {__("【任务监控】")}
            </a>
          }
        />
      );
    } catch (e: any) {
      if (e?.status === 412 && e?.code === 23) {
        const serviceArr = JSON.parse(e?.message).Detail;
        setServiceErrList(serviceArr);
        modal = Modal.confirm({
          title: __("卸载提醒"),
          content: getModalContent(showDependence, serviceArr),
          okText: __("强制卸载"),
          onOk: () => {
            handleUninstallConfirm(record, true);
            setShowDependence(false);
          },
          onCancel: () => {
            setShowDependence(false);
          },
        });
      } else {
        handleError(e);
      }
    }
  };

  const clickHerf = () => {
    assignTo(deployMiniPathname.taskMonitorPathname);
  };

  /**
   * @description 处理查看详情
   * @param {ServiceItem} record 选中行的数据信息
   */
  const handleDetailClick = (record: ServiceItem) => {
    changeServiceMode(ServiceMode.Service);
    changeServiceId(record.id);
    changeMainServiceId(record.id);
  };

  /**
   * @description 更新输入框数据
   * @param e 输入框change事件
   */
  const onFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilter(e.target.value);
    setParams({
      title: e.target.value,
    });
  };

  const header = (
    <Toolbar
      left={
        <Space>
          {/* 安装更新 */}
          <Button
            type="primary"
            disabled={selectedRows.length > 0}
            onClick={() => handleInstallClick()}
          >
            {__("安装更新")}
          </Button>
        </Space>
      }
      right={
        <React.Fragment>
          <Search value={filter} onChange={onFilterChange} debounce />
          <Refresh onClick={() => reload()} />
        </React.Fragment>
      }
      cols={[{ span: 12 }, { span: 12 }]}
      moduleName={SERVICE_PREFIX}
    />
  );

  return (
    <>
      <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
        <Table {...state} columns={columns} />
      </ContentLayout>
      {revertTableOpen ? (
        <RevertTable
          serviceName={serviceName}
          changeRevertTableOpen={setRevertTableOpen}
          sid={revertServiceSid}
        />
      ) : null}
    </>
  );
};

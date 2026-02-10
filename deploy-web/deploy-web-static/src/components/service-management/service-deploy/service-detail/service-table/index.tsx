import React, { FC, useState, useMemo } from "react";
import { Toolbar, ContentLayout } from "../../../../common/components";
import { Table, Search, Button } from "@kweaver-ai/ui";
import { Dot } from "../../../../common/components/text/dot";
import {
  ComponentDefineTypeEnum,
  componentDefineTypeFilter,
  serviceCategoryStatus,
  serviceCategoryStatusItems,
  serviceConfigStatus,
  ServiceConfigStatusEnum,
} from "../../type.d";
import type { ServiceModeType } from "../type";
import { ServiceTableType } from "../../../utils/formatTable";
import { ServiceMode } from "../../../../../core/service-management/service-deploy";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { safetyStr } from "../../../../common/utils/string";
import { SERVICE_PREFIX } from "../../../config";
import { IBaseProps } from "../../declare";
import __ from "../../service-home/locale";

interface IProps extends IBaseProps {
  // 服务类型
  serviceModeType: ServiceModeType;
  // 表格数据
  dataSource: ServiceTableType[];
  // 修改服务id
  changeServiceId: (id: number) => void;
}
export const ServiceTable: FC<IProps> = ({
  serviceModeType,
  dataSource,
  changeServiceId,
  changeServiceMode,
}) => {
  const [filter, setFilter] = useState<string>("");
  const [current, setCurrent] = useState<number>(1);

  // 过滤后的表格数据
  const filteredDataSource = useMemo(() => {
    return dataSource.filter((item) => {
      return item && item.name.includes(filter);
    });
  }, [dataSource, filter]);
  // 表格的列配置项
  const columns: TableColumnsType<ServiceTableType> = [
    {
      title: __("名称"),
      dataIndex: "name",
      render: (value: string, record: ServiceTableType) => (
        <Button type="link" onClick={() => handleDetailClick(record)}>
          {value}
        </Button>
      ),
    },
    ...(serviceModeType === ServiceMode.Service
      ? [
          {
            title: __("类型"),
            dataIndex: "componentDefineType",
            filters: componentDefineTypeFilter,
            onFilter(value: any, record: any) {
              return value === record.componentDefineType;
            },
            defaultFilteredValue: [ComponentDefineTypeEnum.Service],
            render: (value: ComponentDefineTypeEnum) => {
              return safetyStr(
                componentDefineTypeFilter.find((item) => item.value === value)
                  ?.text || value
              );
            },
            tooltip: (value: ComponentDefineTypeEnum) => {
              return safetyStr(
                componentDefineTypeFilter.find((item) => item.value === value)
                  ?.text || value
              );
            },
          },
        ]
      : []),
    {
      title: __("状态"),
      dataIndex: "status",
      filters: Object.values(serviceCategoryStatus),
      onFilter(value: any, record) {
        return serviceCategoryStatusItems[value].includes(record.status);
      },
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
      render: (value: string) => safetyStr(value),
      tooltip: (value: string) => safetyStr(value),
    },
  ];
  /**
   * @description 查看微服务详情
   * @param {ServiceTableType} record 某一行表格数据信息
   */
  const handleDetailClick = (record: ServiceTableType) => {
    changeServiceMode(ServiceMode.MicroService);
    changeServiceId(record.cid);
  };

  const onFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilter(e.target.value);
    setCurrent(1);
  };

  const header =
    serviceModeType === ServiceMode.Service ? (
      <Toolbar
        right={<Search value={filter} onChange={onFilterChange} debounce />}
        rightSize={24}
        moduleName={SERVICE_PREFIX}
      />
    ) : null;
  return (
    <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
      <Table
        className={`${SERVICE_PREFIX}-table-component`}
        rowKey="cid"
        dataSource={filteredDataSource}
        bordered
        scroll={{
          y:
            serviceModeType === ServiceMode.Service
              ? "calc(100vh - 420px)"
              : "calc(100vh - 390px)",
        }}
        columns={columns}
        pagination={{
          current: current,
          showQuickJumper: true,
          showSizeChanger: true,
          showTotal: (total) => __("共${total}条", { total }),
          onChange: (current) => {
            setCurrent(current);
          },
        }}
      />
    </ContentLayout>
  );
};

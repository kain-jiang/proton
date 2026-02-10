import React, { FC, useState } from "react";
import { ServiceMode } from "../../../../core/service-management/service-deploy";
import { ContentLayout, Toolbar } from "../../../common/components";
import { Dot } from "../../../common/components/text/dot";
import {
    serviceCategoryStatus,
    serviceCategoryStatusItems,
    serviceConfigStatus,
    ServiceConfigStatusEnum,
} from "../type.d";
import { IBaseProps } from "../declare";
import { Button, Table, Space, Search, Refresh } from "@kweaver-ai/ui";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { formatTableResponse } from "../../../common/utils/request";
import { safetyStr } from "../../../common/utils/string";
import { SERVICE_PREFIX } from "../../config";
import __ from "./locale";
import {
    SuiteItem,
    IGetSuiteTableParams,
    IGetSuiteParams,
    ApplicationItem,
} from "../../../../api/suite-management/suite-deploy/declare";
import { suiteManifests } from "../../../../api/suite-management/suite-deploy";

interface IProps extends IBaseProps {
    // 修改服务id
    changeServiceId: (id: number) => void;
    // 修改更新服务信息
    changeUpdateServiceRecord: (record: ApplicationItem) => void;
}
export const ServiceHome: FC<IProps> = ({
    changeServiceMode,
    changeServiceId,
    changeUpdateServiceRecord,
}) => {
    // 输入框过滤
    const [filter, setFilter] = useState<string>("");

    // 表格的基本设置
    const { state, api, data } = Table.useTable<
        SuiteItem,
        IGetSuiteTableParams,
        SuiteItem[]
    >({
        request: (params) => {
            const { _filter, current, pageSize, title } = params;
            const formatedParams: IGetSuiteParams = {
                offset: (current - 1) * pageSize,
                limit: pageSize,
                status: getFilterParams(
                    (_filter as any)?.status,
                    Object.values(serviceCategoryStatus).length
                ),
                title,
            };
            return suiteManifests.get(formatedParams);
        },
        rowKey: "jid",
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
    const columns: TableColumnsType<SuiteItem> = [
        {
            title: __("名称"),
            dataIndex: "title",
            render: (value: string, record: SuiteItem) => (
                <Button type="link" onClick={() => handleDetailClick(record)}>
                    {safetyStr(value)}
                </Button>
            ),
            tooltip: (value: string) => safetyStr(value),
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
            dataIndex: "mversion",
        },
        {
            title: __("操作"),
            render: (_, record: SuiteItem) => {
                return (
                    <Space>
                        <Button type="link" disabled>
                            {__("安装")}
                        </Button>
                        <Button
                            type="link"
                            onClick={() => handleUpdateClick(record)}
                        >
                            {__("更新")}
                        </Button>
                    </Space>
                );
            },
            tooltip: () => null,
        },
    ];
    /**
     * @description 处理安装
     */
    const handleInstallClick = () => {
        changeServiceMode(ServiceMode.Install);
    };
    /**
     * @description 处理更新
     * @param {ServiceItem} record 选中行的数据信息
     */
    const handleUpdateClick = (record?: SuiteItem) => {
        changeServiceMode(ServiceMode.Update);
        const updateServiceRecord = record ? record : selectedRows[0];
        changeUpdateServiceRecord({
            mname: updateServiceRecord.jname,
            title: updateServiceRecord.title!,
            mversion: updateServiceRecord.mversion,
        });
    };
    /**
     * @description 处理查看详情
     * @param {ServiceItem} record 选中行的数据信息
     */
    const handleDetailClick = (record: SuiteItem) => {
        changeServiceMode(ServiceMode.Service);
        changeServiceId(record.jid);
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
                    {/* 安装 */}
                    <Button
                        type="primary"
                        disabled={selectedRows.length > 0}
                        onClick={() => handleInstallClick()}
                    >
                        {__("安装")}
                    </Button>
                    {/* 更新 */}
                    <Button
                        type="default"
                        disabled={selectedRows.length !== 1}
                        onClick={() => handleUpdateClick()}
                    >
                        {__("更新")}
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
        <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
            <Table {...state} columns={columns} />
        </ContentLayout>
    );
};

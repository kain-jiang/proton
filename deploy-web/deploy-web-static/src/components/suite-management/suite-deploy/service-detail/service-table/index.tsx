import React, { FC, useState } from "react";
import { ContentLayout } from "../../../../common/components";
import { Table } from "@kweaver-ai/ui";
import { Dot } from "../../../../common/components/text/dot";
import { serviceConfigStatus, ServiceConfigStatusEnum } from "../../type.d";
import type { TableColumnsType } from "@kweaver-ai/ui";
import { safetyStr } from "../../../../common/utils/string";
import { SERVICE_PREFIX } from "../../../config";
import __ from "../../service-home/locale";
import { ServiceSchemaItem } from "../../../../../api/suite-management/suite-deploy/declare";

interface IProps {
    // 表格数据
    dataSource: ServiceSchemaItem[];
}

export const ServiceTable: FC<IProps> = ({ dataSource }) => {
    const [current, setCurrent] = useState<number>(1);

    // 表格的列配置项
    const columns: TableColumnsType<ServiceSchemaItem> = [
        {
            title: __("名称"),
            dataIndex: "title",
            render: (value: string) => safetyStr(value),
        },
        {
            title: __("状态"),
            dataIndex: "status",
            render: (value: ServiceConfigStatusEnum) => {
                return (
                    <Dot
                        color={
                            serviceConfigStatus[ServiceConfigStatusEnum.SUCCESS]
                                .color
                        }
                    >
                        {
                            serviceConfigStatus[ServiceConfigStatusEnum.SUCCESS]
                                .text
                        }
                    </Dot>
                );
            },
            tooltip: (value: ServiceConfigStatusEnum) => {
                return serviceConfigStatus[ServiceConfigStatusEnum.SUCCESS]
                    .text;
            },
        },
        {
            title: __("版本"),
            dataIndex: "version",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value),
        },
    ];

    return (
        <ContentLayout moduleName={SERVICE_PREFIX}>
            <Table
                rowKey="id"
                dataSource={dataSource}
                scroll={{
                    y: "calc(100vh - 390px)",
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

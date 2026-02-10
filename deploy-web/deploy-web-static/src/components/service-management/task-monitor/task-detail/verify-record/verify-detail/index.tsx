import React, { FC } from "react";
import { ContentLayout } from "../../../../../common/components";
import { Drawer, Table, TableColumnsType } from "@kweaver-ai/ui";
import { verify } from "../../../../../../api/service-management/task-monitor";
import { formatTableResponse } from "../../../../../common/utils/request";
import { safetyStr } from "../../../../../common/utils/string";
import {
  DataDetailItem,
  FuncDetailItem,
} from "../../../../../../api/service-management/task-monitor/declare";
import { ActiveEnum } from "../type.d";
import { SERVICE_PREFIX } from "../../../../config";
import __ from "./locale";

interface IProps {
  // 验证记录id
  verifyId: number;
  // 是否展示滑窗
  open: boolean;
  // 验证记录类型
  verifyType: ActiveEnum;
  // 关闭滑窗
  onCancel: () => void;
}
export const VerifyDetail: FC<IProps> = ({
  verifyId,
  open,
  verifyType,
  onCancel,
}) => {
  // 表格的基本设置
  const { state } = Table.useTable<
    FuncDetailItem | DataDetailItem,
    null,
    Array<FuncDetailItem | DataDetailItem>
  >({
    request: (params) => {
      const { current, pageSize } = params;
      const formatedParams = {
        offset: (current - 1) * pageSize,
        limit: pageSize,
      };
      if (verifyType === ActiveEnum.FUNCVERIFY) {
        return verify.getFuncDetail({ ...formatedParams, fid: verifyId });
      } else {
        return verify.getDataDetail({ ...formatedParams, did: verifyId });
      }
    },
    rowKey: "tid",
    pagination: {
      showTotal: (total) => __("共${total}条", { total }),
    },
    ...formatTableResponse(),
  });

  // 将tooltip提示文字按规则换行显示
  const formatTooltip = (str: string): React.JSX.Element[] => {
    const strArr = str.split("\n");
    return strArr.map((item) => {
      return <div>{item}</div>;
    });
  };

  // 表格的列配置项
  const columns: TableColumnsType<FuncDetailItem | DataDetailItem> =
    verifyType === ActiveEnum.FUNCVERIFY
      ? [
          {
            title: __("功能名称"),
            dataIndex: "testFunctionName",
          },
          {
            title: __("详情"),
            dataIndex: "testDescription",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value),
          },
        ]
      : [
          {
            title: __("服务名称"),
            dataIndex: "serviceName",
          },
          {
            title: __("详情"),
            dataIndex: "testResultDetail",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value, formatTooltip),
          },
        ];

  return (
    <Drawer
      title={
        verifyType === ActiveEnum.FUNCVERIFY
          ? __("功能验证详情：ID${verifyId}", { verifyId })
          : __("数据验证详情：ID${verifyId}", { verifyId })
      }
      onClose={onCancel}
      open={open}
      width={900}
      showFooter={false}
    >
      <ContentLayout moduleName={SERVICE_PREFIX}>
        <Table {...state} columns={columns} />
      </ContentLayout>
    </Drawer>
  );
};

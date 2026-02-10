import React, { FC } from "react";
import { Space } from "@kweaver-ai/ui";
import { Dot } from "../../../../common/components/text/dot";
import { serviceConfigStatus } from "../../type.d";
import { safetyTime } from "../../../utils/timer";
import { safetyStr } from "../../../../common/utils/string";
import styles from "./styles.module.less";
import __ from "./locale";
import { SuiteItem } from "../../../../../api/suite-management/suite-deploy/declare";

interface IProps {
    // 服务信息
    serviceDetailInfo: SuiteItem;
}
export const ServiceTitle: FC<IProps> = ({ serviceDetailInfo }) => {
    return (
        <Space direction="vertical">
            <Space size="large">
                <span className={styles["service-name"]}>
                    {serviceDetailInfo.title}
                </span>
                <Dot
                    color={serviceConfigStatus[serviceDetailInfo.status].color}
                >
                    {serviceConfigStatus[serviceDetailInfo.status].text}
                </Dot>
            </Space>
            <Space size="large" className={styles["title-info"]}>
                <span>
                    {__("套件标识：${name}", {
                        name: safetyStr(serviceDetailInfo.jname),
                    })}
                </span>
                <span>
                    {__("版本：${version}", {
                        version: safetyStr(serviceDetailInfo.mversion),
                    })}
                </span>
                <span>
                    {__("更新时间：${updateTime}", {
                        updateTime: safetyTime(serviceDetailInfo.startTime!),
                    })}
                </span>
            </Space>
        </Space>
    );
};

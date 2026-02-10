import React, { FC, useEffect, useState } from "react";
import { Drawer } from "@kweaver-ai/ui";
import { formChangeConfigComfirm, handleError } from "../../utils/handleError";
import styles from "./styles.module.less";
import __ from "./locale";
import { composeJob } from "../../../../api/suite-management/suite-deploy";
import {
    ServiceSchemaItem,
    SuiteItem,
} from "../../../../api/suite-management/suite-deploy/declare";
import { ConfigEditStatusEnum } from "../../suite-config/helper";
import { TaskConfigStatusEnum } from "../type.d";
import { SuiteConfig } from "../../suite-config";
import { SchemaOperateType } from "../../../../core/suite-management/suite-deploy";

interface IProps {
    // 是否展示滑窗
    open: boolean;
    // 任务id
    jid: number;
    // 关闭滑窗
    onCancel: () => void;
}
export const ChangeConfig: FC<IProps> = ({ open, jid, onCancel }) => {
    // 套件配置项（apps）
    const [suiteConfig, setSuiteConfig] = useState<ServiceSchemaItem[]>([]);
    // 原始任务的完整配置
    const [originJobConfig, setOriginJobConfig] = useState<SuiteItem>(
        {} as SuiteItem
    );

    useEffect(() => {
        getFormerJSONSchema(jid);
    }, [jid]);

    /**
     * @description 获取先前配置项
     * @param {number} jid 任务id
     */
    const getFormerJSONSchema = async (jid: number) => {
        try {
            const res = await composeJob.getJobInfo(jid);
            const apps = res?.config?.apps || [];
            setOriginJobConfig(res);
            setSuiteConfig(
                apps.map((app) => {
                    return {
                        ...app,
                        editStatus:
                            app.status === TaskConfigStatusEnum.SUCCESS
                                ? ConfigEditStatusEnum.Disabled
                                : ConfigEditStatusEnum.Init,
                    };
                })
            );
        } catch (error: any) {
            handleError(error);
        }
    };
    /**
     * @description 确定更改配置项
     */
    const onOk = async () => {
        if (
            suiteConfig.some(
                (config) =>
                    config.editStatus === ConfigEditStatusEnum.Unsubmitted
            )
        ) {
            formChangeConfigComfirm();
            return;
        }
        const appsPayload = suiteConfig.map((config) => {
            const newConfig = { ...config };
            delete newConfig.editStatus;
            return newConfig;
        });
        const payload = {
            ...originJobConfig,
            config: { ...originJobConfig.config, apps: appsPayload },
        };
        try {
            await composeJob.patchJob(jid, payload);
            await composeJob.startJob(jid);
            onCancel();
        } catch (error: any) {
            handleError(error);
        }
    };
    return (
        <Drawer
            title={__("更改配置")}
            onOk={onOk}
            onClose={onCancel}
            open={open}
            width={1000}
        >
            <div className={styles["config-drawer-body"]}>
                {suiteConfig.length ? (
                    <SuiteConfig
                        suiteConfig={suiteConfig}
                        isSynchronousUpdate={false}
                        changeSuiteConfig={setSuiteConfig}
                        operateType={SchemaOperateType.ChangeConfig}
                        sid={originJobConfig.sid}
                    />
                ) : null}
            </div>
        </Drawer>
    );
};

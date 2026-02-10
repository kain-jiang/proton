import React, { FC, useEffect, useState } from "react";
import {
    Button,
    Drawer,
    Form,
    Refresh,
    Table,
    TableColumnsType,
} from "@kweaver-ai/ui";
import { handleError } from "../../utils/handleError";
import styles from "./styles.module.less";
import __ from "./locale";
import { SERVICE_PREFIX } from "../../config";
import { TaskConfigStatusEnum, taskConfigStatus } from "../type.d";
import { ContentLayout, Text, Toolbar } from "../../../common/components";
import { safetyRunningTime, safetyTime } from "../../utils/timer";
import { safetyStr } from "../../../common/utils/string";
import {
    ServiceSchemaItem,
    SuiteItem,
} from "../../../../api/suite-management/suite-deploy/declare";
import { composeJob } from "../../../../api/suite-management/suite-deploy";
import { assignTo } from "../../../../tools/browser";
import { deployMiniPathname } from "../../../../core/path";

interface IProps {
    // 滑窗状态
    open: boolean;
    // 任务id
    jid: number;
    // 关闭滑窗
    onCancel: () => void;
}
export const TaskDetail: FC<IProps> = ({ open, jid, onCancel }) => {
    // 任务信息
    const [taskInfo, setTaskInfo] = useState<SuiteItem>({} as SuiteItem);
    // 控制刷新数据
    const [refresh, setRefresh] = useState<boolean>(false);

    useEffect(() => {
        getTaskDetailInfo();
    }, [jid, refresh]);

    /**
     * @description 获取任务信息和验证记录
     */
    const getTaskDetailInfo = async () => {
        try {
            const taskRes = await composeJob.getJobInfo(jid);
            setTaskInfo(taskRes);
        } catch (error: any) {
            handleError(error);
        }
    };

    // 表格的列配置项
    const columns: TableColumnsType<ServiceSchemaItem> = [
        {
            title: __("名称"),
            dataIndex: "title",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value),
        },
        {
            title: __("状态"),
            dataIndex: "status",
            render: (value: TaskConfigStatusEnum) => {
                return (
                    <Text textColor={taskConfigStatus[value].color}>
                        {taskConfigStatus[value].categoryText}
                    </Text>
                );
            },
            tooltip: (value: TaskConfigStatusEnum) =>
                taskConfigStatus[value].categoryText,
        },
        {
            title: __("版本"),
            dataIndex: "version",
            render: (value: string) => safetyStr(value),
            tooltip: (value: string) => safetyStr(value),
        },
        {
            title: __("ID"),
            dataIndex: "id",
        },
        {
            title: __("执行信息"),
            render: () => {
                return (
                    <Button
                        type="link"
                        onClick={() => {
                            assignTo(deployMiniPathname.taskMonitorPathname);
                        }}
                    >
                        {__("查看")}
                    </Button>
                );
            },
            tooltip: () => __("查看"),
        },
    ];

    const header = (
        <Toolbar
            left={<div style={{ fontWeight: 700 }}>{__("基本信息")}</div>}
            right={
                <React.Fragment>
                    <Refresh
                        onClick={() => setRefresh((refresh) => !refresh)}
                    />
                </React.Fragment>
            }
            cols={[{ span: 16 }, { span: 8 }]}
            moduleName={SERVICE_PREFIX}
        />
    );

    const windowHeight = window.innerHeight > 720 ? "100vh" : "720px";

    return (
        <Drawer
            title={taskInfo.title}
            width={1000}
            onClose={onCancel}
            open={open}
            push={false}
            showFooter={false}
            destroyOnClose
        >
            <ContentLayout header={header} moduleName={SERVICE_PREFIX}>
                <Form
                    labelCol={{ span: 3 }}
                    wrapperCol={{ span: 16 }}
                    labelAlign="left"
                >
                    <Form.Item
                        label={__("任务类型")}
                        className={styles["form-item"]}
                    >
                        {taskInfo?.description === "安装"
                            ? __("安装")
                            : taskInfo.description === "更新"
                            ? __("更新")
                            : "---"}
                    </Form.Item>
                    <Form.Item
                        label={__("开始时间")}
                        className={styles["form-item"]}
                    >
                        {safetyTime(taskInfo.startTime!)}
                    </Form.Item>
                    <Form.Item
                        label={__("结束时间")}
                        className={styles["form-item"]}
                    >
                        {safetyTime(taskInfo.endTime!)}
                    </Form.Item>
                    <Form.Item
                        label={__("运行时间")}
                        className={styles["form-item"]}
                    >
                        {safetyRunningTime(
                            taskInfo.endTime!,
                            taskInfo.startTime!
                        )}
                    </Form.Item>
                    <Form.Item
                        label={__("状态")}
                        className={styles["form-item"]}
                    >
                        <Text
                            textColor={taskConfigStatus[taskInfo.status]?.color}
                        >
                            {taskConfigStatus[taskInfo.status]?.categoryText}
                        </Text>
                    </Form.Item>
                    <Form.Item
                        label={__("执行阶段")}
                        className={styles["form-item"]}
                    >
                        {taskConfigStatus[taskInfo.status]?.text}
                    </Form.Item>
                </Form>
                <div className={styles["title"]}>{__("服务信息")}</div>
                <Table
                    dataSource={taskInfo?.config?.apps || []}
                    columns={columns}
                    scroll={{
                        y: `calc(${windowHeight} - 515px)`,
                    }}
                    pagination={{
                        showQuickJumper: true,
                        showSizeChanger: true,
                        showTotal: (total) => __("共${total}条", { total }),
                    }}
                />
            </ContentLayout>
        </Drawer>
    );
};

import React from "react";
import ErrorCode from "../../../tools/request/errorCode/errorCode";
import { Modal } from "@kweaver-ai/ui";
import __ from "../locale";

export const handleError = (error: any) => {
    Modal.error({
        title: __("错误"),
        okText: __("确定"),
        content: (
            <ErrorCode
                cause={error.cause}
                errorCode={error.code || "1"}
                description={error.message}
            />
        ),
    });
};
export const formValidatorComfirm = () => {
    Modal.confirm({
        title: __("配置项验证失败"),
        content: __("存在未提交配置项内容，请确定提交全部服务配置项。"),
        cancelButtonProps: {
            style: {
                display: "none",
            },
        },
    });
};

export const formChangeConfigComfirm = () => {
    Modal.confirm({
        title: __("更改配置项验证失败"),
        content: __("存在未提交配置项内容，请确定提交全部更改配置项。"),
        cancelButtonProps: {
            style: {
                display: "none",
            },
        },
    });
};

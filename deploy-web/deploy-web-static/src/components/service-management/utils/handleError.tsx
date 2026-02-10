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
    content: __(
      "请先正确填写配置项，并点击最下方【提交】按钮进行提交，才能进入下一步操作。"
    ),
    cancelButtonProps: {
      style: {
        display: "none",
      },
    },
  });
};

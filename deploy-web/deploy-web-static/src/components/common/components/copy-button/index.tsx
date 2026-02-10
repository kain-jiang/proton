import React from "react";
import { Button } from "@kweaver-ai/ui";
import type { ButtonProps } from "@kweaver-ai/ui";
import { CopySvg } from "./copy";

interface CopyButtonProps extends Omit<ButtonProps, "icon" | "type"> {
  copyText?: string;
  copyDom?: string;
  onCopySuccess?: () => void;
}

export const copyByDom = (domName: string) => {
  // 执行前先清空
  window?.getSelection()?.removeAllRanges(); //清除页面中已有的selection
  const copyEle = document.querySelector(domName); // 获取要复制的节点
  const range = document.createRange(); // 创造range
  window?.getSelection()?.removeAllRanges(); //清除页面中已有的selection
  if (copyEle) {
    range.selectNode(copyEle); // 选中需要复制的节点
    window?.getSelection()?.addRange(range); // 执行选中元素
    document.execCommand("Copy"); // 执行copy操作
    return true;
  }
  return false;
};
const copyByText = (text: string) => {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.style.position = "absolute";
  textarea.style.left = "-9999px";
  textarea.style.top = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  document.body.removeChild(textarea);
};
export const CopyButton: React.FC<CopyButtonProps> = (props) => {
  const { copyText, copyDom, onCopySuccess, onClick, ...others } = props;
  const onClickHandler = (e: React.MouseEvent<HTMLElement, MouseEvent>) => {
    onClick?.(e);
    if (copyDom) {
      const result = copyByDom(copyDom);
      if (result) {
        onCopySuccess?.();
      }
    } else if (copyText !== undefined) {
      copyByText(copyText);
      onCopySuccess?.();
    }
  };
  return (
    <Button
      icon={<CopySvg />}
      type="text"
      onClick={onClickHandler}
      {...others}
    />
  );
};

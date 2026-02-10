/*
 * @File: Describe the file
 * @Author: Iven.Han
 * @Date: 2023-07-31 18:17:26
 */
import React from "react";
import { Tooltip } from "@kweaver-ai/ui";
import { Help } from "../icons";

interface ITipsProps {
  content: string | React.ReactElement;
  svgIcon?: React.ReactElement;
}

const Tips: React.FC<ITipsProps> = ({ content, svgIcon }) => {
  return (
    <Tooltip title={content} placement="right">
      <span>{svgIcon ? svgIcon : <Help />}</span>
    </Tooltip>
  );
};

export default Tips;

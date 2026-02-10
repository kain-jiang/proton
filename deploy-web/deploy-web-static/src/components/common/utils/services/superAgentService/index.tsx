import React from "react";
import { Space, Tooltip } from "@kweaver-ai/ui";
import { Linux, Windows, Unix } from "../../../components/icons";
import __ from "./locale";

/**
 * 系统类型枚举
 */
export enum SystemTypeEnum {
  /**
   * 全部
   */
  ALL = 0,
  /**
   * LINUX
   */
  LINUX = 1,
  /**
   * WINDOWS
   */
  WINDOWS = 2,
  /**
   * UNIX
   */
  UNIX = 3,
}

interface Node {
  /**
   * 类型
   */
  type: number;
  /**
   * 客户端名称
   */
  hostName: string;
}

/**
 * 主机类型映射
 */
const HostTypeMap = {
  [SystemTypeEnum.LINUX]: {
    text: __("Linux"),
    value: SystemTypeEnum.LINUX,
    icon: "linux-online",
  },
  [SystemTypeEnum.WINDOWS]: {
    text: __("Windows"),
    value: SystemTypeEnum.WINDOWS,
    icon: "windows-online",
  },
  [SystemTypeEnum.UNIX]: {
    text: __("Unix"),
    value: SystemTypeEnum.UNIX,
    icon: "unix-online",
  },
};

/**
 * 图标映射
 */
export const SystemIconMap = (value: SystemTypeEnum, record: Node) => {
  switch (value) {
    case SystemTypeEnum.LINUX:
      return (
        <Space>
          <Tooltip title={HostTypeMap[record.type].text}>
            <Linux />
          </Tooltip>
          <span>{__("Linux")}</span>
        </Space>
      );
    case SystemTypeEnum.WINDOWS:
      return (
        <Space>
          <Tooltip title={HostTypeMap[record.type].text}>
            <Windows />
          </Tooltip>
          <span>{__("Windows")}</span>
        </Space>
      );
    case SystemTypeEnum.UNIX:
      return (
        <Space>
          <Tooltip title={HostTypeMap[record.type].text}>
            <Unix />
          </Tooltip>
          <span>{__("Unix")}</span>
        </Space>
      );
  }
};

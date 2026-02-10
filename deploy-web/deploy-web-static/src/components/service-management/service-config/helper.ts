import __ from "./locale";
import { Color } from "../../suite-management/components/color";

// 服务配置项填写状态
export enum ConfigEditStatusEnum {
  // 初始状态
  Init,
  // 未提交
  Unsubmitted,
  // 已提交
  Submitted,
  // 禁用
  Disabled,
}

export const ConfigEditStatus = {
  [ConfigEditStatusEnum.Unsubmitted]: {
    text: __("未提交"),
    color: Color.TagGrey,
  },
  [ConfigEditStatusEnum.Submitted]: {
    text: __("已提交"),
    color: Color.TagGreen,
  },
};

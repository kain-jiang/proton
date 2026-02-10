import { Color } from "../../../../common/components";
import __ from "../locale";

/**
 * @enum ActiveEnum
 * @description 验证类型枚举
 * @param FUNCVERIFY 0 功能验证
 * @param DATASCHEMAVERIFY 1 数据验证
 */
export enum ActiveEnum {
  FUNCVERIFY,
  DATASCHEMAVERIFY,
}

/**
 * @enum VerifyResultEnum
 * @description 验证结果枚举
 * @param PASS pass 成功
 * @param FAIL fail 失败
 */
export enum VerifyResultEnum {
  PASS = "pass",
  FAIL = "fail",
}

// 验证结果状态
export const verifyResultStatus = {
  [VerifyResultEnum.PASS]: {
    text: __("成功"),
    value: VerifyResultEnum.PASS,
    color: Color.SERVICE_GREEN,
  },
  [VerifyResultEnum.FAIL]: {
    text: __("失败"),
    value: VerifyResultEnum.FAIL,
    color: Color.SERVICE_RED,
  },
};

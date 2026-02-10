/*
 * @File: 数字相关工具
 * @Author: Iven.Han
 * @Date: 2022-03-09 15:26:42
 */
import { CommonEnum } from './common.type'

/**
 * @function
 * @description 数字类型默认处理,无值显示占位符,有值使用回调函数处理
 * @param value 目标值
 * @param callback 有值的回调处理
 */
export const safetyNum = (value: number, callback?: (value: number) => string) => {
    return ![undefined, null, -1].includes(value) ? (callback ? callback(value) : value) : CommonEnum.PLACEHOLDER
}

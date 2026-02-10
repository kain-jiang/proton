/*
 * @File: 字符串相关工具
 * @Author: Iven.Han
 * @Date: 2022-03-09 15:26:42
 */
import { CommonEnum } from './common.type'
import __ from './locale'

/**
 * @function
 * @description 字符串类型默认处理,无值显示占位符,有值且有回调函数，则使用回调函数处理，有值返回值
 * @param value 目标值
 * @param callback 有值的回调处理
 * @example safetyStr('') => '-'
 * @example safetyStr('123') => 123
 * @example safetyStr('123', (value) => value + ':') => 123:
 */
export const safetyStr = (value: string, callback?: (value: string) => any) => {
    return !['', undefined, null].includes(value) ? (callback ? callback(value) : value) : CommonEnum.PLACEHOLDER
}

/**
 * @description: 向字符串左侧填充字符
 * @param {string} text - 待处理字符串
 * @param {number} len - 填充后字符串长度
 * @param {number | string} charStr - 填充字符
 * @return {string}
 */
export const padLeft = (text: string, len: number, charStr: number | string): string => {
    const s = String(text)
    return new Array(len - s.length + 1).join(charStr.toString() || '') + s
}

/**
 * @function
 * @description: 生成uuid标识
 * @return {string}
 */
export const uuid = (): string => {
    return Date.now().toString(16) + Math.random().toString(16).slice(2)
}

export const isEmptyString = (value: any) => {
    return typeof value === 'string' && value.length > 0
}

export const getSelectTitle = (title: string) => {
    return __("请输入${fieldName}", { title })
}

export const addSuffix = (name: string, content: string) => {
    return name + __("（${content}）", { content })
}

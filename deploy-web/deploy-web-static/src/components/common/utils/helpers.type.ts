/*
 * @File: Describe the file
 * @Author: Iven.Han
 * @Date: 2023-07-04 14:49:41
 */
/**
 * @interface
 * @description: 校验是否存在的传参类型定义
 */
export interface IExistedValidatorParams {
    request: any
    existedField?: string
    originName?: string
    message?: string
    params?: any
}

/**
 * @interface
 * @description: 校验传参类型定义
 */
export interface IValidatorParams {
    chinese?: boolean
    uppercase?: boolean
    lowercase?: boolean
    number?: boolean
    chars?: string[]
    anyChars?: boolean //是否匹配任意字符串
    charsRegex?: string //允许的字符正则表达式字符串
    regx?: any
    regex?: RegExp //自定义正则
    excludeChars?: string[]
    message?: string
    minLength?: number
    maxLength?: number
}

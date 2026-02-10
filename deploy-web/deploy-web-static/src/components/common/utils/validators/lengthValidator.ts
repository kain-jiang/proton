/*
 * @File: 字符长度校验工具
 * @Author: TiAmo.han
 * @Date: 2022-07-07 15:06:47
 */

/**
 * @function minValidator
 * @description 校验最小长度
 * @return 用于Form的rule.validator的方法
 */
const minLengthValidator = (min: number) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            // 小于最小字符 有值校验
            if (value !== null && value !== '' && value.length < min) {
                reject()
            } else {
                resolve()
            }
        })
    }
}

/**
 * @function maxValidator
 * @description 校验最大长度
 * @return 用于Form的rule.validator的方法
 */
const maxLengthValidator = (max: number) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            // 大于最大字符 有值校验
            if (value !== null && value !== '' && value.length > max) {
                reject()
            } else {
                resolve()
            }
        })
    }
}

/**
 * @function maxValidator
 * @description 校验长度
 * @return 用于Form的rule.validator的方法
 */
const lengthValidator = (length: number) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            // 大于最大字符 有值校验
            if (value !== null && value !== '' && value.length !== length) {
                reject()
            } else {
                resolve()
            }
        })
    }
}

export { minLengthValidator, maxLengthValidator, lengthValidator }

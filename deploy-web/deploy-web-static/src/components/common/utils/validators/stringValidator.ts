/*
 * @File: 字符串正则校验工具
 * @Author: TiAmo.han
 * @Date: 2022-07-07 14:30:26
 */

import { IExistedValidatorParams, IValidatorParams } from '../helpers.type'
import { minLengthValidator, maxLengthValidator } from './lengthValidator'
import { chartsValidator, excludeCharsValidator } from './chartValidator'
import { chineseReg, uppercaseReg, lowercaseReg } from './common/validatorRegex'

/**
 * @function validator
 * @description 公共校验方法 必填正则
 * @return 用于Form的rule.validator的方法
 */
const requireValidator = () => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            if (['', null, undefined].includes(value) || (Array.isArray(value) && value.length === 0)) {
                reject()
            } else {
                resolve()
            }
        })
    }
}

const requireDataFieldValidator = (fn: (value: any) => any) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            if (!fn(value)) {
                reject()
            } else {
                resolve()
            }
        })
    }
}

/**
 * @function validator
 * @description 公共校验方法 根据传入的正则校验，支持单个正则校验和多个正则满足其一验证，当传入多个正则时，符合其中一个正则，则校验通过
 * @param regx 正则校验 可输入单个或者多个数组形式
 * @return 用于Form的rule.validator的方法
 */
const validator = (regx: RegExp | RegExp[]) => {
    let result = false
    let regxArr: RegExp[] = []
    return (rule: any, value: any) => {
        // 处理正则数组对象
        if (regx instanceof Array) {
            regxArr = regx
        } else {
            regxArr.push(regx)
        }
        // 处理正则校验结果 是否某一项是校验成功 成功返回true
        result = regxArr.some((reg) => {
            const regex = new RegExp(reg)
            return regex.test(value)
        })
        return new Promise<void>((resolve, reject) => {
            // 内容为空不进行校验，由requireValidator处理
            if (value === null || value === undefined || value === '') {
                resolve()
            }
            if (result) {
                resolve()
            } else {
                reject()
            }
        })
    }
}

const negativeValidator = (regx: RegExp) => {
    return (rule: any, value: any) => {
        const regex = new RegExp(regx)
        return new Promise<void>((resolve, reject) => {
            if (regex.test(value)) {
                reject()
            } else {
                resolve()
            }
        })
    }
}

/**
 * @function validatorExisted
 * @description: 公共名称校验是否存在请求param
 * @param {existedField} 接口返回responseData[filed]字段名称
 * @param {originName} 原来的名称
 * @param {request} 请求
 * @return 用于Form的rule.validator的方法
 */
const existedValidator = ({ existedField = 'existed', originName, request, params = {} }: IExistedValidatorParams) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            // 内容为空不进行校验，由requireValidator处理
            if (value === null || value === undefined || value === '') {
                return resolve()
            }
            if (originName !== undefined && value === originName) {
                resolve()
                return
            }
            request(value)
                .then((res: any) => {
                    if (res[existedField]) {
                        reject()
                    } else {
                        resolve()
                    }
                })
                .catch((e: any) => {
                    reject()
                })
        })
    }
}

const localExistedValidator = ({ list }: any) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            list.some((item: any) => item === value) ? reject() : resolve()
        })
    }
}

/**
 * @function stringValidator
 * @description: 字符串校验方法
 * @param {boolean} chinese 是否是中文
 * @param {boolean} uppercase 是否大写
 * @param {boolean} lowercase 是否小写
 * @param {string[]} chars 包含的特殊字符
 * @param {string[]} excludeChars 不包含字符
 * @param {number} minLength 最小长度/这里不涉及信息输出，所以只需要导入数字
 * @param {number} maxLength 最大长度/这里不涉及信息输出，所以只需要导入数字
 * @return 用于Form的rule.validator的方法
 */
const stringValidator = ({
    chinese,
    uppercase,
    lowercase,
    chars,
    regx,
    excludeChars,
    minLength,
    maxLength,
}: IValidatorParams) => {
    if (chinese) {
        return validator(chineseReg)
    }
    if (uppercase) {
        return validator(uppercaseReg)
    }
    if (lowercase) {
        return validator(lowercaseReg)
    }
    if (chars && chars.length > 0) {
        return chartsValidator(chars)
    }
    if (excludeChars && excludeChars.length > 0) {
        return excludeCharsValidator(excludeChars)
    }
    if (minLength) {
        return minLengthValidator(minLength)
    }
    if (maxLength) {
        return maxLengthValidator(maxLength)
    }
    return validator(regx)
}

//正则中需要转义的字符
const needEscape = (char: any) => ['[', ']', '\\', '^', '$', '.', '|', ',', '?', '*', '+', '(', ')'].includes(char)

/**
 * @function stringValidator
 * @description: 自定义字符串校验方法，返回对应正则
 * @param {boolean} chinese 是否是中文
 * @param {boolean} uppercase 是否大写
 * @param {boolean} lowercase 是否小写
 * @param {string[]} chars 包含的特殊字符
 * @param {string[]} excludeChars 不包含字符
 * @param {number} minLength 最小长度/这里不涉及信息输出，所以只需要导入数字
 * @param {number} maxLength 最大长度/这里不涉及信息输出，所以只需要导入数字
 * @return regexp
 * @example
 * 包含中文、数字
 *  customStringRegExp({chinese: true, number: true})
 *
 * 包含中文、数字、大小写、特殊字符-、_，长度为5-14
 *  customStringRegExp({
 *      chinese: true,
 *      number: true,
 *      uppercase: true,
 *      lowercase: true,
 *      chars: ['-', '_'],
 *      minLength: 5,
 *      maxLength: 14
 *  })
 *
 * 长度为5-14任意字符
 *  customStringRegExp({
 *      anyChars: true,
 *      minLength: 5,
 *      maxLength: 14
 *  })
 * => new RegExp('.{5,14}')
 *
 *  A-F的字符串
 *  customStringRegExp({
 *      charsRegex: 'A-F'
 *  })
 * => new RegExp('[A-F]{5,14}')
 */
export const customStringRegExp = ({
    chinese = false,
    uppercase = false,
    lowercase = false,
    chars = [],
    number,
    excludeChars = [],
    minLength = 0,
    maxLength,
    charsRegex,
    anyChars,
}: Omit<IValidatorParams, 'message'>) => {
    let finalRegex
    if (anyChars) {
        finalRegex = '^.'
    } else {
        const chineseReg = chinese ? '\u4e00-\u9fa5' : ''
        const uppercaseReg = uppercase ? 'A-Z' : ''
        const lowercaseReg = lowercase ? 'a-z' : ''
        const numberReg = number ? '0-9' : ''
        const customCharRegex = charsRegex || ''
        const charsReg = chars.length > 0 ? chars.map((char) => (needEscape(char) ? '\\' : '') + char).join('') : ''
        const excludeCharsReg =
            excludeChars.length > 0 ? excludeChars.map((char) => (needEscape(char) ? '\\' : '') + char).join() : ''

        finalRegex = `^[${chineseReg}${uppercaseReg}${lowercaseReg}${numberReg}${charsReg}${customCharRegex}`
        if (excludeCharsReg) {
            finalRegex = finalRegex + `|(?!${excludeChars})`
        }
        finalRegex = finalRegex + ']'
    }

    if (typeof maxLength === 'number') {
        finalRegex = finalRegex + `{${minLength},${maxLength}}`
    } else {
        finalRegex = finalRegex + `{${minLength},}`
    }
    finalRegex = finalRegex + '$'
    return new RegExp(finalRegex, 'g')
}
/**
 * @function stringValidator
 * @description: 字符串校验方法
 * @param {boolean} chinese 是否是中文
 * @param {boolean} uppercase 是否大写
 * @param {boolean} lowercase 是否小写
 * @param {string[]} chars 包含的特殊字符
 * @param {string[]} excludeChars 不包含字符
 * @param {number} minLength 最小长度/这里不涉及信息输出，所以只需要导入数字
 * @param {number} maxLength 最大长度/这里不涉及信息输出，所以只需要导入数字
 * @return 用于Form的rule.validator的方法
 */
export const customStringValidator = (params: Omit<IValidatorParams, 'message'>) => {
    return validator(customStringRegExp(params))
}

export {
    validator,
    existedValidator,
    stringValidator,
    requireValidator,
    localExistedValidator,
    requireDataFieldValidator,
    negativeValidator,
}

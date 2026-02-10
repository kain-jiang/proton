/*
 * @File: 字符校验工具
 * @Author: TiAmo.han
 * @Date: 2022-07-11 18:06:06
 */
/**
 * @function chartValidator
 * @description 包含特殊字符
 * @return 用于Form的rule.validator的方法
 */
const chartsValidator = (charts: any) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            if (value !== null && value !== '' && charts.every((item: any) => value.indexOf(item) < -1)) {
                reject()
            } else {
                resolve()
            }
        })
    }
}
/**
 * @function excludeCharsValidator
 * @description 不包含特殊字符
 * @return 用于Form的rule.validator的方法
 */
const excludeCharsValidator = (charts: any) => {
    return (rule: any, value: any) => {
        return new Promise<void>((resolve, reject) => {
            if (value !== null && value !== '' && charts.every((item: any) => value.indexOf(item) > -1)) {
                reject()
            } else {
                resolve()
            }
        })
    }
}
export { chartsValidator, excludeCharsValidator }

/*
 * @File: 最大最小值校验工具
 * @Author: TiAmo.han
 * @Date: 2022-07-07 15:06:47
 */

/**
 * @function minValueValidator
 * @description 校验最小值
 * @return 用于Form的rule.validator的方法
 */
const minValueValidator = (min: number) => {
    return (rule: any, value: any) => {
        return new Promise((resolve, reject) => {
            // 小于最小值 有值校验
            if (value !== null && value !== '' && Number(value) < min) {
                reject()
            } else {
                resolve(true)
            }
        })
    }
}

/**
 * @function maxValueValidator
 * @description 校验最大值
 * @return 用于Form的rule.validator的方法
 */
const maxValueValidator = (max: number) => {
    return (rule: any, value: any) => {
        return new Promise((resolve, reject) => {
            // 大于最大值 有值校验
            if (value !== null && value !== '' && Number(value) > max) {
                reject()
            } else {
                resolve(true)
            }
        })
    }
}

/**
 * @function uniqueValueValidator
 * @description 校验目标值不与给定数组中的值重复
 * @return 用于Form的rule.validator的方法
 */
const uniqueValueValidator = (values: any[]) => {
    return (rule: any, value: any) => {
        return new Promise((resolve, reject) => {
            // 大于最大值 有值校验
            if (value !== null && value !== '' && values.includes(value)) {
                reject()
            } else {
                resolve(true)
            }
        })
    }
}

export { minValueValidator, maxValueValidator, uniqueValueValidator }

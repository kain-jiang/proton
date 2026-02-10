/*
 * @File: 防抖处理工具
 * @Author: Iven.Han
 * @Date: 2022-06-14 20:35:51
 */

/**
 * @function
 * @description: 防抖函数，`delay`之后执行回调，在`delay`内又触发，则重新计时
 * @param {any} fn 回调函数，业务逻辑
 * @param {number} delay 延迟时间（毫秒）
 * @return {any}
 * @example debounce(check, 500)
 */
const debounce = (fn: (...rest: any[]) => any, delay = 500): ((...rest: any[]) => any) => {
    let ctx: any
    let args: any
    let timer: any = null

    const later = function () {
        fn.apply(ctx, args)
        // 当事件真正执行后，清空定时器
        timer = null
    }

    return (...rest) => {
        args = rest

        if (timer) {
            clearTimeout(timer)
            timer = null
        }

        timer = setTimeout(later, delay)
    }
}

export { debounce }

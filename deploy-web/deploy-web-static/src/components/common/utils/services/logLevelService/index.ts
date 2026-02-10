import { LogLevelEnum } from './declare.d'
import __ from './locale'

/**
 * @interface
 * @description 类型映射值接口
 * @param {string} name 级别名称
 * @param {LogLevelEnum} value 级别value
 * @param {string} styleClass 级别样式类名称
 */
interface IMapValue {
    name: string
    value: LogLevelEnum
    styleClass: string
}
/**
 * @namespace
 * @description 引擎类型映射
 */
const logLevelMap: Record<LogLevelEnum, IMapValue> = {
    [LogLevelEnum.INFORMATION]: {
        name: __('信息'),
        value: LogLevelEnum.INFORMATION,
        styleClass: 'blue-point',
    },
    [LogLevelEnum.WARNING]: {
        name: __('警告'),
        value: LogLevelEnum.WARNING,
        styleClass: 'orange-point',
    },
    [LogLevelEnum.ERROR]: {
        name: __('错误'),
        value: LogLevelEnum.ERROR,
        styleClass: 'red-point',
    },
}

/**
 * @namespace
 * @description 日志级别服务
 */
const logLevelService = {
    isInformation: (level: LogLevelEnum): boolean => {
        return level === LogLevelEnum.INFORMATION
    },
    isWarning: (level: LogLevelEnum): boolean => {
        return level === LogLevelEnum.WARNING
    },
    isError: (level: LogLevelEnum): boolean => {
        return level === LogLevelEnum.ERROR
    },
    getName: (level: LogLevelEnum): string => {
        return logLevelMap[level].name
    },
    getClass: (level: LogLevelEnum): string => {
        return logLevelMap[level].styleClass
    },
    // getOptions: (): IMapValue[] => {
    //     return Object.keys(logLevelMap).map((level) => ({
    //         name: logLevelMap[level].name,
    //         value: logLevelMap[level].value,
    //     }))
    // },
}

export default logLevelService

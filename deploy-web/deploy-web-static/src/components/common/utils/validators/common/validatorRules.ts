/*
 * @File: 校验规则rules
 * @description: 校验表单每次只校验一个规则，但是可以支持包含多种规则,
 *  是否一个校验方法只包含一种规则，校验长度的只包含长度，校验名称的只包含名称的正则（正则里面可包含长度）,校验是否包含特殊字符和不包含特殊字符
 *  这里只包含所有规则的合集，可以在regex里面增加正则在此处引用
 * @Author: TiAmo.han
 * @Date: 2022-07-11 17:13:12
 */
import {
    validator,
    requireValidator,
    existedValidator,
    customStringValidator,
    localExistedValidator,
    requireDataFieldValidator,
    negativeValidator,
} from '../stringValidator'
import { minLengthValidator, maxLengthValidator, lengthValidator } from '../lengthValidator'
import { minValueValidator, maxValueValidator } from '../valueValidator'
import { IValidatorParams } from '../../helpers.type'

import {
    chineseReg,
    uppercaseReg,
    lowercaseReg,
    emailReg,
    nameReg,
    workflowNameReg,
    sanNameReg,
    sharelistNameReg,
    ipv4Reg,
    ipv6Reg,
    superAdminNameReg,
    phoneReg,
    domainReg,
    urlRegExp,
    labelTypeName,
    passwordRegCustom,
    passwordReg1,
    pathReg,
    linuxForcePathReg,
    windowForcePath,
    windowsPathReg
} from './validatorRegex'
import moment from 'moment'
import __ from './locale'

/**
 * @interface
 * @member filedName 字段名称
 * @member message 自定义提示
 */
interface IRequiredRuleParams {
    fieldName?: string
    message?: string
    type?: 'select' | 'input'
}

/**
 * 必填提示：请输入xxx
 * @param param.fieldName 字段名称“请输入xxx中”的xxx
 * @param param.message 自定义提示（一般用不到）
 * @param param.type 默认为'input'，验证组件为下拉框时的提示变为“请选择xxx”，此时更改type为'select'
 */
export const requiredRule = ({ fieldName, message, type = 'input' }: IRequiredRuleParams) => {
    const map = {
        input: __("请输入${fieldName}", {
            fieldName: fieldName,
        }),
        select: __("请选择${fieldName}", {
            fieldName: fieldName,
        }),
    }
    return {
        required: true,
        message: message ? message : map[type],
    }
}

/**
 * @function
 * @description: 判断SSH IP/计算机是否合法
 */
export const sshIpValidator = ({ message }: any) => {
    return [
        {
            validator: validator([ipv4Reg, domainReg, ipv6Reg]),
            message: message,
        },
    ]
}

/**
 * @function
 * @description 名称校验方法
 * @param message 校验错误信息 请输入正确的名称。
 * @param t i18n
 * @return form rules
 */
export const nameValidatorRule = ({
    t,
    min,
    max,
    regx = nameReg(min, max),
    message = __("请输入正确名称。"),
}: any) => {
    return [
        {
            validator: validator(regx),
            message: t(message),
        },
    ]
}

/**
 * @function
 * @description: 判断主机配置的名称是否符合校验
 */
export const validConfigNameValidator = ({ request, originName, t }: any) => {
    return [
        {
            required: true,
            message: __('请输入正确名称。'),
        },
        {
            validator: validator(nameReg(1, 128)),
            message: __('请输入正确的主机配置名称。'),
        },
        {
            validator: existedValidator({ existedField: 'exist', originName, request }),
            message: __('已存在相同主机配置名称。'),
        },
    ]
}

/**
 * @function
 * @description: 判断推送客户端的名称是否符合校验
 */
export const validPushClientNameValidator = ({ request, originName, t }: any) => {
    return [
        {
            required: true,
            message: __('请输入任务名称。'),
        },
        {
            validator: validator(nameReg(1, 256)),
            message: __('请输入正确的任务名称。'),
        },
        {
            validator: existedValidator({ existedField: 'exist', originName, request }),
            message: __('任务名称已存在，请重新输入。'),
        },
    ]
}

/**
 * @function
 * @description 目的端代理管理IP
 * @param message 校验错误信息 请输入目的端代理管理IP。/请输入正确的目的端代理管理IP
 * @return form rules
 */
export const proxyIpValidatorRule = ({
    requiredMessage = __('请输入目的端代理管理IP'),
    ipMessage = __('请输入正确的目的端代理管理IP'),
}) => {
    return [
        {
            validator: requireValidator(),
            message: requiredMessage,
        },
        {
            validator: validator([ipv4Reg, ipv6Reg]),
            message: ipMessage,
        },
    ]
}

/**
 * @function
 * @description 名称校验方法
 * @param message 校验错误信息 请输入正确的名称。
 * @return form rules
 */
export const nameValidatorRuleSingle = ({
    min,
    max,
    regx = nameReg(min, max),
    message = __('请输入正确名称。'),
}: any) => {
    return {
        validator: validator(regx),
        message: message,
    }
}

/**
 * @function
 * @description: 判断推送客户端的linux路径
 */
export const validPushClientLinuxPath = () => {
  return [
        {
            required: true,
            message: __('此项不能为空'),
        },
        {
            validator: validator([pathReg]),
            message: __('安装路径中只允许数字、英文、“-” “_” “/”'),
        },
        {
            validator: negativeValidator(linuxForcePathReg),
            message: __('安装路径中只允许数字、英文、“-” “_” “/”'),
        },
    ]
}

/**
 * @function
 * @description: 判断推送客户端的linux路径-不校验必填
 */
export const validNotRequiredPushClientLinuxPath = () => {
    return [
          {
              validator: validator([pathReg]),
              message: __('安装路径中只允许数字、英文、“-” “_” “/”'),
          },
          {
              validator: negativeValidator(linuxForcePathReg),
              message: __('安装路径中只允许数字、英文、“-” “_” “/”'),
          },
      ]
}

/**
 * @function
 * @description: 判断推送客户端的Windows路径-不校验必填
 */
export const validNotRequiredPushClientWindowsPath = () => {
    return [
        {
            validator: validator([windowsPathReg]),
            message: __('安装路径中只允许数字、英文、“-” “_” “/” “:” “\\”'),
        },
        {
            validator: negativeValidator(windowForcePath),
            message: __('安装路径中只允许数字、英文、“-” “_” “/” “:” “\\”'),
        },
    ]
}

/**
 * @function
 * @description: 判断推送客户端的Windows路径
 */
export const validPushClientWindowsPath = () => {
    return [
        {
            required: true,
            message: __('此项不能为空'),
        },
        {
            validator: validator([windowsPathReg]),
            message: __('安装路径中只允许数字、英文、“-” “_” “/” “:” “\\”'),
        },
        {
            validator: negativeValidator(windowForcePath),
            message: __('安装路径中只允许数字、英文、“-” “_” “/” “:” “\\”'),
        },
    ]
}

/**
 * @function
 * @description: 判断主机是否符合校验
 */
export const validHostValidator = ({ request }: any) => {
    return [
        {
            required: true,
            message: __('请输入SSH IP地址。'),
        },
        {
            validator: validator([ipv4Reg, domainReg, ipv6Reg]),
            message: __('请输入正确的SSH IP地址。'),
        },
        {
            validator: existedValidator({ existedField: 'exist', request }),
            message: __('已存在相同的SSH IP地址。'),
        },
    ]
}


/**
 * @function
 * @description: 判断ip是否符合规范--传入具体的信息即可
 */
export const validIpValidator = ({ message }: { message: string}) => {
    return [
        {   required: true,
            message: __("请输入${fieldName}", {
                fieldName: message,
            }),
        },
        {
            validator: validator([ipv4Reg, ipv6Reg]),
            message: __('请输入正确的${name}', {name: message}),
        },

    ]
}

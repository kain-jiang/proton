/*
 * @File: 校验规则regx
 * @Author: TiAmo.han
 * @Date: 2022-07-11 17:13:12
 */
/**
 * @const chineseReg
 * @description 中文
 */
export const chineseReg = /[\u4e00-\u9fa5]/

/**
 * @const uppercaseReg
 * @description 大写
 */
export const uppercaseReg = /^[A-Z]+$/

/**
 * @const lowercaseReg
 * @description 小写
 */
export const lowercaseReg = /^[a-z]+$/

/**
 * @const emailReg
 * @description 校验邮件格式 第一部分:由字母、数字、下划线、短线“-”、点号“.”组成 “. _ -”不能为最后一个字符;第二部分:为一个域名，域名由字母、数字、短线“-”、域名后缀组成
 */
export const emailReg = /^\w+((\.\w+)|(-\w+)|(_\w+))*@(\w-?)+(\.\w{2,})+$/

/**
 * @const nameReg
 * @description 中文、大小写字母、数字、\“-\”、\“_\”、\“.\”、\“@\”组成，长度为{min}~{max}个字符，全局不可重复。
 */
export const nameReg = (min: number, max: number) => {
    return new RegExp(`^[\u4e00-\u9fa5\\w\\d.\\-_@]{${min},${max}}$`, 'ig')
}

/**
 * @function workflowNameReg
 * @description 中文、大小写字母、数字、“_”组成，长度为{min}~{max}个字符，全局不可重复。
 */
export const workflowNameReg = ({ min, max }: { min: number; max: number }) =>
    new RegExp(`^[\u4e00-\u9fa5\\w\\d_]{${min},${max}}$`, 'ig')

/**
 * @function workflowNameReg
 * @description 大小写字母、数字、“_”组成，长度为{min}~{max}个字符，全局不可重复。
 */
export const workflowExcuParamsNameReg = ({ min, max }: { min: number; max: number }) =>
    new RegExp(`^[\\w\\d_]{${min},${max}}$`, 'ig')

/**
 * @const ipReg
 * @description ip。
 */
export const ipReg = /^((\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.){3}(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])$/

/**
 * @funtion nameReg2
 * @description 中文、大小写字母、数字组成，长度为{min}~{max}（缺省为{3}~{256}）个字符。
 * @returns RegExp
 */
export const nameReg2 = (min = 3, max = 256, Modifiers = 'g') =>
    new RegExp(`^[0-9a-zA-Z\u4E00-\u9FA5]{${min},${max}}$`, Modifiers)

/**
 * @const ipv6Reg
 * @description ipv6Reg
 */
export const ipv6Reg = new RegExp(
    '^\\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0' +
    '-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)){3})|:))' +
    '|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]' +
    '|2[0-4]\\d|1\\d\\d|[1-9]?\\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f' +
    ']{1,4})?:((25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)){3}))|:))|(([0' +
    '-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\\d|1\\d\\d|[1' +
    '-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}' +
    '){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d\\d|' +
    '[1-9]?\\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[' +
    '0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)){3}))|:))|(:(((:[0-9A-Fa-f]{' +
    '1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d' +
    '\\d|[1-9]?\\d)){3}))|:)))(%.+)?\\s*$'
)

/**
 * @const ipv4Reg
 * @description ipv4Reg
 */
export const ipv4Reg = new RegExp(
    '^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])' +
    '\\' +
    '.' +
    '){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$'
)

/**
 * @function
 * @description url正则
 */
export const urlRegExp = (): RegExp => {
    const strRegex =
        '^((https|http)://)?' +
        '(([0-9]{1,3}\\.){3}[0-9]{1,3}' + // IP形式的URL- 3位数字.3位数字.3位数字.3位数字
        '|' + // 允许IP和DOMAIN（域名）
        '(localhost)|' + //匹配localhost
        "([\\w_!~*'()-]+\\.)*" + // 域名- 至少一个[英文或数字_!~*\'()-]加上.
        '\\w+\\.' + // 一级域名 -英文或数字  加上.
        '[a-zA-Z]{1,6})' + // 顶级域名- 1-6位英文
        '(:[0-9]{1,5})?' + // 端口- :80 ,1-5位数字
        '((/?)|' + // url无参数结尾 - 斜杆或这没有
        "(/[\\w_!~*'()\\.;?:@&=+$,%#-]+)+/?)$" //请求参数结尾- 英文或数字和[]内的各种字符

    return new RegExp(strRegex, 'i') //i不区分大小写
}

/**
 * @const wwn
 * @description WWN是以“0x”开头，共18位的16进制字符。
 */
export const wwnReg = /^0x[0-9a-fA-F]{16}$/i

/**
 * @const sanNameReg
 * @description san网关-仅包含中文、字母、数字或 - _ @特殊字符。
 */
export const sanNameReg = (min: number, max: number) => {
    return new RegExp(`^[\u4e00-\u9fa5\\w\\d@\\-_]{${min},${max}}$`, 'ig')
}

/**
 * @const sharelistNameReg
 * @description 共享目录名称仅包含字母、数字或_-特殊字符，长度为3~64个字符。
 */
export const sharelistNameReg = (min: number, max: number) => {
    return new RegExp(`^[\\w\\-]{${min},${max}}$`, 'ig')
}

/**
 * @const phoneReg
 * @description 手机号校验 数字类型
 */
export const phoneReg = /^[0-9]*$/
/**
 * @funtion upLowerNumReg
 * @description 匹配必须同时包含大小写字母和数字，长度为{min}~{max}（缺省为{8}~{16}）个字符。
 * @returns RegExp
 */
export const upLowerNumReg = (min = 8, max = 16, Modifiers = 'g') =>
    new RegExp(`^(?=.*[0-9].*)(?=.*[A-Z].*)(?=.*[a-z].*).{${min},${max}}$`, Modifiers)

/**
 * @funtion upLowerNumCharReg
 * @description 大小写字母、数字或 - _ . @特殊字符，长度为{min}~{max}（缺省为{8}~{16}）个字符。
 * @returns RegExp
 */
export const upLowerNumCharReg = (min = 8, max = 16, Modifiers = 'g') =>
    new RegExp(`^[0-9a-zA-Z.\\-_@]{${min},${max}}$`, Modifiers)

/**
 * @funtion unUpLowerNumReg
 * @description 匹配非大小写字母和数字。
 * @returns RegExp
 */
export const unUpLowerNumReg = (Modifiers = 'g') => new RegExp(`[^0-9a-zA-Z]`, Modifiers)

/**
 * @const superAdminNameReg
 * @description min-max个字符，可包含英文、数字、“-”，“_”，“·”，“@”，不可重名。
 */
export const superAdminNameReg = (min: number, max: number) => {
    return new RegExp(`^[\\w\\d@\\.\\-_]{${min},${max}}$`, 'ig')
}

/**
 * @const domainReg
 * @description 域名校验。由字母、数字、连字符（-）组成，连字符（-）不得出现在字符串的头部或者尾部。且长度不超过63个字符。
 * 字符串间以点分割，且总长度（包括末尾的点）不超过254个字符。
 */
export const domainReg = /^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/

/**
 * @const passwordReg1
 * @description 大小写字母、数字、特殊字符（“-” “_” “.” “@”）
 */
export const passwordReg1 = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[.\-_@]).{4,}$/g
/**
 * @const passwordReg2
 * @description 大小写字母、数字
 */
export const passwordReg2 = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).{3,}$/g
/**
 * @const passwordReg3
 * @description 字母(不区分大小写)、数字
 */
export const passwordReg3 = /^(?=.*[a-zA-Z])(?=.*[0-9]).{2,}$/g
/**
 * @const passwordReg4
 * @description 大小写字母、特殊字符（“-” “_” “.” “@”）
 */
export const passwordReg4 = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[.\-_@]).{3,}$/g
/**
 * @const passwordReg5
 * @description 字母(不区分大小写)、特殊字符（“-” “_” “.” “@”）
 */
export const passwordReg5 = /^(?=.*[a-zA-Z])(?=.*[.\-_@]).{2,}$/g
/**
 * @const passwordReg6
 * @description 数字、特殊字符（“-” “_” “.” “@”）
 */
export const passwordReg6 = /^(?=.*[0-9])(?=.*[.\-_@]).{2,}$/g
/**
 * @const passwordReg7
 * @description 字母(不区分大小写)、数字、特殊字符（“-” “_” “.” “@”）
 */
export const passwordReg7 = /^(?=.*[a-zA-Z])(?=.*[0-9])(?=.*[.\-_@]).{2,}$/g
/**
 * @const continuousCharacterReg
 * @description 连续字符正则
 */
export const continuousCharacterReg = /(.)\1{1,}/g

/**
 * @const endWithDatReg
 * @description 文件以.dat结尾
 */
export const endWithDatReg = /^.+\.dat$/i

/**
 * @const labelTypeName
 * @description 长度为{min}~{max}个字符，可包含大小写字母、空格、“_”
 */
export const labelTypeName = (min: number, max: number) => {
    return new RegExp(`^[A-Za-z\\s_]{${min},${max}}$`, 'ig')
}

/**
 * @const passwordRegCustom
 * @description 大小写字母、特殊字符（“-” “_” “.” “@”）
 */
export const passwordRegCustom = (min: number, max: number) => {
    return new RegExp(`^[\\w\\d@\\.\\-_]{${min},${max}}$`, 'ig')
}
/**
 * @const notIncludeCodeReg
 * @description 不能包含“-”“_”“.”“@”外的特殊字符
 */
export const notIncludeCodeReg = '[`~!#$^&*()=|{}\':;,\\[\\]<>《》/?~！%#￥……&*（）——|{}【】‘；：\\\\ ”“。，、？]'

/**
* @const userNameReg
* @description 英文、数字、\“-\”、\“_\”、\“.\”、\“@\”组成，长度为{min}~{max}个字符。
*/
export const userNameReg = (min: number, max: number) => {
    return new RegExp(`^[\\w\\d.\\-_@]{${min},${max}}$`, 'ig')
}

/**
* @const letterReg
* @description 大小写字母
*/
export const letterReg = /[^A-z]/gi

/**
* @const numWithPointReg
* @description 数字包含小数点
*/
export const numWithPointReg = /[^0-9.]/gi
/**
* @const mustContainCharAndDigit
* @description 必须有字母(不区分大小写)、数字，可以有其他字符
*/
export const mustContainCharAndDigit = /^(?=.*[a-zA-Z])(?=.*[0-9]).{2,}$/

/**
 * @const passwordReg8
 * @description 可输入：字母(不区分大小写)、数字、特殊字符（“-” “_” “.” “@”）
 */
export const passwordReg8 = /^[a-zA-Z0-9.\-_@]+$/


/**
 * @function workflowNameReg
 * @description 中文、大小写字母、数字、“-”、“.”组成，长度为{min}~{max}个字符
 */
export const nameReg3 = ({ min, max }: { min: number; max: number }) =>
    new RegExp(`^[\u4e00-\u9fa5a-zA-Z\\d-.]{${min},${max}}$`, 'ig')

/**
 * @const pathReg
 * @description 可输入：字母(不区分大小写)、数字、特殊字符（“-” “_” “/”）
 */
export const pathReg = /^[a-zA-Z0-9\-_/]+$/

/**
 * @const pathReg
 * @description Linux系统限制路径输入内容 /、/tmp/*、/root/*、/home/*、/var/*、/etc/*、/proc/*、/boot/*、/bin/*、/sbin/*、/dev/*、/usr/*
 */
export const linuxForcePathReg = /^(\/|\/tmp|\/root|\/home|\/var|\/etc|\/proc|\/boot|\/bin|\/sbin|\/dev|\/usr)($|\/.*?)$/g

/**
 * @const pathReg
 * @description Window系统限制路径输入内容 $TEMP、$TMP
 */
export const windowForcePath = /^(\/\$TEMP|\/\$TMP)$/g

/**
 * @const windowsPathReg
 * @description 可输入：字母(不区分大小写)、数字、特殊字符（“-” “_” “/”“:”“\”）
 */
export const windowsPathReg = /^[a-zA-Z0-9\-_/:\\]+$/
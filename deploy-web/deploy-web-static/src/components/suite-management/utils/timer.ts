import { CommonEnum } from "./common.type";
import __ from "../locale";
/**
 * @function
 * @description 有效时间值安全处理，如时间戳、持续时长
 * @param value 目标值(秒级)
 * @param callback 有值的回调处理，默认处理是时间戳转换为YYYY-MM-DD HH:mm:ss格式
 * @example safetyTime(undefinded/null) => "---"
 * @example safetyTime(-1) => "---"
 * @example safetyTime(1671621478) => "2022-12-21 19:17:58"
 * @example safetyTime(1671621478, (value) => value + "000") => 1671621478000
 * @example safetyTime(1) => "1970-01-01 08:00:01"
 */
export const safetyTime = (
  value: number,
  callback = (value: number) => timer(value, CommonEnum.DATE_TIME) as any
) => {
  return value !== null && value > 0
    ? callback(value * 1000)
    : CommonEnum.PLACEHOLDER;
};

/**
 * @function
 * @description: 给时间时分秒加上前导0
 * @return {string}
 * @example addZeroPrefix(2) => "02"
 */
export const addZeroPrefix = (timeParam: number): string => {
  return String(timeParam).padStart(2, "0");
};

/**
 * @function
 * @description: 格式化数据为日期格式
 * @return {string}
 * @example timer(1689583526006) => "2023-7-17 16:45:26"
 */
export const timer = (timestamp: number, format = "YYYY-MM-DD hh:mm:ss") => {
  const {
    yearMonthSeparator,
    monthDaySeparator,
    hourMinuteSeparator,
    minuteSecondSeparator,
  } = getSeparators(format);
  const value = new Date(timestamp);

  const date =
    value.getFullYear() +
    yearMonthSeparator +
    (value.getMonth() + 1) +
    monthDaySeparator +
    value.getDate();
  const time =
    addZeroPrefix(value.getHours()) +
    hourMinuteSeparator +
    addZeroPrefix(value.getMinutes()) +
    minuteSecondSeparator +
    addZeroPrefix(value.getSeconds());
  return `${date} ${time}`;
};

/**
 * @function
 * @description: 获取年月日 时分秒分隔符
 * @return {object}
 * @example getSeparators("YYYY-MM-DD hh:mm:ss") => {
 *      yearMonthSeparator: "-",
 *      monthDaySeparator: "-",
 *      hourMinuteSeparator: "-",
 *      minuteSecondSeparator: "-"
 * }
 */
const getSeparators = (format: string) => {
  // 年月分隔符正则
  const yearMonthSeparatorReg = /Y([^YM]*)M/g;
  // 月日分隔符正则
  const monthDaySeparatorReg = /M([^MD]*)D/g;
  // 时分分隔符正则
  const hourMinuteSeparatorReg = /h([^hm]*)m/g;
  // 分秒分隔符正则
  const minuteSecondSeparatorReg = /m([^ms]*)s/g;

  const yearMonthSeparator = getSeparator(yearMonthSeparatorReg, format);
  const monthDaySeparator = getSeparator(monthDaySeparatorReg, format);
  const hourMinuteSeparator = getSeparator(hourMinuteSeparatorReg, format);
  const minuteSecondSeparator = getSeparator(minuteSecondSeparatorReg, format);

  return {
    yearMonthSeparator,
    monthDaySeparator,
    hourMinuteSeparator,
    minuteSecondSeparator,
  };
};

/**
 * @function
 * @description: 获取分隔符
 * @return {object}
 * @example getSeparator(/Y([^YM]*)M/g, "YYYY-MM-DD hh:mm:ss") => "-"
 * @param reg 分隔符匹配正则
 * @param format 格式字符串
 */
const getSeparator = (reg: RegExp, format: string): string => {
  const result = reg.exec(format);
  return result ? result[1] : "-";
};

/**
 * @function
 * @description: 计算时钟
 * @return {object}
 * @example clockDate(123123) => "34小时12分钟3秒"
 * @param totalSeconds 总秒数
 */
export function clockDate(totalSeconds: number) {
  totalSeconds = Math.floor(totalSeconds);
  const days = Math.floor(totalSeconds / 3600 / 24);
  const hours = Math.floor((totalSeconds / 3600) % 24);
  const minutes = Math.floor((totalSeconds / 60) % 60);
  const seconds = totalSeconds % 60;

  let content = `${seconds}${__("秒")}`;

  if (minutes) {
    content = `${minutes}${__("分钟")}` + content;
  }

  if (hours) {
    content = `${hours}${__("小时")}` + content;
  }

  if (days) {
    content = `${days}${__("天")}` + content;
  }

  return content;
}

/**
 * @function
 * @description 有效运行时间值安全处理
 * @param endTime 结束时间
 * @param startTime 开始时间
 * @example safetyRunningTime(undefinded/null, undefinded/null) => "---"
 * @example safetyRunningTime(-1, -1) => "---"
 * @example safetyRunningTime(1689583526, 1671621475) => "207天21小时27分钟31秒"
 */
export const safetyRunningTime = (endTime: number, startTime: number) => {
  if (endTime !== null && endTime > 0 && startTime !== null && startTime > 0) {
    return clockDate(endTime - startTime);
  } else if (startTime !== null && startTime > 0) {
    return clockDate(Date.now() / 1000 - startTime);
  } else {
    return CommonEnum.PLACEHOLDER;
  }
};

/**
 * @function
 * @description 日期格式处理，统一以/连接
 * @param value 目标值
 * @example formatDateTime("2023-12-01 15:03:29") => "2023/12/01 15:03:29"
 */
export const formatDateTime = (value: string) => {
  return value.replace(/-/g, "/");
};

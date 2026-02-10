/**
 * 从日期字符串转换毫秒数
 * @params dateString 日期字符串
 */
export function getUTCTime(dateString: any) {
    let y, M, d, h, m, s;

    // (ISO 8601标准) 例：2017-12-14 , 2017-12-11T14:50:55+08:00
    if (
        dateString.match(
            /^\d{4}(-?\d{2}){2}([\sT]\d{2}:\d{2}:\d{2}([+\-\s]\d{2}:\d{2})?)?/
        )
    ) {
        const {
            date = "1970-01-01",
            time = "00:00:00",
            zone = "+00:00",
        } = dateString
            .match(
                /^(\d{4}(-?\d{2}){2})|([\sT]\d{2}:\d{2}:\d{2})|([+\-\s]\d{2}:\d{2})/g
            )
            .reduce((prev: any, currentValue: any, index: number) => {
                return {
                    ...prev,
                    date: currentValue.match(/\d{4}(-?\d{2}){2}/)
                        ? currentValue
                        : prev["date"],
                    time: currentValue.match(/[\sT]\d{2}:\d{2}:\d{2}/)
                        ? currentValue
                        : prev["time"],
                    zone: currentValue.match(/[+\-\s]\d{2}:\d{2}/)
                        ? currentValue
                        : prev["zone"],
                };
            }, {});
        // zone指定时区，可以是：Z (UTC)、+hh:mm、-hh:mm
        const [hh, mm] = zone.split(":");
        const [, t = "00:00:00"] = time.split(/[\sT]/);

        [y = 0, M = 0, d = 0] = date.split("-").map(Number);
        [h = 0, m = 0, s = 0] = t.split(":").map(Number);
        h = h - Number(hh);
        m = m - Number(mm);
    } else {
        let [fullDate, time] = dateString.split(/\s+/);

        [h = 0, m = 0, s = 0] = time ? time.split(":").map(Number) : [];
        if (fullDate.match(/^\d{1,2}\/\d{1,2}\/\d{4}$/)) {
            [M = 0, d = 0, y = 0] = fullDate.split("/").map(Number);
        } else if (fullDate.match(/\d{4}(-\d{1,2}){2}/)) {
            [y = 0, M = 0, d = 0] = fullDate.split("-").map(Number);
        } else if (fullDate.match(/\d{4}(\.\d{1,2}){2}/)) {
            [y = 0, M = 0, d = 0] = fullDate.split(".").map(Number);
        }
    }

    return Date.UTC(y, M - 1, d, h, m, s);
}

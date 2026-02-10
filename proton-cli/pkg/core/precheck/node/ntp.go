package node

import (
	"fmt"
	"time"

	"github.com/beevik/ntp"
)

func NTPCheck(ntpServer string) [][]string {
	// 获取远程NTP服务器的时间
	ntpInfo := [][]string{}

	// 获取本地系统当前时间
	localTime := time.Now()

	if ntpServer != "" {
		remoteTime, err := ntp.Time(ntpServer)
		if err != nil {
			ntpInfo = append(ntpInfo, []string{"Time", err.Error(), "\033[31mNO PASS\033[0m", "check local time"})
			return ntpInfo
		}

		// 计算两者之间的差异（单位为纳秒）
		timeDiff := remoteTime.Sub(localTime).Seconds()
		if int(timeDiff) > 10 {
			ntpInfo = append(ntpInfo, []string{"Time", fmt.Sprintf("%.2fs", timeDiff), "\033[31mNO PASS\033[0m", "Local time different with NTP server time > 10s, must update local time"})
		} else if int(timeDiff) <= 10 && int(timeDiff) > 3 {
			ntpInfo = append(ntpInfo, []string{"Time", fmt.Sprintf("%.2fs", timeDiff), "\033[33mWARN\033[0m", "Local time different with NTP server time > 3s, should update local time"})
		} else {
			ntpInfo = append(ntpInfo, []string{"Time", fmt.Sprintf("different with NTP Server %.2fs", timeDiff), "\033[32mPASS\033[0m", ""})
		}
	} else {
		formattedTime := localTime.Format(time.RFC3339)
		ntpInfo = append(ntpInfo, []string{"Time", formattedTime, "\033[33mWARN\033[0m", "cannot compare with NTP server time, manually confirm the gap between the local and the Internet time"})
	}

	return ntpInfo
}

package cpu

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

func cpuIntensiveWorkload() {
	var result int64
	for i := 0; i < 10000000; i++ {
		result += rand.Int63()
	}
	_ = result
}

func CPUPerf() [][]string {
	cpuPerf := [][]string{}
	runtime.GOMAXPROCS(runtime.NumCPU())

	cpuIntensiveWorkload()

	start := time.Now()

	for i := 0; i < 10; i++ {
		cpuIntensiveWorkload()
	}

	elapsed := time.Since(start)

	if elapsed.Seconds() > 3.00 {
		cpuPerf = append(cpuPerf, []string{"Perf CPU", fmt.Sprintf("%.2fs", elapsed.Seconds()), "\033[33mWARN\033[0m", "single cpu perf time should < 3s, suggest a better cpu"})
	} else {
		cpuPerf = append(cpuPerf, []string{"Perf CPU", fmt.Sprintf("%.2fs", elapsed.Seconds()), "\033[32mPASS\033[0m", ""})
	}
	return cpuPerf
}

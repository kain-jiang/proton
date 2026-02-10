package disk

import (
	"crypto/rand"
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func DirectDD() [][]string {
	ddInfo := [][]string{}
	filename := "testfile"
	fileSize := 1024 * 1024 * 128 // 128M
	blockSize := 4096
	concurrentNum := 10

	data := make([]byte, blockSize)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}

	flags := os.O_RDWR | os.O_CREATE | os.O_TRUNC | unix.O_DIRECT
	fd, err := unix.Open(filename, flags, 0644)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)
	defer os.Remove(filename)

	done := make(chan bool)
	startTime := time.Now()

	for i := 0; i < concurrentNum; i++ {
		go func() {
			buffer := make([]byte, blockSize)
			for j := 0; j < fileSize/blockSize/concurrentNum; j++ {
				// 进行写入操作
				_, err := syscall.Write(fd, buffer)
				if err != nil {
					fmt.Println("Failed to write:", err)
					done <- false
					return
				}
			}
			done <- true
		}()
	}

	// 等待所有任务完成
	for i := 0; i < concurrentNum; i++ {
		if !<-done {
			ddInfo = append(ddInfo, []string{"Perf IO", "Write failed", "\033[31mNO PASS\033[0m", "try again"})
			return ddInfo
		}
	}

	err = syscall.Fsync(fd)
	if err != nil {
		ddInfo = append(ddInfo, []string{"Perf IO", err.Error(), "\033[31mNO PASS\033[0m", "try again"})
		return ddInfo
	}

	duration := time.Since(startTime)
	bytesPerSec := float64(fileSize) / duration.Seconds()
	result := fmt.Sprintf("IO direct write %.2f MB with %d Block in %v (%.2f MB/s)\n", float64(fileSize)/(1<<20), blockSize, duration, bytesPerSec/(1<<20))
	if int(bytesPerSec/(1<<20)) <= 20 {
		ddInfo = append(ddInfo, []string{"Perf IO", result, "\033[31mNO PASS\033[0m", "Testing env: 20MB/s ~ 100MB/s\nProduction env: >= 100MB/s"})
	} else if int(bytesPerSec/(1<<20)) > 20 && int(bytesPerSec/(1<<20)) < 100 {
		ddInfo = append(ddInfo, []string{"Perf IO", result, "\033[33mWARN\033[0m", "Testing env: 20MB/s ~ 100MB/s\nProduction env: >= 100MB/s"})
	} else {
		ddInfo = append(ddInfo, []string{"Perf IO", result, "\033[32mPASS\033[0m", ""})
	}
	return ddInfo
}

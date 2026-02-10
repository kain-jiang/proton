package test

import (
	"fmt"
	"os"
	"runtime/pprof"
)

// StartCPUPProf start cpu pprof, write into file
func StartCPUPProf(fpath string) (file *os.File, err error) {
	if fpath == "" {
		fpath = "cpu_flamegraph.pprof"
	}
	file, err = os.Create(fpath)
	if err != nil {
		fmt.Println("无法创建火焰图文件:", err)
		return
	}
	if err = pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return
	}
	return
}

// StartMemoryPProf start memory pprof, write into file
func StartMemoryPProf(fpath string) (file *os.File, err error) {
	if fpath == "" {
		fpath = "memory_flamegraph.pprof"
	}
	file, err = os.Create(fpath)
	if err != nil {
		fmt.Println("无法创建火焰图文件:", err)
		return
	}
	if err = pprof.WriteHeapProfile(file); err != nil {
		file.Close()
		return
	}
	return
}

package main

import "component-manage/internal/cmd"

//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go env -w GO111MODULE=on

func main() {
	cmd.SetVersion(versionMap())
	cmd.Execute()
}

var (
	commitId = "unknown"
	version  = "unknown"
	date     = "unknown"
)

func versionMap() map[string]any {
	return map[string]any{
		"version":  version,
		"commitId": commitId,
		"data":     date,
	}
}

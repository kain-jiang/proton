#!/usr/bin/env bash

# shellcheck disable=SC2046
cd $(dirname $(dirname $(dirname "$0"))) || exit

go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
go env -w GO111MODULE=on
go mod tidy
go tool air
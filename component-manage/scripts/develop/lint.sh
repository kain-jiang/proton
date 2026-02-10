#!/usr/bin/env bash

set -e
set -x

# shellcheck disable=SC2046
cd $(dirname $(dirname $(dirname "$0"))) || exit

go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
go env -w GO111MODULE=on
go mod tidy

go tool swag init -g ./internal/server/server.go
go tool swag fmt
go tool gofumpt -l -w .
go mod tidy
go tool golangci-lint run --out-format="colored-line-number:stdout"
go tool gotestsum --packages="./..." -- -gcflags="all=-N -l" -count=1



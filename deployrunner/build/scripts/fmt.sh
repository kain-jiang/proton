#!/bin/bash

set -e

cd $(dirname $(readlink -f $0))/../..

DOCKER_BUILDKIT=1 docker build \
    --target golangci-lint \
    --output type=local,dest=.build/tools \
    -f build/puredocker/lint.dockerfile \
    build/puredocker

.build/tools/golangci-lint fmt

#!/usr/bin/env bash

set -e
source "$(dirname $(realpath "${BASH_SOURCE[0]}"))/_meta.sh"

export DOCKER_BUILDKIT=1
docker build --target test-result --build-arg DEVOPS_PAT=$meta_devopsPat --output type=local,dest=./ut-result .


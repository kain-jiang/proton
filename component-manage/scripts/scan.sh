#!/usr/bin/env bash

set -e
source "$(dirname $(realpath "${BASH_SOURCE[0]}"))/_meta.sh"

export DOCKER_BUILDKIT=1
docker build --target trivy --output type=local,dest=./trivy .
./trivy/trivy image "$meta_remoteImageName:$meta_imageTag.amd64"


#!/usr/bin/env bash

set -e

# shellcheck disable=SC2046
cd $(dirname $(dirname $(realpath "$0"))) || exit 1

registry="acr.aishu.cn"
repository="ict/deploy-service"
branch=${BRANCH_NAME="local"}
buildNumber=${BUILD_NUMBER="0"}
sourceVersion=${BUILD_SOURCEVERSION="unknown"}
gitTag="git.${sourceVersion:0:8}"

# 小写分支名
tags=("${branch,,}.latest" "${branch,,}.${buildNumber}" "${gitTag}")

# shellcheck disable=SC2068
for tag in ${tags[@]} ; do
  docker manifest create --amend "$registry/$repository:$tag" "$registry/$repository:$tag.x86" "$registry/$repository:$tag.arm"
  docker manifest push --purge "$registry/$repository:$tag"
  docker rmi -f "$registry/$repository:$tag"
done
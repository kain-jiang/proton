#!/usr/bin/env bash

set -e

# build/DeployService/script/$0

# shellcheck disable=SC2046
buildDir=$(dirname $(dirname $(dirname $(dirname $(dirname $(realpath "$0"))))))

buildTempDir=$(mktemp -d)
cleanup() {
  local exit_code=$?
  if [[ -d "$buildTempDir" ]]; then
    rm -rf "$buildTempDir"
    echo "删除临时目录：$buildTempDir"
  fi
  exit $exit_code
}
trap cleanup EXIT INT TERM


declare -A buildDirMap=(
  ["DeployService/src/deploy-service"]="$buildTempDir/DeployService"
  ["API"]="$buildTempDir/API"
  ["proton-rds-sdk-py"]="$buildTempDir/proton-rds-sdk-py"
)

for key in "${!buildDirMap[@]}"; do
  value="${buildDirMap[$key]}"
  if [[ ! -d "$buildDir/$key" ]]; then
    printf "请检查项目目录：%s\n" "$buildDir/$key"
    exit 1
  else
    cp -r "$buildDir/$key" "$value"
  fi
done

cd $buildTempDir || exit 1


##### 原有构建流程

#projectPath=$(pwd)
projectName="deploy-service"

registry="acr.aishu.cn"
repository="ict/deploy-service"
branch=${BRANCH_NAME="local"}
buildNumber=${BUILD_NUMBER="0"}
arch=${BUILD_ARCH="x86"}
sourceVersion=${BUILD_SOURCEVERSION="unknown"}
gitTag="git.${sourceVersion:0:8}"

# 小写分支
tags=("${branch,,}.latest" "${branch,,}.${buildNumber}" "${gitTag}")


########################################
## Docker Build                       ##
########################################
set -x

# shellcheck disable=SC2068
for tag in ${tags[@]} ; do
  docker build --pull -t "$registry/$repository:$tag.$arch" -f DeployService/Dockerfile .
  docker push "$registry/$repository:$tag.$arch"
  docker rmi -f "$registry/$repository:$tag.$arch"
done
docker image prune -f --filter label=stage=builder || true

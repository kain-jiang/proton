#!/usr/bin/env bash

set -e

# shellcheck disable=SC2046
cd $(dirname $(dirname $(realpath "$0"))) || exit 1
projectPath=$(pwd)

git_rev=$(git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
if [[ -n $(git status --porcelain) ]]; then
    git_dirty="dirty"
else
    git_dirty="clean"
fi
local_tag="git.${git_rev}-${git_dirty}"

temp_dir=$(mktemp -d)
cleanup() {
    echo "正在清理临时目录..."
    rm -rf "$temp_dir"
    echo "已删除临时目录: $temp_dir"
}
trap cleanup EXIT
cd $temp_dir

echo "正在下载源码...API"
git clone ssh://devops.aishu.cn:22/AISHUDevOps/AnyShareFamily/_git/API -b MISSION --depth=1
echo "正在下载源码...proton-rds-sdk-py"
git clone ssh://devops.aishu.cn:22/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-py -b 1.4.2 --depth=1
echo "正在拷贝源码...${projectPath}"
cp -rf ${projectPath} DeployService

echo "源码目录结构:"
ls -al $temp_dir

local_img="localhost/ict/deploy-service:${local_tag}"

echo "开始构建镜像...${local_img}"
docker build --pull -t ${local_img} -f DeployService/Dockerfile .
echo "镜像构建完成...${local_img}"
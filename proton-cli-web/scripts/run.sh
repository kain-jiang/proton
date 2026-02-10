# 脚本只要发生错误，就终止执行
set -e

# 定义路径
abspath=$(dirname $(dirname `readlink -f $0`))
tag=""

BRANCH_NAME=$(echo $BRANCH_NAME | sed 's/refs\/[a-z]*\///')
BRANCH_NAME=${BRANCH_NAME//\//-}
if [[ $BRANCH_NAME =~ ^v[0-9]+\.[0-9]+\.[0-9]*\-alpha.* ]]; then
    # 用tag管理的分支：main/master/MISSION
    arr=(${BRANCH_NAME//-/ })
    tag="$SERVICE_VERSION-${arr[1]}"
elif [[ $BRANCH_NAME =~ ^v[0-9]+\.[0-9]+\.[0-9]*$ ]]; then
    # 用tag管理的分支：release 稳定版（最终发行版）
    tag="$RELEASE_VERSION"
elif [[ $BRANCH_NAME =~ ^v[0-9]+\.[0-9]+\.[0-9]*\-.* ]]; then
    # 用tag管理的分支：release 候选版|公测版
    arr=(${BRANCH_NAME//-/ })
    tag="$RELEASE_VERSION-${arr[1]}"
else
    # 用分支管理的分支：feature/bug等
    tag="$SERVICE_VERSION-$BRANCH_NAME"
fi

# 打包
docker run --rm \
-v $abspath/:/build/ \
node:18.7.0 \
/bin/bash -c "bash /build/scripts/build.sh"

# 进入产物路径
cd ./dist/

# 压缩
tar -czvf proton-cli-web.$tag.$BUILD_NUMBER.tar.gz ./*

# 创建ftp专用路径
mkdir source

# 移动压缩包至ftp路径
mv proton-cli-web.$tag.$BUILD_NUMBER.tar.gz ./source

# 创建latest
cd ./source && cp proton-cli-web.$tag.$BUILD_NUMBER.tar.gz proton-cli-web.latest.tar.gz
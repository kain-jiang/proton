
# 获取当前脚本的绝对路径
script_path=$(realpath "${BASH_SOURCE[0]}")
# 切换到项目根目录
cd "$(dirname "$(dirname "$script_path")")" || exit

# 项目信息
meta_projectName="component-manage"
meta_registry="acr.aishu.cn"
meta_repository="ict/component-manage"

meta_acrUser=${REGISTRY_USERNAME="unknown"}
meta_acrPass=${REGISTRY_PASSWORD="unknown"}

declare -A meta_archMap
meta_archMap["x86_64"]="amd64"
meta_archMap["aarch64"]="arm64"
meta_archMap["x86"]="amd64"
meta_archMap["arm"]="arm64"
meta_archMap["amd64"]="amd64"
meta_archMap["arm64"]="arm64"
now_arch="${meta_archMap[${ARCH:-$(uname -m)}]}"

meta_projectPath=$(pwd)
meta_devopsPat=${AZURE_TOKEN}
meta_sourceVersion="${BUILD_SOURCEVERSION:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
meta_branchName="${BUILD_SOURCEBRANCHNAME:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')}"
meta_branchName=${meta_branchName,,}
meta_branch="${BUILD_SOURCEBRANCH:-$(git symbolic-ref -q HEAD || git describe --tags --exact-match 2>/dev/null || echo 'unknown')}"

meta_localImage="localhost/${meta_projectName}:latest"
meta_remoteImageName="${meta_registry}/${meta_repository}"

# 优化 sed 命令，使用单个命令替换多个模式
meta_semVersion=$(git describe --tags --match='v*' --abbrev=0 2>/dev/null | sed -r 's/^v//g')
meta_semVersion=${meta_semVersion:-"0.0.0"}

meta_imageTag="git.${meta_sourceVersion:0:8}"
meta_imageTags=("$meta_imageTag")

# meta_branch="refs/tags/v1.1.1"
if [[ "$meta_branch" == refs/heads/* ]]; then
  # 分支触发
  meta_imageTags+=("${meta_branchName,,}.latest")
  # 分支触发只保留分支版本
  meta_version="${meta_semVersion}-${meta_branchName//-/.}"
  meta_versions=("$meta_version" "${meta_version}.git.${meta_sourceVersion:0:8}")
else
  # 标签触发
  meta_version="${meta_semVersion}"
  meta_versions=("$meta_version" "${meta_version}-git.${meta_sourceVersion:0:8}")
fi

# 版本信息
echo "项目版本: ${meta_version}，镜像版本：$meta_imageTag"
echo "Chart版本: ${meta_versions[*]}"
echo "镜像Tags: ${meta_imageTags[*]}"

echo "提交ID: ${meta_sourceVersion}"
echo "分支名: ${meta_branchName}"
echo "分支: ${meta_branch}"

echo "本地镜像: ${meta_localImage}"
echo "远程镜像名: ${meta_remoteImageName}"
echo "当前架构: $now_arch"
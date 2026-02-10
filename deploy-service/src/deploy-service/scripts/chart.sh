#!/usr/bin/env bash

set -e

# shellcheck disable=SC2046
cd $(dirname $(dirname $(realpath "$0"))) || exit 1
cd chart || exit 1

projectPath=$(pwd)
chartName="deploy-service"

# from $(registry.username)
acrUser=${REGISTRY_USERNAME="unknown"}

# from $(registry.password) by env
acrPass=${REGISTRY_PASSWORD="unknown"}

registry="acr.aishu.cn"
version=${VERSION="3.0.0"}

# from $(Build.SourceBranchName)
readonly branch=${BUILD_SOURCEBRANCHNAME="local"}
# from $(Build.BuildID)
readonly buildID=${BUILD_BUILDID="0"}
# replace by $(Build.SourceVersion) # from $(Build.BuildID)
readonly sourceVersion=${BUILD_SOURCEVERSION="unknown"}
readonly buildGit="git+${sourceVersion:0:8}"
readonly gitTag="git.${sourceVersion:0:8}"

imageTag="${gitTag}"


# 通过分支处理版本
# 如果不是release分支，则添加分支后缀
if [[ ${branch,,} == "release"* ]];
then
  # release分支
  versions=("$version-$buildID" "$version-$buildGit")
else
  # 其他分支
  versions=("$version-${branch,,}" "$version-$buildID" "$version-$buildGit")
fi


echo "ChartVersion: ${versions[*]}"
echo "ImageTag: $imageTag"

yq="docker run --user root --rm -v $projectPath:/workdir acr.aishu.cn/as/yq:universal"
helm="docker run --user root --rm -v $projectPath:/apps -v $projectPath/.helm:/root/.helm acr.aishu.cn/public/alpine-helm-2.16.6:universal"
$helm init --client-only --skip-refresh > /dev/null

$yq e -i '.image.tag="'$imageTag'"' $chartName/values.yaml
$yq e -i '.image.registry="'$registry'"' $chartName/values.yaml

$helm lint $chartName


# shellcheck disable=SC2068
for v in ${versions[@]}; do
  echo "Package Chart Version: $v"
  $helm package $chartName --version $v || true
  # shellcheck disable=SC2046
  chartPath=$(realpath "$chartName-$v.tgz")
  curl -s -f "https://$acrUser:$acrPass@acr.aishu.cn/api/chartrepo/ict/charts?force=true" -XPOST -F "chart=@$chartPath"
  echo
  echo "Pushed Package Version: $v"
  echo
done

chartName="deploy-shadow"
# shellcheck disable=SC2068
for v in ${versions[@]}; do
  echo "Package ShadowChart Version: $v"
  $helm package $chartName --version $v || true
  # shellcheck disable=SC2046
  chartPath=$(realpath "$chartName-$v.tgz")
  curl -s -f "https://$acrUser:$acrPass@acr.aishu.cn/api/chartrepo/ict/charts?force=true" -XPOST -F "chart=@$chartPath"
  echo
  echo "Pushed Package Version: $v"
  echo
done
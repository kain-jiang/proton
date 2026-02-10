#!/usr/bin/env bash

set -e

# shellcheck disable=SC2046
cd $(dirname $(dirname $(realpath "$0"))) || exit 1
cd charts
projectPath=$(pwd)

# from $(registry.username)
acrUser=${REGISTRY_USERNAME="unknown"}
# from $(registry.password) by env
acrPass=${REGISTRY_PASSWORD="unknown"}

registry="acr.aishu.cn"
version=${VERSION="1.0.0"}
imageTag=${TAG="auto"}
chartID=${CHART_ID="0"}

# 通过分支处理版本
# 如果不是release分支，则添加分支后缀
if [[ ${BRANCH,,} == "release-"* ]];
then
  # release分支
  versions=("$version" "$version-$chartID")
else
  # 其他分支
  versions=("$version-${BRANCH,,}" "$version-${BRANCH,,}.$chartID")
fi


echo "ChartVersion: ${versions[*]}"
echo "ImageTag: $imageTag"


for chartName in "deploy-service" "deploy-shadow"; do
    echo "Building Chart: $chartName"

    yq="docker run --user root --rm -v $projectPath:/workdir acr.aishu.cn/as/yq:universal"
    helm="docker run --rm -v $projectPath:/apps -v $projectPath/.helm:/root/.helm --workdir /apps acr.aishu.cn/public/helm:3.11.3-scratch"

    $yq e -i '.image.tag="'$imageTag'"' $chartName/values.yaml
    $yq e -i '.image.registry="'$registry'"' $chartName/values.yaml

    $helm lint $chartName

    # shellcheck disable=SC2068
    for v in ${versions[@]}; do
    echo "Package Chart Version: $v"
    $helm package $chartName --version $v || true
    # shellcheck disable=SC2046
    chartPath=$(realpath "$chartName-$v.tgz")
    curl -s "https://$acrUser:$acrPass@acr.aishu.cn/api/chartrepo/ict/charts?force=true" -XPOST -F "chart=@$chartPath"
    echo
    echo "Pushed Package Version: $v"
    echo
    done

done
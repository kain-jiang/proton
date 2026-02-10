#!/usr/bin/env bash

set -e
source "$(dirname $(realpath "${BASH_SOURCE[0]}"))/_meta.sh"

cd chart || exit 1

# 工具
yq="docker run --user root --rm -v $(pwd):/workdir acr.aishu.cn/as/yq:universal"
helm="docker run --user 0:0 --rm -v $(pwd):/work -w /work acr.aishu.cn/public/helm:3.11.3-scratch"

for chartName in `ls`
do 
  $yq e -i '.image.tag="'$meta_imageTag'"' $chartName/values.yaml
  $yq e -i '.image.registry="'$meta_registry'"' $chartName/values.yaml
  $helm lint $chartName

  # shellcheck disable=SC2068
  for v in ${meta_versions[@]}; do
    echo "Package Chart Version: $v"
    $helm package $chartName --version $v --app-version ${meta_imageTags[0]} || true
    # shellcheck disable=SC2046
    chartPath=$(realpath "$chartName-$v.tgz")
    curl -s -f "https://$meta_acrUser:$meta_acrPass@acr.aishu.cn/api/chartrepo/ict/charts?force=true" -XPOST -F "chart=@$chartPath"
    echo
    echo "Pushed Package Version: $v"
    echo
    rm -f $chartPath
  done
done
#!/usr/bin/env bash

set -e
source "$(dirname $(realpath "${BASH_SOURCE[0]}"))/_meta.sh"

# 构建
export DOCKER_BUILDKIT=1
docker build --target build-result --build-arg VERSION=$meta_version --build-arg DEVOPS_PAT=$meta_devopsPat -t $meta_localImage .

# 推送
# shellcheck disable=SC2068
for tag in ${meta_imageTags[@]} ; do
  docker tag $meta_localImage "$meta_remoteImageName:$tag.$now_arch"
  docker push "$meta_remoteImageName:$tag.$now_arch"
  docker rmi -f "$meta_remoteImageName:$tag.$now_arch"
done
docker rmi -f $meta_localImage

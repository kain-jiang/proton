#!/usr/bin/env bash

set -e
source "$(dirname $(realpath "${BASH_SOURCE[0]}"))/_meta.sh"


for tag in ${meta_imageTags[@]} ; do
  docker manifest create --amend "$meta_remoteImageName:$tag" "$meta_remoteImageName:$tag.amd64" "$meta_remoteImageName:$tag.arm64"
  docker manifest push --purge "$meta_remoteImageName:$tag"
  docker rmi -f "$meta_remoteImageName:$tag"
done

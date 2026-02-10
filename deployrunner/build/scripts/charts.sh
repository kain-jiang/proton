#!/bin/bash

libsPath=$(dirname `readlink -f $0`)/meta.sh
source $libsPath

set -x 
set -e 
set -o errexit
set -o nounset
set -o pipefail

buildDir=${workDir}/.build
mkdir -p ${buildDir}

push_charts() {
    acrUser=${REGISTRY_USERNAME:?"REGISTRY_USERNAME is required"}
    acrPass=${REGISTRY_PASSWORD:?"REGISTRY_PASSWORD is required"}
    for chart in `ls ${buildDir}/cpck`
    do
        curl -s -F "chart=@${buildDir}/cpck/${chart};type=application/x-compressed-tar" \
        -X POST "https://$acrUser:$acrPass@acr.aishu.cn/api/chartrepo/ict/charts?force=true"
    done
}

push_charts
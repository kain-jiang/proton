#!/bin/bash

libsPath=$(dirname `readlink -f $0`)/meta.sh
source $libsPath

set -x 
set -e 
set -o errexit
set -o nounset
set -o pipefail

make_manifest(){
    cmds=("builder" "installer" "mproxy")
    local tags=("${DOCKER_TAG}")
#    if [ ${GIT_BRANCHNAME} == "master" ]
#    then
#        tags=("${tags[@]}" "${CURRENT_TAG}")
#    else 
#        tags=("${tags[@]}" "${CURRENT_TAG}-${GIT_BRANCHNAME}")
#    fi

    for i in ${cmds[@]}
    do
        imgs=""
        for arc in ${ARCHES[@]}
        do
        imgs="${imgs} ${IMAGE_PREFIX}${i}:${DOCKER_TAG}.${arc}"
        done


        # special cmd version control
        local cmd_latest=$(meta_containe_tag ${i})
        if [ ! -z ${cmd_latest} ]
        # if [ ${i} == "builder" ] && [ ${GIT_BRANCHNAME} == "master" ]
        then
            # docker manifest create --amend ${IMAGE_PREFIX}${i}:v1.0.0 ${imgs}
            docker manifest create --amend ${IMAGE_PREFIX}${i}:${cmd_latest} ${imgs}
            docker manifest push --purge ${IMAGE_PREFIX}${i}:${cmd_latest}
            # docker manifest push --purge ${IMAGE_PREFIX}${i}:v1.0.0
        fi

        for tag in ${tags[@]}
        do
        docker manifest create --amend ${IMAGE_PREFIX}${i}:${tag} ${imgs}
        docker manifest push --purge ${IMAGE_PREFIX}${i}:${tag}
        done

    done

}

# debug_meta
make_manifest
